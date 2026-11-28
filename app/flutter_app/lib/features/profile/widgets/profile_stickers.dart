import 'dart:math' as math;

import 'package:flutter/material.dart';

class ProfileStickerDefinition {
  const ProfileStickerDefinition({
    required this.id,
    required this.label,
    required this.requiredLevel,
    required this.accentColor,
  });

  final String id;
  final String label;
  final int requiredLevel;
  final Color accentColor;
}

const List<ProfileStickerDefinition> kProfileStickerCatalog = [
  ProfileStickerDefinition(
    id: 'cat_shy',
    label: 'Gatito curioso',
    requiredLevel: 1,
    accentColor: Color(0xFF9EB0BA),
  ),
  ProfileStickerDefinition(
    id: 'music_note',
    label: 'Nota Meloop',
    requiredLevel: 1,
    accentColor: Color(0xFF00AF92),
  ),
  ProfileStickerDefinition(
    id: 'cloud_blush',
    label: 'Nube kawaii',
    requiredLevel: 2,
    accentColor: Color(0xFFFF8EC4),
  ),
  ProfileStickerDefinition(
    id: 'heart_hug',
    label: 'Abrazo corazón',
    requiredLevel: 3,
    accentColor: Color(0xFFFFC233),
  ),
  ProfileStickerDefinition(
    id: 'vinyl_spin',
    label: 'Vinilo retro',
    requiredLevel: 4,
    accentColor: Color(0xFFE05A7A),
  ),
  ProfileStickerDefinition(
    id: 'star_sparkle',
    label: 'Estrella pop',
    requiredLevel: 5,
    accentColor: Color(0xFFF6B73C),
  ),
  ProfileStickerDefinition(
    id: 'headphones',
    label: 'Audífonos',
    requiredLevel: 6,
    accentColor: Color(0xFF39B7E8),
  ),
  ProfileStickerDefinition(
    id: 'paw_print',
    label: 'Huellita',
    requiredLevel: 8,
    accentColor: Color(0xFFF48FB1),
  ),
  ProfileStickerDefinition(
    id: 'aero_bubble',
    label: 'Burbuja Aero',
    requiredLevel: 10,
    accentColor: Color(0xFF33C7F4),
  ),
  ProfileStickerDefinition(
    id: 'cassette',
    label: 'Casete vintage',
    requiredLevel: 12,
    accentColor: Color(0xFF8B6CE8),
  ),
];

ProfileStickerDefinition stickerDefinitionById(String id) {
  for (final item in kProfileStickerCatalog) {
    if (item.id == id) return item;
  }
  return kProfileStickerCatalog.first;
}

/// Renderiza stickers vectoriales con borde blanco estilo calcomanía recortada
/// (die-cut sticker) para que luzcan nítidos sobre cualquier papel o tarjeta.
class ProfileStickerArtwork extends StatelessWidget {
  const ProfileStickerArtwork({
    super.key,
    required this.stickerId,
    this.size = 76,
  });

  final String stickerId;
  final double size;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: size,
      height: size,
      child: CustomPaint(painter: _StickerVectorPainter(stickerId: stickerId)),
    );
  }
}

class _StickerVectorPainter extends CustomPainter {
  const _StickerVectorPainter({required this.stickerId});

  final String stickerId;

  @override
  void paint(Canvas canvas, Size size) {
    final s = math.min(size.width, size.height);
    final cx = size.width / 2;
    final cy = size.height / 2;

    // Sombra suave estilo sticker pegado encima de la página
    final shadowPaint = Paint()
      ..color = Colors.black.withValues(alpha: 0.16)
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 5);
    canvas.drawCircle(Offset(cx, cy + s * 0.05), s * 0.37, shadowPaint);

    // Borde blanco die-cut
    canvas.drawCircle(Offset(cx, cy), s * 0.42, Paint()..color = Colors.white);

