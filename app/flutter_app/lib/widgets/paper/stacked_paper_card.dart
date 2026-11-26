import 'dart:math' as math;
import 'dart:ui' as ui;

import 'package:flutter/material.dart';

/// Tarjeta con textura de papel apilado (`assets/textures/paper.png` o
/// `assets/textures/photopaper.png`):
/// - Una hoja trasera rotada aleatoriamente entre -2° y 2° (evitando ángulos
///   demasiado cercanos a 0° para que siempre se note la hoja trasera),
///   renderizada con antialiasing para evitar bordes dentados.
/// - Una hoja frontal siempre recta (0°) con esquinas cuadradas de papel.
/// - La textura se estira (`BoxFit.fill`) para ajustarse exactamente a cada
///   contenedor.
class StackedPaperCard extends StatefulWidget {
  const StackedPaperCard({
    super.key,
    required this.child,
    this.textureAsset = 'assets/textures/paper.png',
    this.tintColor,
    this.padding = EdgeInsets.zero,
    this.backAngleDegrees,
    this.paperShadowBlur = 14,
  });

  final Widget child;
  final String textureAsset;

  /// Tinte opcional del papel (p. ej. crema para "Escuchando" o celeste suave
  /// para "Nivel"). Si es null, se usa papel blanco natural.
  final Color? tintColor;
  final EdgeInsetsGeometry padding;

  /// Ángulo fijo opcional en grados para pruebas determinísticas.
  final double? backAngleDegrees;
  final double paperShadowBlur;

  /// Ancho de decodificación en memoria para las texturas de papel.
  /// Reduce `paper.png` (1341×1739, ~9 MB RGBA) a un tamaño óptimo para GPU
  /// sin pérdida visual de grano, eliminando pausas al cambiar de pestaña.
  static int? cacheWidthForTexture(String assetPath) {
    if (assetPath.endsWith('paper.png') &&
        !assetPath.endsWith('photopaper.png')) {
      return 540;
    }
    if (assetPath.endsWith('photopaper.png')) {
      return 360;
    }
    return null;
  }

  /// Precarga las texturas base del tablero de papel una sola vez al iniciar
  /// el shell principal para evitar decodificación síncrona al navegar.
  static Future<void> preloadTextures(BuildContext context) async {
    const assets = <String>[
      'assets/textures/paper.png',
      'assets/textures/photopaper.png',
      'assets/textures/pin.png',
      'assets/textures/tape.png',
    ];
    await Future.wait(
      assets.map((asset) {
        final cacheWidth = cacheWidthForTexture(asset);
        final provider = cacheWidth != null
            ? ResizeImage.resizeIfNeeded(cacheWidth, null, AssetImage(asset))
            : AssetImage(asset) as ImageProvider;
        return precacheImage(provider, context, onError: (_, _) {});
      }),
    );
  }

  @override
  State<StackedPaperCard> createState() => _StackedPaperCardState();
}

class _StackedPaperCardState extends State<StackedPaperCard> {
  static final math.Random _random = math.Random();
  late final double _backAngleRadians;

  @override
  void initState() {
    super.initState();
    final degrees = widget.backAngleDegrees ?? _pickVisiblePaperAngle();
    _backAngleRadians = degrees * math.pi / 180.0;
  }

  /// Genera un ángulo en [-2.0, -0.95] U [0.95, 2.0] para que la hoja trasera
  /// nunca quede alineada a 0° con la hoja frontal.
  static double _pickVisiblePaperAngle() {
    final sign = _random.nextBool() ? 1.0 : -1.0;
    final magnitude = 0.95 + _random.nextDouble() * 1.05;
    return sign * magnitude;
  }

