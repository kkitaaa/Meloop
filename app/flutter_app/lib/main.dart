import 'package:flutter/material.dart';

import 'core/navigation/desktop_navigation_shell.dart';
import 'features/notifications/presentation/notification_feedback_overlay.dart';
import 'features/notifications/presentation/notification_scope.dart';
import 'features/notifications/state/notification_controller.dart';
import 'features/notifications/state/notification_repository.dart';
import 'features/gamification/presentation/level_up_feedback_overlay.dart';
import 'features/profile/presentation/profile_screen.dart';
import 'services/auth_session.dart';
import 'screens/auth_screen.dart';
import 'theme/meloop_typography.dart';
import 'widgets/app_background.dart';
import 'widgets/desktop_viewport_scaler.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AuthSession.shared.restore();
  runApp(const MeloopApp());
}

class MeloopApp extends StatefulWidget {
  const MeloopApp({super.key, this.notificationRepository});

  final NotificationRepository? notificationRepository;

  @override
  State<MeloopApp> createState() => _MeloopAppState();
}

class _MeloopAppState extends State<MeloopApp> {
  final GlobalKey<NavigatorState> _navigatorKey = GlobalKey<NavigatorState>();
  late final NotificationController _notificationController;

  @override
  void initState() {
    super.initState();
    _notificationController = NotificationController(
      repository: widget.notificationRepository ?? MockNotificationRepository(),
      playSound: false,
    );
  }

  @override
  void dispose() {
    _notificationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      navigatorKey: _navigatorKey,
      title: 'Meloop',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        primaryColor: const Color(0xFF11B49F),
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF00AF92),
          primary: const Color(0xFF00AF92),
          surface: Colors.white,
          surfaceContainerLowest: Colors.white,
          surfaceContainerLow: const Color(0xFFFAFAFA),
          surfaceContainer: Colors.white,
          surfaceContainerHigh: Colors.white,
          surfaceContainerHighest: const Color(0xFFF2F6F5),
          surfaceTint: Colors.transparent,
        ),
        scaffoldBackgroundColor: Colors.transparent,
        popupMenuTheme: const PopupMenuThemeData(
          color: Colors.white,
          surfaceTintColor: Colors.transparent,
        ),
        dialogTheme: const DialogThemeData(
          backgroundColor: Colors.white,
          surfaceTintColor: Colors.transparent,
        ),
        cardTheme: const CardThemeData(
          color: Colors.white,
          surfaceTintColor: Colors.transparent,
        ),
        textTheme: MeloopTypography.lexendTextTheme(),
        pageTransitionsTheme: const PageTransitionsTheme(
          builders: {
            TargetPlatform.android: _MeloopPageTransitionsBuilder(),
            TargetPlatform.iOS: _MeloopPageTransitionsBuilder(),
            TargetPlatform.linux: _MeloopPageTransitionsBuilder(),
            TargetPlatform.macOS: _MeloopPageTransitionsBuilder(),
            TargetPlatform.windows: _MeloopPageTransitionsBuilder(),
          },
        ),
      ),
      home: const AuthScreen(),
      builder: (context, child) => DesktopViewportScaler(
        child: NotificationScope(
          controller: _notificationController,
          child: Stack(
            fit: StackFit.expand,
            children: [
              const AppBackground(),
              LevelUpFeedbackOverlay(
                onEquip: _openInventory,
                child: NotificationFeedbackOverlay(
                  controller: _notificationController,
                  child: child ?? const SizedBox.shrink(),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _openInventory() {
    final desktopShell = DesktopNavigationShell.activeController;
    if (desktopShell != null) {
      desktopShell.openProfile(openInventory: true);
      return;
    }
    final navigator = _navigatorKey.currentState;
    if (navigator == null) {
      throw StateError('No hay un navegador disponible para abrir el perfil.');
    }
    navigator.push<void>(
      MaterialPageRoute<void>(
        builder: (_) => const ProfileScreen(openInventory: true),
      ),
    );
  }
}

class _MeloopPageTransitionsBuilder extends PageTransitionsBuilder {
  const _MeloopPageTransitionsBuilder();

  @override
  Widget buildTransitions<T>(
    PageRoute<T> route,
    BuildContext context,
    Animation<double> animation,
    Animation<double> secondaryAnimation,
    Widget child,
  ) {
    final curvedAnimation = CurvedAnimation(
      parent: animation,
      curve: Curves.easeOutCubic,
      reverseCurve: Curves.easeInCubic,
    );

    return FadeTransition(
      opacity: curvedAnimation,
      child: SlideTransition(
        position: Tween<Offset>(
          begin: const Offset(0.035, 0),
          end: Offset.zero,
        ).animate(curvedAnimation),
        child: child,
      ),
    );
  }
}
