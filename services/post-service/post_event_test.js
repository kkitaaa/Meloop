import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10,
    duration: '5s',
};

export default function () {
    const url = 'http://localhost:8084/api/v1/posts'; // <- CORREGIDO AQUÍ
    const payload = JSON.stringify({
        content: `Post de prueba asíncrona desde K6 - VU ${__VU}`,
        media_urls: []
    });

    const params = {
        headers: {
            'X-User-ID': '1', 
            'Content-Type': 'application/json',
        },
    };

    let res = http.post(url, payload, params);

    check(res, {
        'Post creado exitosamente (status 201)': (r) => r.status === 201,
        'Respuesta tiene ID del post': (r) => {
            const body = JSON.parse(r.body);
            return body.ID !== undefined || body.id_publicacion !== undefined || body.id !== undefined; 
        },
    });

    sleep(0.5);
}