import 'dart:math' as math;
import 'dart:ui' show lerpDouble;

import 'package:flutter/material.dart';

import '../core/navigation/desktop_navigation_shell.dart';
import '../features/settings/presentation/account_settings_screen.dart';
import '../features/music/presentation/music_preferences_screen.dart';
import '../features/profile/presentation/blocked_users_screen.dart';
import '../services/api_client.dart';
import '../services/post_audio_playback.dart';
import '../theme/meloop_typography.dart';
import 'meloop_pill_tabs.dart';
import 'notifications_menu.dart';
import 'paper/stacked_paper_card.dart';
import 'session_logout.dart';

enum DesktopNavigationSection { home, friends, artists, none }

void navigateToDesktopSection(BuildContext context, WidgetBuilder builder) {
  final shell = DesktopShellScope.maybeOf(context);
  if (shell != null) {
    shell.navigateToSectionWidget(builder);
    return;
  }
  Navigator.of(context).pushAndRemoveUntil(
    PageRouteBuilder<void>(
      pageBuilder: (context, animation, secondaryAnimation) => builder(context),
      transitionDuration: Duration.zero,
      reverseTransitionDuration: Duration.zero,
    ),
    (route) => route.isFirst,
  );
}

void navigateToDesktopHome(BuildContext context) {
  final shell = DesktopShellScope.maybeOf(context);
  if (shell != null) {
    shell.goHome();
    return;
  }
  Navigator.of(context).popUntil((route) => route.isFirst);
}

enum _ProfileMenuAction {
  viewProfile,
  accountSettings,
  musicPreferences,
  blockedUsers,
  moderationQueue,
  logout,
}

class DesktopAppHeader extends StatelessWidget {
  const DesktopAppHeader({
    super.key,
    required this.activeSection,
    required this.onHomeTap,
    required this.onFriendsTap,
    required this.onArtistsTap,
    this.onProfileTap,
    this.onModerationTap,
    this.onSearchTap,
    this.onPushRoute,
    this.apiClient,
    this.searchOpen = false,
    this.searchController,
    this.onSearchChanged,
    this.onSearchSubmitted,
    this.onSearchBack,
  });

  final DesktopNavigationSection activeSection;
  final VoidCallback onHomeTap;
  final VoidCallback onFriendsTap;
  final VoidCallback onArtistsTap;
  final VoidCallback? onProfileTap;
  final VoidCallback? onModerationTap;
  final VoidCallback? onSearchTap;
  final void Function(WidgetBuilder builder)? onPushRoute;
  final ApiClient? apiClient;
  final bool searchOpen;
  final TextEditingController? searchController;
  final ValueChanged<String>? onSearchChanged;
  final ValueChanged<String>? onSearchSubmitted;
  final VoidCallback? onSearchBack;

  static const _headerGradient = LinearGradient(
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
    colors: [Color(0xFF59E5C9), Color(0xFF00AF92)],
  );
  static const _pillColor = Color(0xFF16877F);
  static const _navFontSize = 14.5;
  static const _navButtonWidth = 192.0;
  static const _navHorizontalPadding = 16.0;
  static const _navOuterPadding = 4.5;
  static const _navTotalWidth = _navButtonWidth * 3 + _navOuterPadding * 2;
  static const _searchGap = 12.0;
  static const _searchButtonSize = 48.0;

  void _pushScreen(BuildContext context, WidgetBuilder builder) {
    if (onPushRoute != null) {
      onPushRoute!(builder);
      return;
    }
    final shell = DesktopShellScope.maybeOf(context);
    if (shell != null) {
      shell.pushPage(builder);
      return;
    }
    Navigator.of(context).push(MaterialPageRoute<void>(builder: builder));
  }

