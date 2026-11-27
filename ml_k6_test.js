// End-to-end recommendation test using alphanumeric Auth-shaped users from supabase/seed.sql.
// No accounts or database records are created by this test.
// Run: k6 run -e VUS=5 -e DURATION=30s .\ml_k6_test.js
// Services: ML, Recommendation Service, Redis, and a DB loaded with supabase/seed.sql.
// For ML-only validation, set RECOMMENDATION_URL=none.

import http from 'k6/http';
import { check, fail, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const VUS = Number(__ENV.VUS || 5);
const DURATION = __ENV.DURATION || '30s';
const ML_URL = (__ENV.ML_URL || 'http://127.0.0.1:8001').replace(/\/$/, '');
const configuredRecommendationURL = __ENV.RECOMMENDATION_URL;
const RECOMMENDATION_URL = configuredRecommendationURL === 'none'
  ? ''
  : (configuredRecommendationURL === undefined
    ? 'http://127.0.0.1:8087'
    : configuredRecommendationURL).replace(/\/$/, '');
const RECOMMENDATION_USER_ID = __ENV.RECOMMENDATION_USER_ID ||
  'a0000000000000000000000000000001';
const LIKED_SONG_ID = __ENV.LIKED_SONG_ID || 'seed-track-002';

const directMLErrorRate = new Rate('ml_direct_error_rate');
const orchestratorErrorRate = new Rate('recommendation_orchestrator_error_rate');
const musicLatency = new Trend('ml_music_latency', true);
const friendsLatency = new Trend('ml_friends_latency', true);
const loggedOrchestratorFailures = {};

function logOrchestratorFailure(name, response) {
  if (loggedOrchestratorFailures[name]) return;
  loggedOrchestratorFailures[name] = true;
  console.error(
    `[orchestrator:${name}] status=${response.status} ` +
    `source=${response.headers['X-Recommendations-Source'] || '<missing>'} ` +
    `body=${String(response.body || '').slice(0, 1000)}`,
  );
}

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
    ml_music_latency: ['p(95)<1000'],
    ml_friends_latency: ['p(95)<1000'],
  },
};

const seedUserIDs = [
  'a0000000000000000000000000000001',
  'b0000000000000000000000000000002',
  'c0000000000000000000000000000003',
  'd0000000000000000000000000000004',
  'e0000000000000000000000000000005',
];

const personas = [
  { genres: ['indie-pop', 'alternative'], artists: ['Artist A'], songs: ['track-a'] },
  { genres: ['indie-pop', 'rock'], artists: ['Artist B'], songs: ['track-b'] },
  { genres: ['jazz', 'soul'], artists: ['Artist C'], songs: ['track-c'] },
  { genres: ['electronic', 'synth-pop'], artists: ['Artist D'], songs: ['track-d'] },
  { genres: ['rock', 'alternative'], artists: ['Artist E'], songs: ['track-e'] },
];

