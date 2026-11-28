import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

http.setResponseCallback(http.expectedStatuses(200, 201, 401, 409, 429));

const spammerAllowed = new Counter('spammer_allowed_200');
const spammerBlocked = new Counter('spammer_blocked_429');

export const options = {
    scenarios: {
        // 1. Ráfaga concurrente de un mismo usuario autenticado (80 peticiones en paralelo entre 8 VUs)
        concurrent_single_user_burst: {
            executor: 'shared-iterations',
            vus: 8,
            iterations: 80,
            maxDuration: '30s',
            exec: 'spammerBurst',
        },
        // 2. Usuarios concurrentes independientes comprobando aislamiento de cuota, rutas save/share y rechazo sin token
        concurrent_multi_user_isolation: {
            executor: 'per-vu-iterations',
            vus: 4,
            iterations: 5,
            maxDuration: '30s',
            exec: 'isolatedUserTraffic',
        },
    },
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
        spammer_allowed_200: ['count==50'],
        spammer_blocked_429: ['count==30'],
    },
};

const POST_BASE_URL = __ENV.POST_SERVICE_URL || 'http://localhost:8084';
const AUTH_BASE_URL = __ENV.AUTH_SERVICE_URL || 'http://localhost:8083';

function registerAndLogin(suffix) {
    const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };
    const email = `k6_spam_${suffix}@meloop.test`;
    const username = `k6spam_${suffix}`;
    const password = 'Password123!';

    http.post(
        `${AUTH_BASE_URL}/auth/register`,
        JSON.stringify({ username, email, password }),
        jsonHeaders
    );

    const loginRes = http.post(
        `${AUTH_BASE_URL}/auth/login`,
        JSON.stringify({ identifier: email, email, password }),
        jsonHeaders
    );

    if (loginRes.status !== 200) {
        throw new Error(`No se pudo autenticar usuario k6 (${email}): status=${loginRes.status} body=${loginRes.body}`);
    }

    const body = JSON.parse(loginRes.body);
    const token =
        (body.data && (body.data.token || body.data.session_token || body.data.access_token)) ||
        body.token ||
        body.session_token;

    if (!token) {
        throw new Error(`Respuesta de login sin token: ${loginRes.body}`);
    }
    return token;
}

export function setup() {
    const runId = `${Date.now()}`;
    const spammerToken = registerAndLogin(`main_${runId}`);
    const cleanUserTokens = [
        registerAndLogin(`u1_${runId}`),
        registerAndLogin(`u2_${runId}`),
        registerAndLogin(`u3_${runId}`),
        registerAndLogin(`u4_${runId}`),
    ];

    const authHeaders = {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${spammerToken}`,
    };

    const createRes = http.post(
        `${POST_BASE_URL}/api/v1/posts`,
        JSON.stringify({ content: `Post prueba concurrencia anti-spam ${runId}` }),
        { headers: authHeaders }
    );
    const createdPost = JSON.parse(createRes.body);

    if (createRes.status !== 201 || !createdPost.id) {
        throw new Error(`No se pudo crear post de prueba: status=${createRes.status} body=${createRes.body}`);
    }

    return {
        postId: createdPost.id,
        spammerToken,
        cleanUserTokens,
    };
}

export function spammerBurst(data) {
    const headers = {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${data.spammerToken}`,
    };

    // Rotar entre like v1, like legacy, save y share para validar que todas comparten la cuota de 50 interacciones
    let res;
    switch (__ITER % 4) {
        case 0:
            res = http.post(
                `${POST_BASE_URL}/posts/${data.postId}/likes`,
                JSON.stringify({ user_id: 'otro-usuario' }),
                { headers }
            );
            break;
        case 1:
            res = http.post(`${POST_BASE_URL}/v1/posts/${data.postId}/like`, null, { headers });
            break;
        case 2:
            res = http.post(`${POST_BASE_URL}/api/v1/posts/${data.postId}/save`, null, { headers });
            break;
        default:
            res = http.post(
                `${POST_BASE_URL}/api/v1/posts/${data.postId}/share`,
                JSON.stringify({ platform: 'internal' }),
                { headers }
            );
            break;
    }

    if (res.status === 200 || res.status === 201 || res.status === 409) {
        spammerAllowed.add(1);
    } else if (res.status === 429) {
        spammerBlocked.add(1);
        const body = JSON.parse(res.body);
        check(res, {
            '429 incluye retry_after_ms válido (> 0 y <= 15m)': () =>
                body.retry_after_ms > 0 && body.retry_after_ms <= 900000,
        });
    }

    check(res, {
        'Ráfaga concurrente responde 200/201/409 antes del límite o 429 al superarlo': (r) =>
            r.status === 200 || r.status === 201 || r.status === 409 || r.status === 429,
    });
}