  @override
  Widget build(BuildContext context) {
    final backTint = widget.tintColor != null
        ? Color.lerp(widget.tintColor, const Color(0xFFD6D2C4), 0.14)
        : const Color(0xFFEBECEB);

    return Stack(
      clipBehavior: Clip.none,
      children: [
        Positioned.fill(
          child: IgnorePointer(
            child: Transform.rotate(
              angle: _backAngleRadians,
              filterQuality: FilterQuality.medium,
              child: _PaperSheetSurface(
                textureAsset: widget.textureAsset,
                tintColor: backTint,
                shadowBlur: widget.paperShadowBlur,
                shadowAlpha: 0.22,
                antiAliasedEdge: true,
              ),
            ),
          ),
        ),
        _PaperSheetSurface(
          textureAsset: widget.textureAsset,
          tintColor: widget.tintColor,
          shadowBlur: widget.paperShadowBlur * 0.7,
          shadowAlpha: 0.16,
          child: Padding(padding: widget.padding, child: widget.child),
        ),
      ],
    );
  }
}

/// Pila configurable de 1 a 3 hojas de papel fotográfico
/// (`assets/textures/photopaper.png`):
/// - `sheetCount == 1`: una sola hoja limpia sin hojas traseras.
/// - `sheetCount == 2`: una hoja trasera ligeramente rotada + hoja frontal.
/// - `sheetCount >= 3`: dos hojas traseras escalonadas + hoja frontal (máx. 3).
class PhotoPaperSheetStack extends StatelessWidget {
  const PhotoPaperSheetStack({
    super.key,
    required this.child,
    this.sheetCount = 1,
    this.padding = EdgeInsets.zero,
    this.tintColor,
    this.shadowBlur = 14,
  });

  final Widget child;
  final int sheetCount;
  final EdgeInsetsGeometry padding;
  final Color? tintColor;
  final double shadowBlur;

  @override
  Widget build(BuildContext context) {
    final clampedCount = sheetCount.clamp(1, 3);
    const texture = 'assets/textures/photopaper.png';
    final midBackTint = tintColor != null
        ? Color.lerp(tintColor, const Color(0xFFD6D2C4), 0.12)
        : const Color(0xFFECEEED);
    final deepBackTint = tintColor != null
        ? Color.lerp(tintColor, const Color(0xFFC8C5B8), 0.18)
        : const Color(0xFFE3E7E6);

    final frontSheet = _PaperSheetSurface(
      textureAsset: texture,
      tintColor: tintColor,
      shadowBlur: shadowBlur * 0.85,
      shadowAlpha: 0.20,
      child: Padding(padding: padding, child: child),
    );

    if (clampedCount == 1) {
      return frontSheet;
    }

    return Stack(
      clipBehavior: Clip.none,
      children: [
        if (clampedCount >= 3)
          Positioned.fill(
            child: IgnorePointer(
              child: Transform.translate(
                offset: const Offset(3, 6),
                child: Transform.rotate(
                  angle: 2.1 * math.pi / 180.0,
                  filterQuality: FilterQuality.medium,
                  child: _PaperSheetSurface(
                    textureAsset: texture,
                    tintColor: deepBackTint,
                    shadowBlur: shadowBlur,
                    shadowAlpha: 0.20,
                    antiAliasedEdge: true,
                  ),
                ),
              ),
            ),
          ),
        Positioned.fill(
          child: IgnorePointer(
            child: Transform.translate(
              offset: const Offset(-2, 3),
              child: Transform.rotate(
                angle: -1.65 * math.pi / 180.0,
                filterQuality: FilterQuality.medium,
                child: _PaperSheetSurface(
                  textureAsset: texture,
                  tintColor: midBackTint,
                  shadowBlur: shadowBlur,
                  shadowAlpha: 0.22,
                  antiAliasedEdge: true,
                ),
              ),
            ),
          ),
        ),
        frontSheet,
      ],
    );
  }
}

class _PaperSheetSurface extends StatelessWidget {
  const _PaperSheetSurface({
    required this.textureAsset,
    this.tintColor,
    this.child,
    this.shadowBlur = 12,
    this.shadowAlpha = 0.18,
    this.antiAliasedEdge = false,
  });

  final String textureAsset;
  final Color? tintColor;
  final Widget? child;
  final double shadowBlur;
  final double shadowAlpha;
  final bool antiAliasedEdge;

