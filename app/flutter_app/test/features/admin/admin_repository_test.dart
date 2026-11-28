import 'package:flutter_app/features/admin/data/admin_models.dart';
import 'package:flutter_app/features/admin/data/admin_repository.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late DemoAdminRepository repository;

  setUp(() {
    repository = DemoAdminRepository(clock: () => DateTime(2026, 11, 2, 12));
  });

  test(
    'demo catalog contains rewards from the existing profile inventory',
    () async {
      final rewards = await repository.fetchRewards();
      expect(
        rewards.map((reward) => reward.name),
        containsAll([
          'Marco Neón',
          'Marco Dorado',
          'Pin Zafiro',
          'Modo Nocturno',
          'Marco RGB',
        ]),
      );
    },
  );

  test(
    'moderator role assignment, removal, and suspended account error work',
    () async {
      final matches = await repository.searchUsersByUsername('vicente');
      final user = matches.single;

      await repository.updateUserRole(user.id, AdminUserRole.moderator);
      expect(
        (await repository.fetchUsers(
          role: AdminUserRole.moderator,
          pageSize: 20,
        )).users.map((entry) => entry.id),
        contains(user.id),
      );

      await repository.updateUserRole(user.id, AdminUserRole.user);
      expect(
        (await repository.fetchUsers(
          role: AdminUserRole.moderator,
          pageSize: 20,
        )).users.map((entry) => entry.id),
        isNot(contains(user.id)),
      );

      await expectLater(
        repository.updateUserRole('user-5', AdminUserRole.moderator),
        throwsA(
          isA<AdminRepositoryException>().having(
            (error) => error.code,
            'code',
            'USER_SUSPENDED',
          ),
        ),
      );
    },
  );

  test('user filters paginate with cursors and do not expose email', () async {
    final firstPage = await repository.fetchUsers(pageSize: 2);
    expect(firstPage.users, hasLength(2));
    expect(firstPage.nextCursor, isNotNull);
    expect(firstPage.users.every((user) => user.username.isNotEmpty), isTrue);

    final secondPage = await repository.fetchUsers(
      pageSize: 2,
      cursor: firstPage.nextCursor,
    );
    expect(
      secondPage.users.map((user) => user.id),
      isNot(containsAll(firstPage.users.map((user) => user.id))),
    );

    final suspended = await repository.fetchUsers(
      status: AdminUserStatus.suspended,
    );
    expect(suspended.users, hasLength(1));
    expect(suspended.users.single.status, AdminUserStatus.suspended);
  });

  test(
    'reward filters, toggle, duplicate name, and level errors work',
    () async {
      await repository.setRewardEnabled('f1', false);
      expect(
        (await repository.fetchRewards(enabled: false)).map((item) => item.id),
        contains('f1'),
      );
      expect(
        await repository.fetchRewards(type: AdminRewardType.frame),
        isNotEmpty,
      );

      await expectLater(
        repository.saveReward(
          const AdminReward(
            id: 'new',
            name: 'Marco Neón',
            description: 'Duplicada',
            type: AdminRewardType.frame,
            requiredLevel: 1,
            enabled: true,
          ),
        ),
        throwsA(
          isA<AdminRepositoryException>().having(
            (error) => error.code,
            'code',
            'DUPLICATE_NAME',
          ),
        ),
      );

      await expectLater(
        repository.saveReward(
          const AdminReward(
            id: 'new',
            name: 'Nueva',
            description: 'Descripción',
            type: AdminRewardType.badge,
            requiredLevel: 99,
            enabled: true,
          ),
        ),
        throwsA(
          isA<AdminRepositoryException>().having(
            (error) => error.code,
            'code',
            'LEVEL_NOT_FOUND',
          ),
        ),
      );
    },
  );
}
