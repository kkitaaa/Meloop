import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

/// Compensa el escalado de pantalla de Windows (125%, 150%, 175%, etc.) y
/// bloquea el escalado de texto del sistema en escritorio para que la interfaz
/// pensada para 1080p mantenga exactamente las mismas proporciones sin importar
/// el porcentaje de zoom configurado en Windows.
class DesktopViewportScaler extends StatelessWidget {
  const DesktopViewportScaler({super.key, required this.child});

  final Widget child;

  static const double _designWidth = 1920.0;

  @override
  Widget build(BuildContext context) {
    final isDesktopPlatform =
        !kIsWeb &&
        (defaultTargetPlatform == TargetPlatform.windows ||
            defaultTargetPlatform == TargetPlatform.linux ||
            defaultTargetPlatform == TargetPlatform.macOS);
    if (!isDesktopPlatform) return child;

    final mq = MediaQuery.of(context);
    final dpr = mq.devicePixelRatio;
    final logicalSize = mq.size;

    // En 1920x1080 con Windows al 125% (logical 1536x864) o 150% (1280x720),
    // compensation == dpr, devolviendo exactamente el canvas de 1920x1080.
    // En 4K al 200% (logical 1920x1080), logicalSize.width >= 1920 -> 1.0.
    final double compensation = (dpr > 1.0 && logicalSize.width < _designWidth)
        ? (_designWidth / logicalSize.width).clamp(1.0, dpr)
        : 1.0;

    final effectiveSize = Size(
      logicalSize.width * compensation,
      logicalSize.height * compensation,
    );

    final adjustedMq = mq.copyWith(
      size: effectiveSize,
      devicePixelRatio: dpr / compensation,
      textScaler: TextScaler.noScaling,
      padding: mq.padding * compensation,
      viewInsets: mq.viewInsets * compensation,
      viewPadding: mq.viewPadding * compensation,
    );

    if (compensation == 1.0) {
      return MediaQuery(data: adjustedMq, child: child);
    }

    return MediaQuery(
      data: adjustedMq,
      child: FittedBox(
        fit: BoxFit.fill,
        alignment: Alignment.topLeft,
        child: SizedBox(
          width: effectiveSize.width,
          height: effectiveSize.height,
          child: child,
        ),
      ),
    );
  }
}
