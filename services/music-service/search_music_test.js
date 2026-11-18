import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 1,
    iterations: 2,
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate==0.0'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8083/v1/music/search';

export default function () {
    // 1. Búsqueda de "Rosalía" (track)
    const res1 = http.get(`${BASE_URL}?q=${encodeURIComponent('Rosalía')}&type=track&limit=5&offset=0`);
    if (res1.status !== 200) {
        console.error(`Error búsqueda 1 (HTTP ${res1.status}): ${res1.body}`);
    }
    const body1 = JSON.parse(res1.body);

    check(res1, {
        '1. Búsqueda Rosalía retorna 200 con estructura normalizada': (r) =>
            r.status === 200 &&
            Array.isArray(body1.results) &&
            body1.limit === 5 &&
            body1.offset === 0 &&
            typeof body1.total === 'number',
    });

    // 2. Segunda búsqueda idéntica (atendida desde caché Redis)
    const res2 = http.get(`${BASE_URL}?q=${encodeURIComponent('Rosalía')}&type=track&limit=5&offset=0`);
    check(res2, {
        '2. Segunda búsqueda idéntica responde 200 desde caché': (r) => r.status === 200,
    });

    // 3. Validación de parámetro inválido -> 400 INVALID_PARAMETER (esperado)
    const resInvalid = http.get(`${BASE_URL}?q=test&type=invalid_type`, {
        responseCallback: http.expectedStatuses(400),
    });
    const errBody = JSON.parse(resInvalid.body);
    check(resInvalid, {
        '3. Parámetro inválido retorna 400 INVALID_PARAMETER': (r) =>
            r.status === 400 && errBody.error === 'INVALID_PARAMETER',
    });
}