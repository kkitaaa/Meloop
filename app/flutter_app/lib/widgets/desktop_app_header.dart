import 'package:flutter/material.dart';

import '../services/api_client.dart';
import 'app_settings_menu.dart';
import 'notifications_menu.dart';

enum DesktopNavigationSection { home, friends, artists, none }

void navigateToDesktopSection(BuildContext context, WidgetBuilder builder) {
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
  Navigator.of(context).popUntil((route) => route.isFirst);
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
    this.apiClient,
    this.searchOpen = false,
    this.searchController,
    this.onSearchChanged,
    this.onSearchBack,
  });

  final DesktopNavigationSection activeSection;
  final VoidCallback onHomeTap;
  final VoidCallback onFriendsTap;
  final VoidCallback onArtistsTap;
  final VoidCallback? onProfileTap;
  final VoidCallback? onModerationTap;
  final VoidCallback? onSearchTap;
  final ApiClient? apiClient;
  final bool searchOpen;
  final TextEditingController? searchController;
  final ValueChanged<String>? onSearchChanged;
  final VoidCallback? onSearchBack;

  static const _accentColor = Color(0xFF1ABC9C);
  static const _headerGradient = LinearGradient(
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
    colors: [Color(0xFF59E5C9), Color(0xFF00AF92)],
  );
  static const _pillColor = Color(0xFF16877F);
  static const _navigationLabels = ['Inicio', 'Amigos', 'Artistas y canciones'];

  @override
  Widget build(BuildContext context) {
    const navigationTextStyle = TextStyle(
      fontSize: 14,
      fontWeight: FontWeight.w600,
    );
    final longestLabelWidth = _navigationLabels
        .map(
          (label) => TextPainter(
            text: TextSpan(text: label, style: navigationTextStyle),
            textDirection: Directionality.of(context),
            textScaler: MediaQuery.textScalerOf(context),
          )..layout(),
        )
        .map((painter) => painter.width)
        .reduce((first, second) => first > second ? first : second);
    final buttonWidth = longestLabelWidth + 20;
    const navigationPadding = 5.0;
    final navigationWidth =
        buttonWidth * _navigationLabels.length + navigationPadding * 2;

    return Container(
      height: 82,
      decoration: const BoxDecoration(
        gradient: _headerGradient,
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(24)),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 8),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final availableNavigationWidth = constraints.maxWidth - 32;
          final navScale = availableNavigationWidth < navigationWidth
              ? availableNavigationWidth / navigationWidth
              : 1.0;
          return Stack(
            alignment: Alignment.center,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Image.asset(
                    'assets/imagenes/meloop.png',
                    height: 38,
                    fit: BoxFit.contain,
                  ),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 8,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.2),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: const Row(
                          children: [
                            Icon(
                              Icons.play_circle_fill,
                              color: Colors.white,
                              size: 18,
                            ),
                            SizedBox(width: 8),
                            Text(
                              'Playing... DJ Gouz',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 12,
                              ),
                            ),
                            SizedBox(width: 8),
                            Icon(
                              Icons.graphic_eq,
                              color: Colors.white,
                              size: 16,
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      if (onModerationTap != null)
                        IconButton(
                          tooltip: 'Cola de moderación',
                          onPressed: onModerationTap,
                          icon: const Icon(
                            Icons.shield_outlined,
                            color: Colors.white,
                          ),
                        ),
                      AppSettingsMenu(apiClient: apiClient),
                      const SizedBox(width: 4),
                      const NotificationsMenu(),
                      const SizedBox(width: 8),
                      InkWell(
                        onTap: onProfileTap,
                        borderRadius: BorderRadius.circular(20),
                        child: const CircleAvatar(
                          radius: 16,
                          backgroundColor: Colors.white,
                          child: Icon(
                            Icons.person,
                            color: _accentColor,
                            size: 20,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              Center(
                child: Transform.scale(
                  scale: navScale,
                  child: _DesktopHeaderCenter(
                    activeSection: activeSection,
                    buttonWidth: buttonWidth,
                    navigationPadding: navigationPadding,
                    navigationWidth: navigationWidth,
                    onHomeTap: onHomeTap,
                    onFriendsTap: onFriendsTap,
                    onArtistsTap: onArtistsTap,
                    searchOpen: searchOpen,
                    searchController: searchController,
                    onSearchChanged: onSearchChanged,
                    onSearchBack: onSearchBack,
                    navigationBuilder: _navigationButton,
                    pillColor: _pillColor,
                  ),
                ),
              ),
              if (!searchOpen && onSearchTap != null)
                Positioned(
                  left:
                      constraints.maxWidth / 2 +
                      navigationWidth * navScale / 2 +
                      12,
                  child: _HeaderCircleButton(
                    tooltip: 'Búsqueda global',
                    onPressed: onSearchTap,
                    icon: Icons.search,
                    color: _pillColor,
                  ),
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _navigationButton(
    String label,
    DesktopNavigationSection section,
    VoidCallback onPressed,
    double width,
  ) {
    final isSelected = activeSection == section;
    return SizedBox(
      width: width,
      child: Center(
        child: SizedBox(
          height: 42,
          child: Semantics(
            button: true,
            selected: isSelected,
            label: label,
            child: Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: onPressed,
                borderRadius: BorderRadius.circular(24),
                child: Container(
                  width: width,
                  height: 42,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: isSelected ? Colors.white : Colors.transparent,
                    gradient: isSelected
                        ? const LinearGradient(
                            begin: Alignment.topCenter,
                            end: Alignment.bottomCenter,
                            colors: [Color(0xFFFFFFFF), Color(0xFFE6E6E6)],
                          )
                        : null,
                    borderRadius: BorderRadius.circular(24),
                  ),
                  child: Text(
                    label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: isSelected ? Colors.black : Colors.white,
                      fontSize: 14,
                      fontWeight: isSelected
                          ? FontWeight.w600
                          : FontWeight.w400,
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _DesktopHeaderCenter extends StatefulWidget {
  const _DesktopHeaderCenter({
    required this.activeSection,
    required this.buttonWidth,
    required this.navigationPadding,
    required this.navigationWidth,
    required this.onHomeTap,
    required this.onFriendsTap,
    required this.onArtistsTap,
    required this.searchOpen,
    required this.navigationBuilder,
    required this.pillColor,
    this.searchController,
    this.onSearchChanged,
    this.onSearchBack,
  });

  final DesktopNavigationSection activeSection;
  final double buttonWidth;
  final double navigationPadding;
  final double navigationWidth;
  final VoidCallback onHomeTap;
  final VoidCallback onFriendsTap;
  final VoidCallback onArtistsTap;
  final bool searchOpen;
  final TextEditingController? searchController;
  final ValueChanged<String>? onSearchChanged;
  final VoidCallback? onSearchBack;
  final Widget Function(String, DesktopNavigationSection, VoidCallback, double)
  navigationBuilder;
  final Color pillColor;

  @override
  State<_DesktopHeaderCenter> createState() => _DesktopHeaderCenterState();
}

class _DesktopHeaderCenterState extends State<_DesktopHeaderCenter> {
  late bool _searchOpen;

  @override
  void initState() {
    super.initState();
    _searchOpen = false;
    if (widget.searchOpen) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) setState(() => _searchOpen = true);
      });
    }
  }

  @override
  void didUpdateWidget(covariant _DesktopHeaderCenter oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.searchOpen != widget.searchOpen && widget.searchOpen) {
      setState(() => _searchOpen = true);
    }
  }

  void _closeSearch() {
    if (!_searchOpen) {
      widget.onSearchBack?.call();
      return;
    }

    setState(() => _searchOpen = false);
    Future<void>.delayed(const Duration(milliseconds: 280), () {
      if (mounted && !_searchOpen) widget.onSearchBack?.call();
    });
  }

  @override
  Widget build(BuildContext context) {
    final width = widget.navigationWidth + 64;
    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 280),
      switchInCurve: Curves.easeOutCubic,
      switchOutCurve: Curves.easeInCubic,
      transitionBuilder: (child, animation) => SizeTransition(
        axis: Axis.horizontal,
        alignment: Alignment.centerRight,
        sizeFactor: animation,
        child: FadeTransition(opacity: animation, child: child),
      ),
      child: SizedBox(
        key: ValueKey(_searchOpen),
        width: width,
        height: 66,
        child: _searchOpen ? _buildSearchBar() : _buildNavigation(),
      ),
    );
  }

  Widget _buildNavigation() {
    return Stack(
      alignment: Alignment.center,
      children: [
        SizedBox(
          width: widget.navigationWidth,
          child: Container(
            height: 54,
            padding: EdgeInsets.all(widget.navigationPadding),
            decoration: BoxDecoration(
              color: widget.pillColor,
              borderRadius: BorderRadius.circular(36),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.15),
                  blurRadius: 8,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Row(
              children: [
                widget.navigationBuilder(
                  'Inicio',
                  DesktopNavigationSection.home,
                  widget.onHomeTap,
                  widget.buttonWidth,
                ),
                widget.navigationBuilder(
                  'Amigos',
                  DesktopNavigationSection.friends,
                  widget.onFriendsTap,
                  widget.buttonWidth,
                ),
                widget.navigationBuilder(
                  'Artistas y canciones',
                  DesktopNavigationSection.artists,
                  widget.onArtistsTap,
                  widget.buttonWidth,
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSearchBar() {
    return Row(
      children: [
        Expanded(
          child: Container(
            height: 44,
            alignment: Alignment.center,
            padding: const EdgeInsets.symmetric(horizontal: 18),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
            ),
            child: TextField(
              controller: widget.searchController,
              onChanged: widget.onSearchChanged,
              autofocus: true,
              textInputAction: TextInputAction.search,
              decoration: const InputDecoration(
                hintText: 'Buscar en Meloop',
                border: InputBorder.none,
                isDense: true,
              ),
            ),
          ),
        ),
        _HeaderCircleButton(
          tooltip: 'Volver',
          onPressed: _closeSearch,
          icon: Icons.arrow_back,
          color: widget.pillColor,
        ),
      ],
    );
  }
}

class _HeaderCircleButton extends StatelessWidget {
  const _HeaderCircleButton({
    required this.tooltip,
    required this.onPressed,
    required this.icon,
    required this.color,
  });

  final String tooltip;
  final VoidCallback? onPressed;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return IconButton(
      tooltip: tooltip,
      onPressed: onPressed,
      constraints: const BoxConstraints.tightFor(width: 48, height: 48),
      padding: EdgeInsets.zero,
      style: IconButton.styleFrom(
        backgroundColor: color,
        foregroundColor: Colors.white,
        shadowColor: Colors.black.withValues(alpha: 0.15),
        elevation: 4,
        shape: const CircleBorder(),
      ),
      icon: Icon(icon, size: 30),
    );
  }
}