  @override
  Widget build(BuildContext context) {
    final cacheWidth = StackedPaperCard.cacheWidthForTexture(textureAsset);
    final backgroundTexture = Image.asset(
      textureAsset,
      cacheWidth: cacheWidth,
      fit: BoxFit.fill,
      isAntiAlias: true,
      filterQuality: FilterQuality.medium,
      color: tintColor,
      colorBlendMode: tintColor != null ? BlendMode.multiply : null,
      errorBuilder: (context, error, stackTrace) =>
          ColoredBox(color: tintColor ?? const Color(0xFFF9FAF9)),
    );

    if (antiAliasedEdge) {
      // Un ClipRRect con radio sub-píxel (0.5) y Clip.antiAlias suaviza los
      // bordes rotados eliminando el efecto serrucho manteniendo esquinas rectas.
      return DecoratedBox(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(0.5),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: shadowAlpha),
              blurRadius: shadowBlur,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(0.5),
          clipBehavior: Clip.antiAlias,
          child: ColoredBox(
            color: tintColor ?? const Color(0xFFF9FAF9),
            child: Stack(
              fit: StackFit.passthrough,
              children: [
                Positioned.fill(child: IgnorePointer(child: backgroundTexture)),
                ?child,
              ],
            ),
          ),
        ),
      );
    }

    // Dibujamos directo al framebuffer sin RepaintBoundary de pantalla completa
    // para evitar asignar render targets MSAA (15-40 MB por hoja en Windows).
    return Container(
      decoration: BoxDecoration(
        color: tintColor ?? const Color(0xFFF9FAF9),
        borderRadius: BorderRadius.zero,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: shadowAlpha),
            blurRadius: shadowBlur,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Stack(
        fit: StackFit.passthrough,
        children: [
          Positioned.fill(child: IgnorePointer(child: backgroundTexture)),
          ?child,
        ],
      ),
    );
  }
}

/// Pin de textura (`assets/textures/pin.png`) con cambio dinámico de color
/// preservando los brillos blancos y la sombra del asset original mediante
/// rotación de tono + ajuste de saturación.
class CustomizablePin extends StatelessWidget {
  const CustomizablePin({super.key, this.color, this.size = 52});

  /// Color personalizado desbloqueado (null = color rojo original del pin).
  final Color? color;
  final double size;

  /// Tono base aproximado de `assets/textures/pin.png` (rojo carmesí ~346°).
  static const double _basePinHue = 346.0;

  @override
  Widget build(BuildContext context) {
    final pinImage = Image.asset(
      'assets/textures/pin.png',
      width: size,
      height: size,
      fit: BoxFit.contain,
      isAntiAlias: true,
      filterQuality: FilterQuality.medium,
      errorBuilder: (context, error, stackTrace) => Icon(
        Icons.push_pin_rounded,
        size: size * 0.65,
        color: color ?? const Color(0xFFE12755),
      ),
    );

    final targetColor = color;
    if (targetColor == null || _isDefaultRed(targetColor)) {
      return RepaintBoundary(child: IgnorePointer(child: pinImage));
    }

    final targetHsl = HSLColor.fromColor(targetColor);
    final hueDeltaDeg = targetHsl.hue - _basePinHue;
    final saturationScale = (targetHsl.saturation / 0.78).clamp(0.15, 1.35);

    return RepaintBoundary(
      child: IgnorePointer(
        child: ColorFiltered(
          colorFilter: ColorFilter.matrix(
            _buildHueSaturationMatrix(
              hueDegrees: hueDeltaDeg,
              saturation: saturationScale,
            ),
          ),
          child: pinImage,
        ),
      ),
    );
  }

  static bool _isDefaultRed(Color c) {
    final hsl = HSLColor.fromColor(c);
    final isRedHue = hsl.hue >= 332 || hsl.hue <= 12;
    return isRedHue && hsl.saturation > 0.55 && hsl.lightness > 0.3;
  }

