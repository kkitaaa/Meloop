import 'dart:ui' as ui;

import 'package:flutter/material.dart';

class AppBackground extends StatelessWidget {
  const AppBackground({super.key});

  @override
  Widget build(BuildContext context) => const CustomPaint(
    painter: _AppBackgroundPainter(),
    child: SizedBox.expand(),
  );
}

class _AppBackgroundPainter extends CustomPainter {
  const _AppBackgroundPainter();

  static const _startColor = Color(0xFF248686);
  static const _endColor = Color(0xFF004562);

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width * 0.1621, size.height * 0.2061);
    final radiusX = size.width * 0.7567;
    final radiusY = size.height * 0.7854;

    canvas.save();
    canvas.translate(center.dx, center.dy);
    canvas.scale(radiusX, radiusY);
    final shader = ui.Gradient.radial(
      Offset.zero,
      1,
      [_startColor, _endColor],
      const [0, 1],
    );
    canvas.drawRect(
      Rect.fromLTRB(
        -center.dx / radiusX,
        -center.dy / radiusY,
        (size.width - center.dx) / radiusX,
        (size.height - center.dy) / radiusY,
      ),
      Paint()..shader = shader,
    );
    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant _AppBackgroundPainter oldDelegate) => false;
}
