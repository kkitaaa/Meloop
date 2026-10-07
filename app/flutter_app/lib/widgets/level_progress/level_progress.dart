import 'package:flutter/material.dart';

/// Widget reutilizable para mostrar nivel, experiencia y progreso.
class LevelProgress extends StatelessWidget {
  const LevelProgress({
    super.key,
    required this.level,
    required this.xp,
    required this.xpNeeded,
    this.compact = false,
    this.accentColor = const Color(0xFF00AF92),
  });

  final int level;
  final int xp;
  final int xpNeeded;
  final bool compact;
  final Color accentColor;

  @override
  Widget build(BuildContext context) {
    final progress = xpNeeded == 0 ? 0.0 : (xp / xpNeeded).clamp(0, 1).toDouble();
    final percentage = (progress * 100).round();
    return LayoutBuilder(builder: (context, constraints) {
      final small = compact || constraints.maxWidth < 270;
      return Container(
        padding: EdgeInsets.all(small ? 12 : 18),
        decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(10), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: .05), blurRadius: 8)]),
        child: Row(children: [
          Container(width: small ? 30 : 44, height: small ? 30 : 44, alignment: Alignment.center, decoration: BoxDecoration(color: accentColor.withValues(alpha: .12), shape: BoxShape.circle), child: Icon(Icons.star_rounded, color: accentColor, size: small ? 18 : 26)),
          SizedBox(width: small ? 9 : 14),
          Expanded(child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [Text('Nivel $level', style: TextStyle(fontSize: small ? 12 : 16, fontWeight: FontWeight.bold)), const Spacer(), Text('$xp / $xpNeeded XP', style: TextStyle(color: Colors.black54, fontSize: small ? 10 : 12))]),
            SizedBox(height: small ? 7 : 10),
            TweenAnimationBuilder<double>(
              tween: Tween(end: progress), duration: const Duration(milliseconds: 420), curve: Curves.easeOutCubic,
              builder: (context, value, _) => Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                ClipRRect(borderRadius: BorderRadius.circular(6), child: LinearProgressIndicator(value: value, minHeight: small ? 6 : 9, backgroundColor: Colors.grey[200], valueColor: AlwaysStoppedAnimation(accentColor))),
                if (!small) ...[const SizedBox(height: 5), Text('$percentage% hacia el siguiente nivel', style: const TextStyle(color: Colors.black45, fontSize: 11))],
              ]),
            ),
          ])),
        ]),
      );
    });
  }
}
