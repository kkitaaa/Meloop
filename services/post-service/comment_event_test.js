import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 1,
    iterations: 2,
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
    },
};

const POSTS_URL = __ENV.POSTS_URL || 'http://localhost:8084/api/v1/posts';
const POST_ID = __ENV.TEST_POST_ID || 'post-1794264562726';
const RABBIT_MGMT = __ENV.RABBIT_MGMT_URL || 'http://guest:guest@localhost:15672/api';
const TEMP_QUEUE = 'k6.verify.comment.created';

export function setup() {
    const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };

    // 1. Verificar que RabbitMQ Management y el exchange meloop.events estén activos
    const exRes = http.get(`${RABBIT_MGMT}/exchanges/%2F/meloop.events`);
    if (exRes.status !== 200) {
        throw new Error(`RabbitMQ exchange meloop.events no disponible (status=${exRes.status})`);
    }

    // 2. Crear cola temporal sin consumidor para inspeccionar los payloads entregados
    const qRes = http.put(
        `${RABBIT_MGMT}/queues/%2F/${encodeURIComponent(TEMP_QUEUE)}`,
        JSON.stringify({ auto_delete: false, durable: false }),
        jsonHeaders
    );
    if (qRes.status !== 201 && qRes.status !== 204) {
        throw new Error(`No se pudo crear cola temporal en RabbitMQ (status=${qRes.status})`);
    }

    // Limpiar mensajes residuales de ejecuciones previas
    http.del(`${RABBIT_MGMT}/queues/%2F/${encodeURIComponent(TEMP_QUEUE)}/contents`);

    // 3. Enlazar la cola temporal a comment.created en meloop.events
    const bindRes = http.post(
        `${RABBIT_MGMT}/bindings/%2F/e/meloop.events/q/${encodeURIComponent(TEMP_QUEUE)}`,
        JSON.stringify({ routing_key: 'comment.created' }),
        jsonHeaders
    );
    if (bindRes.status !== 201 && bindRes.status !== 204) {
        throw new Error(`No se pudo enlazar cola temporal a comment.created (status=${bindRes.status})`);
    }

    return { queueName: TEMP_QUEUE };
}

function fetchQueuePayloads(queueName) {
    const getMsgsRes = http.post(
        `${RABBIT_MGMT}/queues/%2F/${encodeURIComponent(queueName)}/get`,
        JSON.stringify({
            count: 50,
            ackmode: 'ack_requeue_true',
            encoding: 'auto',
        }),
        { headers: { 'Content-Type': 'application/json' } }
    );

    if (getMsgsRes.status !== 200) {
        return [];
    }

    const msgs = JSON.parse(getMsgsRes.body);
    return msgs.map((m) => {
        try {
            return JSON.parse(m.payload);
        } catch (_) {
            return {};
        }
    });
}

export default function (data) {
    const authorHeaders = {
        'X-User-ID': '1',
        'Content-Type': 'application/json',
    };

    // 1. Crear comentario principal (201)
    const commentRes = http.post(
        `${POSTS_URL}/${POST_ID}/comments`,
        JSON.stringify({ content: `Comentario prueba RabbitMQ k6 - iter ${__ITER}` }),
        { headers: authorHeaders }
    );
    const createdComment = JSON.parse(commentRes.body);

    check(commentRes, {
        'Comentario creado (status 201)': (r) => r.status === 201 && Boolean(createdComment.id),
    });

    // 2. Crear respuesta al comentario (201)
    const replyRes = http.post(
        `${POSTS_URL}/${POST_ID}/comments/${createdComment.id}/replies`,
        JSON.stringify({ content: `Respuesta prueba RabbitMQ k6 - iter ${__ITER}` }),
        { headers: authorHeaders }
    );
    const createdReply = JSON.parse(replyRes.body);

    check(replyRes, {
        'Respuesta creada (status 201)': (r) => r.status === 201 && Boolean(createdReply.id),
    });

    // 3. Esperar despacho del OutboxWorker y verificar ambos eventos en la cola de RabbitMQ
    let foundCommentEvent = false;
    let foundReplyEvent = false;

    for (let attempt = 0; attempt < 10; attempt++) {
        sleep(0.3);
        const payloads = fetchQueuePayloads(data.queueName);
        foundCommentEvent = payloads.some(
            (p) => p.event === 'comment.created' && p.comment_id === createdComment.id
        );
        foundReplyEvent = payloads.some(
            (p) => p.event === 'comment.created' && p.comment_id === createdReply.id
        );
        if (foundCommentEvent && foundReplyEvent) {
            break;
        }
    }

    check(null, {
        'RabbitMQ recibió evento comment.created del comentario': () => foundCommentEvent,
        'RabbitMQ recibió evento comment.created de la respuesta': () => foundReplyEvent,
    });

    // 4. Limpiar respuesta y comentario creados en PostgreSQL
    if (createdReply && createdReply.id) {
        const delReply = http.del(`${POSTS_URL}/${POST_ID}/comments/${createdReply.id}`, null, {
            headers: authorHeaders,
        });
        check(delReply, { 'Respuesta de prueba eliminada (200)': (r) => r.status === 200 });
    }

    if (createdComment && createdComment.id) {
        const delComment = http.del(`${POSTS_URL}/${POST_ID}/comments/${createdComment.id}`, null, {
            headers: authorHeaders,
        });
        check(delComment, { 'Comentario de prueba eliminado (200)': (r) => r.status === 200 });
    }
}

export function teardown(data) {
    if (data && data.queueName) {
        http.del(`${RABBIT_MGMT}/queues/%2F/${encodeURIComponent(data.queueName)}`);
    }
}