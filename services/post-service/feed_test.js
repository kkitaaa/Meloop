import http from 'k6/http';
import { check, sleep } from 'k6';

// 200, 400 y 401 son respuestas esperadas según el caso de prueba
http.setResponseCallback(http.expectedStatuses(200, 400, 401));

export const options = {
    vus: 20,
    duration: '15s',
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
    thresholds: {
        checks: ['rate==1.0'],            // Falla la ejecución si cualquier check falla
        http_req_duration: ['p(95)<200'], // El 95% de las peticiones deben responder en < 200ms
        http_req_failed: ['rate<0.01'],   // Sin errores 5xx ni fallos de red
    },
};

function parseJSON(res) {
    try {
        return JSON.parse(res.body);
    } catch (_) {
        return null;
    }
}

export function setup() {
    if (__ENV.AUTH_TOKEN) {
        return { token: __ENV.AUTH_TOKEN };
    }

    const authUrl = __ENV.AUTH_URL || 'http://localhost:8083';
    const email = __ENV.TEST_EMAIL || 'usuariotest@gmail.com';
    const password = __ENV.TEST_PASSWORD || 'Password123!';

    const loginRes = http.post(
        `${authUrl}/auth/login`,
        JSON.stringify({ email: email, password: password }),
        { headers: { 'Content-Type': 'application/json' } }
    );

    const body = parseJSON(loginRes);
    const token =
        body?.data?.session_token ||
        body?.data?.token ||
        body?.data?.access_token ||
        body?.session_token ||
        body?.token ||
        body?.access_token ||
        '';

    if (!token) {
        throw new Error(
            `No se pudo obtener token para ${email} desde ${authUrl}/auth/login (status ${loginRes.status}): ${loginRes.body}`
        );
    }

    return { token };
}

export default function (data) {
    const baseUrl = __ENV.BASE_URL || 'http://localhost:8084';
    const token = data.token;

    // 0. Seguridad: intentar usar X-User-ID = 1 sin token Bearer debe fallar con 401
    const resSpoofed = http.get(`${baseUrl}/v1/feed?algorithm=relevant&page=1&limit=20`, {
        headers: {
            'X-User-ID': '1',
            'Content-Type': 'application/json',
        },
    });
    check(resSpoofed, {
        'Seguridad: rechaza X-User-ID sin token Bearer con 401': (r) => r.status === 401,
    });

    const params = {
        headers: {
            'Authorization': `Bearer ${token}`,
            // Se envía un X-User-ID falso para verificar que el servidor lo ignore y use la sesión de id=1
            'X-User-ID': 'usuario-suplantado-999',
            'Content-Type': 'application/json',
        },
    };

    // 1. Prueba de carga y ordenamiento: Algoritmo Reciente (Cronológico)
    const resRecent = http.get(`${baseUrl}/v1/feed?algorithm=recent&page=1&limit=20`, params);
    const bodyRecent = parseJSON(resRecent);

    check(resRecent, {
        'Feed recent status 200': (r) => r.status === 200,
        'Feed recent devuelve algorithm=recent': () => bodyRecent !== null && bodyRecent.algorithm === 'recent',
        'Feed recent devuelve arreglo data con >= 2 posts': () =>
            bodyRecent !== null && Array.isArray(bodyRecent.data) && bodyRecent.data.length >= 2,
        'Feed recent ordenado cronológicamente en todo el listado': () => {
            if (!bodyRecent || !Array.isArray(bodyRecent.data) || bodyRecent.data.length < 2) return false;
            for (let i = 0; i < bodyRecent.data.length - 1; i++) {
                const curr = new Date(bodyRecent.data[i].created_at).getTime();
                const next = new Date(bodyRecent.data[i + 1].created_at).getTime();
                if (curr < next) return false;
            }
            return true;
        },
    });

    sleep(0.3);

    // 2. Prueba de carga y ordenamiento: Algoritmo Relevante (Popularidad + Preferencias + desempate por ID)
    const resRelevant = http.get(`${baseUrl}/v1/feed?algorithm=relevant&page=1&limit=20`, params);
    const bodyRelevant = parseJSON(resRelevant);

    check(resRelevant, {
        'Feed relevant status 200': (r) => r.status === 200,
        'Feed relevant devuelve algorithm=relevant': () => bodyRelevant !== null && bodyRelevant.algorithm === 'relevant',
        'Feed relevant devuelve arreglo data con >= 2 posts': () =>
            bodyRelevant !== null && Array.isArray(bodyRelevant.data) && bodyRelevant.data.length >= 2,
        'Feed relevant incluye relevance_score numérico en cada item': () => {
            if (!bodyRelevant || !Array.isArray(bodyRelevant.data) || bodyRelevant.data.length === 0) return false;
            return bodyRelevant.data.every((item) => typeof item.relevance_score === 'number');
        },
        'Feed relevant ordenado por relevance_score DESC, created_at DESC e id DESC': () => {
            if (!bodyRelevant || !Array.isArray(bodyRelevant.data) || bodyRelevant.data.length < 2) return false;
            for (let i = 0; i < bodyRelevant.data.length - 1; i++) {
                const curr = bodyRelevant.data[i];
                const next = bodyRelevant.data[i + 1];
                if (curr.relevance_score < next.relevance_score) return false;
                if (curr.relevance_score === next.relevance_score) {
                    const currTime = new Date(curr.created_at).getTime();
                    const nextTime = new Date(next.created_at).getTime();
                    if (currTime < nextTime) return false;
                    if (currTime === nextTime && String(curr.id) < String(next.id)) return false;
                }
            }
            return true;
        },
    });

    sleep(0.3);

    // 3. Validación de parámetros inválidos (debe retornar 400)
    const resInvalid = http.get(`${baseUrl}/v1/feed?algorithm=invalido`, params);
    check(resInvalid, {
        'Algoritmo inválido rechazado con 400': (r) => r.status === 400,
    });
}