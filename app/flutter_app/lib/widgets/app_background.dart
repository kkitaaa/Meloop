import 'dart:math' as math;
import 'dart:ui' as ui;

import 'package:flutter/material.dart';

import '../features/profile/models/profile_customization.dart';
import '../theme/meloop_theme_controller.dart';

class AppBackground extends StatelessWidget {
  const AppBackground({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = MeloopThemeController.instance;
    return AnimatedBuilder(
      animation: controller,
      builder: (context, _) => CustomPaint(
        painter: _AppBackgroundPainter(
          theme: controller.colorTheme,
          pattern: controller.backgroundPattern,
        ),
        child: const SizedBox.expand(),
      ),
    );
  }
}

class _AppBackgroundPainter extends CustomPainter {
  const _AppBackgroundPainter({
    required this.theme,
    required this.pattern,
  });

  final MeloopColorTheme theme;
  final MeloopBackgroundPattern pattern;

  @override
  void paint(Canvas canvas, Size size) {
    switch (pattern) {
      case MeloopBackgroundPattern.tealRadial:
        _paintRadialGradient(canvas, size, theme.bgStart, theme.bgEnd);
      case MeloopBackgroundPattern.pastelGrid:
        _paintPastelGrid(canvas, size);
      case MeloopBackgroundPattern.aeroBubbles:
        _paintAeroBubbles(canvas, size);
      case MeloopBackgroundPattern.dotNotebook:
        _paintDotNotebook(canvas, size);
      case MeloopBackgroundPattern.starlightNight:
        _paintStarlightNight(canvas, size);
    }
  }

  void _paintRadialGradient(
    Canvas canvas,
    Size size,
    Color startColor,
    Color endColor,
  ) {
    final center = Offset(size.width * 0.1621, size.height * 0.2061);
    final radiusX = math.max(1.0, size.width * 0.7567);
    final radiusY = math.max(1.0, size.height * 0.7854);

    canvas.save();
    canvas.translate(center.dx, center.dy);
    canvas.scale(radiusX, radiusY);
    final shader = ui.Gradient.radial(
      Offset.zero,
      1,
      [startColor, endColor],
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

  void _paintPastelGrid(Canvas canvas, Size size) {
    canvas.drawRect(
      Offset.zero & size,
      Paint()..color = theme.gridBg,
    );
    final linePaint = Paint()
      ..color = theme.gridLine
      ..strokeWidth = 2.0;
    const step = 44.0;
    for (double x = 0; x <= size.width; x += step) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), linePaint);
    }
    for (double y = 0; y <= size.height; y += step) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), linePaint);
    }
  }

  void _paintAeroBubbles(Canvas canvas, Size size) {
    final bgShader = ui.Gradient.linear(
      Offset.zero,
      Offset(0, size.height),
      [
        Color.lerp(theme.headerTop, Colors.white, 0.28)!,
        theme.bgStart,
        theme.bgEnd,
      ],
      const [0.0, 0.52, 1.0],
    );
    canvas.drawRect(Offset.zero & size, Paint()..shader = bgShader);

    // Onda luminosa suave estilo Frutiger Aero
    final wavePath = Path()
      ..moveTo(0, size.height * 0.62)
      ..cubicTo(
        size.width * 0.28,
        size.height * 0.48,
        size.width * 0.66,
        size.height * 0.76,
        size.width,
        size.height * 0.55,
      )
      ..lineTo(size.width, size.height)
      ..lineTo(0, size.height)
      ..close();
    canvas.drawPath(
      wavePath,
      Paint()..color = Colors.white.withValues(alpha: 0.10),
    );

    const bubbles = <(double, double, double)>[
      (0.10, 0.24, 68),
      (0.22, 0.74, 92),
      (0.46, 0.16, 46),
      (0.78, 0.26, 84),
      (0.88, 0.72, 110),
      (0.62, 0.84, 58),
      (0.34, 0.42, 34),
    ];
    for (final (fx, fy, r) in bubbles) {
      final center = Offset(size.width * fx, size.height * fy);
      final bubbleGradient = ui.Gradient.radial(
        center - Offset(r * 0.3, r * 0.3),
        r * 1.1,
        [
          Colors.white.withValues(alpha: 0.48),
          Colors.white.withValues(alpha: 0.12),
          Colors.white.withValues(alpha: 0.03),
        ],
        const [0.0, 0.55, 1.0],
      );
      canvas.drawCircle(center, r, Paint()..shader = bubbleGradient);
      canvas.drawCircle(
        center,
        r,
        Paint()
          ..style = PaintingStyle.stroke
          ..strokeWidth = 1.4
          ..color = Colors.white.withValues(alpha: 0.34),
      );
      canvas.drawOval(
        Rect.fromCenter(
          center: center - Offset(r * 0.28, r * 0.34),
          width: r * 0.48,
          height: r * 0.22,
        ),
        Paint()..color = Colors.white.withValues(alpha: 0.42),
      );
    }
  }

  void _paintDotNotebook(Canvas canvas, Size size) {
    canvas.drawRect(Offset.zero & size, Paint()..color = theme.gridBg);
    final dotPaint = Paint()..color = theme.gridLine;
    const step = 30.0;
    for (double y = 16; y < size.height; y += step) {
      for (double x = 16; x < size.width; x += step) {
        canvas.drawCircle(Offset(x, y), 2.1, dotPaint);
      }
    }
  }

  void _paintStarlightNight(Canvas canvas, Size size) {
    final sky = ui.Gradient.linear(
      Offset.zero,
      Offset(size.width * 0.4, size.height),
      [
        Color.lerp(const Color(0xFF0E1428), theme.bgEnd, 0.35)!,
        const Color(0xFF090D1C),
      ],
    );
    canvas.drawRect(Offset.zero & size, Paint()..shader = sky);

    final rng = math.Random(2026);
    for (var i = 0; i < 95; i++) {
      final dx = rng.nextDouble() * size.width;
      final dy = rng.nextDouble() * size.height;
      final radius = 0.8 + rng.nextDouble() * 1.8;
      final alpha = 0.28 + rng.nextDouble() * 0.65;
      canvas.drawCircle(
        Offset(dx, dy),
        radius,
        Paint()
          ..color = (i % 6 == 0 ? theme.headerTop : Colors.white).withValues(
            alpha: alpha,
          ),
      );
    }
  }

  @override
  bool shouldRepaint(covariant _AppBackgroundPainter oldDelegate) =>
      oldDelegate.theme != theme || oldDelegate.pattern != pattern;
}
