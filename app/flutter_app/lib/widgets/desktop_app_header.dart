import 'package:flutter/material.dart';

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
  });

  final DesktopNavigationSection activeSection;
  final VoidCallback onHomeTap;
  final VoidCallback onFriendsTap;
  final VoidCallback onArtistsTap;
  final VoidCallback? onProfileTap;
  final VoidCallback? onModerationTap;

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
                  child: Container(
                    width: navigationWidth,
                    height: 54,
                    padding: const EdgeInsets.all(navigationPadding),
                    decoration: BoxDecoration(
                      color: _pillColor,
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
                        _navigationButton(
                          'Inicio',
                          DesktopNavigationSection.home,
                          onHomeTap,
                          buttonWidth,
                        ),
                        _navigationButton(
                          'Amigos',
                          DesktopNavigationSection.friends,
                          onFriendsTap,
                          buttonWidth,
                        ),
                        _navigationButton(
                          'Artistas y canciones',
                          DesktopNavigationSection.artists,
                          onArtistsTap,
                          buttonWidth,
                        ),
                      ],
                    ),
                  ),
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
