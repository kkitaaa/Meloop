import 'package:flutter/material.dart';

import 'features/notifications/presentation/notification_feedback_overlay.dart';
import 'features/notifications/presentation/notification_scope.dart';
import 'features/notifications/state/notification_controller.dart';
import 'features/notifications/state/notification_repository.dart';
import 'screens/auth_screen.dart';

void main() {
  runApp(const MeloopApp());
}

class MeloopApp extends StatefulWidget {
  const MeloopApp({super.key, this.notificationRepository});

  final NotificationRepository? notificationRepository;

  @override
  State<MeloopApp> createState() => _MeloopAppState();
}

class _MeloopAppState extends State<MeloopApp> {
  late final NotificationController _notificationController;

  @override
  void initState() {
    super.initState();
    _notificationController = NotificationController(
      repository: widget.notificationRepository ?? MockNotificationRepository(),
      playSound: false,
    )..start();
  }

  @override
  void dispose() {
    _notificationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Meloop',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(primaryColor: const Color(0xFF11B49F)),
      home: const AuthScreen(),
      builder: (context, child) => NotificationScope(
        controller: _notificationController,
        child: NotificationFeedbackOverlay(
          controller: _notificationController,
          child: child ?? const SizedBox.shrink(),
        ),
      ),
    );
  }
}
