import 'package:flutter/material.dart';

import '../../services/auth_session.dart';

enum AppPermission { moderation, administration }

class RoleGuard extends StatelessWidget {
  const RoleGuard({
    super.key,
    required this.permission,
    required this.child,
    this.session,
  });

  final AppPermission permission;
  final Widget child;
  final AuthSession? session;

  @override
  Widget build(BuildContext context) {
    final authSession = session ?? AuthSession.shared;
    return AnimatedBuilder(
      animation: authSession,
      builder: (context, _) {
        final permissions = authSession.permissions;
        final allowed = switch (permission) {
          AppPermission.moderation => permissions.canModerateReports,
          AppPermission.administration => permissions.canAccessAdministration,
        };
        if (allowed) return child;

        final title = permission == AppPermission.administration
            ? 'Administración'
            : 'Moderación';
        final message = permission == AppPermission.administration
            ? 'Esta sección está disponible solo para administradores.'
            : 'Esta sección está disponible solo para moderadores.';
        return Scaffold(
          backgroundColor: Colors.transparent,
          appBar: AppBar(title: Text(title)),
          body: Center(child: Text(message)),
        );
      },
    );
  }
}
