import 'package:flutter/material.dart';

abstract final class SettingsColors {
  static const background = Color(0xFF0D5C5E);
  static const accent = Color(0xFF1ABC9C);
  static const panel = Color(0xFFF5F8F7);
  static const text = Color(0xFF193A3A);
  static const border = Color(0xFFDCE8E5);
}

class SettingsPageShell extends StatelessWidget {
  const SettingsPageShell({
    super.key,
    required this.title,
    required this.icon,
    required this.child,
    this.maxWidth = 920,
  });

  final String title;
  final IconData icon;
  final Widget child;
  final double maxWidth;

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: SettingsColors.background,
    appBar: PreferredSize(
      preferredSize: const Size.fromHeight(56),
      child: Container(
        color: SettingsColors.accent,
        padding: const EdgeInsets.symmetric(horizontal: 20),
        child: Row(
          children: [
            IconButton(
              tooltip: 'Volver',
              onPressed: () => Navigator.maybePop(context),
              icon: const Icon(Icons.arrow_back_rounded, color: Colors.white),
            ),
            const SizedBox(width: 8),
            Image.asset(
              'assets/imagenes/meloop.png',
              height: 26,
              fit: BoxFit.contain,
            ),
            const SizedBox(width: 18),
            Expanded(
              child: Text(
                title,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
            Icon(icon, color: Colors.white, size: 20),
          ],
        ),
      ),
    ),
    body: Center(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: ConstrainedBox(
          constraints: BoxConstraints(maxWidth: maxWidth),
          child: Container(
            width: double.infinity,
            height: double.infinity,
            decoration: BoxDecoration(
              color: SettingsColors.panel,
              borderRadius: BorderRadius.circular(16),
            ),
            clipBehavior: Clip.antiAlias,
            child: child,
          ),
        ),
      ),
    ),
  );
}

class SettingsSectionCard extends StatelessWidget {
  const SettingsSectionCard({
    super.key,
    required this.title,
    required this.icon,
    required this.child,
  });

  final String title;
  final IconData icon;
  final Widget child;

  @override
  Widget build(BuildContext context) => Card(
    color: Colors.white,
    elevation: 0,
    margin: const EdgeInsets.only(bottom: 14),
    shape: RoundedRectangleBorder(
      side: const BorderSide(color: SettingsColors.border),
      borderRadius: BorderRadius.circular(14),
    ),
    child: Padding(
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: SettingsColors.accent, size: 21),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  title,
                  style: const TextStyle(
                    color: SettingsColors.text,
                    fontSize: 16,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          child,
        ],
      ),
    ),
  );
}