  @override
  Widget build(BuildContext context) {
    const navigationWidth = _navTotalWidth;

    return Container(
      height: 84,
      decoration: BoxDecoration(
        gradient: _headerGradient,
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(24)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.18),
            blurRadius: 14,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 10),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final availableCenterSpace = constraints.maxWidth - 360;
          final navScale = availableCenterSpace < navigationWidth
              ? (availableCenterSpace / navigationWidth).clamp(0.65, 1.0)
              : 1.0;
          final scaledNavHalf = navigationWidth * navScale / 2;
          final searchRightEdge =
              constraints.maxWidth / 2 +
              scaledNavHalf +
              (onSearchTap != null ? _searchGap + _searchButtonSize : 0);
          final rightAvailable = constraints.maxWidth - searchRightEdge - 16;

          final showFullMiniPlayer = rightAvailable >= 470;
          final showCompactMiniPlayer =
              !showFullMiniPlayer && rightAvailable >= 340;

          return Stack(
            alignment: Alignment.center,
            children: [
              // Izquierda (logo) y Derecha (mini reproductor + notificaciones + perfil)
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  MouseRegion(
                    cursor: SystemMouseCursors.click,
                    child: GestureDetector(
                      onTap: onHomeTap,
                      child: const MeloopLogoSticker(
                        height: 46,
                        strokeWidth: 2.2,
                      ),
                    ),
                  ),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (showFullMiniPlayer) ...[
                        const _HeaderMiniPlayer(width: 340),
                        const SizedBox(width: 14),
                      ] else if (showCompactMiniPlayer) ...[
                        const _HeaderMiniPlayer(width: 210, compact: true),
                        const SizedBox(width: 12),
                      ],
                      const NotificationsMenu(),
                      const SizedBox(width: 10),
                      _HeaderProfileDropdown(
                        onProfileTap: onProfileTap,
                        onModerationTap: onModerationTap,
                        apiClient: apiClient,
                        onMusicPreferencesTap: () => _pushScreen(
                          context,
                          (_) => MusicPreferencesScreen(apiClient: apiClient),
                        ),
                        onBlockedUsersTap: () => _pushScreen(
                          context,
                          (_) => const BlockedUsersScreen(),
                        ),
                        onAccountSettingsTap: () => _pushScreen(
                          context,
                          (_) => AccountSettingsScreen(apiClient: apiClient),
                        ),
                      ),
                    ],
                  ),
                ],
              ),

              // Botones de navegación centrados + barra de búsqueda unificada con
              // transición fluida (sin saltos ni desaparición brusca de la lupa)
              Center(
                child: Transform.scale(
                  scale: navScale,
                  child: _DesktopHeaderNavAndSearch(
                    activeSection: activeSection,
                    navigationWidth: navigationWidth,
                    onHomeTap: onHomeTap,
                    onFriendsTap: onFriendsTap,
                    onArtistsTap: onArtistsTap,
                    onSearchTap: onSearchTap,
                    searchOpen: searchOpen,
                    searchController: searchController,
                    onSearchChanged: onSearchChanged,
                    onSearchSubmitted: onSearchSubmitted,
                    onSearchBack: onSearchBack,
                    pillColor: _pillColor,
                    fontSize: _navFontSize,
                    horizontalItemPadding: _navHorizontalPadding,
                    outerPadding: _navOuterPadding,
                    searchGap: _searchGap,
                    searchButtonSize: _searchButtonSize,
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _DesktopHeaderNavAndSearch extends StatefulWidget {
  const _DesktopHeaderNavAndSearch({
    required this.activeSection,
    required this.navigationWidth,
    required this.onHomeTap,
    required this.onFriendsTap,
    required this.onArtistsTap,
    required this.searchOpen,
    required this.pillColor,
    required this.fontSize,
    required this.horizontalItemPadding,
    required this.outerPadding,
    required this.searchGap,
    required this.searchButtonSize,
    this.onSearchTap,
    this.searchController,
    this.onSearchChanged,
    this.onSearchSubmitted,
    this.onSearchBack,
  });

  final DesktopNavigationSection activeSection;
  final double navigationWidth;
  final VoidCallback onHomeTap;
  final VoidCallback onFriendsTap;
  final VoidCallback onArtistsTap;
  final VoidCallback? onSearchTap;
  final bool searchOpen;
  final TextEditingController? searchController;
  final ValueChanged<String>? onSearchChanged;
  final ValueChanged<String>? onSearchSubmitted;
  final VoidCallback? onSearchBack;
  final Color pillColor;
  final double fontSize;
  final double horizontalItemPadding;
  final double outerPadding;
  final double searchGap;
  final double searchButtonSize;

  @override
  State<_DesktopHeaderNavAndSearch> createState() =>
      _DesktopHeaderNavAndSearchState();
}

class _DesktopHeaderNavAndSearchState extends State<_DesktopHeaderNavAndSearch>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final FocusNode _searchFocusNode;

  @override
  void initState() {
    super.initState();
    _searchFocusNode = FocusNode();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 260),
      reverseDuration: const Duration(milliseconds: 220),
      value: widget.searchOpen ? 1.0 : 0.0,
    );
    if (widget.searchOpen) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && widget.searchOpen) {
          _searchFocusNode.requestFocus();
        }
      });
    }
  }

  @override
  void didUpdateWidget(covariant _DesktopHeaderNavAndSearch oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.searchOpen != widget.searchOpen) {
      if (widget.searchOpen) {
        _controller.forward();
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted && widget.searchOpen) {
            _searchFocusNode.requestFocus();
          }
        });
      } else {
        _searchFocusNode.unfocus();
        _controller.reverse();
      }
    }
  }

  @override
  void dispose() {
    _searchFocusNode.dispose();
    _controller.dispose();
    super.dispose();
  }

  void _handleCircleButtonTap() {
    if (widget.searchOpen) {
      _searchFocusNode.unfocus();
      _controller.reverse();
      widget.onSearchBack?.call();
    } else {
      widget.onSearchTap?.call();
    }
  }

  @override
  Widget build(BuildContext context) {
    final hasSearchAction = widget.onSearchTap != null || widget.searchOpen;
    final sideSlot = hasSearchAction
        ? widget.searchGap + widget.searchButtonSize
        : 0.0;
    final totalWidth = widget.navigationWidth + sideSlot * 2;

    return SizedBox(
      width: totalWidth,
      height: 50,
      child: AnimatedBuilder(
        animation: _controller,
        builder: (context, _) {
          final t = Curves.easeOutCubic.transform(_controller.value);
          final navOpacity = (1.0 - t * 1.65).clamp(0.0, 1.0);
          final navScale = lerpDouble(1.0, 0.96, t)!;
          final navShiftX = lerpDouble(0.0, -14.0, t)!;

          final searchOpacity = ((t - 0.05) * 1.45).clamp(0.0, 1.0);
          final searchWidthFactor = lerpDouble(0.14, 1.0, t)!.clamp(0.0, 1.0);

          return Stack(
            alignment: Alignment.center,
            children: [
              // Píldora de navegación exactamente centrada en la barra
              Positioned(
                left: sideSlot,
                width: widget.navigationWidth,
                top: 0,
                bottom: 0,
                child: IgnorePointer(
                  ignoring: t > 0.35,
                  child: Opacity(
                    opacity: navOpacity,
                    child: Transform.translate(
                      offset: Offset(navShiftX, 0),
                      child: Transform.scale(
                        scale: navScale,
                        child: _buildNavigationPill(),
                      ),
                    ),
                  ),
                ),
              ),

              // Barra de búsqueda desplegándose desde la lupa hacia la izquierda
              if (_controller.value > 0.01)
                Positioned(
                  left: sideSlot,
                  width: widget.navigationWidth,
                  top: 2,
                  bottom: 2,
                  child: IgnorePointer(
                    ignoring: !widget.searchOpen && t < 0.5,
                    child: Opacity(
                      opacity: searchOpacity,
                      child: Align(
                        alignment: Alignment.centerRight,
                        child: FractionallySizedBox(
                          widthFactor: searchWidthFactor,
                          child: _buildSearchBar(),
                        ),
                      ),
                    ),
                  ),
                ),

              // Botón circular fijo a la derecha (lupa <-> flecha volver)
              if (hasSearchAction)
                Positioned(
                  right: 0,
                  top: 1,
                  bottom: 1,
                  width: widget.searchButtonSize,
                  child: _MorphingSearchCircleButton(
                    progress: t,
                    isSearchOpen: widget.searchOpen,
                    color: widget.pillColor,
                    onPressed: _handleCircleButtonTap,
                  ),
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildNavigationPill() {
    return SizedBox(
      key: const ValueKey('nav-bar-pill-row'),
      width: widget.navigationWidth,
      child: MeloopPillTabs<DesktopNavigationSection>(
        items: const [
          MeloopPillItem(value: DesktopNavigationSection.home, label: 'Inicio'),
          MeloopPillItem(
            value: DesktopNavigationSection.friends,
            label: 'Amigos',
          ),
          MeloopPillItem(
            value: DesktopNavigationSection.artists,
            label: 'Artistas y canciones',
          ),
        ],
        selectedValue: widget.activeSection == DesktopNavigationSection.none
            ? null
            : widget.activeSection,
        onSelected: (section) {
          switch (section) {
            case DesktopNavigationSection.home:
              widget.onHomeTap();
            case DesktopNavigationSection.friends:
              widget.onFriendsTap();
            case DesktopNavigationSection.artists:
              widget.onArtistsTap();
            case DesktopNavigationSection.none:
              break;
          }
        },
        height: 50,
        padding: widget.outerPadding,
        fontSize: widget.fontSize,
        horizontalItemPadding: widget.horizontalItemPadding,
        fixedButtonWidth: DesktopAppHeader._navButtonWidth,
        pillColor: widget.pillColor,
        equalWidths: true,
      ),
    );
  }

  Widget _buildSearchBar() {
    final controller = widget.searchController;
    return Container(
      key: const ValueKey('nav-bar-search-row'),
      height: 46,
      alignment: Alignment.center,
      padding: const EdgeInsets.only(left: 18, right: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(26),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.10),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          IconButton(
            tooltip: 'Buscar',
            onPressed: () =>
                widget.onSearchSubmitted?.call(controller?.text ?? ''),
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints.tightFor(width: 30, height: 30),
            icon: const Icon(
              Icons.search_rounded,
              color: Color(0xFF16877F),
              size: 21,
            ),
          ),
          const SizedBox(width: 6),
          Expanded(
            child: TextField(
              controller: controller,
              focusNode: _searchFocusNode,
              onChanged: widget.onSearchChanged,
              onSubmitted: widget.onSearchSubmitted,
              textInputAction: TextInputAction.search,
              style: MeloopTypography.lexend(
                fontSize: 14.5,
                color: const Color(0xFF1A2424),
              ),
              decoration: InputDecoration(
                hintText:
                    'Buscar publicaciones, canciones o amigos (Enter)...',
                hintStyle: MeloopTypography.lexend(
                  fontSize: 13.5,
                  color: const Color(0xFF7D908E),
                ),
                border: InputBorder.none,
                isDense: true,
                contentPadding: const EdgeInsets.symmetric(vertical: 10),
              ),
            ),
          ),
          if (controller != null)
            ValueListenableBuilder<TextEditingValue>(
              valueListenable: controller,
              builder: (context, value, _) {
                if (value.text.isEmpty) return const SizedBox(width: 6);
                return IconButton(
                  tooltip: 'Limpiar búsqueda',
                  onPressed: controller.clear,
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints.tightFor(
                    width: 32,
                    height: 32,
                  ),
                  icon: const Icon(
                    Icons.close_rounded,
                    size: 18,
                    color: Color(0xFF6B807D),
                  ),
                );
              },
            ),
        ],
      ),
    );
  }
}

class _MorphingSearchCircleButton extends StatelessWidget {
  const _MorphingSearchCircleButton({
    required this.progress,
    required this.isSearchOpen,
    required this.color,
    required this.onPressed,
  });

  final double progress;
  final bool isSearchOpen;
  final Color color;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final searchOpacity = (1.0 - progress * 2.0).clamp(0.0, 1.0);
    final backOpacity = ((progress - 0.45) * 1.9).clamp(0.0, 1.0);

    return IconButton(
      tooltip: isSearchOpen ? 'Volver' : 'Búsqueda global',
      onPressed: onPressed,
      constraints: const BoxConstraints.tightFor(width: 48, height: 48),
      padding: EdgeInsets.zero,
      style: IconButton.styleFrom(
        backgroundColor: color,
        foregroundColor: Colors.white,
        elevation: 0,
        shadowColor: Colors.transparent,
        shape: const CircleBorder(),
      ),
      icon: Stack(
        alignment: Alignment.center,
        children: [
          Opacity(
            opacity: searchOpacity,
            child: Transform.rotate(
              angle: progress * math.pi * 0.35,
              child: const Icon(Icons.search_rounded, size: 25),
            ),
          ),
          Opacity(
            opacity: backOpacity,
            child: Transform.rotate(
              angle: (1.0 - progress) * -math.pi * 0.35,
              child: const Icon(Icons.arrow_back_rounded, size: 24),
            ),
          ),
        ],
      ),
    );
  }
}

class _HeaderMiniPlayer extends StatefulWidget {
  const _HeaderMiniPlayer({required this.width, this.compact = false});

  final double width;
  final bool compact;

  @override
  State<_HeaderMiniPlayer> createState() => _HeaderMiniPlayerState();
}

class _HeaderMiniPlayerState extends State<_HeaderMiniPlayer> {
  bool _localPlaying = false;

  @override
  Widget build(BuildContext context) {
    final playback = PostAudioPlaybackController.shared;
    return AnimatedBuilder(
      animation: playback,
      builder: (context, _) {
        final hasActiveAudio = playback.activePostId != null;
        final isPlaying = hasActiveAudio ? playback.isPlaying : _localPlaying;
        final position = hasActiveAudio
            ? playback.position
            : const Duration(minutes: 2, seconds: 36);
        final duration = hasActiveAudio && playback.duration != null
            ? playback.duration!
            : const Duration(minutes: 3, seconds: 57);
        final progress = duration.inMilliseconds > 0
            ? (position.inMilliseconds / duration.inMilliseconds).clamp(
                0.0,
                1.0,
              )
            : 0.0;

        return Container(
          width: widget.width,
          height: 58,
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [Color(0xFF167B74), Color(0xFF125E5B)],
            ),
            borderRadius: BorderRadius.circular(12),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.16),
                blurRadius: 8,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Row(
            children: [
              // Controles prev / play / next
              if (!widget.compact) ...[
                Icon(
                  Icons.skip_previous_rounded,
                  color: Colors.white.withValues(alpha: 0.8),
                  size: 20,
                ),
                const SizedBox(width: 4),
              ],
              InkWell(
                onTap: () => setState(() => _localPlaying = !_localPlaying),
                borderRadius: BorderRadius.circular(18),
                child: Container(
                  width: 32,
                  height: 32,
                  decoration: const BoxDecoration(
                    color: Colors.white,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded,
                    color: const Color(0xFF136862),
                    size: 21,
                  ),
                ),
              ),
              if (!widget.compact) ...[
                const SizedBox(width: 4),
                Icon(
                  Icons.skip_next_rounded,
                  color: Colors.white.withValues(alpha: 0.8),
                  size: 20,
                ),
              ],
              const SizedBox(width: 12),

              // Información de pista + barra de progreso en Roboto Condensed Regular
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text(
                                'Flashing Lights',
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: MeloopTypography.musicPlayer(
                                  fontSize: 13.5,
                                  color: Colors.white,
                                  height: 1.05,
                                ),
                              ),
                              Row(
                                children: [
                                  Flexible(
                                    child: Text(
                                      'Graduation',
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: MeloopTypography.musicPlayer(
                                        fontSize: 12,
                                        color: Colors.white.withValues(
                                          alpha: 0.88,
                                        ),
                                        height: 1.1,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Flexible(
                                    child: Text(
                                      'Kanye West',
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: MeloopTypography.musicPlayer(
                                        fontSize: 11.5,
                                        color: Colors.white.withValues(
                                          alpha: 0.62,
                                        ),
                                        height: 1.1,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                        if (!widget.compact)
                          Padding(
                            padding: const EdgeInsets.only(left: 8),
                            child: Text(
                              '${_formatMiniDuration(position)} / ${_formatMiniDuration(duration)}',
                              style: MeloopTypography.musicPlayer(
                                fontSize: 11.5,
                                color: Colors.white.withValues(alpha: 0.85),
                              ),
                            ),
                          ),
                      ],
                    ),
                    const SizedBox(height: 5),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(3),
                      child: SizedBox(
                        height: 3.5,
                        child: LinearProgressIndicator(
                          value: progress,
                          backgroundColor: Colors.white.withValues(alpha: 0.25),
                          valueColor: const AlwaysStoppedAnimation<Color>(
                            Colors.white,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(width: 10),
              Container(
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.16),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.graphic_eq_rounded,
                  color: Colors.white,
                  size: 15,
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  static String _formatMiniDuration(Duration d) {
    final m = d.inMinutes;
    final s = d.inSeconds.remainder(60).toString().padLeft(2, '0');
    return '$m:$s';
  }
}

class _HeaderProfileDropdown extends StatelessWidget {
  const _HeaderProfileDropdown({
    required this.onProfileTap,
    required this.onModerationTap,
    required this.onMusicPreferencesTap,
    required this.onBlockedUsersTap,
    required this.onAccountSettingsTap,
    required this.apiClient,
  });

  final VoidCallback? onProfileTap;
  final VoidCallback? onModerationTap;
  final VoidCallback onMusicPreferencesTap;
  final VoidCallback onBlockedUsersTap;
  final VoidCallback onAccountSettingsTap;
  final ApiClient? apiClient;

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<_ProfileMenuAction>(
      tooltip: 'Perfil y opciones',
      offset: const Offset(0, 52),
      color: Colors.white,
      elevation: 10,
      constraints: const BoxConstraints(minWidth: 250, maxWidth: 340),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      onSelected: (action) async {
        switch (action) {
          case _ProfileMenuAction.viewProfile:
            onProfileTap?.call();
          case _ProfileMenuAction.accountSettings:
            onAccountSettingsTap();
          case _ProfileMenuAction.musicPreferences:
            onMusicPreferencesTap();
          case _ProfileMenuAction.blockedUsers:
            onBlockedUsersTap();
          case _ProfileMenuAction.moderationQueue:
            onModerationTap?.call();
          case _ProfileMenuAction.logout:
            await SessionLogout.perform(context, apiClient: apiClient);
        }
      },
      itemBuilder: (context) => [
        PopupMenuItem(
          value: _ProfileMenuAction.viewProfile,
          child: Row(
            children: [
              const Icon(
                Icons.person_outline_rounded,
                color: Color(0xFF16877F),
                size: 20,
              ),
              const SizedBox(width: 12),
              Flexible(
                child: Text(
                  'Mirar Perfil',
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                    color: const Color(0xFF182222),
                  ),
                ),
              ),
            ],
          ),
        ),
        const PopupMenuDivider(height: 8),
        PopupMenuItem(
          value: _ProfileMenuAction.accountSettings,
          child: Row(
            children: [
              const Icon(
                Icons.manage_accounts_outlined,
                color: Color(0xFF16877F),
                size: 20,
              ),
              const SizedBox(width: 12),
              Flexible(
                child: Text(
                  'Configuración de cuenta',
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 14,
                    color: const Color(0xFF182222),
                  ),
                ),
              ),
            ],
          ),
        ),
        PopupMenuItem(
          value: _ProfileMenuAction.musicPreferences,
          child: Row(
            children: [
              const Icon(
                Icons.queue_music_rounded,
                color: Color(0xFF16877F),
                size: 20,
              ),
              const SizedBox(width: 12),
              Flexible(
                child: Text(
                  'Preferencias musicales',
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 14,
                    color: const Color(0xFF182222),
                  ),
                ),
              ),
            ],
          ),
        ),
        PopupMenuItem(
          value: _ProfileMenuAction.blockedUsers,
          child: Row(
            children: [
              const Icon(
                Icons.block_rounded,
                color: Color(0xFF16877F),
                size: 20,
              ),
              const SizedBox(width: 12),
              Flexible(
                child: Text(
                  'Usuarios bloqueados',
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 14,
                    color: const Color(0xFF182222),
                  ),
                ),
              ),
            ],
          ),
        ),
        if (onModerationTap != null) ...[
          const PopupMenuDivider(height: 8),
          PopupMenuItem(
            value: _ProfileMenuAction.moderationQueue,
            child: Row(
              children: [
                const Icon(
                  Icons.shield_outlined,
                  color: Color(0xFF16877F),
                  size: 20,
                ),
                const SizedBox(width: 12),
                Flexible(
                  child: Text(
                    'Cola de moderación',
                    overflow: TextOverflow.ellipsis,
                    style: MeloopTypography.lexend(
                      fontSize: 14,
                      color: const Color(0xFF182222),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
        const PopupMenuDivider(height: 8),
        PopupMenuItem(
          value: _ProfileMenuAction.logout,
          child: Row(
            children: [
              const Icon(Icons.logout_rounded, color: Colors.red, size: 20),
              const SizedBox(width: 12),
              Flexible(
                child: Text(
                  'Cerrar sesión',
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 14,
                    color: Colors.red,
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
      child: Container(
        width: 44,
        height: 44,
        padding: const EdgeInsets.all(2.5),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(6),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.16),
              blurRadius: 6,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: const MeloopImagePlaceholder(
          iconSize: 20,
          borderRadius: BorderRadius.all(Radius.circular(4)),
          backgroundColor: Color(0xFFE7EFEE),
          iconColor: Color(0xFF16877F),
          borderColor: null,
        ),
      ),
    );
  }
}
