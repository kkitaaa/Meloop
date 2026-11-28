import 'package:flutter/material.dart';

import '../features/notifications/presentation/notification_scope.dart';
import '../screens/auth_screen.dart';
import '../services/api_client.dart';

class SessionLogout {
  static Future<void> perform(
    BuildContext context, {
    ApiClient? apiClient,
  }) async {
    final messenger = ScaffoldMessenger.maybeOf(context);
    final navigator = Navigator.of(context, rootNavigator: true);
    final client = apiClient ?? ApiClient();
    Object? logoutError;

    NotificationScope.maybeOf(context)?.stop();
    try {
      await client.logout();
    } on Exception catch (error) {
      logoutError = error;
    } finally {
      if (apiClient == null) client.close();
    }

    if (!navigator.mounted) return;
    navigator.pushAndRemoveUntil<void>(
      MaterialPageRoute<void>(builder: (_) => const AuthScreen()),
      (route) => false,
    );

    if (logoutError != null) {
      messenger?.showSnackBar(
        SnackBar(
          content: Text(
            'Se cerró la sesión localmente, pero no se pudo notificar al '
            'servidor: $logoutError',
          ),
        ),
      );
    }
  }
}
