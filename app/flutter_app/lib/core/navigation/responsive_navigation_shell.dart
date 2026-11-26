import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../features/moderation/data/moderation_models.dart';
import '../../features/moderation/presentation/moderation_queue_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/search/presentation/global_search_route.dart';
import '../../screens/feed_screen.dart';
import '../../screens/friends_screen.dart';
import '../../services/api_client.dart';
import '../../theme/meloop_theme_controller.dart';
import '../../widgets/mobile_app_header.dart';
import '../../widgets/paper/stacked_paper_card.dart';
import 'artists_and_songs_screen.dart';
import 'sliding_tab_stack.dart';

/// Controlador compartido del shell móvil/Android para mantener una sola barra
/// superior fija mientras las pestañas deslizan suavemente debajo, y permitir
/// que pantallas como Búsqueda global salten directo a Amigos o Artistas.
abstract class MobileShellController {
  int get selectedIndex;

  void goHome();
  void openFriends({String? initialQuery});
  void openArtists({String? initialQuery});
  void openProfile();
}

class MobileShellScope extends InheritedWidget {
  const MobileShellScope({
    super.key,
    required this.controller,
    required super.child,
  });

  final MobileShellController controller;

  static MobileShellController? maybeOf(BuildContext context) =>
      context.dependOnInheritedWidgetOfExactType<MobileShellScope>()?.controller;

  @override
  bool updateShouldNotify(MobileShellScope oldWidget) => true;
}

class ResponsiveNavigationShell extends StatefulWidget {
  const ResponsiveNavigationShell({
    super.key,
    this.userRole = ModerationRole.user,
    this.apiClient,
  });

  final ModerationRole userRole;
  final ApiClient? apiClient;

  @override
  State<ResponsiveNavigationShell> createState() =>
      _ResponsiveNavigationShellState();
}