export function isolatedUserTraffic(data) {
    // 1. Verificar que suplantación con X-User-ID sin Bearer token es rechazada (401) en like, save y share
    const spoofHeaders = { headers: { 'X-User-ID': 'attacker-spoof', 'Content-Type': 'application/json' } };
    const spoofLike = http.post(`${POST_BASE_URL}/v1/posts/${data.postId}/like`, null, spoofHeaders);
    const spoofSave = http.post(`${POST_BASE_URL}/api/v1/posts/${data.postId}/save`, null, spoofHeaders);
    const spoofShare = http.post(
        `${POST_BASE_URL}/api/v1/posts/${data.postId}/share`,
        JSON.stringify({ platform: 'whatsapp' }),
        spoofHeaders
    );

    check(spoofLike, {
        'Like sin sesión (solo X-User-ID) rechazado con 401': (r) => r.status === 401,
    });
    check(spoofSave, {
        'Save sin sesión (solo X-User-ID) rechazado con 401': (r) => r.status === 401,
    });
    check(spoofShare, {
        'Share sin sesión (solo X-User-ID) rechazado con 401': (r) => r.status === 401,
    });

    // 2. Usuario limpio concurrente ejecuta comentario, save, share, unsave y consulta de guardados
    const token = data.cleanUserTokens[(__VU - 1) % data.cleanUserTokens.length];
    const cleanHeaders = {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
    };

    const commentRes = http.post(
        `${POST_BASE_URL}/api/v1/posts/${data.postId}/comments`,
        JSON.stringify({ content: `Comentario concurrente VU ${__VU} iter ${__ITER}` }),
        { headers: cleanHeaders }
    );
    const saveRes = http.post(`${POST_BASE_URL}/api/v1/posts/${data.postId}/save`, null, { headers: cleanHeaders });
    const shareRes = http.post(
        `${POST_BASE_URL}/api/v1/posts/${data.postId}/share`,
        JSON.stringify({ platform: 'internal' }),
        { headers: cleanHeaders }
    );
    const unsaveRes = http.del(`${POST_BASE_URL}/api/v1/posts/${data.postId}/save`, null, { headers: cleanHeaders });
    const savedListRes = http.get(
        `${POST_BASE_URL}/api/v1/users/me/saved?user_id=attacker-spoof`,
        { headers: cleanHeaders }
    );

    let savedBody = {};
    try {
        savedBody = JSON.parse(savedListRes.body);
    } catch (_) {}

    check(commentRes, {
        'Usuario autenticado distinto conserva su cuota en comentarios (201)': (r) => r.status === 201,
    });
    check(saveRes, {
        'Save autenticado responde 200': (r) => r.status === 200,
    });
    check(shareRes, {
        'Share autenticado responde 200': (r) => r.status === 200,
    });
    check(unsaveRes, {
        'Unsave autenticado responde 200': (r) => r.status === 200,
    });
    check(savedListRes, {
        'GET /api/v1/users/me/saved usa identidad Bearer e ignora query user_id (200)': (r) =>
            r.status === 200 &&
            savedBody.user_id &&
            savedBody.user_id !== 'attacker-spoof' &&
            savedBody.user_id !== 'current_user_123',
    });
}

export function teardown(data) {
    if (data && data.postId && data.spammerToken) {
        http.del(`${POST_BASE_URL}/api/v1/posts/${data.postId}`, null, {
            headers: { Authorization: `Bearer ${data.spammerToken}` },
        });
    }
}