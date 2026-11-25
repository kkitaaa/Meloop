import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../theme/meloop_typography.dart';

/// Muestra una píldora flotante plana (en sintonía con la barra de navegación
/// inferior de Meloop) que aparece desde abajo para confirmar acciones breves.
void showMeloopToastPill(
  BuildContext context, {
  required String message,
  int? xpAwarded,
  IconData icon = Icons.check_rounded,
  Duration duration = const Duration(milliseconds: 2400),
}) {
  final overlay = Overlay.maybeOf(context, rootOverlay: true);
  if (overlay == null) return;

  late final OverlayEntry entry;
  entry = OverlayEntry(
    builder: (overlayContext) => _MeloopToastPillOverlay(
      message: message,
      xpAwarded: xpAwarded,
      icon: icon,
      duration: duration,
      onDismissed: () {
        if (entry.mounted) {
          entry.remove();
        }
      },
    ),
  );

  overlay.insert(entry);
}

class _MeloopToastPillOverlay extends StatefulWidget {
  const _MeloopToastPillOverlay({
    required this.message,
    required this.icon,
    required this.duration,
    required this.onDismissed,
    this.xpAwarded,
  });

  final String message;
  final int? xpAwarded;
  final IconData icon;
  final Duration duration;
  final VoidCallback onDismissed;

  @override
  State<_MeloopToastPillOverlay> createState() =>
      _MeloopToastPillOverlayState();
}

class _MeloopToastPillOverlayState extends State<_MeloopToastPillOverlay>
    with SingleTickerProviderStateMixin {
  static const _pillColor = Color(0xFF00AF92);

  late final AnimationController _controller;
  late final Animation<double> _opacity;
  late final Animation<Offset> _slide;
  late final Animation<double> _scale;
  Timer? _dismissTimer;

  bool get _isMobilePlatform =>
      defaultTargetPlatform == TargetPlatform.android ||
      defaultTargetPlatform == TargetPlatform.iOS;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 240),
      reverseDuration: const Duration(milliseconds: 190),
    );
    final curved = CurvedAnimation(
      parent: _controller,
      curve: Curves.easeOutCubic,
      reverseCurve: Curves.easeInCubic,
    );
    _opacity = curved;
    _slide = Tween<Offset>(
      begin: const Offset(0, 0.45),
      end: Offset.zero,
    ).animate(curved);
    _scale = Tween<double>(begin: 0.94, end: 1.0).animate(curved);

    _controller.forward();
    _dismissTimer = Timer(widget.duration, _dismiss);
  }

  Future<void> _dismiss() async {
    if (!mounted) return;
    await _controller.reverse();
    if (mounted) {
      widget.onDismissed();
    }
  }

  @override
  void dispose() {
    _dismissTimer?.cancel();
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.viewPaddingOf(context).bottom;
    final pill = Container(
      padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 11),
      decoration: BoxDecoration(
        color: _pillColor,
        borderRadius: BorderRadius.circular(999),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.24),
            blurRadius: 14,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(widget.icon, color: Colors.white, size: 18),
          const SizedBox(width: 8),
          Text(
            widget.message,
            style: MeloopTypography.lexend(
              color: Colors.white,
              fontSize: 13.5,
              fontWeight: FontWeight.w600,
            ),
          ),
          if (widget.xpAwarded != null) ...[
            const SizedBox(width: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.22),
                borderRadius: BorderRadius.circular(999),
              ),
              child: Text(
                '+${widget.xpAwarded} XP',
                style: MeloopTypography.lexend(
                  color: Colors.white,
                  fontSize: 11.5,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          ],
        ],
      ),
    );

    return Positioned(
      left: 16,
      right: 16,
      bottom: bottomInset + (_isMobilePlatform ? 78 : 28),
      child: IgnorePointer(
        child: Material(
          color: Colors.transparent,
          child: Center(
            child: FadeTransition(
              opacity: _opacity,
              child: SlideTransition(
                position: _slide,
                child: ScaleTransition(scale: _scale, child: pill),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
