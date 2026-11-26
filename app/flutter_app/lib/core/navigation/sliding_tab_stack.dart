import 'package:flutter/material.dart';

/// Contenedor de pestañas con animación de deslizamiento direccional (izquierda/
/// derecha según el índice).
///
/// Política de memoria:
/// - Las pestañas en [keepAliveIndices] (por defecto solo `{0}`, Inicio/Feed)
///   permanecen montadas en `Offstage` para conservar el feed y la posición de
///   scroll.
/// - El resto de pestañas (`Amigos`, `Artistas y canciones`, etc.) se montan
///   bajo demanda al abrirlas y se desmontan automáticamente al terminar el
///   deslizamiento de salida, evitando retener widgets y memoria en segundo
///   plano cuando no están en uso.
class SlidingTabStack extends StatefulWidget {
  const SlidingTabStack({
    super.key,
    required this.currentIndex,
    required this.children,
    this.duration = const Duration(milliseconds: 260),
    this.curve = Curves.easeOutCubic,
    this.keepAliveIndices = const {0},
  });

  final int currentIndex;
  final List<Widget> children;
  final Duration duration;
  final Curve curve;
  final Set<int> keepAliveIndices;

  @override
  State<SlidingTabStack> createState() => SlidingTabStackState();
}

class SlidingTabStackState extends State<SlidingTabStack>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late CurvedAnimation _curvedAnimation;

  late int _currentIndex;
  int? _previousIndex;
  double _direction = 1.0;

  /// Cambia inmediatamente a [index] sin ejecutar animación de deslizamiento.
  /// Útil cuando el stack está oculto debajo de una subruta (como Perfil o
  /// Búsqueda) y se está haciendo pop hacia una pestaña específica, evitando
  /// que dos animaciones se superpongan y se vean cortadas.
  void jumpWithoutAnimation(int index) {
    final targetIndex = index.clamp(0, widget.children.length - 1);
    _controller.stop();
    _controller.value = 1.0;
    setState(() {
      _previousIndex = null;
      _currentIndex = targetIndex;
    });
  }

  @override
  void initState() {
    super.initState();
    _currentIndex = widget.currentIndex.clamp(0, widget.children.length - 1);
    _controller = AnimationController(vsync: this, duration: widget.duration)
      ..value = 1.0
      ..addStatusListener(_handleAnimationStatus);
    _curvedAnimation = CurvedAnimation(
      parent: _controller,
      curve: widget.curve,
    );
  }

  void _handleAnimationStatus(AnimationStatus status) {
    if (status == AnimationStatus.completed && _previousIndex != null) {
      if (mounted) {
        setState(() {
          _previousIndex = null;
        });
      }
    }
  }

  @override
  void didUpdateWidget(covariant SlidingTabStack oldWidget) {
    super.didUpdateWidget(oldWidget);

    if (oldWidget.duration != widget.duration) {
      _controller.duration = widget.duration;
    }
    if (oldWidget.curve != widget.curve) {
      _curvedAnimation.dispose();
      _curvedAnimation = CurvedAnimation(
        parent: _controller,
        curve: widget.curve,
      );
    }

    final targetIndex = widget.currentIndex.clamp(
      0,
      widget.children.length - 1,
    );
    if (targetIndex != _currentIndex) {
      _previousIndex = _currentIndex;
      _direction = targetIndex > _currentIndex ? 1.0 : -1.0;
      _currentIndex = targetIndex;
      _controller.forward(from: 0.0);
    }
  }

  @override
  void dispose() {
    _controller.removeStatusListener(_handleAnimationStatus);
    _curvedAnimation.dispose();
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ClipRect(
      child: AnimatedBuilder(
        animation: _curvedAnimation,
        builder: (context, _) {
          final isAnimating =
              _previousIndex != null && _controller.value < 1.0;
          final t = _curvedAnimation.value;

          return Stack(
            fit: StackFit.expand,
            children: List<Widget>.generate(widget.children.length, (index) {
              final isCurrent = index == _currentIndex;
              final isOutgoing = isAnimating && index == _previousIndex;
              final isVisible = isCurrent || isOutgoing;
              final shouldStayMounted =
                  isVisible || widget.keepAliveIndices.contains(index);

              if (!shouldStayMounted) {
                return const SizedBox.shrink();
              }

              Offset slideOffset = Offset.zero;
              if (isAnimating) {
                if (isCurrent) {
                  slideOffset = Offset(_direction * (1.0 - t), 0);
                } else if (isOutgoing) {
                  slideOffset = Offset(-_direction * t, 0);
                }
              }

              Widget tabChild = widget.children[index];

              if (isVisible && slideOffset != Offset.zero) {
                // Solo creamos capa compuesta durante los 260 ms del deslizamiento
                // para no retener render targets de pantalla completa en reposo.
                tabChild = RepaintBoundary(
                  child: FractionalTranslation(
                    translation: slideOffset,
                    transformHitTests: isCurrent,
                    child: tabChild,
                  ),
                );
              }

              return Offstage(
                offstage: !isVisible,
                child: TickerMode(
                  enabled: isVisible,
                  child: IgnorePointer(ignoring: !isCurrent, child: tabChild),
                ),
              );
            }),
          );
        },
      ),
    );
  }
}
