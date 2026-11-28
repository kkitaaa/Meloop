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
import '../../widgets/paper/stacked_paper_card.dart';
import 'artists_and_songs_screen.dart';
import 'sliding_tab_stack.dart';

/// Controlador de navegación global del shell de escritorio.
/// Mantiene la barra superior (`DesktopAppHeader`) fija e inamovible en todas
/// las secciones y submenús para que ninguna transición desplace o duplique la
/// navbar.
abstract class DesktopShellController {
  DesktopNavigationSection get activeSection;
  bool get searchOpen;
  TextEditingController get searchController;

  void goHome();
  void openFriends({String? initialQuery});
  void openArtists({String? initialQuery});
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
  final GlobalKey<SlidingTabStackState> _tabStackKey =
      GlobalKey<SlidingTabStackState>();
  final GlobalKey<FriendsScreenState> _friendsKey =
      GlobalKey<FriendsScreenState>();
  final GlobalKey<ArtistsAndSongsScreenState> _artistsKey =
      GlobalKey<ArtistsAndSongsScreenState>();
  final TextEditingController _searchController = TextEditingController();
  final ValueNotifier<int> _searchSubmitNonce = ValueNotifier<int>(0);
  final ValueNotifier<int> _primaryTabIndex = ValueNotifier<int>(0);

  late final ApiClient _apiClient = widget.apiClient ?? ApiClient();
  late final bool _ownsApiClient = widget.apiClient == null;
  late final _ShellNavigatorObserver _routeObserver = _ShellNavigatorObserver(
    onRouteStackChanged: _handleRouteStackChanged,
  );
  late final Widget _homeTab = FeedScreen(
    apiClient: _apiClient,
    userRole: widget.userRole,
  );
  String? _friendsInitialQuery;
  String? _artistsInitialQuery;

