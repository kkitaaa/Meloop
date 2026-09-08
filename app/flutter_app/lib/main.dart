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
      ),
      home: const AuthScreen(),
    );
  }
}