  static List<double> _buildHueSaturationMatrix({
    required double hueDegrees,
    required double saturation,
  }) {
    final rad = hueDegrees * math.pi / 180.0;
    final cosVal = math.cos(rad);
    final sinVal = math.sin(rad);

    const lumR = 0.213;
    const lumG = 0.715;
    const lumB = 0.072;

    final h00 = lumR + cosVal * (1 - lumR) + sinVal * (-lumR);
    final h01 = lumG + cosVal * (-lumG) + sinVal * (-lumG);
    final h02 = lumB + cosVal * (-lumB) + sinVal * (1 - lumB);

    final h10 = lumR + cosVal * (-lumR) + sinVal * 0.143;
    final h11 = lumG + cosVal * (1 - lumG) + sinVal * 0.140;
    final h12 = lumB + cosVal * (-lumB) + sinVal * (-0.283);

    final h20 = lumR + cosVal * (-lumR) + sinVal * (-(1 - lumR));
    final h21 = lumG + cosVal * (-lumG) + sinVal * lumG;
    final h22 = lumB + cosVal * (1 - lumB) + sinVal * lumB;

    final s = saturation;
    final sr = (1 - s) * lumR;
    final sg = (1 - s) * lumG;
    final sb = (1 - s) * lumB;

    return <double>[
      (sr + s) * h00 + sg * h10 + sb * h20,
      (sr + s) * h01 + sg * h11 + sb * h21,
      (sr + s) * h02 + sg * h12 + sb * h22,
      0,
      0,
      sr * h00 + (sg + s) * h10 + sb * h20,
      sr * h01 + (sg + s) * h11 + sb * h21,
      sr * h02 + (sg + s) * h12 + sb * h22,
      0,
      0,
      sr * h00 + sg * h10 + (sb + s) * h20,
      sr * h01 + sg * h11 + (sb + s) * h21,
      sr * h02 + sg * h12 + (sb + s) * h22,
      0,
      0,
      0,
      0,
      0,
      1,
      0,
    ];
  }
}

/// Cintas adhesivas (`assets/textures/tape.png`, relación natural 472x311) para
/// las esquinas superiores de la hoja principal de blogs:
/// - Esquina superior izquierda: pegada directamente sin rotación extra, con
///   tamaño realista que sujeta firmemente la esquina del papel.
/// - Esquina superior derecha: ligeramente más pequeña (~80%) y rotada a un
///   ángulo natural complementario para sostener la esquina derecha.
class TapeCornerOverlay extends StatelessWidget {
  const TapeCornerOverlay({
    super.key,
    required this.child,
    this.leftWidth = 175,
    this.rightWidth = 152,
  });

  final Widget child;
  final double leftWidth;
  final double rightWidth;

  static const double _tapeAspect = 311.0 / 472.0;

