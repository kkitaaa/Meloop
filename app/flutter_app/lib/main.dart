import 'package:flutter/material.dart';

import 'services/api_client.dart';
import 'services/auth_session.dart';
import 'features/notifications/presentation/notification_feedback_overlay.dart';
import 'features/notifications/presentation/notification_scope.dart';
import 'features/notifications/state/notification_controller.dart';
import 'features/notifications/state/notification_repository.dart';
import 'features/gamification/presentation/level_up_feedback_overlay.dart';
import 'features/profile/presentation/profile_screen.dart';
import 'screens/auth_screen.dart';
import 'widgets/app_background.dart';

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

class _MeloopAppState extends State<MeloopApp> with WidgetsBindingObserver {
  final GlobalKey<NavigatorState> _navigatorKey = GlobalKey<NavigatorState>();
  final _routeObserver = _AppRouteObserver();
  late final NotificationController _notificationController;
  final AuthSession _session = AuthSession.shared;
  late final ApiClient _apiClient = ApiClient(session: _session);
  late bool _hadActiveSession;

  @override
  void initState() {
    super.initState();
    _hadActiveSession = _session.isAuthenticated;
    _session.addListener(_onSessionChanged);
    WidgetsBinding.instance.addObserver(this);
    _notificationController = NotificationController(
      repository: widget.notificationRepository ?? MockNotificationRepository(),
      playSound: false,
    );
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _session.removeListener(_onSessionChanged);
    _apiClient.close();
    _notificationController.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && _session.hasToken) {
      _syncSession();
    }
  }

  Future<void> _syncSession() async {
    try {
      await _apiClient.validateSession();
    } on ApiException catch (error) {
      if (error.statusCode != 401) {
        _showSessionMessage(
          'No se pudieron actualizar los permisos: ${error.message}',
        );
      }
    } on Exception catch (error) {
      _showSessionMessage('No se pudo validar la sesión: $error');
    }
  }

  void _onSessionChanged() {
    if (_session.isAuthenticated) {
      _hadActiveSession = true;
      return;
    }
    if (!_hadActiveSession || _routeObserver.currentRouteName == '/login') {
      _hadActiveSession = false;
      return;
    }

    _hadActiveSession = false;
    _navigatorKey.currentState?.pushAndRemoveUntil<void>(
      MaterialPageRoute<void>(
        settings: const RouteSettings(name: '/login'),
        builder: (_) => const AuthScreen(),
      ),
      (route) => false,
    );
  }

  void _showSessionMessage(String message) {
    final context = _navigatorKey.currentContext;
    if (context == null) return;
    ScaffoldMessenger.maybeOf(context)
        ?.showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      navigatorKey: _navigatorKey,
      navigatorObservers: [_routeObserver],
      title: 'Meloop',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        primaryColor: const Color(0xFF11B49F),
        scaffoldBackgroundColor: Colors.transparent,
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
      initialRoute: '/login',
      routes: {'/login': (_) => const AuthScreen()},
      builder: (context, child) => NotificationScope(
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
    );
  }

  void _openInventory() {
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

class _AppRouteObserver extends NavigatorObserver {
  String? currentRouteName;

  @override
  void didPush(Route<dynamic> route, Route<dynamic>? previousRoute) {
    currentRouteName = route.settings.name;
  }

  @override
  void didReplace({Route<dynamic>? newRoute, Route<dynamic>? oldRoute}) {
    currentRouteName = newRoute?.settings.name;
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
