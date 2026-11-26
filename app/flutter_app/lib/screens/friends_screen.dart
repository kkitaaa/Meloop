import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../core/navigation/artists_and_songs_screen.dart';
import '../core/navigation/desktop_navigation_shell.dart';
import '../core/navigation/responsive_navigation_shell.dart';
import '../features/gamification/presentation/gamification_demo_controller.dart';
import '../features/profile/presentation/profile_screen.dart';
import '../features/search/presentation/global_search_route.dart';
import '../theme/meloop_theme_controller.dart';
import '../theme/meloop_typography.dart';
import '../widgets/desktop_app_header.dart';
import '../widgets/meloop_pill_tabs.dart';
import '../widgets/meloop_toast_pill.dart';
import '../widgets/mobile_app_header.dart';
import '../widgets/paper/stacked_paper_card.dart';

enum _FriendsTab { friends, requests, suggestions }

class FriendsScreen extends StatefulWidget {
  const FriendsScreen({super.key, this.initialQuery});

  final String? initialQuery;

  @override
  State<FriendsScreen> createState() => FriendsScreenState();
}

class FriendsScreenState extends State<FriendsScreen> {
  Color get _tealAccent => MeloopThemeController.instance.primaryColor;
  Color get _pillTrackColor => MeloopThemeController.instance.pillTrackColor;

  final TextEditingController _searchController = TextEditingController();
  _FriendsTab _activeTab = _FriendsTab.friends;
  bool _onlyOnline = false;
  String _searchQuery = '';

  // TODO(backend): Reemplazar estos datos locales con las respuestas reales del
  // social-service vía API Gateway (amigos, solicitudes entrantes y perfiles
  // sugeridos por afinidad musical en Spotify).
  final List<_FriendItem> _friends = [
    const _FriendItem(
      id: '1',
      username: 'Vicente',
      status: 'Escuchando a Skrillex',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 14,
      blogsCount: 6,
      isOnline: true,
    ),
    const _FriendItem(
      id: '2',
      username: 'Andrea',
      status: 'Conectada hace 2 min',
      joinedLabel: 'Se unió hace 3 días.',
      friendsCount: 22,
      blogsCount: 11,
      isOnline: true,
    ),
    const _FriendItem(
      id: '3',
      username: 'Felipe',
      status: 'Escuchando electrónica',
      joinedLabel: 'Se unió hace 1 semana.',
      friendsCount: 9,
      blogsCount: 3,
      isOnline: true,
    ),
    const _FriendItem(
      id: '4',
      username: 'Valentina',
      status: 'Ausente',
      joinedLabel: 'Se unió hace 2 semanas.',
      friendsCount: 18,
      blogsCount: 8,
      isOnline: false,
    ),
  ];

  final List<_FriendRequestItem> _requests = [
    const _FriendRequestItem(
      id: '101',
      username: 'NuevoUsuario_01',
      description: 'Te ha enviado una solicitud de amistad',
      mutualLabel: '3 amigos en común · Escucha Indie Pop',
    ),
    const _FriendRequestItem(
      id: '102',
      username: 'DJ_Productor',
      description: 'Te ha enviado una solicitud de amistad',
      mutualLabel: '5 amigos en común · Escucha Electrónica',
    ),
  ];

  final List<_SuggestedProfileItem> _suggestions = [
    const _SuggestedProfileItem(
      id: '201',
      username: 'Martín',
      description: 'Ambos escuchan a DJ Gouz y Skrillex',
      joinedLabel: '12 amigos · 5 blogs',
      compatibility: 95,
    ),
    const _SuggestedProfileItem(
      id: '202',
      username: 'Camila',
      description: 'Tienen 12 artistas en común',
      joinedLabel: '19 amigos · 8 blogs',
      compatibility: 82,
    ),
    const _SuggestedProfileItem(
      id: '203',
      username: 'Pedro',
      description: 'Ambos escuchan electrónica y breakbeat',
      joinedLabel: '7 amigos · 2 blogs',
      compatibility: 74,
    ),
    const _SuggestedProfileItem(
      id: '204',
      username: 'Sofía',
      description: 'Le gusta Kanye West y Daft Punk',
      joinedLabel: '15 amigos · 9 blogs',
      compatibility: 68,
    ),
  ];

  @override
  void initState() {
    super.initState();
    final initial = widget.initialQuery?.trim();
    if (initial != null && initial.isNotEmpty) {
      _searchController.text = initial;
      _searchQuery = initial;
    }
  }

