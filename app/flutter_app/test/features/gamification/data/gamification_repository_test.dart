import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/features/gamification/data/gamification_repository.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  const endpoint = 'https://example.test/v1/gamification/me/progress';

  test('fetches and parses user progress and rewards', () async {
    final repository = HttpGamificationRepository(
      endpoint: Uri.parse(endpoint),
      accessToken: () async => 'token',
      client: MockClient((request) async {
        expect(request.method, 'GET');
        expect(request.url, Uri.parse(endpoint));
        expect(request.headers['Authorization'], 'Bearer token');
        return http.Response(
          jsonEncode({
            'success': true,
            'data': {
              'user_id': 'user-1',
              'id_nivel': 2,
              'level_number': 3,
              'experiencia': 70,
              'rewards': [
                {
                  'id_recompensa': 'reward-1',
                  'tipo': 'Marco RGB',
                  'id_nivel': 3,
                  'equipada': true,
                },
              ],
            },
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      }),
    );

    final progress = await repository.fetchUserProgress();

    expect(progress.userId, 'user-1');
    expect(progress.levelId, 2);
    expect(progress.levelNumber, 3);
    expect(progress.experience, 70);
    expect(progress.rewards, hasLength(1));
    expect(progress.rewards.single.id, 'reward-1');
    expect(progress.rewards.single.type, 'Marco RGB');
    expect(progress.rewards.single.levelId, 3);
    expect(progress.rewards.single.equipped, isTrue);
    repository.dispose();
  });

  test('returns an empty rewards list when a user has no rewards', () async {
    final repository = HttpGamificationRepository(
      endpoint: Uri.parse(endpoint),
      accessToken: () async => null,
      client: MockClient(
        (_) async => http.Response(
          jsonEncode({
            'success': true,
            'data': {
              'user_id': 'user-1',
              'id_nivel': 1,
              'level_number': 1,
              'experiencia': 0,
            },
          }),
          200,
        ),
      ),
    );

    final progress = await repository.fetchUserProgress();

    expect(progress.rewards, isEmpty);
    repository.dispose();
  });

  test(
    'throws a standard HTTP client error for unsuccessful responses',
    () async {
      final repository = HttpGamificationRepository(
        endpoint: Uri.parse(endpoint),
        accessToken: () async => null,
        client: MockClient((_) async => http.Response('', 401)),
      );

      await expectLater(
        repository.fetchUserProgress(),
        throwsA(isA<http.ClientException>()),
      );
      repository.dispose();
    },
  );
}
