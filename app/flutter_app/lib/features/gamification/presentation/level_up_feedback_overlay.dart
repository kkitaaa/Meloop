import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../../../theme/meloop_typography.dart';
import '../../../widgets/paper/stacked_paper_card.dart';
import 'gamification_demo_controller.dart';

/// Presenta la notificación de subida de nivel sobre una hoja de papel
/// fotográfico con cañones laterales de confeti que salen disparados desde
/// ambos costados de la tarjeta y caen por gravedad sin bloquear la app.
class LevelUpFeedbackOverlay extends StatefulWidget {
  const LevelUpFeedbackOverlay({
    required this.child,
    required this.onEquip,
    super.key,
  });

  final Widget child;
  final VoidCallback onEquip;

  @override
  State<LevelUpFeedbackOverlay> createState() => _LevelUpFeedbackOverlayState();
}

class _LevelUpFeedbackOverlayState extends State<LevelUpFeedbackOverlay> {
  final GamificationDemoController _controller =
      GamificationDemoController.instance;
  LevelUpReward? _reward;

  @override
  void initState() {
    super.initState();
    _controller.addListener(_onProgressChanged);
  }

  @override
  void dispose() {
    _controller.removeListener(_onProgressChanged);
    super.dispose();
  }

  void _onProgressChanged() {
    final reward = _controller.takePendingLevelUp();
    if (reward == null || !mounted) return;
    setState(() => _reward = reward);
  }

  void _closeReward() {
    if (!mounted) return;
    setState(() => _reward = null);
  }

  void _equipReward() {
    _controller.equipRgbFrame();
    _closeReward();
    widget.onEquip();
  }

  @override
  Widget build(BuildContext context) {
    final reward = _reward;
    final screenWidth = MediaQuery.sizeOf(context).width;
    final isDesktop = screenWidth > 900;
    final bannerTop =
        MediaQuery.viewPaddingOf(context).top + (isDesktop ? 98.0 : 56.0);
    final bannerWidth = math.min(screenWidth - 32.0, 520.0);

    return Stack(
      fit: StackFit.expand,
      children: [
        widget.child,
        if (reward != null)
          Positioned.fill(
            child: _SideCannonConfetti(
              key: const ValueKey('level-up-confetti'),
              bannerTop: bannerTop,
              bannerWidth: bannerWidth,
            ),
          ),
        if (reward != null)
          Positioned(
            top: bannerTop,
            left: 0,
            right: 0,
            child: _RewardLevelUpBanner(
              key: ValueKey('level-up-banner-${reward.level}-${reward.name}'),
              reward: reward,
              onClose: _closeReward,
              onEquip: _equipReward,
            ),
          ),
      ],
    );
  }
}

class _RewardLevelUpBanner extends StatefulWidget {
  const _RewardLevelUpBanner({
    required this.reward,
    required this.onClose,
    required this.onEquip,
    super.key,
  });

  final LevelUpReward reward;
  final VoidCallback onClose;
  final VoidCallback onEquip;

  @override
  State<_RewardLevelUpBanner> createState() => _RewardLevelUpBannerState();
}