    switch (stickerId) {
      case 'cat_shy':
        _paintCat(canvas, cx, cy, s);
      case 'cloud_blush':
        _paintCloudBlush(canvas, cx, cy, s);
      case 'heart_hug':
        _paintHeartHug(canvas, cx, cy, s);
      case 'vinyl_spin':
        _paintVinyl(canvas, cx, cy, s);
      case 'star_sparkle':
        _paintStar(canvas, cx, cy, s);
      case 'headphones':
        _paintHeadphones(canvas, cx, cy, s);
      case 'paw_print':
        _paintPaw(canvas, cx, cy, s);
      case 'aero_bubble':
        _paintAeroBubble(canvas, cx, cy, s);
      case 'cassette':
        _paintCassette(canvas, cx, cy, s);
      case 'music_note':
      default:
        _paintMusicNote(canvas, cx, cy, s);
    }
  }

  void _paintCat(Canvas canvas, double cx, double cy, double s) {
    final outline = Paint()
      ..color = const Color(0xFF3E4B55)
      ..style = PaintingStyle.stroke
      ..strokeWidth = s * 0.036
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final body = Paint()..color = const Color(0xFFB5C4CD);

    // Orejas
    final leftEar = Path()
      ..moveTo(cx - s * 0.25, cy - s * 0.10)
      ..lineTo(cx - s * 0.29, cy - s * 0.32)
      ..lineTo(cx - s * 0.08, cy - s * 0.22)
      ..close();
    final rightEar = Path()
      ..moveTo(cx + s * 0.25, cy - s * 0.10)
      ..lineTo(cx + s * 0.29, cy - s * 0.32)
      ..lineTo(cx + s * 0.08, cy - s * 0.22)
      ..close();
    canvas.drawPath(leftEar, body);
    canvas.drawPath(rightEar, body);
    canvas.drawPath(leftEar, outline);
    canvas.drawPath(rightEar, outline);

    // Cabeza
    final headRect = Rect.fromCenter(
      center: Offset(cx, cy + s * 0.02),
      width: s * 0.56,
      height: s * 0.48,
    );
    canvas.drawOval(headRect, body);
    canvas.drawOval(headRect, outline);

    // Rayitas en la frente
    for (final dx in [-0.07, 0.0, 0.07]) {
      canvas.drawLine(
        Offset(cx + s * dx, cy - s * 0.20),
        Offset(cx + s * dx, cy - s * 0.12),
        outline..strokeWidth = s * 0.026,
      );
    }

    // Ojos y chapitas
    final eyePaint = Paint()..color = const Color(0xFF2B353D);
    canvas.drawCircle(
      Offset(cx - s * 0.11, cy - s * 0.01),
      s * 0.036,
      eyePaint,
    );
    canvas.drawCircle(
      Offset(cx + s * 0.11, cy - s * 0.01),
      s * 0.036,
      eyePaint,
    );
    final blush = Paint()
      ..color = const Color(0xFFFF9DBB).withValues(alpha: 0.7);
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx - s * 0.17, cy + s * 0.06),
        width: s * 0.10,
        height: s * 0.055,
      ),
      blush,
    );
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx + s * 0.17, cy + s * 0.06),
        width: s * 0.10,
        height: s * 0.055,
      ),
      blush,
    );

    // Patitas juntas al frente
    final pawFill = Paint()..color = const Color(0xFFDCE6EC);
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx - s * 0.06, cy + s * 0.21),
        width: s * 0.13,
        height: s * 0.12,
      ),
      pawFill,
    );
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx + s * 0.06, cy + s * 0.21),
        width: s * 0.13,
        height: s * 0.12,
      ),
      pawFill,
    );
  }

  void _paintCloudBlush(Canvas canvas, double cx, double cy, double s) {
    final fill = Paint()..color = const Color(0xFFFFFBFD);
    final outline = Paint()
      ..color = const Color(0xFF473A46)
      ..style = PaintingStyle.stroke
      ..strokeWidth = s * 0.034;

    final bodyRect = RRect.fromRectAndRadius(
      Rect.fromCenter(
        center: Offset(cx, cy + s * 0.04),
        width: s * 0.56,
        height: s * 0.40,
      ),
      Radius.circular(s * 0.20),
    );
    canvas.drawRRect(bodyRect, fill);
    canvas.drawRRect(bodyRect, outline);

    // Corazones flotando arriba
    _drawMiniHeart(
      canvas,
      Offset(cx - s * 0.22, cy - s * 0.22),
      s * 0.12,
      const Color(0xFFFF7AA8),
    );
    _drawMiniHeart(
      canvas,
      Offset(cx + s * 0.19, cy - s * 0.25),
      s * 0.10,
      const Color(0xFFFF5C94),
    );

    // Carita kawaii
    final dark = Paint()..color = const Color(0xFF3E303D);
    canvas.drawCircle(Offset(cx - s * 0.10, cy + s * 0.02), s * 0.032, dark);
    canvas.drawCircle(Offset(cx + s * 0.10, cy + s * 0.02), s * 0.032, dark);
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx - s * 0.17, cy + s * 0.09),
        width: s * 0.09,
        height: s * 0.05,
      ),
      Paint()..color = const Color(0xFFFFA5C7),
    );
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx + s * 0.17, cy + s * 0.09),
        width: s * 0.09,
        height: s * 0.05,
      ),
      Paint()..color = const Color(0xFFFFA5C7),
    );
    final smile = Path()
      ..moveTo(cx - s * 0.04, cy + s * 0.07)
      ..quadraticBezierTo(cx, cy + s * 0.12, cx + s * 0.04, cy + s * 0.07);
    canvas.drawPath(smile, outline..strokeWidth = s * 0.028);
  }

  void _paintHeartHug(Canvas canvas, double cx, double cy, double s) {
    final yellow = Paint()..color = const Color(0xFFFFD54F);
    final outline = Paint()
      ..color = const Color(0xFF4A3114)
      ..style = PaintingStyle.stroke
      ..strokeWidth = s * 0.034;

    canvas.drawCircle(Offset(cx, cy - s * 0.02), s * 0.30, yellow);
    canvas.drawCircle(Offset(cx, cy - s * 0.02), s * 0.30, outline);

    // Ojitos cerrados felices (^ ^)
    final arcPaint = Paint()
      ..color = const Color(0xFF4A3114)
      ..style = PaintingStyle.stroke
      ..strokeWidth = s * 0.032
      ..strokeCap = StrokeCap.round;
    canvas.drawArc(
      Rect.fromCenter(
        center: Offset(cx - s * 0.10, cy - s * 0.08),
        width: s * 0.11,
        height: s * 0.09,
      ),
      math.pi,
      math.pi,
      false,
      arcPaint,
    );
    canvas.drawArc(
      Rect.fromCenter(
        center: Offset(cx + s * 0.10, cy - s * 0.08),
        width: s * 0.11,
        height: s * 0.09,
      ),
      math.pi,
      math.pi,
      false,
      arcPaint,
    );

    // Corazón abrazado al centro-abajo
    _drawMiniHeart(
      canvas,
      Offset(cx, cy + s * 0.14),
      s * 0.22,
      const Color(0xFFFF4D7E),
    );
  }

  void _paintVinyl(Canvas canvas, double cx, double cy, double s) {
    canvas.drawCircle(
      Offset(cx, cy),
      s * 0.34,
      Paint()..color = const Color(0xFF1C2124),
    );
    final groove = Paint()
      ..color = Colors.white.withValues(alpha: 0.14)
      ..style = PaintingStyle.stroke
      ..strokeWidth = s * 0.015;
    canvas.drawCircle(Offset(cx, cy), s * 0.27, groove);
    canvas.drawCircle(Offset(cx, cy), s * 0.21, groove);
    canvas.drawCircle(
      Offset(cx, cy),
      s * 0.13,
      Paint()..color = const Color(0xFFFF6699),
    );
    canvas.drawCircle(Offset(cx, cy), s * 0.035, Paint()..color = Colors.white);
  }

  void _paintStar(Canvas canvas, double cx, double cy, double s) {
    final path = Path();
    const points = 5;
    final outer = s * 0.34;
    final inner = s * 0.16;
    for (var i = 0; i < points * 2; i++) {
      final r = i.isEven ? outer : inner;
      final angle = -math.pi / 2 + i * math.pi / points;
      final pt = Offset(cx + math.cos(angle) * r, cy + math.sin(angle) * r);
      if (i == 0) {
        path.moveTo(pt.dx, pt.dy);
      } else {
        path.lineTo(pt.dx, pt.dy);
      }
    }
    path.close();
    canvas.drawPath(path, Paint()..color = const Color(0xFFFFCA3A));
    canvas.drawPath(
      path,
      Paint()
        ..color = const Color(0xFF7A4B00)
        ..style = PaintingStyle.stroke
        ..strokeWidth = s * 0.03
        ..strokeJoin = StrokeJoin.round,
    );
  }

  void _paintHeadphones(Canvas canvas, double cx, double cy, double s) {
    final band = Paint()
      ..color = const Color(0xFF1D9BF0)
      ..style = PaintingStyle.stroke
      ..strokeWidth = s * 0.07
      ..strokeCap = StrokeCap.round;
    canvas.drawArc(
      Rect.fromCenter(
        center: Offset(cx, cy + s * 0.04),
        width: s * 0.50,
        height: s * 0.52,
      ),
      math.pi,
      math.pi,
      false,
      band,
    );
    final cupPaint = Paint()..color = const Color(0xFF125B9A);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromCenter(
          center: Offset(cx - s * 0.24, cy + s * 0.11),
          width: s * 0.14,
          height: s * 0.24,
        ),
        Radius.circular(s * 0.06),
      ),
      cupPaint,
    );
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromCenter(
          center: Offset(cx + s * 0.24, cy + s * 0.11),
          width: s * 0.14,
          height: s * 0.24,
        ),
        Radius.circular(s * 0.06),
      ),
      cupPaint,
    );
  }

  void _paintPaw(Canvas canvas, double cx, double cy, double s) {
    final pink = Paint()..color = const Color(0xFFFF85B3);
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx, cy + s * 0.11),
        width: s * 0.34,
        height: s * 0.25,
      ),
      pink,
    );
    for (final offset in [
      Offset(-0.20, -0.06),
      Offset(-0.07, -0.17),
      Offset(0.07, -0.17),
      Offset(0.20, -0.06),
    ]) {
      canvas.drawCircle(
        Offset(cx + s * offset.dx, cy + s * offset.dy),
        s * 0.075,
        pink,
      );
    }
  }

  void _paintAeroBubble(Canvas canvas, double cx, double cy, double s) {
    final rect = Rect.fromCircle(center: Offset(cx, cy), radius: s * 0.34);
    final gradient = const RadialGradient(
      center: Alignment(-0.35, -0.35),
      radius: 0.9,
      colors: [Color(0xFFE6FAFF), Color(0xFF63D5FF), Color(0xFF0E88D4)],
      stops: [0.0, 0.55, 1.0],
    );
    canvas.drawCircle(
      Offset(cx, cy),
      s * 0.34,
      Paint()..shader = gradient.createShader(rect),
    );
    // Reflejo especular típico de Frutiger Aero
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx - s * 0.11, cy - s * 0.14),
        width: s * 0.22,
        height: s * 0.12,
      ),
      Paint()..color = Colors.white.withValues(alpha: 0.72),
    );
    canvas.drawCircle(
      Offset(cx + s * 0.15, cy + s * 0.14),
      s * 0.05,
      Paint()..color = Colors.white.withValues(alpha: 0.55),
    );
  }

  void _paintCassette(Canvas canvas, double cx, double cy, double s) {
    final body = RRect.fromRectAndRadius(
      Rect.fromCenter(
        center: Offset(cx, cy),
        width: s * 0.62,
        height: s * 0.42,
      ),
      Radius.circular(s * 0.06),
    );
    canvas.drawRRect(body, Paint()..color = const Color(0xFF7E5CE6));
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromCenter(
          center: Offset(cx, cy - s * 0.05),
          width: s * 0.50,
          height: s * 0.20,
        ),
        Radius.circular(s * 0.03),
      ),
      Paint()..color = const Color(0xFFFFF6D6),
    );
    canvas.drawCircle(
      Offset(cx - s * 0.11, cy - s * 0.04),
      s * 0.055,
      Paint()..color = const Color(0xFF2B2440),
    );
    canvas.drawCircle(
      Offset(cx + s * 0.11, cy - s * 0.04),
      s * 0.055,
      Paint()..color = const Color(0xFF2B2440),
    );
  }

  void _paintMusicNote(Canvas canvas, double cx, double cy, double s) {
    final teal = Paint()..color = const Color(0xFF00AF92);
    final stem = Paint()
      ..color = const Color(0xFF00AF92)
      ..strokeWidth = s * 0.065
      ..strokeCap = StrokeCap.round;
    canvas.drawLine(
      Offset(cx - s * 0.10, cy + s * 0.15),
      Offset(cx - s * 0.10, cy - s * 0.18),
      stem,
    );
    canvas.drawLine(
      Offset(cx + s * 0.16, cy + s * 0.10),
      Offset(cx + s * 0.16, cy - s * 0.23),
      stem,
    );
    canvas.drawLine(
      Offset(cx - s * 0.10, cy - s * 0.18),
      Offset(cx + s * 0.16, cy - s * 0.23),
      stem..strokeWidth = s * 0.085,
    );
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx - s * 0.17, cy + s * 0.16),
        width: s * 0.17,
        height: s * 0.13,
      ),
      teal,
    );
    canvas.drawOval(
      Rect.fromCenter(
        center: Offset(cx + s * 0.09, cy + s * 0.11),
        width: s * 0.17,
        height: s * 0.13,
      ),
      teal,
    );
  }

  void _drawMiniHeart(Canvas canvas, Offset center, double size, Color color) {
    final path = Path()
      ..moveTo(center.dx, center.dy + size * 0.35)
      ..cubicTo(
        center.dx - size * 0.55,
        center.dy - size * 0.05,
        center.dx - size * 0.35,
        center.dy - size * 0.52,
        center.dx,
        center.dy - size * 0.18,
      )
      ..cubicTo(
        center.dx + size * 0.35,
        center.dy - size * 0.52,
        center.dx + size * 0.55,
        center.dy - size * 0.05,
        center.dx,
        center.dy + size * 0.35,
      )
      ..close();
    canvas.drawPath(path, Paint()..color = color);
  }

  @override
  bool shouldRepaint(covariant _StickerVectorPainter oldDelegate) =>
      oldDelegate.stickerId != stickerId;
}
