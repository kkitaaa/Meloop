enum AdminUserRole {
  user('user', 'Usuario'),
  moderator('moderator', 'Moderador'),
  admin('admin', 'Administrador');

  const AdminUserRole(this.apiValue, this.label);

  final String apiValue;
  final String label;
}

enum AdminUserStatus {
  active('active', 'Activa'),
  suspended('suspended', 'Suspendida'),
  inactive('inactive', 'Inactiva'),
  pendingDeletion('pending_deletion', 'Pendiente de eliminación');

  const AdminUserStatus(this.apiValue, this.label);

  final String apiValue;
  final String label;
}

class AdminUser {
  const AdminUser({
    required this.id,
    required this.username,
    required this.role,
    required this.status,
    required this.level,
    required this.lastActivity,
    this.avatarUrl,
  });

  final String id;
  final String username;
  final AdminUserRole role;
  final AdminUserStatus status;
  final int level;
  final DateTime lastActivity;
  final String? avatarUrl;

  AdminUser copyWith({AdminUserRole? role, AdminUserStatus? status}) =>
      AdminUser(
        id: id,
        username: username,
        role: role ?? this.role,
        status: status ?? this.status,
        level: level,
        lastActivity: lastActivity,
        avatarUrl: avatarUrl,
      );
}

class AdminUserPage {
  const AdminUserPage({required this.users, this.nextCursor});

  final List<AdminUser> users;
  final String? nextCursor;
}

class AdminHistoryEntry {
  const AdminHistoryEntry({
    required this.action,
    required this.details,
    required this.createdAt,
  });

  final String action;
  final String details;
  final DateTime createdAt;
}

enum AdminRewardType {
  frame('frame', 'Marco'),
  background('background', 'Fondo'),
  banner('banner', 'Banner'),
  badge('badge', 'Insignia'),
  theme('theme', 'Tema'),
  color('color', 'Color'),
  effect('effect', 'Efecto');

  const AdminRewardType(this.apiValue, this.label);

  final String apiValue;
  final String label;
}

class AdminReward {
  const AdminReward({
    required this.id,
    required this.name,
    required this.description,
    required this.type,
    required this.requiredLevel,
    required this.enabled,
    this.imageUrl,
  });

  final String id;
  final String name;
  final String description;
  final AdminRewardType type;
  final int requiredLevel;
  final bool enabled;
  final String? imageUrl;

  AdminReward copyWith({
    String? name,
    String? description,
    AdminRewardType? type,
    int? requiredLevel,
    bool? enabled,
    String? imageUrl,
  }) => AdminReward(
    id: id,
    name: name ?? this.name,
    description: description ?? this.description,
    type: type ?? this.type,
    requiredLevel: requiredLevel ?? this.requiredLevel,
    enabled: enabled ?? this.enabled,
    imageUrl: imageUrl ?? this.imageUrl,
  );
}