  void applySearchQuery(String query) {
    final trimmed = query.trim();
    setState(() {
      _activeTab = _FriendsTab.friends;
      _onlyOnline = false;
      _searchController.text = trimmed;
      _searchQuery = trimmed;
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<_FriendItem> get _filteredFriends {
    final query = _searchQuery.trim().toLowerCase();
    return _friends.where((friend) {
      if (_onlyOnline && !friend.isOnline) return false;
      if (query.isEmpty) return true;
      return friend.username.toLowerCase().contains(query) ||
          friend.status.toLowerCase().contains(query);
    }).toList();
  }

  int get _onlineFriendsCount => _friends.where((f) => f.isOnline).length;

  void _openUserProfile(String username, String userId) {
    final shell = DesktopShellScope.maybeOf(context);
    if (shell != null) {
      shell.openProfile(username: username, userId: userId);
      return;
    }
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (_) => ProfileScreen(username: username, userId: userId),
      ),
    );
  }

  Future<void> _confirmRemoveFriend(_FriendItem friend) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
        title: Row(
          children: [
            Container(
              width: 38,
              height: 38,
              decoration: BoxDecoration(
                color: const Color(0xFFFFEAEA),
                borderRadius: BorderRadius.circular(10),
              ),
              child: const Icon(
                Icons.person_remove_rounded,
                color: Color(0xFFE04848),
                size: 20,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                'Eliminar amigo',
                style: MeloopTypography.lexend(
                  fontSize: 17,
                  fontWeight: FontWeight.w700,
                  color: const Color(0xFF182222),
                ),
              ),
            ),
          ],
        ),
        content: Text(
          '¿Seguro que quieres eliminar a ${friend.username} de tu red de amigos? Tendrás que enviar una nueva solicitud si quieres volver a conectarte.',
          style: MeloopTypography.lexend(
            fontSize: 13.5,
            color: const Color(0xFF556363),
            height: 1.45,
          ),
        ),
        actionsPadding: const EdgeInsets.fromLTRB(18, 0, 18, 16),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: Text(
              'Cancelar',
              style: MeloopTypography.lexend(
                fontSize: 13.5,
                fontWeight: FontWeight.w600,
                color: const Color(0xFF607070),
              ),
            ),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFE04848),
              foregroundColor: Colors.white,
              elevation: 0,
              padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 10),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20),
              ),
            ),
            onPressed: () => Navigator.pop(dialogContext, true),
            child: Text(
              'Eliminar',
              style: MeloopTypography.lexend(
                fontSize: 13.5,
                fontWeight: FontWeight.w600,
                color: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;
    setState(() {
      _friends.removeWhere((item) => item.id == friend.id);
    });
    showMeloopToastPill(
      context,
      message: '${friend.username} eliminado de amigos',
      icon: Icons.person_remove_outlined,
    );
  }

  void _handleRequest(_FriendRequestItem request, {required bool accepted}) {
    setState(() {
      _requests.removeWhere((item) => item.id == request.id);
      if (accepted) {
        _friends.insert(
          0,
          _FriendItem(
            id: request.id,
            username: request.username,
            status: 'Nuevo amigo en Meloop',
            joinedLabel: 'Conectado ahora',
            friendsCount: 1,
            blogsCount: 0,
            isOnline: true,
          ),
        );
      }
    });
    showMeloopToastPill(
      context,
      message: accepted
          ? 'Aceptaste a ${request.username}'
          : 'Solicitud de ${request.username} rechazada',
      icon: accepted ? Icons.check_rounded : Icons.close_rounded,
    );
  }

  void _sendFriendRequest(_SuggestedProfileItem profile) {
    setState(() {
      _suggestions.removeWhere((item) => item.id == profile.id);
    });
    showMeloopToastPill(
      context,
      message: 'Solicitud enviada a ${profile.username}',
      icon: Icons.person_add_alt_1_rounded,
    );
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.sizeOf(context).width;
    final bool isDesktop = screenWidth > 900;
    final bool insideDesktopShell = DesktopShellScope.maybeOf(context) != null;
    final bool insideMobileShell = MobileShellScope.maybeOf(context) != null;

    // En escritorio la columna derecha muestra "Perfiles que te podrían interesar",
    // así que si el usuario venía de la pestaña móvil "Sugeridos" volvemos a "Amigos".
    if (isDesktop && _activeTab == _FriendsTab.suggestions) {
      _activeTab = _FriendsTab.friends;
    }

    return AnimatedBuilder(
      animation: MeloopThemeController.instance,
      builder: (context, _) => Scaffold(
        backgroundColor: Colors.transparent,
        body: Column(
          children: [
            if (!insideDesktopShell && !insideMobileShell)
              _buildTopBar(isDesktop, context),
            Expanded(
              child: isDesktop ? _buildDesktopLayout() : _buildMobileLayout(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTopBar(bool isDesktop, BuildContext context) {
    if (!isDesktop && defaultTargetPlatform == TargetPlatform.android) {
      return MobileAppHeader(onSearchTap: () => openGlobalSearch(context));
    }

    return DesktopAppHeader(
      activeSection: DesktopNavigationSection.friends,
      onHomeTap: () => navigateToDesktopHome(context),
      onFriendsTap: () {},
      onArtistsTap: () => navigateToDesktopSection(
        context,
        (_) => const ArtistsAndSongsScreen(),
      ),
      onProfileTap: () =>
          navigateToDesktopSection(context, (_) => const ProfileScreen()),
      onSearchTap: () => openGlobalSearch(context),
    );
  }

  Widget _buildDesktopLayout() {
    final gamification = GamificationDemoController.instance;

    return LayoutBuilder(
      builder: (context, constraints) {
        final double horizontalPadding = (constraints.maxWidth * 0.032).clamp(
          24.0,
          56.0,
        );
        final double contentWidth = math.min(
          constraints.maxWidth - horizontalPadding * 2,
          1560.0,
        );
        final double columnGap = (contentWidth * 0.026).clamp(24.0, 42.0);
        final double rightPanelWidth = (contentWidth * 0.34).clamp(
          330.0,
          470.0,
        );

        return Center(
          child: SizedBox(
            width: contentWidth,
            child: Padding(
              padding: const EdgeInsets.only(top: 32, bottom: 22),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Hoja principal izquierda: Red de amigos (Amigos / Solicitudes)
                  Expanded(
                    child: TapeCornerOverlay(
                      leftWidth: 162,
                      rightWidth: 140,
                      child: StackedPaperCard(
                        textureAsset: 'assets/textures/paper.png',
                        padding: const EdgeInsets.fromLTRB(38, 30, 38, 22),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            _buildDesktopLeftHeader(),
                            const SizedBox(height: 18),
                            if (_activeTab == _FriendsTab.friends) ...[
                              _buildFriendsToolbar(isMobile: false),
                              const SizedBox(height: 16),
                            ],
                            Expanded(
                              child: AnimatedSwitcher(
                                duration: const Duration(milliseconds: 220),
                                switchInCurve: Curves.easeOutCubic,
                                switchOutCurve: Curves.easeInCubic,
                                child: _activeTab == _FriendsTab.friends
                                    ? _buildFriendsList(
                                        key: const ValueKey('desktop-friends'),
                                        isMobile: false,
                                      )
                                    : _buildRequestsList(
                                        key: const ValueKey('desktop-requests'),
                                        isMobile: false,
                                      ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                  SizedBox(width: columnGap),
                  // Hoja secundaria derecha: Perfiles que te podrían interesar
                  SizedBox(
                    width: rightPanelWidth,
                    child: AnimatedBuilder(
                      animation: gamification,
                      builder: (context, _) => Stack(
                        alignment: Alignment.topCenter,
                        clipBehavior: Clip.none,
                        children: [
                          Positioned.fill(
                            child: StackedPaperCard(
                              textureAsset: 'assets/textures/paper.png',
                              tintColor: MeloopThemeController
                                  .instance
                                  .colorTheme
                                  .softCardTint,
                              padding: const EdgeInsets.fromLTRB(
                                24,
                                28,
                                24,
                                20,
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.stretch,
                                children: [
                                  Text(
                                    'Perfiles que te podrían interesar',
                                    textAlign: TextAlign.center,
                                    style: MeloopTypography.lexend(
                                      fontSize: 18.5,
                                      fontWeight: FontWeight.w700,
                                      color: MeloopThemeController
                                          .instance
                                          .cardTitleColor,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    'Basado en tus gustos y artistas en Spotify',
                                    textAlign: TextAlign.center,
                                    style: MeloopTypography.lexend(
                                      fontSize: 12,
                                      color: MeloopThemeController
                                          .instance
                                          .cardMutedColor,
                                    ),
                                  ),
                                  const SizedBox(height: 14),
                                  Divider(
                                    height: 1,
                                    thickness: 1,
                                    color: MeloopThemeController
                                        .instance
                                        .cardDividerColor,
                                  ),
                                  const SizedBox(height: 14),
                                  Expanded(
                                    child: _buildSuggestionsList(
                                      isMobile: false,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                          Positioned(
                            top: -20,
                            child: CustomizablePin(
                              color: gamification.activePinColor,
                              size: 48,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildDesktopLeftHeader() {
    final requestsLabel = _requests.isEmpty
        ? 'Solicitudes'
        : 'Solicitudes (${_requests.length})';

    final titleBlock = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Transform.rotate(
          angle: -2.2 * math.pi / 180.0,
          child: Text(
            'Red de amigos',
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: MeloopTypography.momoDisplay(
              fontSize: 30,
              color: MeloopThemeController.instance.cardTitleColor,
            ),
          ),
        ),
        const SizedBox(height: 2),
        Text(
          '${_friends.length} amigos · $_onlineFriendsCount en línea',
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: MeloopTypography.lexend(
            fontSize: 12.5,
            color: MeloopThemeController.instance.cardMutedColor,
          ),
        ),
      ],
    );

    final pillTabs = MeloopPillTabs<_FriendsTab>(
      items: [
        const MeloopPillItem(value: _FriendsTab.friends, label: 'Amigos'),
        MeloopPillItem(value: _FriendsTab.requests, label: requestsLabel),
      ],
      selectedValue: _activeTab == _FriendsTab.suggestions
          ? _FriendsTab.friends
          : _activeTab,
      onSelected: (tab) => setState(() => _activeTab = tab),
      height: 44,
      padding: 4,
      fontSize: 14,
      horizontalItemPadding: 18,
      fixedButtonWidth: 148,
      pillColor: _pillTrackColor,
      equalWidths: true,
    );

    return LayoutBuilder(
      builder: (context, constraints) {
        if (constraints.maxWidth < 660) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              titleBlock,
              const SizedBox(height: 14),
              Center(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: pillTabs,
                ),
              ),
            ],
          );
        }

        return Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Expanded(child: titleBlock),
            const SizedBox(width: 16),
            pillTabs,
          ],
        );
      },
    );
  }

  Widget _buildMobileLayout() {
    final isAndroid = defaultTargetPlatform == TargetPlatform.android;
    final requestsLabel = _requests.isEmpty
        ? 'Solicitudes'
        : 'Solicitudes (${_requests.length})';

    return Padding(
      padding: const EdgeInsets.fromLTRB(10, 12, 10, 0),
      child: TapeCornerOverlay(
        leftWidth: 96,
        rightWidth: 84,
        child: StackedPaperCard(
          textureAsset: 'assets/textures/paper.png',
          padding: const EdgeInsets.fromLTRB(14, 16, 14, 0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  Expanded(
                    child: Align(
                      alignment: Alignment.centerLeft,
                      child: Transform.rotate(
                        angle: -2.0 * math.pi / 180.0,
                        child: Text(
                          'Red de amigos',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: MeloopTypography.momoDisplay(
                            fontSize: 22,
                            color:
                                MeloopThemeController.instance.cardTitleColor,
                          ),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: MeloopThemeController.instance.softBadgeBgColor,
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: Text(
                      '${_friends.length} amigos',
                      style: MeloopTypography.lexend(
                        fontSize: 11.5,
                        fontWeight: FontWeight.w600,
                        color: MeloopThemeController.instance.softBadgeFgColor,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              LayoutBuilder(
                builder: (context, tabConstraints) {
                  final buttonWidth = ((tabConstraints.maxWidth - 8) / 3).clamp(
                    84.0,
                    140.0,
                  );
                  return Center(
                    child: MeloopPillTabs<_FriendsTab>(
                      items: [
                        const MeloopPillItem(
                          value: _FriendsTab.friends,
                          label: 'Amigos',
                        ),
                        MeloopPillItem(
                          value: _FriendsTab.requests,
                          label: requestsLabel,
                        ),
                        const MeloopPillItem(
                          value: _FriendsTab.suggestions,
                          label: 'Sugeridos',
                        ),
                      ],
                      selectedValue: _activeTab,
                      onSelected: (tab) => setState(() => _activeTab = tab),
                      height: 38,
                      padding: 4,
                      fontSize: 12,
                      horizontalItemPadding: 8,
                      fixedButtonWidth: buttonWidth,
                      pillColor: _pillTrackColor,
                      equalWidths: true,
                    ),
                  );
                },
              ),
              const SizedBox(height: 12),
              if (_activeTab == _FriendsTab.friends) ...[
                _buildFriendsToolbar(isMobile: true),
                const SizedBox(height: 10),
              ],
              Expanded(
                child: AnimatedSwitcher(
                  duration: const Duration(milliseconds: 200),
                  child: switch (_activeTab) {
                    _FriendsTab.friends => _buildFriendsList(
                      key: const ValueKey('mobile-friends'),
                      isMobile: true,
                      bottomPadding: isAndroid ? 142 : 24,
                    ),
                    _FriendsTab.requests => _buildRequestsList(
                      key: const ValueKey('mobile-requests'),
                      isMobile: true,
                      bottomPadding: isAndroid ? 142 : 24,
                    ),
                    _FriendsTab.suggestions => _buildSuggestionsList(
                      key: const ValueKey('mobile-suggestions'),
                      isMobile: true,
                      bottomPadding: isAndroid ? 142 : 24,
                    ),
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFriendsToolbar({required bool isMobile}) {
    final themeCtrl = MeloopThemeController.instance;
    final searchField = Container(
      height: isMobile ? 38 : 40,
      decoration: BoxDecoration(
        color: themeCtrl.inputFillColor,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: themeCtrl.inputBorderColor),
      ),
      child: TextField(
        controller: _searchController,
        onChanged: (value) => setState(() => _searchQuery = value),
        style: MeloopTypography.lexend(
          fontSize: isMobile ? 12.5 : 13.5,
          color: themeCtrl.cardTitleColor,
        ),
        decoration: InputDecoration(
          hintText: 'Buscar en tus amigos...',
          hintStyle: MeloopTypography.lexend(
            fontSize: isMobile ? 12.5 : 13.5,
            color: themeCtrl.cardSubtleMutedColor,
          ),
          prefixIcon: Icon(
            Icons.search_rounded,
            size: 19,
            color: themeCtrl.cardMutedColor,
          ),
          suffixIcon: _searchQuery.isNotEmpty
              ? IconButton(
                  tooltip: 'Limpiar búsqueda',
                  padding: EdgeInsets.zero,
                  icon: Icon(
                    Icons.close_rounded,
                    size: 17,
                    color: themeCtrl.cardMutedColor,
                  ),
                  onPressed: () {
                    _searchController.clear();
                    setState(() => _searchQuery = '');
                  },
                )
              : null,
          border: InputBorder.none,
          contentPadding: EdgeInsets.symmetric(
            vertical: isMobile ? 9 : 10,
            horizontal: 8,
          ),
          isDense: true,
        ),
      ),
    );

    final onlineChip = FilterChip(
      selected: _onlyOnline,
      onSelected: (selected) => setState(() => _onlyOnline = selected),
      showCheckmark: false,
      avatar: Container(
        width: 9,
        height: 9,
        decoration: BoxDecoration(
          color: _onlyOnline ? Colors.white : const Color(0xFF16A34A),
          shape: BoxShape.circle,
        ),
      ),
      label: Text(
        'En línea ($_onlineFriendsCount)',
        style: MeloopTypography.lexend(
          fontSize: isMobile ? 11.5 : 12.5,
          fontWeight: FontWeight.w600,
          color: _onlyOnline ? Colors.white : themeCtrl.cardBodyColor,
        ),
      ),
      backgroundColor: themeCtrl.inputFillColor,
      selectedColor: _tealAccent,
      side: BorderSide(
        color: _onlyOnline ? _tealAccent : themeCtrl.inputBorderColor,
      ),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );

    return Row(
      children: [
        Expanded(child: searchField),
        const SizedBox(width: 10),
        onlineChip,
      ],
    );
  }

  Widget _buildFriendsList({
    Key? key,
    required bool isMobile,
    double bottomPadding = 16,
  }) {
    final visibleFriends = _filteredFriends;
    if (_friends.isEmpty) {
      return _buildEmptyPanel(
        key: key,
        icon: Icons.people_outline_rounded,
        title: 'Aún no tienes amigos',
        subtitle:
            'Explora los perfiles sugeridos para empezar a construir tu red musical.',
      );
    }

    if (visibleFriends.isEmpty) {
      return _buildEmptyPanel(
        key: key,
        icon: Icons.search_off_rounded,
        title: 'Sin resultados',
        subtitle: 'No encontramos amigos que coincidan con tu filtro actual.',
      );
    }

    return ListView.builder(
      key: key,
      padding: EdgeInsets.only(bottom: bottomPadding, right: isMobile ? 0 : 4),
      itemCount: visibleFriends.length,
      itemBuilder: (context, index) {
        final friend = visibleFriends[index];
        return _FriendProfileTile(
          friend: friend,
          isMobile: isMobile,
          onTap: () => _openUserProfile(friend.username, friend.id),
          onDeleteTap: () => _confirmRemoveFriend(friend),
        );
      },
    );
  }

  Widget _buildRequestsList({
    Key? key,
    required bool isMobile,
    double bottomPadding = 16,
  }) {
    if (_requests.isEmpty) {
      return _buildEmptyPanel(
        key: key,
        icon: Icons.mark_email_read_outlined,
        title: 'No tienes solicitudes pendientes',
        subtitle:
            'Cuando alguien quiera conectar contigo, aparecerá aquí para que aceptes o rechaces.',
      );
    }

    return ListView.builder(
      key: key,
      padding: EdgeInsets.only(bottom: bottomPadding, right: isMobile ? 0 : 4),
      itemCount: _requests.length,
      itemBuilder: (context, index) {
        final request = _requests[index];
        return _FriendRequestTile(
          request: request,
          isMobile: isMobile,
          onTap: () => _openUserProfile(request.username, request.id),
          onAccept: () => _handleRequest(request, accepted: true),
          onReject: () => _handleRequest(request, accepted: false),
        );
      },
    );
  }

  Widget _buildSuggestionsList({
    Key? key,
    required bool isMobile,
    double bottomPadding = 16,
  }) {
    if (_suggestions.isEmpty) {
      return _buildEmptyPanel(
        key: key,
        icon: Icons.auto_awesome_outlined,
        title: '¡Al día con las sugerencias!',
        subtitle:
            'Pronto te recomendaremos más perfiles según tus playlists y artistas favoritos.',
      );
    }

    return ListView.builder(
      key: key,
      padding: EdgeInsets.only(bottom: bottomPadding, right: isMobile ? 0 : 2),
      itemCount: _suggestions.length,
      itemBuilder: (context, index) {
        final profile = _suggestions[index];
        return _SuggestedProfileTile(
          profile: profile,
          isMobile: isMobile,
          onTap: () => _openUserProfile(profile.username, profile.id),
          onAddTap: () => _sendFriendRequest(profile),
        );
      },
    );
  }

  Widget _buildEmptyPanel({
    Key? key,
    required IconData icon,
    required String title,
    required String subtitle,
  }) {
    final themeCtrl = MeloopThemeController.instance;
    return Center(
      key: key,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 46, color: themeCtrl.cardSubtleMutedColor),
            const SizedBox(height: 12),
            Text(
              title,
              textAlign: TextAlign.center,
              style: MeloopTypography.lexend(
                fontSize: 15.5,
                fontWeight: FontWeight.w600,
                color: themeCtrl.cardTitleColor,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              subtitle,
              textAlign: TextAlign.center,
              style: MeloopTypography.lexend(
                fontSize: 12.5,
                color: themeCtrl.cardMutedColor,
                height: 1.4,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _FriendProfileTile extends StatelessWidget {
  const _FriendProfileTile({
    required this.friend,
    required this.isMobile,
    required this.onTap,
    required this.onDeleteTap,
  });

  final _FriendItem friend;
  final bool isMobile;
  final VoidCallback onTap;
  final VoidCallback onDeleteTap;

  @override
  Widget build(BuildContext context) {
    final themeCtrl = MeloopThemeController.instance;
    final isDark = themeCtrl.isDarkCardStyle;
    return Container(
      margin: EdgeInsets.only(bottom: isMobile ? 10 : 12),
      decoration: BoxDecoration(
        color: themeCtrl.innerTileColor,
        border: Border.all(color: themeCtrl.innerTileBorderColor),
        boxShadow: isDark
            ? null
            : [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.055),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          child: Padding(
            padding: EdgeInsets.symmetric(
              horizontal: isMobile ? 12 : 16,
              vertical: isMobile ? 11 : 13,
            ),
            child: Row(
              children: [
                _FramedAvatar(
                  size: isMobile ? 46 : 54,
                  isOnline: friend.isOnline,
                ),
                SizedBox(width: isMobile ? 12 : 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Row(
                        children: [
                          Flexible(
                            child: Text(
                              friend.username,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: MeloopTypography.lexend(
                                fontSize: isMobile ? 15 : 16,
                                fontWeight: FontWeight.w700,
                                color: themeCtrl.cardTitleColor,
                                height: 1.2,
                              ),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Container(
                            width: 9,
                            height: 9,
                            decoration: BoxDecoration(
                              color: friend.isOnline
                                  ? const Color(0xFF16A34A)
                                  : const Color(0xFFB6C2C2),
                              shape: BoxShape.circle,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 2),
                      Text(
                        friend.status,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: isMobile ? 12 : 13,
                          color: themeCtrl.cardMutedColor,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        '${friend.joinedLabel} · ${friend.friendsCount} amigos · ${friend.blogsCount} blogs',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: isMobile ? 10.5 : 11.5,
                          color: themeCtrl.cardSubtleMutedColor,
                        ),
                      ),
                    ],
                  ),
                ),
                SizedBox(width: isMobile ? 6 : 12),
                if (!isMobile) ...[
                  OutlinedButton(
                    style: OutlinedButton.styleFrom(
                      foregroundColor: themeCtrl.accentLinkColor,
                      side: BorderSide(
                        color: isDark
                            ? themeCtrl.colorTheme.headerTop.withValues(
                                alpha: 0.45,
                              )
                            : themeCtrl.colorTheme.gridLine,
                      ),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 14,
                        vertical: 9,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(20),
                      ),
                    ),
                    onPressed: onTap,
                    child: Text(
                      'Ver perfil',
                      style: MeloopTypography.lexend(
                        fontSize: 12.5,
                        fontWeight: FontWeight.w600,
                        color: themeCtrl.accentLinkColor,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                ],
                Tooltip(
                  message: 'Eliminar amigo',
                  child: Material(
                    color: isDark
                        ? const Color(0xFFE04848).withValues(alpha: 0.20)
                        : const Color(0xFFFFF1F1),
                    borderRadius: BorderRadius.circular(20),
                    child: InkWell(
                      onTap: onDeleteTap,
                      borderRadius: BorderRadius.circular(20),
                      child: Padding(
                        padding: EdgeInsets.symmetric(
                          horizontal: isMobile ? 9 : 12,
                          vertical: 8,
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              Icons.person_remove_outlined,
                              color: isDark
                                  ? const Color(0xFFFF7A7A)
                                  : const Color(0xFFD94444),
                              size: 18,
                            ),
                            if (!isMobile) ...[
                              const SizedBox(width: 6),
                              Text(
                                'Eliminar',
                                style: MeloopTypography.lexend(
                                  fontSize: 12.5,
                                  fontWeight: FontWeight.w600,
                                  color: isDark
                                      ? const Color(0xFFFF7A7A)
                                      : const Color(0xFFD94444),
                                ),
                              ),
                            ],
                          ],
                        ),
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
  }
}

class _FriendRequestTile extends StatelessWidget {
  const _FriendRequestTile({
    required this.request,
    required this.isMobile,
    required this.onTap,
    required this.onAccept,
    required this.onReject,
  });

  final _FriendRequestItem request;
  final bool isMobile;
  final VoidCallback onTap;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    final themeCtrl = MeloopThemeController.instance;
    final isDark = themeCtrl.isDarkCardStyle;
    final actions = Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        OutlinedButton(
          style: OutlinedButton.styleFrom(
            foregroundColor: themeCtrl.cardMutedColor,
            side: BorderSide(color: themeCtrl.inputBorderColor),
            padding: EdgeInsets.symmetric(
              horizontal: isMobile ? 12 : 16,
              vertical: isMobile ? 7 : 9,
            ),
            minimumSize: isMobile ? const Size(0, 34) : null,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(20),
            ),
          ),
          onPressed: onReject,
          child: Text(
            'Rechazar',
            style: MeloopTypography.lexend(
              fontSize: isMobile ? 12 : 12.5,
              fontWeight: FontWeight.w600,
              color: themeCtrl.cardMutedColor,
            ),
          ),
        ),
        const SizedBox(width: 8),
        ElevatedButton(
          style: ElevatedButton.styleFrom(
            backgroundColor: themeCtrl.primaryColor,
            foregroundColor: Colors.white,
            elevation: 0,
            padding: EdgeInsets.symmetric(
              horizontal: isMobile ? 14 : 18,
              vertical: isMobile ? 7 : 9,
            ),
            minimumSize: isMobile ? const Size(0, 34) : null,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(20),
            ),
          ),
          onPressed: onAccept,
          child: Text(
            'Aceptar',
            style: MeloopTypography.lexend(
              fontSize: isMobile ? 12 : 12.5,
              fontWeight: FontWeight.w600,
              color: Colors.white,
            ),
          ),
        ),
      ],
    );

    return Container(
      margin: EdgeInsets.only(bottom: isMobile ? 10 : 12),
      decoration: BoxDecoration(
        color: themeCtrl.innerTileColor,
        border: Border.all(color: themeCtrl.innerTileBorderColor),
        boxShadow: isDark
            ? null
            : [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.055),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          child: Padding(
            padding: EdgeInsets.symmetric(
              horizontal: isMobile ? 12 : 16,
              vertical: isMobile ? 11 : 13,
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    _FramedAvatar(size: isMobile ? 44 : 52, isOnline: null),
                    SizedBox(width: isMobile ? 12 : 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            request.username,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: isMobile ? 15 : 16,
                              fontWeight: FontWeight.w700,
                              color: themeCtrl.cardTitleColor,
                              height: 1.2,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            request.description,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: isMobile ? 12 : 13,
                              color: themeCtrl.cardMutedColor,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            request.mutualLabel,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: isMobile ? 10.5 : 11.5,
                              color: themeCtrl.cardSubtleMutedColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                    if (!isMobile) ...[const SizedBox(width: 12), actions],
                  ],
                ),
                if (isMobile) ...[
                  const SizedBox(height: 10),
                  Align(alignment: Alignment.centerRight, child: actions),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _SuggestedProfileTile extends StatelessWidget {
  const _SuggestedProfileTile({
    required this.profile,
    required this.isMobile,
    required this.onTap,
    required this.onAddTap,
  });

  final _SuggestedProfileItem profile;
  final bool isMobile;
  final VoidCallback onTap;
  final VoidCallback onAddTap;

  @override
  Widget build(BuildContext context) {
    final themeCtrl = MeloopThemeController.instance;
    final isDark = themeCtrl.isDarkCardStyle;
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: themeCtrl.innerTileColor,
        border: Border.all(color: themeCtrl.innerTileBorderColor),
        boxShadow: isDark
            ? null
            : [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 6,
                  offset: const Offset(0, 2),
                ),
              ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 11),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const _FramedAvatar(size: 46, isOnline: null),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            profile.username,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: 15,
                              fontWeight: FontWeight.w700,
                              color: themeCtrl.cardTitleColor,
                              height: 1.2,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            profile.description,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: 11.5,
                              color: themeCtrl.cardMutedColor,
                              height: 1.3,
                            ),
                          ),
                          const SizedBox(height: 3),
                          Text(
                            profile.joinedLabel,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: 10.5,
                              color: themeCtrl.cardSubtleMutedColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: themeCtrl.softBadgeBgColor,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.graphic_eq_rounded,
                            color: themeCtrl.softBadgeFgColor,
                            size: 13,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            '${profile.compatibility}% afinidad',
                            style: MeloopTypography.lexend(
                              fontSize: 11,
                              fontWeight: FontWeight.w600,
                              color: themeCtrl.softBadgeFgColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const Spacer(),
                    ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: themeCtrl.primaryColor,
                        foregroundColor: Colors.white,
                        elevation: 0,
                        minimumSize: const Size(0, 32),
                        padding: EdgeInsets.symmetric(
                          horizontal: isMobile ? 12 : 14,
                          vertical: 6,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(18),
                        ),
                      ),
                      onPressed: onAddTap,
                      icon: const Icon(Icons.person_add_alt_1_rounded, size: 15),
                      label: Text(
                        'Agregar',
                        style: MeloopTypography.lexend(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: Colors.white,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _FramedAvatar extends StatelessWidget {
  const _FramedAvatar({required this.size, required this.isOnline});

  final double size;
  final bool? isOnline;

  @override
  Widget build(BuildContext context) {
    final themeCtrl = MeloopThemeController.instance;
    final isDark = themeCtrl.isDarkCardStyle;
    return Container(
      width: size,
      height: size,
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        color: isDark ? Colors.white.withValues(alpha: 0.14) : Colors.white,
        boxShadow: isDark
            ? null
            : [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.14),
                  blurRadius: 4,
                  offset: const Offset(0, 1),
                ),
              ],
      ),
      child: ColoredBox(
        color: themeCtrl.placeholderBgColor,
        child: Icon(
          Icons.person,
          size: size * 0.56,
          color: themeCtrl.placeholderIconColor,
        ),
      ),
    );
  }
}

class _FriendItem {
  const _FriendItem({
    required this.id,
    required this.username,
    required this.status,
    required this.joinedLabel,
    required this.friendsCount,
    required this.blogsCount,
    required this.isOnline,
  });

  final String id;
  final String username;
  final String status;
  final String joinedLabel;
  final int friendsCount;
  final int blogsCount;
  final bool isOnline;
}

class _FriendRequestItem {
  const _FriendRequestItem({
    required this.id,
    required this.username,
    required this.description,
    required this.mutualLabel,
  });

  final String id;
  final String username;
  final String description;
  final String mutualLabel;
}

class _SuggestedProfileItem {
  const _SuggestedProfileItem({
    required this.id,
    required this.username,
    required this.description,
    required this.joinedLabel,
    required this.compatibility,
  });

  final String id;
  final String username;
  final String description;
  final String joinedLabel;
  final int compatibility;
}
