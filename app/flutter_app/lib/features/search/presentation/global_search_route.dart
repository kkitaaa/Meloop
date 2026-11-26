import 'package:flutter/material.dart';

import '../../../core/navigation/desktop_navigation_shell.dart';
import '../../../core/navigation/responsive_navigation_shell.dart';
import 'global_search_screen.dart';

void openGlobalSearch(BuildContext context) {
  final shell = DesktopShellScope.maybeOf(context);
  if (shell != null) {
    shell.openSearch();
    return;
  }
  final mobileShell = MobileShellScope.maybeOf(context);
  Navigator.of(context).push<void>(
    PageRouteBuilder<void>(
      settings: const RouteSettings(name: '/search'),
      pageBuilder: (context, animation, secondaryAnimation) =>
          GlobalSearchScreen(mobileShellController: mobileShell),
      transitionDuration: const Duration(milliseconds: 320),
      reverseTransitionDuration: const Duration(milliseconds: 260),
      transitionsBuilder: (context, animation, secondaryAnimation, child) {
        final fadeIn = CurvedAnimation(
          parent: animation,
          curve: const Interval(0.42, 1.0, curve: Curves.easeOutCubic),
          reverseCurve: const Interval(0.48, 1.0, curve: Curves.easeInCubic),
        );
        final slide = Tween<Offset>(
          begin: const Offset(0, 0.025),
          end: Offset.zero,
        ).animate(fadeIn);
        return FadeTransition(
          opacity: fadeIn,
          child: SlideTransition(position: slide, child: child),
        );
      },
    ),
  );
}
