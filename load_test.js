import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Metricas personalizadas
const errorRate = new Rate('error_rate');
const profileLatency = new Trend('profile_latency');
const socialLatency = new Trend('social_latency');

// Configuracion del escenario de carga concurrente
export const options = {
  stages: [
    { duration: '5s', target: 10 },  // Rampa de subida a 10 usuarios simultaneos
    { duration: '20s', target: 25 }, // Carga sostenida con 25 usuarios concurrentes
    { duration: '5s', target: 0 },   // Rampa de bajada
  ],
  thresholds: {
    // Definicion de objetivos de rendimiento (SLAs)
    http_req_failed: ['rate<0.05'],           // Menos del 5% de errores
    http_req_duration: ['p(95)<400'],         // 95% de peticiones por debajo de 400ms
    error_rate: ['rate<0.05'],
  },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  // Generar identificadores unicos por usuario virtual (VU) e iteracion
  const uniqueId = `${__VU}_${__ITER}_${Date.now()}`;
  const username = `load_user_${uniqueId}`;
  const email = `user_${uniqueId}@loadtest.com`;
  const password = 'Password123!';

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  let userId = null;

  // -------------------------------------------------------------
  // GRUPO 1: Health Check del API Gateway
  // -------------------------------------------------------------
  group('01. Gateway Health', function () {
    const res = http.get(`${BASE_URL}/health`);
    const ok = check(res, {
      'health status es 200': (r) => r.status === 200,
    });
    errorRate.add(!ok);
  });

  // -------------------------------------------------------------
  // GRUPO 2: Registro de Usuario (auth-service / Postgres)
  // -------------------------------------------------------------
  group('02. Registro de Usuario', function () {
    const payload = JSON.stringify({
      username: username,
      email: email,
      password: password,
    });

    const res = http.post(`${BASE_URL}/auth/register`, payload, params);
    const ok = check(res, {
      'registro exitoso (201)': (r) => r.status === 201,
      'devuelve id de usuario': (r) => {
        try {
          const body = JSON.parse(r.body);
          if (body.data && body.data.id) {
            userId = body.data.id;
            return true;
          }
        } catch (e) {}
        return false;
      },
    });
    errorRate.add(!ok);
  });

  // Si no se obtuvo ID, generamos uno temporal para continuar las pruebas
  if (!userId) {
    userId = `fallback_${uniqueId}`;
  }

  const authParams = {
    headers: {
      'Content-Type': 'application/json',
      'X-User-ID': userId,
    },
  };

  // -------------------------------------------------------------
  // GRUPO 3: Perfil y Privacidad (user-service)
  // -------------------------------------------------------------
  group('03. Gestion de Usuario (Perfil y Privacidad)', function () {
    // 3.1 Consultar cuenta
    const accRes = http.get(`${BASE_URL}/users/me`, authParams);
    check(accRes, {
      'consulta de cuenta exitosa (200)': (r) => r.status === 200,
    });

    // 3.2 Actualizar biografia de perfil
    const startProfile = Date.now();
    const profilePayload = JSON.stringify({
      biografia: `Usuario concurrente en prueba de carga ${uniqueId}`,
      tema: 'neon_dark',
    });
    const profRes = http.put(`${BASE_URL}/users/me/profile`, profilePayload, authParams);
    profileLatency.add(Date.now() - startProfile);

    const profOk = check(profRes, {
      'actualizacion de perfil exitosa (200)': (r) => r.status === 200,
    });
    errorRate.add(!profOk);

    // 3.3 Actualizar privacidad
    const privPayload = JSON.stringify({
      visibilidad_perfil: 'PRIVADO',
      recepcion_mensajes: 'AMIGOS',
    });
    const privRes = http.patch(`${BASE_URL}/users/me/privacy`, privPayload, authParams);
    const privOk = check(privRes, {
      'actualizacion de privacidad exitosa (200)': (r) => r.status === 200,
    });
    errorRate.add(!privOk);
  });

  // -------------------------------------------------------------
  // GRUPO 4: Operaciones Sociales (social-service)
  // -------------------------------------------------------------
  group('04. Operaciones Sociales', function () {
    const startSocial = Date.now();

    // 4.1 Listar amigos
    const friendsRes = http.get(`${BASE_URL}/friends`, authParams);
    check(friendsRes, {
      'listado de amigos exitoso (200)': (r) => r.status === 200,
    });

    // 4.2 Listar solicitudes recibidas
    const reqsRes = http.get(`${BASE_URL}/friends/requests/received`, authParams);
    check(reqsRes, {
      'listado de solicitudes exitoso (200)': (r) => r.status === 200,
    });

    socialLatency.add(Date.now() - startSocial);
  });

  // Pausa realista entre acciones de usuario (entre 0.5s y 1s)
  sleep(0.5);
}
