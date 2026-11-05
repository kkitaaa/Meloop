import 'admin_models.dart';

class AdminRepositoryException implements Exception {
  const AdminRepositoryException(this.code);

  final String code;

  @override
  String toString() => 'AdminRepositoryException($code)';
}

abstract interface class AdminRepository {
  Future<AdminUserPage> fetchUsers({
    AdminUserStatus? status,
    AdminUserRole? role,
    String usernameQuery = '',
    String? cursor,
    int pageSize = 8,
  });

  Future<List<AdminUser>> searchUsersByUsername(String username);

  Future<void> updateUserRole(String userId, AdminUserRole role);

  Future<void> updateUserStatus(
    String userId,
    AdminUserStatus status, {
    String? reason,
  });

  Future<List<AdminHistoryEntry>> fetchUserHistory(String userId);

  Future<List<AdminReward>> fetchRewards({
    AdminRewardType? type,
    bool? enabled,
  });

  Future<AdminReward> saveReward(AdminReward reward);

  Future<void> setRewardEnabled(String rewardId, bool enabled);

  Future<String> createRewardImageUploadUrl(String fileName);
}

class DemoAdminRepository implements AdminRepository {
  DemoAdminRepository({DateTime Function()? clock})
    : _clock = clock ?? DateTime.now {
    final now = _clock();
    _users = List.generate(18, (index) {
      final role = switch (index) {
        0 => AdminUserRole.admin,
        1 || 2 => AdminUserRole.moderator,
        _ => AdminUserRole.user,
      };
      final status = switch (index) {
        4 => AdminUserStatus.suspended,
        5 => AdminUserStatus.inactive,
        6 => AdminUserStatus.pendingDeletion,
        _ => AdminUserStatus.active,
      };
      return AdminUser(
        id: 'user-${index + 1}',
        username: [
          'admin_meloop',
          'mod_luna',
          'mod_gouz',
          'usuario_04',
          'cuenta_suspendida',
          'usuario_inactivo',
          'cuenta_en_baja',
          'vicente',
          'maria',
          'rockfan',
          'jazzy',
          'sonic',
          'ambient',
          'catalina',
          'leo',
          'camila',
          'nicolas',
          'valentina',
        ][index],
        role: role,
        status: status,
        level: 1 + (index * 3 % 20),
        lastActivity: now.subtract(Duration(hours: index * 5 + 1)),
      );
    });
    _rewards = [
      const AdminReward(
        id: 'f1',
        name: 'Marco Neón',
        description: 'Un marco brillante para tu perfil.',
        type: AdminRewardType.frame,
        requiredLevel: 2,
        enabled: true,
      ),
      const AdminReward(
        id: 'f2',
        name: 'Marco Dorado',
        description: 'Un marco de color dorado.',
        type: AdminRewardType.frame,
        requiredLevel: 8,
        enabled: true,
      ),
      const AdminReward(
        id: 'p1',
        name: 'Pin Zafiro',
        description: 'Un pin azul para personalizar tu perfil.',
        type: AdminRewardType.badge,
        requiredLevel: 3,
        enabled: true,
      ),
      const AdminReward(
        id: 'bg1',
        name: 'Modo Nocturno',
        description: 'Fondo oscuro para tu perfil.',
        type: AdminRewardType.background,
        requiredLevel: 5,
        enabled: true,
      ),
      const AdminReward(
        id: 'frame-rgb',
        name: 'Marco RGB',
        description: 'Un marco animado multicolor para tu perfil.',
        type: AdminRewardType.frame,
        requiredLevel: 4,
        enabled: true,
      ),
    ];
  }

  final DateTime Function() _clock;
  late final List<AdminUser> _users;
  late final List<AdminReward> _rewards;
  int _nextRewardId = 1;

  @override
  Future<AdminUserPage> fetchUsers({
    AdminUserStatus? status,
    AdminUserRole? role,
    String usernameQuery = '',
    String? cursor,
    int pageSize = 8,
  }) async {
    await Future<void>.delayed(const Duration(milliseconds: 120));
    final normalizedQuery = usernameQuery.trim().toLowerCase();
    final filtered = _users.where((user) {
      return (status == null || user.status == status) &&
          (role == null || user.role == role) &&
          (normalizedQuery.isEmpty ||
              user.username.toLowerCase().contains(normalizedQuery));
    }).toList();

    final start = int.tryParse(cursor ?? '') ?? 0;
    final page = filtered.skip(start).take(pageSize).toList(growable: false);
    final nextIndex = start + page.length;
    return AdminUserPage(
      users: List.unmodifiable(page),
      nextCursor: nextIndex < filtered.length ? '$nextIndex' : null,
    );
  }

