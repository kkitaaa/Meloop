import 'package:flutter/material.dart';

import '../../features/moderation/data/moderation_models.dart';
import '../../features/moderation/presentation/moderation_queue_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/search/presentation/global_search_screen.dart';
import '../../screens/feed_screen.dart';
import '../../screens/friends_screen.dart';
import '../../services/api_client.dart';
import '../../theme/meloop_typography.dart';
import '../../widgets/desktop_app_header.dart';
import 'artists_and_songs_screen.dart';

/// Controlador de navegación global del shell de escritorio.
/// Mantiene la barra superior (`DesktopAppHeader`) fija e inamovible en todas
/// las secciones y submenús para que ninguna transición desplace o duplique la
/// navbar.
abstract class DesktopShellController {
  DesktopNavigationSection get activeSection;
  bool get searchOpen;
  TextEditingController get searchController;

  void goHome();
  void openFriends();
  void openArtists();
  void openProfile({
    String? username,
    String? userId,
    bool openInventory = false,
  });
  void openSearch();
  void closeSearch();
  void openModerationQueue();
  void pushPage(WidgetBuilder builder);
  void navigateToSectionWidget(WidgetBuilder builder);
}

class DesktopShellScope extends InheritedWidget {
  const DesktopShellScope({
    super.key,
    required this.controller,
    required super.child,
  });

  final DesktopShellController controller;

  static DesktopShellController? maybeOf(BuildContext context) => context
      .dependOnInheritedWidgetOfExactType<DesktopShellScope>()
      ?.controller;

  @override
  bool updateShouldNotify(DesktopShellScope oldWidget) => true;
}

class DesktopNavigationShell extends StatefulWidget {
  const DesktopNavigationShell({
    super.key,
    this.userRole = ModerationRole.user,
    this.apiClient,
  });

  final ModerationRole userRole;
  final ApiClient? apiClient;

  /// Referencia al shell activo para que los overlays globales (como
  /// LevelUpFeedbackOverlay -> Equipar) naveguen dentro del shell sin cubrir la
  /// navbar global.
  static DesktopShellController? activeController;

  @override
  State<DesktopNavigationShell> createState() => _DesktopNavigationShellState();
}