class _ResponsiveNavigationShellState extends State<ResponsiveNavigationShell>
    implements MobileShellController {
  int _selectedIndex = 0;
  bool _texturesPreloaded = false;
  late final ApiClient _apiClient = widget.apiClient ?? ApiClient();
  late final bool _ownsApiClient = widget.apiClient == null;
  final GlobalKey<FeedScreenState> _feedKey = GlobalKey<FeedScreenState>();
  final GlobalKey<FriendsScreenState> _friendsKey =
      GlobalKey<FriendsScreenState>();
  final GlobalKey<ArtistsAndSongsScreenState> _artistsKey =
      GlobalKey<ArtistsAndSongsScreenState>();

  late final List<Widget> _screens = [
    FeedScreen(key: _feedKey, apiClient: _apiClient, userRole: widget.userRole),
    FriendsScreen(key: _friendsKey),
    ArtistsAndSongsScreen(key: _artistsKey),
    ProfileScreen(apiClient: _apiClient),
  ];

  @override
  int get selectedIndex => _selectedIndex;

  @override
  void goHome() => setState(() => _selectedIndex = 0);

  @override
  void openFriends({String? initialQuery}) {
    setState(() => _selectedIndex = 1);
    final trimmed = initialQuery?.trim();
    if (trimmed != null && trimmed.isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        _friendsKey.currentState?.applySearchQuery(trimmed);
      });
    }
  }

  @override
  void openArtists({String? initialQuery}) {
    setState(() => _selectedIndex = 2);
    final trimmed = initialQuery?.trim();
    if (trimmed != null && trimmed.isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        _artistsKey.currentState?.applySearchQuery(trimmed);
      });
    }
  }

  @override
  void openProfile() => setState(() => _selectedIndex = 3);

  Future<void> _createBlog() async {
    await _feedKey.currentState?.createPost();
  }

  void _openModerationQueue() {
    Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (_) => ModerationQueueScreen(role: widget.userRole),
      ),
    );
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
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.viewPaddingOf(context).bottom;
    final canModerate = widget.userRole != ModerationRole.user;
    final themeCtrl = MeloopThemeController.instance;

    return AnimatedBuilder(
      animation: themeCtrl,
      builder: (context, _) {
        final accent = themeCtrl.primaryColor;
        final highlight = themeCtrl.colorTheme.headerTop;
        return MobileShellScope(
          controller: this,
          child: AnnotatedRegion<SystemUiOverlayStyle>(
            value: SystemUiOverlayStyle(
              statusBarColor: themeCtrl.colorTheme.bgStart,
              statusBarIconBrightness: Brightness.light,
            ),
            child: Scaffold(
              backgroundColor: Colors.transparent,
              extendBody: true,
              body: SafeArea(
                bottom: false,
                child: Column(
                  children: [
                    Builder(
                      builder: (headerContext) => MobileAppHeader(
                        onSearchTap: () => openGlobalSearch(headerContext),
                        onModerationTap: canModerate
                            ? _openModerationQueue
                            : null,
                        apiClient: _apiClient,
                      ),
                    ),
                    Expanded(
                      child: SlidingTabStack(
                        currentIndex: _selectedIndex,
                        children: _screens,
                      ),
                    ),
                  ],
                ),
              ),
              bottomNavigationBar: Container(
                height: bottomInset + 132,
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [Color(0x00000000), Color(0x88000000)],
                  ),
                ),
                child: Stack(
                  clipBehavior: Clip.none,
                  children: [
                    Positioned(
                      left: 14,
                      right: 14,
                      bottom: bottomInset + 10,
                      height: 52,
                      child: _buildNavigationPill(
                        accent: accent,
                        highlight: highlight,
                      ),
                    ),
                    Positioned(
                      right: 24,
                      bottom: bottomInset + 72,
                      child: Material(
                        color: highlight,
                        shape: const CircleBorder(),
                        elevation: 8,
                        shadowColor: Colors.black.withValues(alpha: 0.3),
                        child: IconButton(
                          tooltip: 'Crear blog',
                          onPressed: _createBlog,
                          icon: const Icon(
                            Icons.edit_note,
                            color: Colors.white,
                          ),
                          iconSize: 26,
                          constraints: const BoxConstraints.tightFor(
                            width: 56,
                            height: 56,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildNavigationPill({
    required Color accent,
    required Color highlight,
  }) {
    const items = <({IconData icon, String label})>[
      (icon: Icons.home_filled, label: 'Inicio'),
      (icon: Icons.people_outline, label: 'Amigos'),
      (icon: Icons.graphic_eq, label: 'Artistas y canciones'),
      (icon: Icons.person_outline, label: 'Perfil'),
    ];

    return Container(
      padding: const EdgeInsets.all(5),
      decoration: BoxDecoration(
        color: accent,
        borderRadius: BorderRadius.circular(32),
        border: Border.all(color: Colors.white.withValues(alpha: 0.16)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.3),
            blurRadius: 18,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Row(
        children: List.generate(items.length, (index) {
          final item = items[index];
          final isSelected = index == _selectedIndex;

          return Expanded(
            child: Semantics(
              button: true,
              selected: isSelected,
              label: item.label,
              child: InkWell(
                borderRadius: BorderRadius.circular(26),
                onTap: () => setState(() => _selectedIndex = index),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 220),
                  curve: Curves.easeOutCubic,
                  height: 42,
                  decoration: BoxDecoration(
                    color: isSelected ? highlight : Colors.transparent,
                    borderRadius: BorderRadius.circular(26),
                  ),
                  child: Center(
                    child: index == 3
                        ? CircleAvatar(
                            radius: 15,
                            backgroundColor: Colors.white,
                            child: Icon(Icons.person, color: accent, size: 19),
                          )
                        : Icon(item.icon, color: Colors.white, size: 23),
                  ),
                ),
              ),
            ),
          );
        }),
      ),
    );
  }
}