  @override
  Widget build(BuildContext context) {
    final leftHeight = leftWidth * _tapeAspect;
    final rightHeight = rightWidth * _tapeAspect;

    return Stack(
      clipBehavior: Clip.none,
      children: [
        child,
        Positioned(
          top: -leftHeight * 0.35,
          left: -leftWidth * 0.38,
          child: IgnorePointer(
            child: Opacity(
              opacity: 0.48,
              child: Image.asset(
                'assets/textures/tape.png',
                width: leftWidth,
                height: leftHeight,
                fit: BoxFit.contain,
                isAntiAlias: true,
                filterQuality: FilterQuality.medium,
                errorBuilder: (context, error, stackTrace) =>
                    const SizedBox.shrink(),
              ),
            ),
          ),
        ),
        Positioned(
          top: -rightHeight * 0.28,
          right: -rightWidth * 0.42,
          child: IgnorePointer(
            child: Opacity(
              opacity: 0.46,
              child: Transform.rotate(
                angle: 66 * math.pi / 180.0,
                filterQuality: FilterQuality.medium,
                child: Image.asset(
                  'assets/textures/tape.png',
                  width: rightWidth,
                  height: rightHeight,
                  fit: BoxFit.contain,
                  isAntiAlias: true,
                  filterQuality: FilterQuality.medium,
                  errorBuilder: (context, error, stackTrace) =>
                      const SizedBox.shrink(),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

/// Placeholder limpio con símbolo de imagen para avatares y carátulas mientras
/// no exista contenido multimedia real desde el backend.
class MeloopImagePlaceholder extends StatelessWidget {
  const MeloopImagePlaceholder({
    super.key,
    this.width,
    this.height,
    this.iconSize = 24,
    this.borderRadius = BorderRadius.zero,
    this.backgroundColor = const Color(0xFFE5EAE9),
    this.iconColor = const Color(0xFF7D908E),
    this.borderColor = const Color(0xFFD0DADA),
  });

  final double? width;
  final double? height;
  final double iconSize;
  final BorderRadius borderRadius;
  final Color backgroundColor;
  final Color iconColor;
  final Color? borderColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: borderRadius,
        border: borderColor != null ? Border.all(color: borderColor!) : null,
      ),
      child: Icon(Icons.image_outlined, size: iconSize, color: iconColor),
    );
  }
}

/// Logo de Meloop con contorno blanco estilo sticker para que resalte sobre la
/// barra turquesa exactamente como en el mockup de Figma.
class MeloopLogoSticker extends StatelessWidget {
  const MeloopLogoSticker({
    super.key,
    this.height = 46,
    this.strokeWidth = 2.4,
    this.whiteOnly = false,
  });

  final double height;
  final double strokeWidth;
  final bool whiteOnly;

  static int decodedHeightFor(BuildContext context, double height) {
    final dpr = MediaQuery.maybeDevicePixelRatioOf(context) ?? 1.0;
    return (height * dpr * 2).round().clamp(48, 605);
  }

  /// Precarga el logo con el mismo `cacheHeight` exacto que usa el widget para
  /// garantizar que esté decodificado en `ImageCache` antes de revelar la vista.
  static Future<void> preload(
    BuildContext context, {
    double height = 56,
  }) async {
    final decodedHeight = decodedHeightFor(context, height);
    await precacheImage(
      ResizeImage.resizeIfNeeded(
        null,
        decodedHeight,
        const AssetImage('assets/imagenes/meloop.png'),
      ),
      context,
      onError: (_, _) {},
    );
  }

  @override
  Widget build(BuildContext context) {
    // meloop.png mide 2075x605 (~13x más grande que 46px). Usar FilterQuality.medium
    // (mipmaps trilineales) junto a cacheHeight evita el aliasing de bicubic 4x4
    // en Skia/CanvasKit (Edge/Chrome) al reducir imágenes muy grandes.
    final decodedHeight = decodedHeightFor(context, height);

    if (whiteOnly) {
      return Image.asset(
        'assets/imagenes/meloop.png',
        height: height,
        cacheHeight: decodedHeight,
        fit: BoxFit.contain,
        isAntiAlias: true,
        filterQuality: FilterQuality.medium,
        color: Colors.white,
        colorBlendMode: BlendMode.srcIn,
      );
    }

    final outlineOffsets = List<Offset>.generate(16, (index) {
      final angle = index * (2 * math.pi / 16);
      return Offset(
        math.cos(angle) * strokeWidth,
        math.sin(angle) * strokeWidth,
      );
    });

    return RepaintBoundary(
      child: Stack(
        alignment: Alignment.center,
        children: [
          ImageFiltered(
            imageFilter: ui.ImageFilter.blur(
              sigmaX: 0.55,
              sigmaY: 0.55,
              tileMode: TileMode.decal,
            ),
            child: Stack(
              alignment: Alignment.center,
              children: [
                for (final offset in outlineOffsets)
                  Transform.translate(
                    offset: offset,
                    filterQuality: FilterQuality.medium,
                    child: Image.asset(
                      'assets/imagenes/meloop.png',
                      height: height,
                      cacheHeight: decodedHeight,
                      fit: BoxFit.contain,
                      isAntiAlias: true,
                      filterQuality: FilterQuality.medium,
                      color: Colors.white,
                      colorBlendMode: BlendMode.srcIn,
                    ),
                  ),
              ],
            ),
          ),
          Image.asset(
            'assets/imagenes/meloop.png',
            height: height,
            cacheHeight: decodedHeight,
            fit: BoxFit.contain,
            isAntiAlias: true,
            filterQuality: FilterQuality.medium,
          ),
        ],
      ),
    );
  }
}
