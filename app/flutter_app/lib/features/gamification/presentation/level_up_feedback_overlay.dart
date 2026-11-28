import 'package:flutter/material.dart';

import 'gamification_demo_controller.dart';

/// Presents the original Jose panels level-up banner without blocking the app.
/// Keep this presentation when demo progression is replaced by API progress.
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
    final isDesktop = MediaQuery.sizeOf(context).width > 900;
    return Stack(
      fit: StackFit.expand,
      children: [
        widget.child,
        if (reward != null)
          const Positioned.fill(
            child: _FullScreenConfetti(key: ValueKey('level-up-confetti')),
          ),
        if (reward != null)
          Positioned(
            top: MediaQuery.viewPaddingOf(context).top + (isDesktop ? 98 : 18),
            left: 0,
            right: 0,
            child: _RewardLevelUpBanner(
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
  Widget build(BuildContext context) => Center(
    child: SlideTransition(
      position: Tween<Offset>(begin: const Offset(0, -1), end: Offset.zero)
          .animate(
            CurvedAnimation(
              parent: _timer,
              curve: const Interval(0, .08, curve: Curves.easeOutBack),
            ),
          ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Container(
          margin: const EdgeInsets.symmetric(horizontal: 16),
          padding: const EdgeInsets.fromLTRB(18, 14, 18, 8),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: .22),
                blurRadius: 18,
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                height: 58,
                child: AnimatedBuilder(
                  animation: _timer,
                  builder: (_, _) {
                    if (_timer.value < .3) {
                      const phrase = '¡Subiste de nivel!';
                      final progress = (_timer.value / .3)
                          .clamp(0, 1)
                          .toDouble();
                      final fade = (1 - (progress - .8).clamp(0, .2) * 5)
                          .toDouble();
                      return Center(
                        child: Opacity(
                          opacity: fade,
                          child: Text(
                            phrase.substring(
                              0,
                              (phrase.length * progress).floor(),
                            ),
                            style: const TextStyle(
                              fontSize: 23,
                              fontWeight: FontWeight.bold,
                              color: Color(0xFF1E7E70),
                            ),
                          ),
                        ),
                      );
                    }
                    return Row(
                      children: [
                        const Icon(
                          Icons.gradient,
                          color: Colors.purple,
                          size: 40,
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                widget.reward.name,
                                style: const TextStyle(
                                  fontSize: 17,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              const Text(
                                'Marco de perfil',
                                style: TextStyle(
                                  color: Colors.black54,
                                  fontSize: 12,
                                ),
                              ),
                            ],
                          ),
                        ),
                        ElevatedButton(
                          onPressed: widget.onEquip,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF00AF92),
                            foregroundColor: Colors.white,
                          ),
                          child: const Text('Equipar'),
                        ),
                      ],
                    );
                  },
                ),
              ),
              AnimatedBuilder(
                animation: _timer,
                builder: (_, _) => LinearProgressIndicator(
                  value: 1 - _timer.value,
                  minHeight: 3,
                  color: const Color(0xFF00AF92),
                  backgroundColor: Colors.grey[200],
                ),
              ),
            ],
          ),
        ),
      ),
    ),
  );
}

class _FullScreenConfetti extends StatefulWidget {
  const _FullScreenConfetti({super.key});

  @override
  State<_FullScreenConfetti> createState() => _FullScreenConfettiState();
}

class _FullScreenConfettiState extends State<_FullScreenConfetti>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1250),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => IgnorePointer(
    child: LayoutBuilder(
      builder: (context, size) => AnimatedBuilder(
        animation: _controller,
        builder: (_, _) => Stack(
          children: List.generate(10, (i) {
            const colors = [
              Colors.amber,
              Color(0xFFFF6262),
              Color(0xFF00AF92),
              Colors.purple,
              Colors.orange,
            ];
            final x = ((i * .173) % 1);
            final fall = (_controller.value + i * .19) % 1;
            return Positioned(
              left: x * size.maxWidth,
              top: -18 + fall * (size.maxHeight + 36),
              child: Transform.rotate(
                angle: fall * 10,
                child: Icon(
                  i.isEven ? Icons.square : Icons.circle,
                  color: colors[i % colors.length],
                  size: 10,
                ),
              ),
            );
          }),
        ),
      ),
    ),
  );
}