  @override
  Future<List<AdminUser>> searchUsersByUsername(String username) async {
    final page = await fetchUsers(usernameQuery: username, pageSize: 50);
    return page.users;
  }

  @override
  Future<void> updateUserRole(String userId, AdminUserRole role) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    final index = _users.indexWhere((user) => user.id == userId);
    if (index < 0) throw const AdminRepositoryException('USER_NOT_FOUND');
    final user = _users[index];
    if (role == AdminUserRole.moderator &&
        user.status == AdminUserStatus.suspended) {
      throw const AdminRepositoryException('USER_SUSPENDED');
    }
    _users[index] = user.copyWith(role: role);
  }

  @override
  Future<void> updateUserStatus(
    String userId,
    AdminUserStatus status, {
    String? reason,
  }) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    final index = _users.indexWhere((user) => user.id == userId);
    if (index < 0) throw const AdminRepositoryException('USER_NOT_FOUND');
    final user = _users[index];
    if (user.status == status) {
      throw const AdminRepositoryException('ALREADY_IN_STATE');
    }
    if (user.role == AdminUserRole.admin) {
      throw const AdminRepositoryException('FORBIDDEN');
    }
    if (status == AdminUserStatus.suspended &&
        (reason == null ||
            reason.trim().length < 5 ||
            reason.trim().length > 300)) {
      throw const AdminRepositoryException('INVALID_REASON');
    }
    _users[index] = user.copyWith(status: status);
  }

  @override
  Future<List<AdminHistoryEntry>> fetchUserHistory(String userId) async {
    await Future<void>.delayed(const Duration(milliseconds: 120));
    if (!_users.any((user) => user.id == userId)) {
      throw const AdminRepositoryException('USER_NOT_FOUND');
    }
    final now = _clock();
    return [
      AdminHistoryEntry(
        action: 'Cuenta creada',
        details: 'El usuario se unió a Meloop.',
        createdAt: now.subtract(const Duration(days: 95)),
      ),
      AdminHistoryEntry(
        action: 'Nivel actualizado',
        details:
            'Alcanzó el nivel ${_users.firstWhere((user) => user.id == userId).level}.',
        createdAt: now.subtract(const Duration(days: 12)),
      ),
      AdminHistoryEntry(
        action: 'Inicio de sesión',
        details: 'Actividad reciente registrada.',
        createdAt: now.subtract(const Duration(hours: 2)),
      ),
    ];
  }

  @override
  Future<List<AdminReward>> fetchRewards({
    AdminRewardType? type,
    bool? enabled,
  }) async {
    await Future<void>.delayed(const Duration(milliseconds: 120));
    return List.unmodifiable(
      _rewards.where(
        (reward) =>
            (type == null || reward.type == type) &&
            (enabled == null || reward.enabled == enabled),
      ),
    );
  }

  @override
  Future<AdminReward> saveReward(AdminReward reward) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    final duplicate = _rewards.any(
      (existing) =>
          existing.id != reward.id &&
          existing.name.toLowerCase() == reward.name.trim().toLowerCase(),
    );
    if (duplicate) throw const AdminRepositoryException('DUPLICATE_NAME');
    if (reward.requiredLevel < 1 || reward.requiredLevel > 10) {
      throw const AdminRepositoryException('LEVEL_NOT_FOUND');
    }
    final index = _rewards.indexWhere((existing) => existing.id == reward.id);
    final saved = reward.copyWith(name: reward.name.trim());
    if (index < 0) {
      final created = AdminReward(
        id: 'reward-${_nextRewardId++}',
        name: saved.name,
        description: saved.description,
        type: saved.type,
        requiredLevel: saved.requiredLevel,
        enabled: saved.enabled,
        imageUrl: saved.imageUrl,
      );
      _rewards.add(created);
      return created;
    }
    _rewards[index] = saved;
    return saved;
  }

  @override
  Future<void> setRewardEnabled(String rewardId, bool enabled) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    final index = _rewards.indexWhere((reward) => reward.id == rewardId);
    if (index < 0) throw const AdminRepositoryException('REWARD_NOT_FOUND');
    _rewards[index] = _rewards[index].copyWith(enabled: enabled);
  }

  @override
  Future<String> createRewardImageUploadUrl(String fileName) async {
    await Future<void>.delayed(const Duration(milliseconds: 120));
    if (fileName.trim().isEmpty) {
      throw const AdminRepositoryException('INVALID_FILE_NAME');
    }
    return 'https://images.example.invalid/rewards/${Uri.encodeComponent(fileName)}';
  }
}
