import http from 'k6/http';
import { check, sleep } from 'k6';

// 200, 201, 204 y 401 son respuestas esperadas según el caso de prueba
http.setResponseCallback(http.expectedStatuses(200, 201, 204, 401));

export const options = {
    vus: 1,
    iterations: 2,
    thresholds: {
        checks: ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
    },
};

const AUTH_URL = __ENV.AUTH_URL || 'http://localhost:8083';
const POSTS_URL = __ENV.POSTS_URL || 'http://localhost:8084/api/v1/posts';
const RABBIT_MGMT = __ENV.RABBIT_MGMT_URL || 'http://guest:guest@localhost:15672/api';
const TEMP_QUEUE = 'k6.verify.comment.created';

function parseJSON(res) {
    try {
        return JSON.parse(res.body);
    } catch (_) {
        return null;
    }
}

function extractToken(res) {
    const body = parseJSON(res);
    return (
        body?.data?.session_token ||
        body?.data?.token ||
        body?.data?.access_token ||
        body?.session_token ||
        body?.token ||
        body?.access_token ||
        ''
    );
}

function createTempAuthenticatedUser(tag) {
    const ts = `${Date.now()}_${tag}`;
    const username = `k6_${ts}`;
    const email = `k6.${ts}@example.test`;
    const password = 'K6CommentTest123!';

    const regRes = http.post(
        `${AUTH_URL}/auth/register`,
        JSON.stringify({
            username,
            email,
            password,
            fecha_nacimiento: '2000-01-01',
        }),
        { headers: { 'Content-Type': 'application/json' } }
    );
    if (regRes.status !== 201 && regRes.status !== 200) {
        throw new Error(`No se pudo registrar usuario temporal (${tag}): status=${regRes.status} body=${regRes.body}`);
    }

    const loginRes = http.post(
        `${AUTH_URL}/auth/login`,
        JSON.stringify({ email, password }),
        { headers: { 'Content-Type': 'application/json' } }
    );
    const token = extractToken(loginRes);
    if (!token) {
        throw new Error(`No se pudo iniciar sesión con usuario temporal (${tag}): status=${loginRes.status}`);
    }

    const valRes = http.get(`${AUTH_URL}/auth/validate`, {
        headers: { Authorization: `Bearer ${token}` },
    });
    const valBody = parseJSON(valRes);
    const userID = valBody?.data?.id || '';
    if (valRes.status !== 200 || !userID) {
        throw new Error(`No se pudo validar sesión temporal (${tag}): status=${valRes.status} body=${valRes.body}`);
    }

    return { token, userID };
}

export function setup() {
    const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };

    // 1. Verificar que RabbitMQ Management y el exchange meloop.events estén activos
    const exRes = http.get(`${RABBIT_MGMT}/exchanges/%2F/meloop.events`);
    if (exRes.status !== 200) {
        throw new Error(`RabbitMQ exchange meloop.events no disponible (status=${exRes.status})`);
    }

    // 2. Crear cola temporal para inspeccionar payloads publicados sin interferir con gamification.events
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

    // 4. Registrar dos usuarios temporales (dueño del post y comentarista)
    const owner = createTempAuthenticatedUser('owner');
    const commenter = createTempAuthenticatedUser('commenter');

    // 5. Crear una publicación real con el dueño para que el test funcione incluso con la BD vacía
    const postRes = http.post(
        POSTS_URL,
        JSON.stringify({ content: 'Publicación temporal k6 para prueba de eventos y seguridad' }),
        {
            headers: {
                Authorization: `Bearer ${owner.token}`,
                'Content-Type': 'application/json',
            },
        }
    );
    const createdPost = parseJSON(postRes);
    const postID = createdPost?.id || '';
    if (postRes.status !== 201 || !postID) {
        throw new Error(`No se pudo crear publicación de prueba: status=${postRes.status} body=${postRes.body}`);
    }

    return {
        queueName: TEMP_QUEUE,
        postID,
        ownerToken: owner.token,
        ownerID: owner.userID,
        commenterToken: commenter.token,
        commenterID: commenter.userID,
    };
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

    const msgs = parseJSON(getMsgsRes) || [];
    return msgs.map((m) => {
        try {
            return JSON.parse(m.payload);
        } catch (_) {
            return {};
        }
    });
}

