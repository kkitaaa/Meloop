import http from 'k6/http';
import { check, sleep } from 'k6';

http.setResponseCallback(http.expectedStatuses(200, 201, 401, 403, 404));

export const options = {
    vus: 1,
    iterations: 5,
    thresholds: {
        http_req_failed: ['rate<0.01'],
        checks: ['rate==1.0'],
    },
};

const BASE_URL = 'http://localhost:8084/api/v1/posts';

export default function () {
    const authorHeaders = {
        'Content-Type': 'application/json',
        'X-User-ID': '1',
    };
    const intruderHeaders = {
        'Content-Type': 'application/json',
        'X-User-ID': '999',
    };
    const anonHeaders = {
        'Content-Type': 'application/json',
    };

    // 1. Crear post temporal con autor legítimo
    const createRes = http.post(
        BASE_URL,
        JSON.stringify({ content: 'Publicación inicial k6' }),
        { headers: authorHeaders }
    );
    const created = JSON.parse(createRes.body);
    const postId = created.id;

    check(createRes, {
        '1. Post creado con 201 y tiene ID UUID': (r) => r.status === 201 && Boolean(postId),
    });

    // 2. Sin identidad (sin X-User-ID) -> 401
    const unauthPut = http.put(
        `${BASE_URL}/${postId}`,
        JSON.stringify({ content: 'Intento sin header' }),
        { headers: anonHeaders }
    );
    const unauthDel = http.del(`${BASE_URL}/${postId}`, null, { headers: anonHeaders });

    check(unauthPut, { '2a. PUT sin X-User-ID rechazado con 401': (r) => r.status === 401 });
    check(unauthDel, { '2b. DELETE sin X-User-ID rechazado con 401': (r) => r.status === 401 });

    // 3. Otro usuario (no autor) -> 403
    const forbiddenPut = http.put(
        `${BASE_URL}/${postId}`,
        JSON.stringify({ content: 'Intento de intruso' }),
        { headers: intruderHeaders }
    );
    const forbiddenDel = http.del(`${BASE_URL}/${postId}`, null, { headers: intruderHeaders });

    check(forbiddenPut, { '3a. PUT de otro usuario rechazado con 403': (r) => r.status === 403 });
    check(forbiddenDel, { '3b. DELETE de otro usuario rechazado con 403': (r) => r.status === 403 });

    // 4. Post inexistente -> 404
    const missingId = '00000000-0000-0000-0000-000000000000';
    const notFoundPut = http.put(
        `${BASE_URL}/${missingId}`,
        JSON.stringify({ content: 'No existe' }),
        { headers: authorHeaders }
    );
    check(notFoundPut, { '4. PUT a post inexistente retorna 404': (r) => r.status === 404 });

    // 5. Edición legítima del autor -> 200 y verificación con GET
    const updatedText = 'Contenido actualizado por su autor';
    const validPut = http.put(
        `${BASE_URL}/${postId}`,
        JSON.stringify({ content: updatedText }),
        { headers: authorHeaders }
    );
    check(validPut, { '5a. PUT del autor retorna 200': (r) => r.status === 200 });

    const getAfterPut = http.get(`${BASE_URL}/${postId}`, { headers: authorHeaders });
    check(getAfterPut, {
        '5b. GET confirma contenido actualizado': (r) =>
            r.status === 200 && JSON.parse(r.body).content === updatedText,
    });

    // 6. Eliminación legítima del autor -> 200 y verificación 404 en GET
    const validDel = http.del(`${BASE_URL}/${postId}`, null, { headers: authorHeaders });
    check(validDel, { '6a. DELETE del autor retorna 200': (r) => r.status === 200 });

    const getAfterDel = http.get(`${BASE_URL}/${postId}`, { headers: authorHeaders });
    check(getAfterDel, { '6b. GET confirma post eliminado (404)': (r) => r.status === 404 });

    sleep(0.5);
}