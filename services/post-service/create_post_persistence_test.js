import http from 'k6/http';
import { check } from 'k6';

http.setResponseCallback(http.expectedStatuses(200, 201, 401));

export const options = {
    vus: 2,
    iterations: 4,
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
    },
};

const POST_BASE_URL = __ENV.BASE_URL || 'http://localhost:8084/api/v1/posts';
const AUTH_BASE_URL = __ENV.AUTH_SERVICE_URL || 'http://localhost:8083';

function registerAndLogin(suffix) {
    const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };
    const email = `k6_persist_${suffix}@meloop.test`;
    const username = `k6persist_${suffix}`;
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
    return {
        tokens: [
            registerAndLogin(`u1_${runId}`),
            registerAndLogin(`u2_${runId}`),
        ],
    };
}

export default function (data) {
    const token = data.tokens[(__VU - 1) % data.tokens.length];
    const authHeaders = {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        'X-User-ID': 'spoofed-header-ignored',
    };

    // 0. Verificar que sin token Bearer (aunque envíe X-User-ID) rechaza con 401
    const unauthRes = http.post(
        POST_BASE_URL,
        JSON.stringify({ content: 'Intento sin token Bearer' }),
        { headers: { 'Content-Type': 'application/json', 'X-User-ID': 'attacker-id' } }
    );
    check(unauthRes, {
        '0. Rechaza X-User-ID sin sesión Bearer con 401': (r) => r.status === 401,
    });

    // 1. Crear post de texto solo (ignorando author_id malicioso en body y X-User-ID falso)
    const resText = http.post(
        POST_BASE_URL,
        JSON.stringify({
            author_id: 'usuario-falso',
            content: 'Publicación persistida de prueba k6 (texto solo)',
        }),
        { headers: authHeaders }
    );
    const postText = JSON.parse(resText.body);

    check(resText, {
        '1a. Texto solo retorna 201 con UUID, fecha_creacion y author_id autenticado': (r) =>
            r.status === 201 &&
            Boolean(postText.id) &&
            Boolean(postText.created_at) &&
            Boolean(postText.author_id) &&
            postText.author_id !== 'usuario-falso' &&
            postText.author_id !== 'spoofed-header-ignored',
    });

    // 1b. Leer desde BD vía GET para confirmar persistencia real
    const getText = http.get(`${POST_BASE_URL}/${postText.id}`, { headers: authHeaders });
    check(getText, {
        '1b. Post de texto existe en BD al consultar GET': (r) =>
            r.status === 200 && JSON.parse(r.body).content === 'Publicación persistida de prueba k6 (texto solo)',
    });

    // 2. Crear post con texto + canción Spotify válida bajo concurrencia (valida ON CONFLICT spotify_id)
    const resMusic = http.post(
        POST_BASE_URL,
        JSON.stringify({
            content: 'Publicación persistida con música k6',
            music_id: '4cOdK2wGLETKBW3PvgPWqT',
        }),
        { headers: authHeaders }
    );
    const postMusic = JSON.parse(resMusic.body);

    check(resMusic, {
        '2a. Texto + música retorna 201 con music_id asociado': (r) =>
            r.status === 201 && Boolean(postMusic.id) && Boolean(postMusic.music_id),
    });

    const getMusic = http.get(`${POST_BASE_URL}/${postMusic.id}`, { headers: authHeaders });
    check(getMusic, {
        '2b. Post con música existe en BD al consultar GET': (r) => r.status === 200,
    });

    // 3. Limpiar posts creados para dejar la BD limpia
    const del1 = http.del(`${POST_BASE_URL}/${postText.id}`, null, { headers: authHeaders });
    const del2 = http.del(`${POST_BASE_URL}/${postMusic.id}`, null, { headers: authHeaders });
    check(null, {
        '3. Registros de prueba limpiados (200)': () => del1.status === 200 && del2.status === 200,
    });
}