function jsonParams() {
  return { headers: { 'Content-Type': 'application/json' }, timeout: '10s' };
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

export function setup() {
  if (VUS < 2 || VUS > seedUserIDs.length) {
    fail(`VUS debe estar entre 2 y ${seedUserIDs.length} (usuarios UUID del seed)`);
  }
  const readiness = http.get(`${ML_URL}/ready`);
  if (!check(readiness, { 'ML service está listo': (r) => r.status === 200 })) {
    fail(`ML service no está listo en ${ML_URL}`);
  }

  if (RECOMMENDATION_URL) {
    const recommendationReadiness = http.get(`${RECOMMENDATION_URL}/ready`);
    if (!check(recommendationReadiness, {
      'Recommendation Service y ML están listos': (r) => r.status === 200,
    })) {
      fail(`Recommendation Service no está listo en ${RECOMMENDATION_URL}`);
    }
  }

  return {
    accounts: seedUserIDs.slice(0, VUS).map((id, index) => ({
      id,
      profile: {
        genres: personas[index].genres,
        artists: personas[index].artists,
        songs: [...personas[index].songs, `ml-k6-track-${index + 1}`],
      },
    })),
  };
}

function makeCatalog(accounts) {
  return accounts.map((account) => ({
    item_id: account.profile.songs[0],
    profile: account.profile,
  }));
}

function testMusic(account, accounts) {
  const accountIndex = accounts.findIndex((candidate) => candidate.id === account.id);
  const likedItemID = accounts[(accountIndex + 1) % accounts.length].profile.songs[0];
  const catalog = makeCatalog(accounts);
  const payload = {
    user_id: account.id,
    limit: Math.min(5, catalog.length),
    profile: account.profile,
    interactions: [{ type: 'like', target_id: likedItemID }],
    catalog,
  };

  const started = Date.now();
  const response = http.post(`${ML_URL}/recommendations/music`, JSON.stringify(payload), jsonParams());
  musicLatency.add(Date.now() - started);
  const body = responseData(response);
  const recs = body && Array.isArray(body.recommendations) ? body.recommendations : [];
  const sorted = recs.every((item, index) => index === 0 || recs[index - 1].score >= item.score);
  const likedExcluded = recs.every((item) => item.item_id !== likedItemID);
  const personalizedTopResult = recs.length > 0 && recs[0].item_id === account.profile.songs[0];
  const ok = check(response, {
    'endpoint musical acepta user_id alfanumérico': (r) => r.status === 200,
    'respuesta musical identifica el modelo': () => body?.model === 'music_content_similarity',
    'ranking musical devuelve resultados ordenados': () => recs.length > 0 && sorted,
    'like excluye el ID de canción del catálogo': () => likedExcluded,
    'perfil obtiene su canción más afín primero': () => personalizedTopResult,
  });
  directMLErrorRate.add(!ok);
}

function testFriends(account, accounts) {
  const payload = {
    user_id: account.id,
    limit: Math.min(5, accounts.length - 1),
    profile: account.profile,
    candidate_profiles: accounts
      .filter((candidate) => candidate.id !== account.id)
      .map((candidate) => ({ user_id: candidate.id, profile: candidate.profile })),
  };

  const started = Date.now();
  const response = http.post(`${ML_URL}/recommendations/friends`, JSON.stringify(payload), jsonParams());
  friendsLatency.add(Date.now() - started);
  const body = responseData(response);
  const recs = body && Array.isArray(body.recommendations) ? body.recommendations : [];
  const noSelf = recs.every((item) => item.item_id !== account.id);
  const sorted = recs.every((item, index) => index === 0 || recs[index - 1].score >= item.score);
  const ok = check(response, {
    'amistades acepta IDs alfanuméricos de usuario y candidatos': (r) => r.status === 200,
    'recomendaciones de amistad contienen candidatos': () => recs.length > 0,
    'usuario no se recomienda a sí mismo': () => noSelf,
    'ranking de amistades está ordenado': () => sorted,
  });
  directMLErrorRate.add(!ok);
}

function testRecommendationOrchestrator() {
  if (!RECOMMENDATION_URL) return;

  const musicResponse = http.get(
    `${RECOMMENDATION_URL}/recommendations/${encodeURIComponent(RECOMMENDATION_USER_ID)}/music?limit=10`,
    { timeout: '10s' },
  );
  const musicBody = responseData(musicResponse);
  const likedSongExcluded = musicBody?.recommendations?.every((item) =>
    item.item_id !== LIKED_SONG_ID && item.item_key !== LIKED_SONG_ID,
  );
  const musicOK = check(musicResponse, {
    'orquestador acepta el ID alfanumérico del usuario del seed': (r) => r.status === 200,
    'orquestador obtiene respuesta de ML': (r) => r.headers['X-Recommendations-Source'] === 'ml',
    'orquestador entrega canciones del catálogo': () =>
      Boolean(musicBody && Array.isArray(musicBody.recommendations) && musicBody.recommendations.length > 0),
    'orquestador excluye la canción con like': () => likedSongExcluded,
  });
  orchestratorErrorRate.add(!musicOK);
  if (!musicOK) logOrchestratorFailure('music', musicResponse);

  const friendsResponse = http.get(
    `${RECOMMENDATION_URL}/recommendations/${encodeURIComponent(RECOMMENDATION_USER_ID)}/friends?limit=10`,
    { timeout: '10s' },
  );
  const friendsBody = responseData(friendsResponse);
  const friendsOK = check(friendsResponse, {
    'orquestador responde recomendaciones de amistad para ID alfanumérico': (r) => r.status === 200,
    'orquestador obtiene amistades de ML': (r) => r.headers['X-Recommendations-Source'] === 'ml',
    'orquestador devuelve candidatos': () =>
      Boolean(friendsBody && Array.isArray(friendsBody.recommendations) && friendsBody.recommendations.length > 0),
  });
  orchestratorErrorRate.add(!friendsOK);
  if (!friendsOK) logOrchestratorFailure('friends', friendsResponse);
}

export default function (data) {
  const account = data.accounts[(__VU - 1) % data.accounts.length];
  testMusic(account, data.accounts);
  testFriends(account, data.accounts);
  testRecommendationOrchestrator();
  sleep(0.2);
}
