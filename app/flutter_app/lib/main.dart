import 'package:flutter/material.dart';

import 'screens/auth_screen.dart';

void main() {
  runApp(const MeloopApp());
}

class MeloopApp extends StatelessWidget {
  const MeloopApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Meloop',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        primaryColor: const Color(0xFF11B49F),
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
