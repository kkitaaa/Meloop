import http from 'k6/http';
import { check, sleep } from 'k6';

// Indicamos a k6 que 200 y 400 son respuestas esperadas en esta prueba
http.setResponseCallback(http.expectedStatuses(200, 400));

export const options = {
    vus: 20,
    duration: '15s',
    thresholds: {
        http_req_duration: ['p(95)<200'], // El 95% de las peticiones deben responder en menos de 200ms
        http_req_failed: ['rate<0.01'],   // Solo fallará si hay errores 500 o de conexión
    },
};

export default function () {
    const params = {
        headers: {
            'X-User-ID': '1',
            'Content-Type': 'application/json',
        },
    };

    // 1. Prueba de carga: Algoritmo Reciente (Cronológico)
    const resRecent = http.get('http://localhost:8084/v1/feed?algorithm=recent&page=1&limit=20', params);
    const bodyRecent = JSON.parse(resRecent.body);

    check(resRecent, {
        'Feed recent status 200': (r) => r.status === 200,
        'Feed recent devuelve algorithm=recent': () => bodyRecent.algorithm === 'recent',
        'Feed recent devuelve arreglo data': () => Array.isArray(bodyRecent.data),
        'Feed recent ordenado cronológicamente': () => {
            if (!Array.isArray(bodyRecent.data) || bodyRecent.data.length < 2) return true;
            const first = new Date(bodyRecent.data[0].created_at).getTime();
            const second = new Date(bodyRecent.data[1].created_at).getTime();
            return first >= second;
        },
    });

    sleep(0.3);

    // 2. Prueba de carga: Algoritmo Relevante (Popularidad + Preferencias)
    const resRelevant = http.get('http://localhost:8084/v1/feed?algorithm=relevant&page=1&limit=20', params);
    const bodyRelevant = JSON.parse(resRelevant.body);

    check(resRelevant, {
        'Feed relevant status 200': (r) => r.status === 200,
        'Feed relevant devuelve algorithm=relevant': () => bodyRelevant.algorithm === 'relevant',
        'Feed relevant devuelve arreglo data': () => Array.isArray(bodyRelevant.data),
        'Feed relevant ordenado por relevance_score': () => {
            if (!Array.isArray(bodyRelevant.data) || bodyRelevant.data.length < 2) return true;
            return bodyRelevant.data[0].relevance_score >= bodyRelevant.data[1].relevance_score;
        },
        'Feed relevant cambia el orden respecto a recent (con datos)': () => {
            if (!Array.isArray(bodyRecent.data) || !Array.isArray(bodyRelevant.data)) return false;
            if (bodyRecent.data.length < 2 || bodyRelevant.data.length < 2) return true;
            return bodyRelevant.data[0].id !== bodyRecent.data[0].id;
        },
    });

    sleep(0.3);

    // 3. Validación de parámetros inválidos (debe retornar 400)
    const resInvalid = http.get('http://localhost:8084/v1/feed?algorithm=invalido', params);
    check(resInvalid, {
        'Algoritmo inválido rechazado con 400': (r) => r.status === 400,
    });
}