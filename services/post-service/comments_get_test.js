import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 5,
    duration: '3s',
};

export default function () {
    // Usamos el ID del post que validamos antes en la base de datos
    const postId = 'post-1794264562726';
    const url = `http://localhost:8084/api/v1/posts/${postId}/comments`;

    let res = http.get(url);

    check(res, {
        'status es 200 OK': (r) => r.status === 200,
        'devuelve la estructura de comentarios': (r) => r.json('comments') !== undefined,
    });

    sleep(0.5);
}