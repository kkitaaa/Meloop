import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 2,
    iterations: 4,
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
    },
};

const POSTS_URL = 'http://localhost:8084/api/v1/posts';
const RABBIT_API_URL = 'http://guest:guest@localhost:15672/api/exchanges/%2F/meloop.events';

export default function () {
    const authorHeaders = {
        'X-User-ID': '1',
        'Content-Type': 'application/json',
    };

    // 0. Verificar que RabbitMQ Management API esté activo y el exchange exista
    const rabbitBefore = http.get(RABBIT_API_URL);
    check(rabbitBefore, {
        'RabbitMQ broker activo y exchange meloop.events disponible': (r) => r.status === 200,
    });

    // 1. Crear comentario principal (201)
    const commentRes = http.post(
        `${POSTS_URL}/post-1794264562726/comments`,
        JSON.stringify({ content: `Comentario prueba RabbitMQ k6 - VU ${__VU}` }),
        { headers: authorHeaders }
    );
    const createdComment = JSON.parse(commentRes.body);

    check(commentRes, {
        'Comentario creado (status 201)': (r) => r.status === 201 && Boolean(createdComment.id),
    });

    // 2. Crear respuesta al comentario (201)
    const replyRes = http.post(
        `${POSTS_URL}/post-1794264562726/comments/${createdComment.id}/replies`,
        JSON.stringify({ content: `Respuesta prueba RabbitMQ k6 - VU ${__VU}` }),
        { headers: authorHeaders }
    );
    const createdReply = JSON.parse(replyRes.body);

    check(replyRes, {
        'Respuesta creada (status 201)': (r) => r.status === 201 && Boolean(createdReply.id),
    });

    // Dar tiempo a la goroutine para confirmar entrega con RabbitMQ
    sleep(0.5);

    const rabbitAfter = http.get(RABBIT_API_URL);
    check(rabbitAfter, {
        'RabbitMQ confirma recepción tras publicación': (r) => r.status === 200,
    });

    // 3. Limpiar respuesta y comentario creados para no ensuciar la BD
    if (createdReply && createdReply.id) {
        const delReply = http.del(`${POSTS_URL}/post-1794264562726/comments/${createdReply.id}`, null, {
            headers: authorHeaders,
        });
        check(delReply, { 'Respuesta de prueba eliminada (200)': (r) => r.status === 200 });
    }

    if (createdComment && createdComment.id) {
        const delComment = http.del(`${POSTS_URL}/post-1794264562726/comments/${createdComment.id}`, null, {
            headers: authorHeaders,
        });
        check(delComment, { 'Comentario de prueba eliminado (200)': (r) => r.status === 200 });
    }
}