class _DesktopNavigationShellState extends State<DesktopNavigationShell>
    implements DesktopShellController {
  final GlobalKey<NavigatorState> _navigatorKey = GlobalKey<NavigatorState>();
  final TextEditingController _searchController = TextEditingController();
  late final ApiClient _apiClient = widget.apiClient ?? ApiClient();
  late final bool _ownsApiClient = widget.apiClient == null;
  late final _ShellNavigatorObserver _routeObserver = _ShellNavigatorObserver(
    onRouteStackChanged: _handleRouteStackChanged,
  );

  DesktopNavigationSection _activeSection = DesktopNavigationSection.home;
  bool _searchOpen = false;

  @override
  DesktopNavigationSection get activeSection => _activeSection;

  @override
  bool get searchOpen => _searchOpen;

  @override
  TextEditingController get searchController => _searchController;

  @override
  void initState() {
    super.initState();
    DesktopNavigationShell.activeController = this;
  }

  @override
  void dispose() {
    if (identical(DesktopNavigationShell.activeController, this)) {
      DesktopNavigationShell.activeController = null;
    }
    _searchController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  void _handleRouteStackChanged(Route<dynamic>? currentRoute) {
    if (!mounted || currentRoute is! PageRoute) return;
    final name = currentRoute.settings.name;
    final nextSection = currentRoute.isFirst
        ? DesktopNavigationSection.home
        : switch (name) {
            '/home' || '/' => DesktopNavigationSection.home,
            '/friends' => DesktopNavigationSection.friends,
            '/artists' => DesktopNavigationSection.artists,
            _ => DesktopNavigationSection.none,
          };
    final nextSearchOpen = name == '/search';
    if (nextSection != _activeSection || nextSearchOpen != _searchOpen) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        setState(() {
          _activeSection = nextSection;
          _searchOpen = nextSearchOpen;
        });
      });
    }
  }

  @override
  void goHome() {
    setState(() {
      _activeSection = DesktopNavigationSection.home;
      _searchOpen = false;
    });
    _navigatorKey.currentState?.popUntil((route) => route.isFirst);
  }

  @override
  void openFriends() {
    if (_activeSection == DesktopNavigationSection.friends && !_searchOpen) {
      return;
    }
    setState(() {
      _activeSection = DesktopNavigationSection.friends;
      _searchOpen = false;
    });
    _replaceSectionRoute(
      name: '/friends',
      builder: (_) => const FriendsScreen(),
    );
  }

  @override
  void openArtists() {
    if (_activeSection == DesktopNavigationSection.artists && !_searchOpen) {
      return;
    }
    setState(() {
      _activeSection = DesktopNavigationSection.artists;
      _searchOpen = false;
    });
    _replaceSectionRoute(
      name: '/artists',
      builder: (_) => const ArtistsAndSongsScreen(),
    );
  }

  @override
  void openProfile({
    String? username,
    String? userId,
    bool openInventory = false,
  }) {
    setState(() {
      _activeSection = DesktopNavigationSection.none;
      _searchOpen = false;
    });
    _navigatorKey.currentState?.push<void>(
      MaterialPageRoute<void>(
        settings: const RouteSettings(name: '/profile'),
        builder: (_) => ProfileScreen(
          username: username,
          userId: userId,
          apiClient: _apiClient,
          openInventory: openInventory,
          insideDesktopShell: true,
        ),
      ),
    );
  }

  @override
  void openSearch() {
    if (_searchOpen) return;
    setState(() {
      _activeSection = DesktopNavigationSection.none;
      _searchOpen = true;
    });
    _navigatorKey.currentState?.push<void>(
      PageRouteBuilder<void>(
        settings: const RouteSettings(name: '/search'),
        pageBuilder: (context, animation, secondaryAnimation) =>
            GlobalSearchScreen(
              apiClient: _apiClient,
              externalQueryController: _searchController,
            ),
        transitionDuration: const Duration(milliseconds: 180),
        reverseTransitionDuration: const Duration(milliseconds: 140),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
    );
  }

  @override
  void closeSearch() {
    if (_searchOpen) {
      setState(() => _searchOpen = false);
      final nav = _navigatorKey.currentState;
      if (nav != null && nav.canPop()) {
        nav.pop();
      } else {
        goHome();
      }
    }
  }

  @override
  void openModerationQueue() {
    setState(() {
      _activeSection = DesktopNavigationSection.none;
      _searchOpen = false;
    });
    _navigatorKey.currentState?.push<void>(
      MaterialPageRoute<void>(
        settings: const RouteSettings(name: '/moderation'),
        builder: (_) => ModerationQueueScreen(role: widget.userRole),
      ),
    );
  }

  @override
  void pushPage(WidgetBuilder builder) {
    setState(() {
      _activeSection = DesktopNavigationSection.none;
      _searchOpen = false;
    });
    _navigatorKey.currentState?.push<void>(
      MaterialPageRoute<void>(
        settings: const RouteSettings(name: '/subpage'),
        builder: builder,
      ),
    );
  }

  @override
  void navigateToSectionWidget(WidgetBuilder builder) {
    setState(() {
      _searchOpen = false;
    });
    _replaceSectionRoute(name: '/section', builder: builder);
  }

  void _replaceSectionRoute({
    required String name,
    required WidgetBuilder builder,
  }) {
    _navigatorKey.currentState?.pushAndRemoveUntil<void>(
      PageRouteBuilder<void>(
        settings: RouteSettings(name: name),
        pageBuilder: (context, animation, secondaryAnimation) =>
            builder(context),
        transitionDuration: const Duration(milliseconds: 160),
        reverseTransitionDuration: Duration.zero,
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
      (route) => route.isFirst,
    );
  }

  @override
  Widget build(BuildContext context) {
    final canModerate = widget.userRole != ModerationRole.user;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        final nav = _navigatorKey.currentState;
        if (nav != null && nav.canPop()) {
          nav.pop();
        }
      },
      child: DesktopShellScope(
        controller: this,
        child: Scaffold(
          backgroundColor: Colors.transparent,
          body: Column(
            children: [
              // Navbar global fija — nunca se mueve ni se duplica
              DesktopAppHeader(
                activeSection: _activeSection,
                onHomeTap: goHome,
                onFriendsTap: openFriends,
                onArtistsTap: openArtists,
                onProfileTap: () => openProfile(),
                onModerationTap: canModerate ? openModerationQueue : null,
                onSearchTap: openSearch,
                onPushRoute: pushPage,
                apiClient: _apiClient,
                searchOpen: _searchOpen,
                searchController: _searchController,
                onSearchBack: closeSearch,
              ),
              Expanded(
                child: ClipRect(
                  child: Navigator(
                    key: _navigatorKey,
                    observers: [_routeObserver],
                    onGenerateRoute: (settings) => PageRouteBuilder<void>(
                      settings: const RouteSettings(name: '/home'),
                      pageBuilder: (context, animation, secondaryAnimation) =>
                          FeedScreen(
                            apiClient: _apiClient,
                            userRole: widget.userRole,
                          ),
                      transitionDuration: Duration.zero,
                      reverseTransitionDuration: Duration.zero,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ShellNavigatorObserver extends NavigatorObserver {
  _ShellNavigatorObserver({required this.onRouteStackChanged});

  final ValueChanged<Route<dynamic>?> onRouteStackChanged;

  @override
  void didPush(Route<dynamic> route, Route<dynamic>? previousRoute) {
    super.didPush(route, previousRoute);
    onRouteStackChanged(route);
  }

  @override
  void didPop(Route<dynamic> route, Route<dynamic>? previousRoute) {
    super.didPop(route, previousRoute);
    onRouteStackChanged(previousRoute);
  }

  @override
  void didReplace({Route<dynamic>? newRoute, Route<dynamic>? oldRoute}) {
    super.didReplace(newRoute: newRoute, oldRoute: oldRoute);
    onRouteStackChanged(newRoute);
  }
}

/// Cabecera interna compacta (sin duplicar la navbar) para subpantallas
/// abiertas dentro del `DesktopNavigationShell`.
class DesktopSubPageBar extends StatelessWidget {
  const DesktopSubPageBar({super.key, required this.title, this.trailing});

  final String title;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: const BoxDecoration(
        color: Color(0xFFF3F8F7),
        border: Border(bottom: BorderSide(color: Color(0xFFDCE7E5))),
      ),
      child: Row(
        children: [
          TextButton.icon(
            onPressed: () => Navigator.maybePop(context),
            style: TextButton.styleFrom(
              foregroundColor: const Color(0xFF16877F),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            ),
            icon: const Icon(Icons.arrow_back_rounded, size: 20),
            label: Text(
              'Volver',
              style: MeloopTypography.lexend(
                fontSize: 14,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              title,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: MeloopTypography.lexend(
                fontSize: 17,
                fontWeight: FontWeight.w700,
                color: const Color(0xFF162525),
              ),
            ),
          ),
          ?trailing,
        ],
      ),
    );
  }
}