class _RewardLevelUpBannerState extends State<_RewardLevelUpBanner>
    with SingleTickerProviderStateMixin {
  late final AnimationController _timer;

  @override
  void initState() {
    super.initState();
    _timer = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 10),
    )..forward();
    _timer.addStatusListener((status) {
      if (status == AnimationStatus.completed) widget.onClose();
    });
  }

  @override
  void dispose() {
    _timer.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: AnimatedBuilder(
        animation: _timer,
        builder: (context, child) {
          final enterT = Curves.easeOutBack.transform(
            (_timer.value / 0.065).clamp(0.0, 1.0),
          );
          final exitFade = _timer.value > 0.94
              ? (1.0 - ((_timer.value - 0.94) / 0.06)).clamp(0.0, 1.0)
              : 1.0;
          final entrySlideY = -1.15 * (1.0 - enterT);
          final entryScale = 0.92 + 0.08 * enterT.clamp(0.0, 1.15);

          return Opacity(
            opacity: exitFade,
            child: FractionalTranslation(
              translation: Offset(0, entrySlideY),
              child: Transform.scale(scale: entryScale, child: child),
            ),
          );
        },
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: PhotoPaperSheetStack(
              sheetCount: 1,
              shadowBlur: 18,
              padding: const EdgeInsets.fromLTRB(18, 14, 18, 10),
              child: Material(
                color: Colors.transparent,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    SizedBox(
                      height: 62,
                      child: AnimatedBuilder(
                        animation: _timer,
                        builder: (_, _) {
                          // Fase 1 (0.0 -> 0.29): celebración entusiasta de texto
                          if (_timer.value < 0.29) {
                            return _EnthusiasticLevelUpHeadline(
                              level: widget.reward.level,
                              phaseProgress: (_timer.value / 0.29).clamp(
                                0.0,
                                1.0,
                              ),
                            );
                          }
                          // Fase 2 (0.29 -> 1.0): recompensa desbloqueada + botón Equipar
                          final revealT = Curves.easeOutCubic.transform(
                            ((_timer.value - 0.29) / 0.07).clamp(0.0, 1.0),
                          );
                          return Opacity(
                            opacity: revealT,
                            child: Transform.translate(
                              offset: Offset(0, 10 * (1.0 - revealT)),
                              child: Row(
                                children: [
                                  Container(
                                    width: 44,
                                    height: 44,
                                    decoration: BoxDecoration(
                                      gradient: const LinearGradient(
                                        begin: Alignment.topLeft,
                                        end: Alignment.bottomRight,
                                        colors: [
                                          Color(0xFF00AF92),
                                          Color(0xFF4BE3C8),
                                          Color(0xFFFFB830),
                                        ],
                                      ),
                                      borderRadius: BorderRadius.circular(10),
                                    ),
                                    child: const Icon(
                                      Icons.auto_awesome_rounded,
                                      color: Colors.white,
                                      size: 24,
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      mainAxisAlignment:
                                          MainAxisAlignment.center,
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          widget.reward.name,
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                          style: MeloopTypography.lexend(
                                            fontSize: 16.5,
                                            fontWeight: FontWeight.w700,
                                            color: const Color(0xFF152020),
                                          ),
                                        ),
                                        const SizedBox(height: 1),
                                        Text(
                                          'Marco de perfil',
                                          style: MeloopTypography.lexend(
                                            color: const Color(0xFF647574),
                                            fontSize: 12,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  ElevatedButton(
                                    onPressed: widget.onEquip,
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: const Color(0xFF00AF92),
                                      foregroundColor: Colors.white,
                                      elevation: 0,
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 18,
                                        vertical: 10,
                                      ),
                                      shape: RoundedRectangleBorder(
                                        borderRadius: BorderRadius.circular(20),
                                      ),
                                    ),
                                    child: Text(
                                      'Equipar',
                                      style: MeloopTypography.lexend(
                                        fontSize: 13.5,
                                        fontWeight: FontWeight.w600,
                                        color: Colors.white,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                    const SizedBox(height: 6),
                    AnimatedBuilder(
                      animation: _timer,
                      builder: (_, _) => ClipRRect(
                        borderRadius: BorderRadius.circular(4),
                        child: LinearProgressIndicator(
                          value: 1 - _timer.value,
                          minHeight: 3.5,
                          color: const Color(0xFF00AF92),
                          backgroundColor: const Color(0xFFE0ECEA),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// Texto animado entusiasta letra por letra con rebote elástico, destellos
/// laterales y transición limpia hacia la recompensa.
class _EnthusiasticLevelUpHeadline extends StatelessWidget {
  const _EnthusiasticLevelUpHeadline({
    required this.level,
    required this.phaseProgress,
  });

  final int level;
  final double phaseProgress;

  static const _phrase = '¡Subiste de nivel!';

  @override
  Widget build(BuildContext context) {
    final fadeOut = phaseProgress > 0.82
        ? (1.0 - ((phaseProgress - 0.82) / 0.18)).clamp(0.0, 1.0)
        : 1.0;
    final sparklePulse = 0.88 + 0.24 * math.sin(phaseProgress * math.pi * 6);
    final sparkleSpin = phaseProgress * math.pi * 1.4;
    final badgeT = Curves.easeOutBack.transform(
      ((phaseProgress - 0.22) / 0.24).clamp(0.0, 1.0),
    );

    return Opacity(
      opacity: fadeOut,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Transform.rotate(
                angle: -sparkleSpin,
                child: Transform.scale(
                  scale: sparklePulse,
                  child: const Icon(
                    Icons.auto_awesome_rounded,
                    color: Color(0xFFFFB830),
                    size: 20,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Flexible(
                child: FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: List.generate(_phrase.length, (index) {
                      final charStart = index * 0.022;
                      final charT = ((phaseProgress - charStart) / 0.24).clamp(
                        0.0,
                        1.0,
                      );
                      final popScale = charT == 0
                          ? 0.0
                          : Curves.elasticOut.transform(charT);
                      final waveBounce =
                          math.sin(
                            (phaseProgress * 5.5 - index * 0.35).clamp(
                              0.0,
                              math.pi * 4,
                            ),
                          ) *
                          2.2 *
                          (1.0 - phaseProgress);
                      final hopY = (1.0 - Curves.easeOut.transform(charT)) * 12;

                      return Transform.translate(
                        offset: Offset(0, hopY - waveBounce),
                        child: Transform.scale(
                          scale: popScale.clamp(0.0, 1.35),
                          child: Text(
                            _phrase[index],
                            style: MeloopTypography.lexend(
                              fontSize: 22,
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF127A6E),
                              letterSpacing: 0.2,
                            ),
                          ),
                        ),
                      );
                    }),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Transform.rotate(
                angle: sparkleSpin,
                child: Transform.scale(
                  scale: sparklePulse,
                  child: const Icon(
                    Icons.auto_awesome_rounded,
                    color: Color(0xFF00AF92),
                    size: 20,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 3),
          Transform.scale(
            scale: badgeT.clamp(0.0, 1.15),
            child: Opacity(
              opacity: badgeT.clamp(0.0, 1.0),
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 2,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFFE1F7F3),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  '¡Ahora eres Nivel $level!',
                  style: MeloopTypography.lexend(
                    fontSize: 11.5,
                    fontWeight: FontWeight.w700,
                    color: const Color(0xFF10857A),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Cañones laterales de confeti con física balística: las piezas salen
/// disparadas desde los costados izquierdo y derecho de la hoja de papel,
/// desaceleran por arrastre del aire y caen con aleteo por gravedad.
class _SideCannonConfetti extends StatefulWidget {
  const _SideCannonConfetti({
    super.key,
    required this.bannerTop,
    required this.bannerWidth,
  });

  final double bannerTop;
  final double bannerWidth;

  @override
  State<_SideCannonConfetti> createState() => _SideCannonConfettiState();
}

class _SideCannonConfettiState extends State<_SideCannonConfetti>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final List<_ConfettiParticle> _particles;

  static const _palette = <Color>[
    Color(0xFF00AF92),
    Color(0xFF50DDC7),
    Color(0xFFFFB830),
    Color(0xFFFF6262),
    Color(0xFF38BDF8),
    Color(0xFFF472B6),
  ];

  @override
  void initState() {
    super.initState();
    final rng = math.Random(2026);
    _particles = List<_ConfettiParticle>.generate(56, (index) {
      final isLeft = index.isEven;
      final dir = isLeft ? -1.0 : 1.0;
      // Velocidad lateral en abanico hacia afuera y ligeramente hacia arriba (escritorio)
      final speedX = (140 + rng.nextDouble() * 340) * dir;
      final speedY = -(130 + rng.nextDouble() * 290);
      // Parámetros para fuente superior en pantallas angostas / Android
      final topSpreadFraction = rng.nextDouble();
      final topVx = (rng.nextDouble() - 0.5) * 230;
      final topVy = -(170 + rng.nextDouble() * 220);
      return _ConfettiParticle(
        fromLeft: isLeft,
        originYOffset: 14 + rng.nextDouble() * 44,
        topSpreadFraction: topSpreadFraction,
        vx: speedX,
        vy: speedY,
        topVx: topVx,
        topVy: topVy,
        drag: 1.55 + rng.nextDouble() * 1.1,
        gravity: 340 + rng.nextDouble() * 180,
        swayAmp: 10 + rng.nextDouble() * 18,
        swayFreq: 3.0 + rng.nextDouble() * 3.5,
        spinSpeed: (rng.nextBool() ? 1 : -1) * (4.0 + rng.nextDouble() * 6.5),
        flipSpeed: 4.5 + rng.nextDouble() * 5.0,
        phase: rng.nextDouble() * math.pi * 2,
        width: 6.5 + rng.nextDouble() * 5.5,
        height: 4.0 + rng.nextDouble() * 4.5,
        isCircle: index % 5 == 0,
        color: _palette[index % _palette.length],
        launchDelay: rng.nextDouble() * 0.18,
      );
    });

    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 3600),
    )..forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: RepaintBoundary(
        child: AnimatedBuilder(
          animation: _controller,
          builder: (context, _) => CustomPaint(
            painter: _SideCannonConfettiPainter(
              particles: _particles,
              elapsedSeconds: _controller.value * 3.6,
              bannerTop: widget.bannerTop,
              bannerWidth: widget.bannerWidth,
            ),
            child: const SizedBox.expand(),
          ),
        ),
      ),
    );
  }
}

class _SideCannonConfettiPainter extends CustomPainter {
  const _SideCannonConfettiPainter({
    required this.particles,
    required this.elapsedSeconds,
    required this.bannerTop,
    required this.bannerWidth,
  });

  final List<_ConfettiParticle> particles;
  final double elapsedSeconds;
  final double bannerTop;
  final double bannerWidth;

  @override
  void paint(Canvas canvas, Size size) {
    if (size.isEmpty) return;

    final leftEdge = ((size.width - bannerWidth) / 2).clamp(12.0, size.width);
    final rightEdge = (leftEdge + bannerWidth).clamp(0.0, size.width - 12.0);
    // En pantallas angostas (Android/móvil) el papel ocupa casi todo el ancho,
    // así que el confeti dispara desde la parte superior de la tarjeta hacia
    // arriba en abanico y luego cae por toda la pantalla sin perderse por los bordes.
    final shootFromTop = size.width <= 700;
    final paint = Paint()..style = PaintingStyle.fill;

    for (final p in particles) {
      final t = elapsedSeconds - p.launchDelay;
      if (t <= 0) continue;

      final lifeProgress = (t / 3.3).clamp(0.0, 1.0);
      final alpha = lifeProgress > 0.72
          ? (1.0 - (lifeProgress - 0.72) / 0.28).clamp(0.0, 1.0)
          : 1.0;
      if (alpha <= 0.01) continue;

      // Posición horizontal con arrastre exponencial + aleteo sinusoidal
      final dragFactor = (1.0 - math.exp(-p.drag * t)) / p.drag;
      final double originX;
      final double originY;
      final double launchVx;
      final double launchVy;
      if (shootFromTop) {
        final usableWidth = math.max(24.0, (rightEdge - leftEdge) - 24.0);
        originX = leftEdge + 12.0 + p.topSpreadFraction * usableWidth;
        originY = bannerTop + 8.0;
        launchVx = p.topVx;
        launchVy = p.topVy;
      } else {
        originX = p.fromLeft ? leftEdge + 6 : rightEdge - 6;
        originY = bannerTop + p.originYOffset;
        launchVx = p.vx;
        launchVy = p.vy;
      }

      final x =
          originX +
          launchVx * dragFactor +
          math.sin(t * p.swayFreq + p.phase) * p.swayAmp * (t / 2.0).clamp(0, 1);
      final y = originY + launchVy * dragFactor + 0.5 * p.gravity * t * t;

      if (y > size.height + 24 || x < -24 || x > size.width + 24) continue;

      paint.color = p.color.withValues(alpha: alpha);
      canvas.save();
      canvas.translate(x, y);
      canvas.rotate(p.phase + t * p.spinSpeed);
      final flipScaleY = 0.25 + 0.75 * math.cos(t * p.flipSpeed + p.phase).abs();
      canvas.scale(1.0, flipScaleY);
      if (p.isCircle) {
        canvas.drawCircle(Offset.zero, p.width * 0.42, paint);
      } else {
        canvas.drawRRect(
          RRect.fromRectAndRadius(
            Rect.fromCenter(
              center: Offset.zero,
              width: p.width,
              height: p.height,
            ),
            const Radius.circular(1.4),
          ),
          paint,
        );
      }
      canvas.restore();
    }
  }

  @override
  bool shouldRepaint(covariant _SideCannonConfettiPainter oldDelegate) =>
      oldDelegate.elapsedSeconds != elapsedSeconds ||
      oldDelegate.bannerTop != bannerTop ||
      oldDelegate.bannerWidth != bannerWidth;
}

class _ConfettiParticle {
  const _ConfettiParticle({
    required this.fromLeft,
    required this.originYOffset,
    required this.topSpreadFraction,
    required this.vx,
    required this.vy,
    required this.topVx,
    required this.topVy,
    required this.drag,
    required this.gravity,
    required this.swayAmp,
    required this.swayFreq,
    required this.spinSpeed,
    required this.flipSpeed,
    required this.phase,
    required this.width,
    required this.height,
    required this.isCircle,
    required this.color,
    required this.launchDelay,
  });

  final bool fromLeft;
  final double originYOffset;
  final double topSpreadFraction;
  final double vx;
  final double vy;
  final double topVx;
  final double topVy;
  final double drag;
  final double gravity;
  final double swayAmp;
  final double swayFreq;
  final double spinSpeed;
  final double flipSpeed;
  final double phase;
  final double width;
  final double height;
  final bool isCircle;
  final Color color;
  final double launchDelay;
}
