import http from 'k6/http';
import { check, fail, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const AUTH_URL = __ENV.AUTH_URL || 'http://localhost:8083';
const RABBITMQ_MANAGEMENT_URL = __ENV.RABBITMQ_MANAGEMENT_URL || 'http://localhost:15672';
const RABBITMQ_USER = __ENV.RABBITMQ_USER || 'guest';
const RABBITMQ_PASSWORD = __ENV.RABBITMQ_PASSWORD || 'guest';
const VHOST = encodeURIComponent(__ENV.RABBITMQ_VHOST || '/');
const POLL_TIMEOUT_SECONDS = Number(__ENV.POLL_TIMEOUT_SECONDS || 15);
const PASSWORD = __ENV.TEST_PASSWORD || 'SecurePassword123!';

const setupFailures = new Rate('notification_setup_failures');
const publishFailures = new Rate('notification_publish_failures');
const deliveryFailures = new Rate('notification_delivery_failures');
const deliveryLatency = new Trend('notification_delivery_latency', true);

const vus = Number(__ENV.VUS || 1);
const iterations = Number(__ENV.ITERATIONS || vus);

export const options = {
  scenarios: {
    notification_e2e: {
      executor: 'shared-iterations',
      vus,
      iterations,
      maxDuration: __ENV.MAX_DURATION || '5m',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    notification_setup_failures: ['rate==0'],
    notification_publish_failures: ['rate==0'],
    notification_delivery_failures: ['rate==0'],
    notification_delivery_latency: ['p(95)<20000'],
  },
};

const jsonParams = {
  headers: { 'Content-Type': 'application/json' },
  timeout: '10s',
};

function parseJSON(response) {
  try {
    return JSON.parse(response.body);
  } catch (_) {
    return null;
  }
}

function createUser(label, runID) {
  const username = `nk6_${label}_${runID}`.slice(0, 50);
  const email = `${username}@example.test`;
  const registration = http.post(
    `${AUTH_URL}/auth/register`,
    JSON.stringify({ username, email, password: PASSWORD }),
    jsonParams,
  );
  if (registration.status !== 201) {
    console.error(`registration failed: status=${registration.status} body=${registration.body}`);
  }
  let ok = check(registration, {
    'registro k6 devuelve 201': (r) => r.status === 201,
  });
  setupFailures.add(!ok);
  if (!ok) return null;

  const login = http.post(
    `${AUTH_URL}/auth/login`,
    JSON.stringify({ email, password: PASSWORD }),
    jsonParams,
  );
  const loginBody = parseJSON(login);
  ok = check(login, {
    'login k6 devuelve 200': (r) => r.status === 200,
    'login devuelve session_token': () => Boolean(loginBody?.data?.session_token),
  });
  setupFailures.add(!ok);
  if (!ok) return null;

  return {
    id: loginBody.data.user.id,
    token: loginBody.data.session_token,
  };
}

function publishEvent(routingKey, eventID, payload) {
  const response = http.post(
    `${RABBITMQ_MANAGEMENT_URL}/api/exchanges/${VHOST}/meloop.events/publish`,
    JSON.stringify({
      properties: {
        delivery_mode: 2,
        message_id: eventID,
        content_type: 'application/json',
      },
      routing_key: routingKey,
      payload: JSON.stringify({ event_id: eventID, ...payload }),
      payload_encoding: 'string',
    }),
    {
      headers: { 'Content-Type': 'application/json' },
      auth: 'basic',
      username: RABBITMQ_USER,
      password: RABBITMQ_PASSWORD,
      timeout: '10s',
    },
  );
  const body = parseJSON(response);
  const ok = check(response, {
    [`RabbitMQ acepta ${routingKey}`]: (r) => r.status === 200 && body?.routed === true,
  });
  publishFailures.add(!ok);
  return ok;
}

function listNotifications(token) {
  return http.get(`${BASE_URL}/notifications?limit=100`, {
    headers: { Authorization: `Bearer ${token}` },
    timeout: '10s',
  });
}

function waitForEvents(token, expectedEventIDs) {
  const startedAt = Date.now();
  const deadline = startedAt + POLL_TIMEOUT_SECONDS * 1000;
  let response;
  let actualByEventID = {};
  let eventCounts = {};

  while (Date.now() < deadline) {
    response = listNotifications(token);
    if (response.status === 200) {
      const body = parseJSON(response);
      const rows = body?.data?.notifications || [];
      actualByEventID = {};
      eventCounts = {};
      for (const row of rows) {
        if (row.id_evento) {
          actualByEventID[row.id_evento] = row;
          eventCounts[row.id_evento] = (eventCounts[row.id_evento] || 0) + 1;
        }
      }
      if (expectedEventIDs.every((id) => actualByEventID[id])) break;
    }
    sleep(0.5);
  }

  const latency = Date.now() - startedAt;
  deliveryLatency.add(latency);
  const ok = check(response, {
    'bandeja autenticada devuelve 200': (r) => r && r.status === 200,
    'el consumidor guardó todos los eventos': () => expectedEventIDs.every((id) => actualByEventID[id]),
    'las alertas nuevas empiezan sin leer': () => expectedEventIDs.every((id) => actualByEventID[id]?.leida === false),
  });
  deliveryFailures.add(!ok);
  return { byEventID: actualByEventID, counts: eventCounts };
}

export default function () {
  const runID = `${__VU}_${__ITER}_${Date.now()}`;
  const actor = createUser('actor', runID);
  const recipient = createUser('dest', runID);
  if (!actor || !recipient) fail('No se pudieron crear y autenticar los usuarios de prueba');

  const events = [
    {
      route: 'post.liked',
      id: `k6-like-${runID}`,
      payload: { userId: actor.id, recipientId: recipient.id, postId: `k6-post-${runID}` },
    },
    {
      route: 'comment.created',
      id: `k6-comment-${runID}`,
      payload: {
        comment_author_id: actor.id,
        target_author_id: recipient.id,
        target_type: 'post',
        target_id: `k6-post-${runID}`,
      },
    },
    {
      route: 'comment.created',
      id: `k6-reply-${runID}`,
      payload: {
        comment_id: `k6-comment-id-${runID}`,
        comment_author_id: actor.id,
        target_author_id: recipient.id,
        target_type: 'comment',
        target_id: `k6-parent-comment-${runID}`,
      },
      expectedType: 'comment.replied',
    },
    {
      route: 'friend.requested',
      id: `k6-friend-request-${runID}`,
      payload: { sender_id: actor.id, receiver_id: recipient.id },
    },
    {
      route: 'friend.accepted',
      id: `k6-friend-accepted-${runID}`,
      payload: { sender_id: recipient.id, receiver_id: actor.id },
      recipient: actor,
    },
    {
      route: 'message.sent',
      id: `k6-message-${runID}`,
      payload: { sender_id: actor.id, receiver_id: recipient.id, message_id: `k6-msg-${runID}` },
    },
    {
      route: 'report.created',
      id: `k6-report-${runID}`,
      payload: { actor_id: actor.id, recipient_id: recipient.id },
    },
    {
      route: 'moderation.action',
      id: `k6-moderation-${runID}`,
      payload: { moderator_id: actor.id, affected_user_id: recipient.id },
    },
    {
      route: 'user.level_up',
      id: `k6-level-${runID}`,
      payload: { user_id: recipient.id, level: 2 },
    },
    {
      route: 'reward.unlocked',
      id: `k6-reward-${runID}`,
      payload: { user_id: recipient.id, reward_id: `k6-reward-id-${runID}`, reward_type: 'INSIGNIA' },
    },
  ];

  const expected = [];
  for (const event of events) {
    const target = event.recipient || recipient;
    if (publishEvent(event.route, event.id, event.payload)) {
      expected.push({ user: target, eventID: event.id, type: event.expectedType || event.route });
    }
  }

  // Republish one event with the same ID. The inbox must still contain one row for it.
  const duplicate = events.find((event) => event.route === 'user.level_up');
  publishEvent(duplicate.route, duplicate.id, duplicate.payload);

  const usersToCheck = new Map();
  for (const item of expected) usersToCheck.set(item.user.id, item.user);
  for (const [userID, user] of usersToCheck) {
    const rows = waitForEvents(user.token, expected
      .filter((item) => item.user.id === userID)
      .map((item) => item.eventID));
    for (const item of expected.values()) {
      if (item.user.id !== userID) continue;
      check(rows.byEventID[item.eventID], {
        [`tipo ${item.type} se guardó correctamente`]: (row) => row?.tipo === item.type,
      });
    }
    if (userID === recipient.id) {
      check(rows.counts[duplicate.id], {
        'evento repetido no duplica la notificación': () => rows.counts[duplicate.id] === 1,
      });
    }
  }
}
