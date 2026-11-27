// Acceptance/load test for the complete available recommendation path.
// Prerequisites: Gateway, ML service, Recommendation Service, and a test DB
// loaded with supabase/seed.sql (the default orchestration user is ID 1).
// Run from the repo root: k6 run ml_k6_test.js
// Override endpoints with GATEWAY_URL, ML_URL, RECOMMENDATION_URL; set
// RECOMMENDATION_URL='' to test registration/profile + direct ML only.
// Each VU registers one account; the current API has no account cleanup route.

import http from 'k6/http';
import { check, fail, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const VUS = Number(__ENV.VUS || 5);
const DURATION = __ENV.DURATION || '30s';
const GATEWAY_URL = (__ENV.GATEWAY_URL || 'http://[::1]:8080').replace(/\/$/, '');
const ML_URL = (__ENV.ML_URL || 'http://127.0.0.1:8001').replace(/\/$/, '');
const RECOMMENDATION_URL = (__ENV.RECOMMENDATION_URL === undefined
  ? 'http://127.0.0.1:8087'
  : __ENV.RECOMMENDATION_URL).replace(/\/$/, '');
const RECOMMENDATION_USER_ID = __ENV.RECOMMENDATION_USER_ID === undefined
  ? '1'
  : __ENV.RECOMMENDATION_USER_ID;
const RUN_ID = __ENV.RUN_ID || `${Date.now()}`;

const setupFailures = new Counter('ml_setup_failures');
const directMLErrorRate = new Rate('ml_direct_error_rate');
const orchestratorErrorRate = new Rate('recommendation_orchestrator_error_rate');
const musicLatency = new Trend('ml_music_latency', true);
const friendsLatency = new Trend('ml_friends_latency', true);

export const options = {
  scenarios: {
    ml_recommendations: {
      executor: 'constant-vus',
      vus: VUS,
      duration: DURATION,
      gracefulStop: '10s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    ml_direct_error_rate: ['rate<0.05'],
    recommendation_orchestrator_error_rate: ['rate<0.05'],
    ml_setup_failures: ['count==0'],
    ml_music_latency: ['p(95)<1000'],
    ml_friends_latency: ['p(95)<1000'],
  },
};

const personas = [
  { genres: ['indie-pop', 'alternative'], artists: ['Artist A'], songs: ['track-a'] },
  { genres: ['indie-pop', 'rock'], artists: ['Artist B'], songs: ['track-b'] },
  { genres: ['jazz', 'soul'], artists: ['Artist C'], songs: ['track-c'] },
  { genres: ['electronic', 'synth-pop'], artists: ['Artist D'], songs: ['track-d'] },
  { genres: ['rock', 'alternative'], artists: ['Artist E'], songs: ['track-e'] },
  { genres: ['jazz', 'indie-pop'], artists: ['Artist F'], songs: ['track-f'] },
  { genres: ['electronic', 'rock'], artists: ['Artist G'], songs: ['track-g'] },
  { genres: ['soul', 'alternative'], artists: ['Artist H'], songs: ['track-h'] },
];

function jsonParams(token = '') {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;
  return { headers, timeout: '10s' };
}

function responseData(response) {
  try {
    const body = response.json();
    if (body && body.success) return body.data;
    return body && !body.error ? body : null;
  } catch (_) {
    return null;
  }
}

function logHTTPFailure(label, response) {
  if (response.status >= 200 && response.status < 400) return;
  let detail = '';
  try {
    const body = response.json();
    detail = body && body.error && body.error.message ? body.error.message : '';
  } catch (_) {
    detail = String(response.body || '').slice(0, 240);
  }
  console.error(`${label} respondió HTTP ${response.status}${detail ? `: ${detail}` : ''}`);
}

function createAndProfileUser(index) {
  const basePersona = personas[index % personas.length];
  const persona = {
    genres: basePersona.genres,
    artists: [...basePersona.artists, `ml-k6-artist-${index + 1}`],
    songs: [...basePersona.songs, `ml-k6-track-${index + 1}`],
  };
  const suffix = `${RUN_ID}-${index + 1}`;
  const credentials = {
    username: `ml_k6_${suffix}`,
    email: `ml-k6-${suffix}@example.test`,
    password: 'MeloopK6-Test-2026!',
  };

  const registered = http.post(
    `${GATEWAY_URL}/auth/register`,
    JSON.stringify(credentials),
    jsonParams(),
  );
  logHTTPFailure(`Registro del usuario k6 #${index + 1}`, registered);
  const registrationData = responseData(registered);
  if (
    !check(registered, {
      'registro de usuario de prueba devuelve 201': (response) => response.status === 201,
      'registro devuelve ID de usuario': () => Boolean(registrationData && registrationData.id),
    })
  ) {
    setupFailures.add(1);
    return null;
  }

  const loggedIn = http.post(
    `${GATEWAY_URL}/auth/login`,
    JSON.stringify({ email: credentials.email, password: credentials.password }),
    jsonParams(),
  );
  logHTTPFailure(`Login del usuario k6 #${index + 1}`, loggedIn);
  const session = responseData(loggedIn);
  if (
    !check(loggedIn, {
      'login devuelve 200': (response) => response.status === 200,
      'login devuelve token de sesión': () => Boolean(session && session.session_token),
    })
  ) {
    setupFailures.add(1);
    return null;
  }

  const profilePayload = {
    biografia: 'Cuenta temporal para prueba de recomendaciones ML',
    informacion_musical: {
      generos: persona.genres,
      artistas: persona.artists,
      canciones: persona.songs,
    },
  };
  const updated = http.put(
    `${GATEWAY_URL}/users/me/profile`,
    JSON.stringify(profilePayload),
    jsonParams(session.session_token),
  );
  logHTTPFailure(`Guardado de perfil k6 #${index + 1}`, updated);
  const updatedProfile = responseData(updated);
  if (
    !check(updated, {
      'perfil musical se persiste mediante Gateway': (response) => response.status === 200,
      'respuesta contiene preferencias musicales': () =>
        Boolean(updatedProfile && updatedProfile.informacion_musical),
    })
  ) {
    setupFailures.add(1);
    return null;
  }

  // Re-read the profile so the ML requests use values that were persisted by user-service.
  const fetched = http.get(`${GATEWAY_URL}/users/me/profile`, jsonParams(session.session_token));
  logHTTPFailure(`Lectura de perfil k6 #${index + 1}`, fetched);
  const fetchedProfile = responseData(fetched);
  if (
    !check(fetched, {
      'perfil musical persistido se puede recuperar': (response) =>
        response.status === 200 &&
        Boolean(responseData(response) && responseData(response).informacion_musical),
      'perfil persistido contiene géneros, artistas y canciones': () => {
        const music = fetchedProfile && fetchedProfile.informacion_musical;
        return Boolean(
          music && music.generos && music.generos.length > 0 &&
          music.artistas && music.artistas.length > 0 &&
          music.canciones && music.canciones.length > 0,
        );
      },
    })
  ) {
    setupFailures.add(1);
    return null;
  }

  return {
    userId: registrationData.id,
    token: session.session_token,
    profile: fetchedProfile.informacion_musical,
    mlUserId: 100000 + index + 1,
  };
}

export function setup() {
  if (VUS < 2) fail('VUS debe ser al menos 2 para probar recomendaciones entre usuarios');

  const readiness = http.get(`${ML_URL}/ready`);
  if (!check(readiness, { 'ML service está listo': (response) => response.status === 200 })) {
    fail(`ML service no está listo en ${ML_URL}`);
  }
  if (RECOMMENDATION_URL) {
    const recommendationReadiness = http.get(`${RECOMMENDATION_URL}/ready`);
    if (!check(recommendationReadiness, {
      'Recommendation Service y su dependencia ML están listos':
        (response) => response.status === 200,
    })) {
      fail(`Recommendation Service no está listo en ${RECOMMENDATION_URL}`);
    }
  }

  const accounts = [];
  for (let index = 0; index < VUS; index += 1) {
    const account = createAndProfileUser(index);
    if (account) accounts.push(account);
  }
  if (accounts.length !== VUS) {
    fail(`Se crearon ${accounts.length} de ${VUS} usuarios de prueba`);
  }
  return { accounts };
}

function makeCatalog(accounts) {
  return accounts.map((account, index) => ({
    item_id: 200000 + index + 1,
    profile: {
      genres: account.profile.generos || [],
      artists: account.profile.artistas || [],
      songs: account.profile.canciones || [],
    },
  }));
}

function testMusic(account, accounts) {
  const userCatalogItem = 200000 + (account.mlUserId - 100001) + 1;
  const likedItemId = 200000 + ((account.mlUserId - 100001 + 1) % accounts.length) + 1;
  const catalog = makeCatalog(accounts);
  const payload = {
    user_id: account.mlUserId,
    limit: Math.min(5, catalog.length),
    profile: {
      genres: account.profile.generos || [],
      artists: account.profile.artistas || [],
      songs: account.profile.canciones || [],
    },
    interactions: [{ type: 'like', target_id: likedItemId }],
    catalog,
  };

  const started = Date.now();
  const response = http.post(
    `${ML_URL}/recommendations/music`,
    JSON.stringify(payload),
    jsonParams(),
  );
  musicLatency.add(Date.now() - started);
  const body = responseData(response);
  const recs = body && Array.isArray(body.recommendations) ? body.recommendations : [];
  const sorted = recs.every((item, index) => index === 0 || recs[index - 1].score >= item.score);
  const likedExcluded = recs.every((item) => item.item_id !== likedItemId);
  const personalizedTopResult = recs.length > 0 && recs[0].item_id === userCatalogItem;
  const ok = check(response, {
    'endpoint musical devuelve 200': (r) => r.status === 200,
    'respuesta musical identifica el modelo de contenido': () =>
      Boolean(body && body.model === 'music_content_similarity'),
    'ranking musical devuelve resultados': () => recs.length > 0,
    'ranking musical está ordenado por score': () => sorted,
    'canción ya marcada con like queda fuera': () => likedExcluded,
    'perfil persistido obtiene su canción más afín primero': () => personalizedTopResult,
  });
  directMLErrorRate.add(!ok);
}

function testFriends(account, accounts) {
  const payload = {
    user_id: account.mlUserId,
    limit: Math.min(5, accounts.length - 1),
    profile: {
      genres: account.profile.generos || [],
      artists: account.profile.artistas || [],
      songs: account.profile.canciones || [],
    },
    candidate_profiles: accounts
      .filter((candidate) => candidate.mlUserId !== account.mlUserId)
      .map((candidate) => ({
        user_id: candidate.mlUserId,
        profile: {
          genres: candidate.profile.generos || [],
          artists: candidate.profile.artistas || [],
          songs: candidate.profile.canciones || [],
        },
      })),
  };

  const started = Date.now();
  const response = http.post(
    `${ML_URL}/recommendations/friends`,
    JSON.stringify(payload),
    jsonParams(),
  );
  friendsLatency.add(Date.now() - started);
  const body = responseData(response);
  const recs = body && Array.isArray(body.recommendations) ? body.recommendations : [];
  const noSelf = recs.every((item) => item.item_id !== account.mlUserId);
  const sorted = recs.every((item, index) => index === 0 || recs[index - 1].score >= item.score);
  const ok = check(response, {
    'endpoint de amistades devuelve 200': (r) => r.status === 200,
    'respuesta de amistades contiene candidatos': () => recs.length > 0,
    'usuario no se recomienda a sí mismo': () => noSelf,
    'ranking de amistades está ordenado por score': () => sorted,
  });
  directMLErrorRate.add(!ok);
}

function testRecommendationOrchestrator() {
  if (!RECOMMENDATION_URL || !RECOMMENDATION_USER_ID) return;

  const musicResponse = http.get(
    `${RECOMMENDATION_URL}/recommendations/${RECOMMENDATION_USER_ID}/music?limit=10`,
    { timeout: '10s' },
  );
  const musicBody = musicResponse.json();
  const musicOK = check(musicResponse, {
    'Recommendation Service responde 200 para música': (r) => r.status === 200,
    'orquestador obtiene respuesta musical directa de ML': (r) =>
      r.headers['X-Recommendations-Source'] === 'ml',
    'orquestador entrega canciones del catálogo': () =>
      Boolean(musicBody && Array.isArray(musicBody.recommendations) && musicBody.recommendations.length > 0),
  });
  orchestratorErrorRate.add(!musicOK);

  const friendsResponse = http.get(
    `${RECOMMENDATION_URL}/recommendations/${RECOMMENDATION_USER_ID}/friends?limit=10`,
    { timeout: '10s' },
  );
  const friendsBody = friendsResponse.json();
  const friendsOK = check(friendsResponse, {
    'Recommendation Service responde 200 para amistades': (r) => r.status === 200,
    'orquestador obtiene respuesta de amistades directa de ML': (r) =>
      r.headers['X-Recommendations-Source'] === 'ml',
    'orquestador devuelve candidatos de amistad': () =>
      Boolean(friendsBody && Array.isArray(friendsBody.recommendations) && friendsBody.recommendations.length > 0),
  });
  orchestratorErrorRate.add(!friendsOK);
}

export default function (data) {
  const account = data.accounts[(__VU - 1) % data.accounts.length];
  testMusic(account, data.accounts);
  testFriends(account, data.accounts);
  testRecommendationOrchestrator();
  sleep(0.2);
}