  DesktopNavigationSection _activeSection = DesktopNavigationSection.home;
  bool _searchOpen = false;
  bool _texturesPreloaded = false;

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
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_texturesPreloaded) {
      _texturesPreloaded = true;
      StackedPaperCard.preloadTextures(context);
    }
  }

  @override
  void dispose() {
    if (identical(DesktopNavigationShell.activeController, this)) {
      DesktopNavigationShell.activeController = null;
    }
    _primaryTabIndex.dispose();
    _searchSubmitNonce.dispose();
    _searchController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  DesktopNavigationSection _sectionForPrimaryIndex(int index) {
    return switch (index) {
      0 => DesktopNavigationSection.home,
      1 => DesktopNavigationSection.friends,
      2 => DesktopNavigationSection.artists,
      _ => DesktopNavigationSection.home,
    };
  }

  void _handleRouteStackChanged(Route<dynamic>? currentRoute) {
    if (!mounted || currentRoute is! PageRoute) return;
    final name = currentRoute.settings.name;
    final nextSection = currentRoute.isFirst
        ? _sectionForPrimaryIndex(_primaryTabIndex.value)
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

  void _selectPrimaryTab(int index, {VoidCallback? onReady}) {
    final targetSection = _sectionForPrimaryIndex(index);
    final nav = _navigatorKey.currentState;
    final wasOnSubRoute = nav != null && nav.canPop();

    // Actualizamos el índice ANTES de hacer popUntil para que el
    // NavigatorObserver (`didPop`) vea inmediatamente la sección destino y no
    // reinicie la píldora a "Inicio". Además, cuando venimos de Perfil/Búsqueda
    // saltamos sin deslizar para que no choquen dos animaciones simultáneas.
    if (wasOnSubRoute) {
      _tabStackKey.currentState?.jumpWithoutAnimation(index);
    }
    _primaryTabIndex.value = index;

    setState(() {
      _activeSection = targetSection;
      _searchOpen = false;
    });

    if (wasOnSubRoute) {
      nav.popUntil((route) => route.isFirst);
    }

    if (onReady != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        onReady();
      });
    }
  }

  @override
  void goHome() {
    _selectPrimaryTab(0);
  }

  @override
  void openFriends({String? initialQuery}) {
    final trimmedQuery = initialQuery?.trim();
    _friendsInitialQuery = trimmedQuery != null && trimmedQuery.isNotEmpty
        ? trimmedQuery
        : null;
    _selectPrimaryTab(
      1,
      onReady: _friendsInitialQuery != null
          ? () => _friendsKey.currentState?.applySearchQuery(
              _friendsInitialQuery!,
            )
          : null,
    );
  }

  @override
  void openArtists({String? initialQuery}) {
    final trimmedQuery = initialQuery?.trim();
    _artistsInitialQuery = trimmedQuery != null && trimmedQuery.isNotEmpty
        ? trimmedQuery
        : null;
    _selectPrimaryTab(
      2,
      onReady: _artistsInitialQuery != null
          ? () => _artistsKey.currentState?.applySearchQuery(
              _artistsInitialQuery!,
            )
          : null,
    );
  }

  PageRouteBuilder<void> _buildSubRoute({
    required String name,
    required WidgetBuilder builder,
  }) {
    return PageRouteBuilder<void>(
      settings: RouteSettings(name: name),
      pageBuilder: (context, animation, secondaryAnimation) => builder(context),
      transitionDuration: const Duration(milliseconds: 300),
      reverseTransitionDuration: const Duration(milliseconds: 260),
      transitionsBuilder: (context, animation, secondaryAnimation, child) {
        // Espera a que la vista inferior termine de desvanecerse (0.0 -> 0.44)
        // antes de revelar la nueva pantalla, y al salir se oculta por completo
        // (1.0 -> 0.50) antes de mostrar la pestaña destino. Así nunca se
        // superponen dos pantallas en ningún fotograma.
        final fadeIn = CurvedAnimation(
          parent: animation,
          curve: const Interval(0.46, 1.0, curve: Curves.easeOutCubic),
          reverseCurve: const Interval(0.50, 1.0, curve: Curves.easeInCubic),
        );
        final slide = Tween<Offset>(
          begin: const Offset(0, 0.018),
          end: Offset.zero,
        ).animate(fadeIn);
        return FadeTransition(
          opacity: fadeIn,
          child: SlideTransition(position: slide, child: child),
        );
      },
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
      _buildSubRoute(
        name: '/profile',
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

  void _submitSearch(String _) {
    _searchSubmitNonce.value++;
  }

  @override
  void openSearch() {
    if (_searchOpen) return;
    setState(() {
      _activeSection = DesktopNavigationSection.none;
      _searchOpen = true;
    });
    _navigatorKey.currentState?.push<void>(
      _buildSubRoute(
        name: '/search',
        builder: (_) => GlobalSearchScreen(
          apiClient: _apiClient,
          externalQueryController: _searchController,
          externalSubmitTrigger: _searchSubmitNonce,
        ),
      ),
    );
  }

  @override
  void closeSearch() {
    if (_searchOpen) {
      final nav = _navigatorKey.currentState;
      final restoredSection = _sectionForPrimaryIndex(_primaryTabIndex.value);
      setState(() {
        _searchOpen = false;
        _activeSection = restoredSection;
      });
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
      _buildSubRoute(
        name: '/moderation',
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
      _buildSubRoute(name: '/subpage', builder: builder),
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
        transitionDuration: const Duration(milliseconds: 220),
        reverseTransitionDuration: const Duration(milliseconds: 180),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(
              opacity: CurvedAnimation(
                parent: animation,
                curve: Curves.easeOutCubic,
              ),
              child: child,
            ),
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
                onSearchSubmitted: _submitSearch,
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
                          ValueListenableBuilder<int>(
                            valueListenable: _primaryTabIndex,
                            builder: (context, index, _) => SlidingTabStack(
                              key: _tabStackKey,
                              currentIndex: index,
                              children: [
                                _homeTab,
                                FriendsScreen(
                                  key: _friendsKey,
                                  initialQuery: _friendsInitialQuery,
                                ),
                                ArtistsAndSongsScreen(
                                  key: _artistsKey,
                                  initialQuery: _artistsInitialQuery,
                                ),
                              ],
                            ),
                          ),
                      transitionDuration: Duration.zero,
                      reverseTransitionDuration: Duration.zero,
                      transitionsBuilder:
                          (context, animation, secondaryAnimation, child) {
                            // Cuando se abre Búsqueda, Perfil o una subpantalla,
                            // primero desvanecemos y ocultamos la pestaña base
                            // para que nunca se vea detrás de la nueva pantalla.
                            final fadeOut = Tween<double>(begin: 1.0, end: 0.0)
                                .animate(
                                  CurvedAnimation(
                                    parent: secondaryAnimation,
                                    curve: const Interval(
                                      0.0,
                                      0.44,
                                      curve: Curves.easeOut,
                                    ),
                                    reverseCurve: const Interval(
                                      0.0,
                                      0.48,
                                      curve: Curves.easeIn,
                                    ),
                                  ),
                                );
                            return AnimatedBuilder(
                              animation: secondaryAnimation,
                              builder: (context, innerChild) {
                                final hidden = secondaryAnimation.value >= 0.45;
                                return Offstage(
                                  offstage: hidden,
                                  child: IgnorePointer(
                                    ignoring: secondaryAnimation.value > 0.05,
                                    child: FadeTransition(
                                      opacity: fadeOut,
                                      child: innerChild,
                                    ),
                                  ),
                                );
                              },
                              child: child,
                            );
                          },
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