export default function (data) {
    // 0. Seguridad: intentar comentar enviando solo X-User-ID sin token Bearer debe fallar con 401
    const spoofRes = http.post(
        `${POSTS_URL}/${data.postID}/comments`,
        JSON.stringify({ content: 'Intento sin token Bearer' }),
        {
            headers: {
                'X-User-ID': data.commenterID,
                'Content-Type': 'application/json',
            },
        }
    );
    check(spoofRes, {
        'Seguridad: rechaza X-User-ID sin token Bearer con 401': (r) => r.status === 401,
    });

    // Headers autenticados del comentarista + X-User-ID falso para comprobar que el backend lo ignora
    const commenterHeaders = {
        Authorization: `Bearer ${data.commenterToken}`,
        'X-User-ID': 'usuario-suplantado-999',
        'Content-Type': 'application/json',
    };

    // 1. Crear comentario principal (201)
    const commentRes = http.post(
        `${POSTS_URL}/${data.postID}/comments`,
        JSON.stringify({ content: `Comentario prueba RabbitMQ k6 - iter ${__ITER}` }),
        { headers: commenterHeaders }
    );
    const createdComment = parseJSON(commentRes);

    check(commentRes, {
        'Comentario creado con sesión Bearer (status 201)': (r) => r.status === 201 && Boolean(createdComment?.id),
    });

    // 2. Crear respuesta al comentario usando el dueño del post (201)
    const ownerHeaders = {
        Authorization: `Bearer ${data.ownerToken}`,
        'Content-Type': 'application/json',
    };
    const replyRes = http.post(
        `${POSTS_URL}/${data.postID}/comments/${createdComment.id}/replies`,
        JSON.stringify({ content: `Respuesta prueba RabbitMQ k6 - iter ${__ITER}` }),
        { headers: ownerHeaders }
    );
    const createdReply = parseJSON(replyRes);

    check(replyRes, {
        'Respuesta creada con sesión Bearer (status 201)': (r) => r.status === 201 && Boolean(createdReply?.id),
    });

    // 3. Esperar despacho del OutboxWorker y validar contrato completo (event_id, actor_id real y target_owner_id)
    let commentEvent = null;
    let replyEvent = null;

    for (let attempt = 0; attempt < 15; attempt++) {
        sleep(0.3);
        const payloads = fetchQueuePayloads(data.queueName);
        commentEvent = payloads.find(
            (p) => p.event === 'comment.created' && p.comment_id === createdComment.id
        );
        replyEvent = payloads.find(
            (p) => p.event === 'comment.created' && p.comment_id === createdReply.id
        );
        if (commentEvent && replyEvent) {
            break;
        }
    }

    check(null, {
        'RabbitMQ recibió evento comment.created del comentario con event_id': () =>
            Boolean(commentEvent && commentEvent.event_id),
        'Evento de comentario usa actor_id de sesión Bearer e ignora X-User-ID falso': () =>
            Boolean(
                commentEvent &&
                commentEvent.actor_id === data.commenterID &&
                commentEvent.target_owner_id === data.ownerID
            ),
        'RabbitMQ recibió evento comment.created de la respuesta con event_id único': () =>
            Boolean(replyEvent && replyEvent.event_id && replyEvent.event_id !== commentEvent?.event_id),
        'Evento de respuesta asigna actor_id y target_owner_id correctos': () =>
            Boolean(
                replyEvent &&
                replyEvent.actor_id === data.ownerID &&
                replyEvent.target_owner_id === data.commenterID
            ),
    });

    // 4. Limpiar respuesta y comentario creados en PostgreSQL
    if (createdReply?.id) {
        const delReply = http.del(`${POSTS_URL}/${data.postID}/comments/${createdReply.id}`, null, {
            headers: ownerHeaders,
        });
        check(delReply, { 'Respuesta de prueba eliminada (200)': (r) => r.status === 200 });
    }

    if (createdComment?.id) {
        const delComment = http.del(`${POSTS_URL}/${data.postID}/comments/${createdComment.id}`, null, {
            headers: commenterHeaders,
        });
        check(delComment, { 'Comentario de prueba eliminado (200)': (r) => r.status === 200 });
    }
}

export function teardown(data) {
    if (!data) return;

    if (data.postID && data.ownerToken) {
        http.del(`${POSTS_URL}/${data.postID}`, null, {
            headers: { Authorization: `Bearer ${data.ownerToken}` },
        });
    }

    if (data.queueName) {
        http.del(`${RABBIT_MGMT}/queues/%2F/${encodeURIComponent(data.queueName)}`);
    }
}