import http from 'k6/http';
import { check } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:18080';
const tokenFile = __ENV.TOKEN_FILE;

if (!tokenFile) {
  throw new Error('TOKEN_FILE must point to a session file outside the repository');
}

const users = JSON.parse(open(tokenFile));
if (!Array.isArray(users) || users.length !== 200) {
  throw new Error('TOKEN_FILE must contain exactly 200 prepared test sessions');
}

const friendsDuration = new Trend('friends_duration', true);
const suggestionsDuration = new Trend('suggestions_duration', true);
const functionalErrors = new Rate('functional_error_rate');
const functionalFailures = new Counter('functional_validation_failures');

export const options = {
  vus: 200,
  duration: '2m',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
    'http_req_duration{endpoint:friends}': ['p(95)<500'],
    'http_req_duration{endpoint:suggestions}': ['p(95)<500'],
  },
};

function payloadOf(response) {
  try {
    return response.json();
  } catch (error) {
    return null;
  }
}

function recordFunctionalResult(endpoint, passed) {
  functionalErrors.add(!passed, { endpoint: endpoint });
  functionalFailures.add(passed ? 0 : 1, { endpoint: endpoint });
}

function isFriendListValid(payload, user) {
  if (!payload || payload.success !== true || !Array.isArray(payload.data) ||
      payload.data.length !== user.expected_friend_ids.length) {
    return false;
  }

  const expected = new Set(user.expected_friend_ids);
  const seen = new Set();
  for (const friend of payload.data) {
    if (!friend || friend.estado !== 'ACEPTADA' ||
        friend.id_usuario === user.user_id ||
        !expected.has(friend.id_usuario) ||
        seen.has(friend.id_usuario)) {
      return false;
    }
    seen.add(friend.id_usuario);
  }
  return seen.size === expected.size;
}

function areSuggestionsValid(payload, user) {
  if (!payload || payload.success !== true || !Array.isArray(payload.data) ||
      payload.data.length !== 469) {
    return false;
  }

  const directFriends = new Set(user.expected_friend_ids);
  const seen = new Set();
  let knownCandidateFound = false;
  for (const suggestion of payload.data) {
    if (!suggestion || typeof suggestion.id_usuario !== 'string' ||
        suggestion.id_usuario === user.user_id ||
        directFriends.has(suggestion.id_usuario) ||
        seen.has(suggestion.id_usuario) ||
        !Number.isInteger(suggestion.amigos_en_comun) ||
        suggestion.amigos_en_comun < 0) {
      return false;
    }
    seen.add(suggestion.id_usuario);
    if (suggestion.id_usuario === user.expected_suggestion_id) {
      knownCandidateFound = suggestion.amigos_en_comun === user.expected_mutual_friends;
    }
  }
  return seen.size === 469 && knownCandidateFound;
}

export default function () {
  const user = users[(__VU - 1) % users.length];
  const headers = {
    headers: { Authorization: `Bearer ${user.token}` },
    tags: { endpoint: 'friends' },
  };

  const friendsResponse = http.get(`${BASE_URL}/friends`, headers);
  friendsDuration.add(friendsResponse.timings.duration);
  const friendsPayload = payloadOf(friendsResponse);
  const friendsValid = check(friendsResponse, {
    'friends returns HTTP 200': (response) => response.status === 200,
    'friends returns a successful JSON envelope': () =>
      friendsResponse.headers['Content-Type']?.includes('application/json') === true &&
      friendsPayload !== null && friendsPayload.success === true,
    'friends match the seeded identities and accepted state': () =>
      isFriendListValid(friendsPayload, user),
  });
  recordFunctionalResult('friends', friendsValid);

  const suggestionResponse = http.get(`${BASE_URL}/friends/suggestions`, {
    headers: { Authorization: `Bearer ${user.token}` },
    tags: { endpoint: 'suggestions' },
  });
  suggestionsDuration.add(suggestionResponse.timings.duration);
  const suggestionPayload = payloadOf(suggestionResponse);
  const suggestionsValid = check(suggestionResponse, {
    'suggestions returns HTTP 200': (response) => response.status === 200,
    'suggestions returns a successful JSON envelope': () =>
      suggestionResponse.headers['Content-Type']?.includes('application/json') === true &&
      suggestionPayload !== null && suggestionPayload.success === true,
    'suggestions contain unique eligible candidates and valid mutual counts': () =>
      areSuggestionsValid(suggestionPayload, user),
  });
  recordFunctionalResult('suggestions', suggestionsValid);
}
