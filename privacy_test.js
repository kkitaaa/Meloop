import http from 'k6/http';
import { check, group } from 'k6';

// Flujo funcional RF-52. Usa cuentas desechables creadas en cada iteración.
export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    checks: ['rate==1'],
  },
};

const BASE_URL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const PASSWORD = __ENV.TEST_PASSWORD || 'PrivacyTest123!';

function requestParams(token) {
  return {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };
}

function readData(response) {
  try {
    return JSON.parse(response.body).data;
  } catch (_) {
    return null;
  }
}

function createAccount(label, suffix) {
  const username = `privacy_${label}_${suffix}`;
  const email = `${username}@example.test`;
  const registration = http.post(
    `${BASE_URL}/auth/register`,
    JSON.stringify({ username, email, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  const registered = check(registration, {
    [`${label}: registro responde 201`]: (r) => r.status === 201,
  });
  if (!registered) return null;

  const login = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  const loginData = readData(login);
  const authenticated = check(login, {
    [`${label}: login responde 200`]: (r) => r.status === 200,
    [`${label}: login entrega token e ID`]: () => Boolean(loginData?.session_token && loginData?.user?.id),
  });
  if (!authenticated) return null;

  return { id: loginData.user.id, token: loginData.session_token };
}

export default function () {
  const suffix = `${__VU}_${__ITER}_${Date.now()}`;
  let owner;
  let sender;

  group('Preparar usuarios de prueba', () => {
    owner = createAccount('owner', suffix);
    sender = createAccount('sender', suffix);
  });
  if (!owner || !sender) return;

  group('Leer y guardar preferencias de privacidad', () => {
    const initial = http.get(`${BASE_URL}/users/me/privacy`, requestParams(owner.token));
    check(initial, {
      'GET privacidad responde 200': (r) => r.status === 200,
      'GET privacidad devuelve valores por defecto': (r) => {
        const data = readData(r);
        return data?.recepcion_mensajes === 'AMIGOS'
          && data?.recepcion_solicitudes_amistad === 'TODOS';
      },
    });

    const restricted = http.patch(
      `${BASE_URL}/users/me/privacy`,
      JSON.stringify({
        visibilidad_perfil: 'PRIVADO',
        visibilidad_publicaciones: 'AMIGOS',
        recepcion_mensajes: 'NADIE',
        recepcion_solicitudes_amistad: 'NADIE',
      }),
      requestParams(owner.token),
    );
    check(restricted, {
      'PATCH privacidad restringida responde 200': (r) => r.status === 200,
      'PATCH devuelve preferencias guardadas': (r) => {
        const data = readData(r);
        return data?.visibilidad_perfil === 'PRIVADO'
          && data?.recepcion_mensajes === 'NADIE'
          && data?.recepcion_solicitudes_amistad === 'NADIE';
      },
    });

    const persisted = http.get(`${BASE_URL}/users/me/privacy`, requestParams(owner.token));
    check(persisted, {
      'GET confirma persistencia de preferencias': (r) => {
        const data = readData(r);
        return r.status === 200
          && data?.visibilidad_perfil === 'PRIVADO'
          && data?.recepcion_mensajes === 'NADIE'
          && data?.recepcion_solicitudes_amistad === 'NADIE';
      },
    });
  });

  group('Aplicar restricciones a interacciones', () => {
    const messageDenied = http.post(
      `${BASE_URL}/friends/validate-message`,
      JSON.stringify({ id_usuario: owner.id }),
      requestParams(sender.token),
    );
    check(messageDenied, {
      'NADIE deniega mensajes con 403': (r) => r.status === 403,
    });

    const requestDenied = http.post(
      `${BASE_URL}/friends/requests`,
      JSON.stringify({ id_usuario: owner.id }),
      requestParams(sender.token),
    );
    check(requestDenied, {
      'NADIE deniega solicitudes con 403': (r) => r.status === 403,
    });
  });

  group('Permitir interacciones tras cambiar preferencias', () => {
    const open = http.patch(
      `${BASE_URL}/users/me/privacy`,
      JSON.stringify({
        visibilidad_perfil: 'PUBLICO',
        visibilidad_publicaciones: 'PUBLICO',
        recepcion_mensajes: 'TODOS',
        recepcion_solicitudes_amistad: 'TODOS',
      }),
      requestParams(owner.token),
    );
    check(open, { 'Cambiar privacidad a TODOS responde 200': (r) => r.status === 200 });

    const messageAllowed = http.post(
      `${BASE_URL}/friends/validate-message`,
      JSON.stringify({ id_usuario: owner.id }),
      requestParams(sender.token),
    );
    check(messageAllowed, {
      'TODOS permite mensajes con 200': (r) => r.status === 200,
    });

    const requestAllowed = http.post(
      `${BASE_URL}/friends/requests`,
      JSON.stringify({ id_usuario: owner.id }),
      requestParams(sender.token),
    );
    check(requestAllowed, {
      'TODOS permite solicitudes con 201': (r) => r.status === 201,
    });
  });
}
