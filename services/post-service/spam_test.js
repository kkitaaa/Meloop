import http from 'k6/http';
import { check } from 'k6';

http.setResponseCallback(http.expectedStatuses(200, 201, 429));

export const options = {
    vus: 1,
    iterations: 1,
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = 'http://localhost:8084';

export default function () {
    const headers = {
        'Content-Type': 'application/json',
        'X-User-ID': '1',
    };

    // 0. Crear publicación temporal para la prueba (no cuenta como interacción)
    const createRes = http.post(
        `${BASE_URL}/api/v1/posts`,
        JSON.stringify({ content: 'Post temporal prueba anti-spam k6' }),
        { headers }
    );
    const createdPost = JSON.parse(createRes.body);
    const postId = createdPost.id;

    check(createRes, {
        '0. Post temporal creado (201)': (r) => r.status === 201 && Boolean(postId),
    });

    // 1. Ejecutar exactamente 50 interacciones permitidas (alternando like y unlike)
    let all50Allowed = true;
    for (let i = 1; i <= 50; i++) {
        const res = i % 2 === 1
            ? http.post(`${BASE_URL}/v1/posts/${postId}/like`, null, { headers })
            : http.del(`${BASE_URL}/v1/posts/${postId}/like`, null, { headers });

        if (res.status !== 200) {
            all50Allowed = false;
            console.error(`Interacción ${i} falló con status ${res.status}: ${res.body}`);
            break;
        }
    }

    check(null, {
        '1. Primeras 50 interacciones permitidas (200)': () => all50Allowed,
    });

    // 2. Interacción #51 debe bloquearse con 429 y bloqueo de 15 minutos
    const res51 = http.post(`${BASE_URL}/v1/posts/${postId}/like`, null, { headers });
    const body51 = JSON.parse(res51.body);

    check(res51, {
        '2a. Interacción 51 bloqueada con 429': (r) => r.status === 429,
        '2b. Incluye retry_after_ms de ~15 minutos': () =>
            body51.retry_after_ms > 890000 && body51.retry_after_ms <= 900000,
    });

    // 3. Interacción #52 (comentario) también bloqueada antes de escribir en BD
    const res52 = http.post(
        `${BASE_URL}/api/v1/posts/${postId}/comments`,
        JSON.stringify({ content: 'Este comentario debe ser bloqueado por anti-spam' }),
        { headers }
    );

    check(res52, {
        '3. Comentario 52 bloqueado con 429 sin ensuciar BD': (r) => r.status === 429,
    });

    // 4. Eliminar publicación temporal (CRUD no bloqueado por anti-spam de interacciones)
    const delRes = http.del(`${BASE_URL}/api/v1/posts/${postId}`, null, { headers });
    check(delRes, {
        '4. Post temporal eliminado limpiamente (200)': (r) => r.status === 200,
    });
}