enum UserRole {
  user,
  moderator,
  admin;

  static UserRole fromBackend(Object? value) {
    if (value is! String) {
      throw const FormatException('El servidor no devolvió el rol del usuario.');
    }

    return switch (value.trim().toLowerCase()) {
      'user' || 'usuario' => UserRole.user,
      'moderator' || 'moderador' => UserRole.moderator,
      'admin' || 'administrator' || 'administrador' => UserRole.admin,
      _ => throw const FormatException(
        'El servidor devolvió un rol de usuario desconocido.',
      ),
    };
  }
}

class UserPermissions {
  const UserPermissions(this.role);

  final UserRole role;

  bool get canModerateReports =>
      role == UserRole.moderator || role == UserRole.admin;

  bool get canAccessAdministration => role == UserRole.admin;
}
