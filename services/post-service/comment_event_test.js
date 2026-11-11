import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10,
    duration: '5s',
};

export default function () {
    const postUrl = 'http://localhost:8084/api/v1/posts/post-1794264562726/comments'; 
    
    // AQUÍ ESTÁ LA MAGIA: Agregamos author_id al JSON
    const payload = JSON.stringify({
        author_id: '1', 
        content: `Comentario de prueba asíncrona desde K6 - VU ${__VU}`
    });

    const params = {
        headers: {
            'X-User-ID': '1', 
            'Content-Type': 'application/json',
        },
    };

    let res = http.post(postUrl, payload, params);

    check(res, {
        'Comentario creado exitosamente (status 201)': (r) => r.status === 201,
    });

    sleep(0.5);
}