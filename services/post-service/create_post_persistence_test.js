import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 1,
    iterations: 3,
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8084/api/v1/posts';
const USER_ID = __ENV.USER_ID || '1';

export default function () {
    const headers = {
        'Content-Type': 'application/json',
        'X-User-ID': USER_ID,
    };

    // 1. Crear post de texto solo (ignorando author_id malicioso en body - PST-02)
    const resText = http.post(
        BASE_URL,
        JSON.stringify({
            author_id: 'usuario-falso',
            content: 'Publicación persistida de prueba k6 (texto solo)',
        }),
        { headers }
    );
    const postText = JSON.parse(resText.body);

    check(resText, {
        '1a. Texto solo retorna 201 con UUID y fecha_creacion': (r) =>
            r.status === 201 &&
            Boolean(postText.id) &&
            Boolean(postText.created_at) &&
            postText.author_id === USER_ID,
    });

    // 1b. Leer desde BD vía GET para confirmar persistencia real
    const getText = http.get(`${BASE_URL}/${postText.id}`, { headers });
    check(getText, {
        '1b. Post de texto existe en BD al consultar GET': (r) =>
            r.status === 200 && JSON.parse(r.body).content === 'Publicación persistida de prueba k6 (texto solo)',
    });

    // 2. Crear post con texto + canción Spotify válida
    const resMusic = http.post(
        BASE_URL,
        JSON.stringify({
            content: 'Publicación persistida con música k6',
            music_id: '4cOdK2wGLETKBW3PvgPWqT',
        }),
        { headers }
    );
    const postMusic = JSON.parse(resMusic.body);

    check(resMusic, {
        '2a. Texto + música retorna 201 con music_id asociado': (r) =>
            r.status === 201 && Boolean(postMusic.id) && Boolean(postMusic.music_id),
    });

    const getMusic = http.get(`${BASE_URL}/${postMusic.id}`, { headers });
    check(getMusic, {
        '2b. Post con música existe en BD al consultar GET': (r) => r.status === 200,
    });

    // 3. Limpiar posts creados para dejar la BD limpia
    const del1 = http.del(`${BASE_URL}/${postText.id}`, null, { headers });
    const del2 = http.del(`${BASE_URL}/${postMusic.id}`, null, { headers });
    check(null, {
        '3. Registros de prueba limpiados (200)': () => del1.status === 200 && del2.status === 200,
    });
}