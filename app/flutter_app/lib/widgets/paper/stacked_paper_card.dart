import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../../features/profile/models/profile_customization.dart';
import '../../theme/meloop_theme_controller.dart';

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
    final themeCtrl = MeloopThemeController.instance;
    return AnimatedBuilder(
      animation: themeCtrl,
      builder: (context, _) {
        final style = themeCtrl.appStyle;
        if (style != MeloopAppStyle.paper) {
          return _StyledAppCard(
            style: style,
            colorTheme: themeCtrl.colorTheme,
            tintColor: widget.tintColor,
            padding: widget.padding,
            shadowBlur: widget.paperShadowBlur,
            child: widget.child,
          );
        }

        final backTint = widget.tintColor != null
            ? Color.lerp(widget.tintColor, const Color(0xFFD6D2C4), 0.14)
            : const Color(0xFFEBECEB);

        return Stack(
          clipBehavior: Clip.none,
          fit: StackFit.passthrough,
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
      },
    );
  }
}

class _StyledAppCard extends StatelessWidget {
  const _StyledAppCard({
    required this.style,
    required this.colorTheme,
    required this.padding,
    required this.shadowBlur,
    required this.child,
    this.tintColor,
  });

  final MeloopAppStyle style;
  final MeloopColorTheme colorTheme;
  final Color? tintColor;
  final EdgeInsetsGeometry padding;
  final double shadowBlur;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    switch (style) {
      case MeloopAppStyle.flatSquare:
        final bg = tintColor != null
            ? Color.lerp(Colors.white, colorTheme.softCardTint, 0.72)!
            : Colors.white;
        return Container(
          decoration: BoxDecoration(
            color: bg,
            borderRadius: BorderRadius.circular(2),
            border: Border.all(
              color: Color.lerp(colorTheme.gridLine, Colors.black12, 0.18)!,
              width: 1,
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.09),
                blurRadius: shadowBlur * 0.85,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          padding: padding,
          child: child,
        );

      case MeloopAppStyle.flatSquircle:
        final bg = tintColor != null
            ? Color.lerp(Colors.white, colorTheme.softCardTint, 0.80)!
            : Colors.white;
        return Container(
          decoration: BoxDecoration(
            color: bg,
            borderRadius: BorderRadius.circular(22),
            border: Border.all(
              color: colorTheme.primary.withValues(alpha: 0.18),
              width: 1.2,
            ),
            boxShadow: [
              BoxShadow(
                color: colorTheme.primary.withValues(alpha: 0.10),
                blurRadius: shadowBlur * 1.05,
                offset: const Offset(0, 6),
              ),
            ],
          ),
          padding: padding,
          child: child,
        );

      case MeloopAppStyle.frutigerAero:
        final baseTint = tintColor ?? colorTheme.softCardTint;
        return Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                Colors.white.withValues(alpha: 0.94),
                Color.lerp(Colors.white, baseTint, 0.55)!.withValues(
                  alpha: 0.88,
                ),
                Color.lerp(baseTint, colorTheme.headerTop, 0.18)!.withValues(
                  alpha: 0.84,
                ),
              ],
              stops: const [0.0, 0.45, 1.0],
            ),
            borderRadius: BorderRadius.circular(18),
            border: Border.all(
              color: Colors.white.withValues(alpha: 0.92),
              width: 1.8,
            ),
            boxShadow: [
              BoxShadow(
                color: colorTheme.primary.withValues(alpha: 0.22),
                blurRadius: shadowBlur * 1.15,
                offset: const Offset(0, 6),
              ),
              BoxShadow(
                color: Colors.white.withValues(alpha: 0.55),
                blurRadius: 1,
                spreadRadius: -1,
                offset: const Offset(0, -1),
              ),
            ],
          ),
          padding: padding,
          child: child,
        );

      case MeloopAppStyle.neonNight:
        final baseDark = tintColor != null
            ? Color.lerp(const Color(0xFF181F2C), colorTheme.primary, 0.12)!
            : const Color(0xFF161C27);
        return Theme(
          data: Theme.of(context).copyWith(
            brightness: Brightness.dark,
            textTheme: Theme.of(context).textTheme.apply(
              bodyColor: const Color(0xFFECF4F3),
              displayColor: Colors.white,
            ),
          ),
          child: DefaultTextStyle.merge(
            style: const TextStyle(color: Color(0xFFEAF3F2)),
            child: Container(
              decoration: BoxDecoration(
                color: baseDark,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: colorTheme.headerTop.withValues(alpha: 0.55),
                  width: 1.4,
                ),
                boxShadow: [
                  BoxShadow(
                    color: colorTheme.primary.withValues(alpha: 0.24),
                    blurRadius: shadowBlur,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              padding: padding,
              child: child,
            ),
          ),
        );

      case MeloopAppStyle.paper:
        return Padding(padding: padding, child: child);
    }
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
    final themeCtrl = MeloopThemeController.instance;
    return AnimatedBuilder(
      animation: themeCtrl,
      builder: (context, _) {
        if (themeCtrl.appStyle != MeloopAppStyle.paper) {
          return child;
        }
        final leftHeight = leftWidth * _tapeAspect;
        final rightHeight = rightWidth * _tapeAspect;

        return Stack(
          clipBehavior: Clip.none,
          fit: StackFit.passthrough,
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
      },
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
/// Logo vectorial de Meloop renderizado directamente desde las curvas oficiales
/// de `assets/imagenes/svg_meloop_invert.svg` y `assets/imagenes/svg_meloop.svg`:
/// - En temas claros/coloridos usa la variante invertida (letras y detalles del
///   vinilo en el color del tema + contorno blanco de alto contraste).
/// - En temas oscuros (`Neón nocturno` / `Medianoche`) usa la variante original
///   (letras blancas + contorno y disco en el acento del tema).
/// - Con `whiteOnly: true` renderiza la silueta blanca con recortes limpios.
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

  static const double _viewBoxWidth = 519.0;
  static const double _viewBoxHeight = 152.0;

  static Future<void> preload(
    BuildContext context, {
    double height = 56,
  }) async {
    // El logo ahora es 100% vectorial: no requiere decodificar PNGs.
    _MeloopLogoPaths.ensureInitialized();
  }

  @override
  Widget build(BuildContext context) {
    final width = height * (_viewBoxWidth / _viewBoxHeight);

    if (whiteOnly) {
      return RepaintBoundary(
        child: CustomPaint(
          size: Size(width, height),
          painter: const _MeloopLogoPainter(
            letterFill: Colors.white,
            outlineAndDisc: Colors.transparent,
            whiteOnlyCutout: true,
          ),
        ),
      );
    }

    final themeCtrl = MeloopThemeController.instance;
    return AnimatedBuilder(
      animation: themeCtrl,
      builder: (context, _) {
        final isDark = themeCtrl.isDarkTheme;
        // Claro/colorido (svg_meloop_invert): letras del color del tema + borde blanco.
        // Oscuro (svg_meloop original): letras blancas + borde/disco en tono oscuro del tema.
        final Color letterFill = isDark
            ? Colors.white
            : themeCtrl.primaryColor;
        final Color outlineAndDisc = isDark
            ? themeCtrl.pillTrackColor
            : Colors.white;

        return RepaintBoundary(
          child: CustomPaint(
            size: Size(width, height),
            painter: _MeloopLogoPainter(
              letterFill: letterFill,
              outlineAndDisc: outlineAndDisc,
              whiteOnlyCutout: false,
            ),
          ),
        );
      },
    );
  }
}

class _MeloopLogoPainter extends CustomPainter {
  const _MeloopLogoPainter({
    required this.letterFill,
    required this.outlineAndDisc,
    required this.whiteOnlyCutout,
  });

  final Color letterFill;
  final Color outlineAndDisc;
  final bool whiteOnlyCutout;

  @override
  void paint(Canvas canvas, Size size) {
    final paths = _MeloopLogoPaths.ensureInitialized();
    final scale = size.height / MeloopLogoSticker._viewBoxHeight;

    canvas.save();
    canvas.scale(scale, scale);

    if (whiteOnlyCutout) {
      canvas.saveLayer(
        const Rect.fromLTWH(
          0,
          0,
          MeloopLogoSticker._viewBoxWidth,
          MeloopLogoSticker._viewBoxHeight,
        ),
        Paint(),
      );
      final whitePaint = Paint()
        ..color = letterFill
        ..style = PaintingStyle.fill
        ..isAntiAlias = true;
      canvas.drawPath(paths.letters, whitePaint);
      canvas.drawPath(paths.vinylDisc, whitePaint);

      final clearPaint = Paint()
        ..blendMode = BlendMode.clear
        ..style = PaintingStyle.fill
        ..isAntiAlias = true;
      canvas.drawPath(paths.vinylDetailA, clearPaint);
      canvas.drawPath(paths.vinylDetailB, clearPaint);
      canvas.drawPath(paths.vinylCenterRing, clearPaint);
      canvas.restore();
      canvas.restore();
      return;
    }

    // 1. Contorno exterior redondeado (stroke-width="22" en viewBox 519x152)
    final outlinePaint = Paint()
      ..color = outlineAndDisc
      ..style = PaintingStyle.stroke
      ..strokeWidth = 22.0
      ..strokeJoin = StrokeJoin.round
      ..strokeCap = StrokeCap.round
      ..isAntiAlias = true;
    canvas.drawPath(paths.letters, outlinePaint);

    // 2. Relleno principal de las letras (cubre la mitad interior del trazo)
    final fillPaint = Paint()
      ..color = letterFill
      ..style = PaintingStyle.fill
      ..isAntiAlias = true;
    canvas.drawPath(paths.letters, fillPaint);

    // 3. Disco de vinilo integrado en la segunda "o"
    final discPaint = Paint()
      ..color = outlineAndDisc
      ..style = PaintingStyle.fill
      ..isAntiAlias = true;
    canvas.drawPath(paths.vinylDisc, discPaint);

    // 4. Brillos del vinilo y etiqueta central (con orificio central evenOdd)
    canvas.drawPath(paths.vinylDetailA, fillPaint);
    canvas.drawPath(paths.vinylDetailB, fillPaint);
    canvas.drawPath(paths.vinylCenterRing, fillPaint);

    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant _MeloopLogoPainter oldDelegate) =>
      oldDelegate.letterFill != letterFill ||
      oldDelegate.outlineAndDisc != outlineAndDisc ||
      oldDelegate.whiteOnlyCutout != whiteOnlyCutout;
}

class _MeloopLogoPaths {
  _MeloopLogoPaths._({
    required this.letters,
    required this.vinylDisc,
    required this.vinylDetailA,
    required this.vinylDetailB,
    required this.vinylCenterRing,
  });

  final Path letters;
  final Path vinylDisc;
  final Path vinylDetailA;
  final Path vinylDetailB;
  final Path vinylCenterRing;

  static _MeloopLogoPaths? _cached;

  static _MeloopLogoPaths ensureInitialized() {
    return _cached ??= _MeloopLogoPaths._(
      letters: _parseSvgPath(_lettersSvgData),
      vinylDisc: _parseSvgPath(_vinylDiscSvgData),
      vinylDetailA: _parseSvgPath(_vinylDetailASvgData),
      vinylDetailB: _parseSvgPath(_vinylDetailBSvgData),
      vinylCenterRing: _parseSvgPath(
        _vinylCenterRingSvgData,
        fillType: PathFillType.evenOdd,
      ),
    );
  }

  static const String _lettersSvgData =
      'M11 113.196V41.3156C11 32.4035 13.7028 25.1229 19.1085 19.4738C24.5142 13.8246 31.5269 11 40.1468 11C50.3738 11 58.2388 15.6995 63.7418 25.0986C67.1021 20.0338 70.7059 16.43 74.5532 14.2872C78.4005 12.0957 83.1487 11 88.7979 11C97.7586 11 104.625 13.8246 109.398 19.4738C114.17 25.1229 116.557 33.2558 116.557 43.8723V113.196H86.8986V48.3283C86.8986 45.1142 86.6307 42.9957 86.095 41.973C85.608 40.9016 84.5853 40.3659 83.0269 40.3659C80.105 40.3659 78.644 42.8496 78.644 47.817V113.196H48.9858V48.3283C48.9858 45.1629 48.718 43.0444 48.1823 41.973C47.6466 40.9016 46.5995 40.3659 45.0411 40.3659C42.1191 40.3659 40.6582 43.0201 40.6582 48.3283V113.196H11Z'
      'M204.143 81.2007H162.724C162.091 79.1066 161.774 77.1829 161.774 75.4297C161.774 72.9948 162.213 70.6085 163.089 68.2709H176.969C175.995 62.8652 172.951 60.1624 167.838 60.1624C164.429 60.1624 161.653 61.6234 159.51 64.5453C157.367 67.4186 156.296 71.1198 156.296 75.6489C156.296 80.3728 157.318 84.1957 159.364 87.1177C161.458 90.0397 164.185 91.5006 167.545 91.5006C170.273 91.5006 172.902 90.1371 175.435 87.4099L190.994 106.111C183.738 111.808 175.678 114.657 166.815 114.657C156.004 114.657 146.921 111.029 139.567 103.773C132.214 96.5167 128.537 87.5316 128.537 76.8177C128.537 66.1524 132.238 57.143 139.64 49.7893C147.091 42.4357 156.223 38.7588 167.034 38.7588C177.602 38.7588 186.465 42.3626 193.624 49.5702C200.832 56.729 204.435 65.568 204.435 76.0872C204.435 77.1586 204.338 78.8631 204.143 81.2007Z'
      'M215.539 13.4837H245.197V72.7269C245.197 77.3534 245.928 80.4458 247.389 82.0042C248.898 83.5139 251.845 84.2687 256.228 84.2687H257.762V113.196H249.872C239.061 113.196 230.636 110.055 224.597 103.773C218.558 97.442 215.539 88.6761 215.539 77.4751V13.4837Z'
      'M300.35 38.8319C310.674 38.8319 319.513 42.5574 326.867 50.0085C334.269 57.4108 337.97 66.2985 337.97 76.6716C337.97 87.1907 334.22 96.1515 326.721 103.554C319.269 110.956 310.26 114.657 299.692 114.657C289.124 114.657 280.091 110.956 272.591 103.554C265.091 96.1028 261.341 87.142 261.341 76.6716C261.341 66.0063 265.091 57.0456 272.591 49.7893C280.091 42.4844 289.343 38.8319 300.35 38.8319ZM299.619 65.7872C296.697 65.7872 294.213 66.8586 292.168 69.0014C290.123 71.0955 289.1 73.6765 289.1 76.7446C289.1 79.764 290.123 82.3451 292.168 84.4879C294.262 86.6307 296.746 87.7021 299.619 87.7021C302.541 87.7021 305.025 86.6307 307.07 84.4879C309.164 82.3451 310.211 79.764 310.211 76.7446C310.211 73.6765 309.189 71.0955 307.143 69.0014C305.098 66.8586 302.59 65.7872 299.619 65.7872Z'
      'M383.918 38.8319C394.243 38.8319 403.082 42.5574 410.435 50.0085C417.838 57.4108 421.539 66.2985 421.539 76.6716C421.539 87.1907 417.789 96.1515 410.289 103.554C402.838 110.956 393.829 114.657 383.261 114.657C372.693 114.657 363.659 110.956 356.16 103.554C348.66 96.1028 344.91 87.142 344.91 76.6716C344.91 66.0063 348.66 57.0456 356.16 49.7893C363.659 42.4844 372.912 38.8319 383.918 38.8319ZM383.188 65.7872C380.266 65.7872 377.782 66.8586 375.737 69.0014C373.691 71.0955 372.669 73.6765 372.669 76.7446C372.669 79.764 373.691 82.3451 375.737 84.4879C377.831 86.6307 380.315 87.7021 383.188 87.7021C386.11 87.7021 388.594 86.6307 390.639 84.4879C392.733 82.3451 393.78 79.764 393.78 76.7446C393.78 73.6765 392.757 71.0955 390.712 69.0014C388.667 66.8586 386.159 65.7872 383.188 65.7872Z'
      'M458.648 140.152H431.328V79.3744C431.328 68.2222 434.25 59.1397 440.094 52.1269C443.454 48.0848 447.715 44.8707 452.877 42.4844C458.088 40.0981 463.421 38.9049 468.875 38.9049C479.735 38.9049 488.891 42.5574 496.342 49.8624C503.793 57.1673 507.518 66.1281 507.518 76.7446C507.518 87.069 503.817 95.981 496.415 103.481C489.013 110.981 480.247 114.73 470.117 114.73C468.461 114.73 466.026 114.56 462.812 114.219V84.1226C465.052 85.9245 467.292 86.8255 469.533 86.8255C472.357 86.8255 474.768 85.7784 476.765 83.6843C478.81 81.5903 479.833 79.1066 479.833 76.2333C479.833 73.36 478.786 70.925 476.691 68.9283C474.597 66.8829 472.089 65.8602 469.167 65.8602C462.155 65.8602 458.648 70.7059 458.648 80.3971V140.152Z';

  static const String _vinylDiscSvgData =
      'M380.748 32.5271C405.125 31.5124 425.707 50.4527 426.718 74.8294C427.73 99.2063 408.787 119.787 384.41 120.795C360.038 121.803 339.463 102.864 338.452 78.4918C337.441 54.1198 356.376 33.5417 380.748 32.5271Z';

  static const String _vinylDetailASvgData =
      'M364.003 89.3006C364.167 89.374 364.728 90.1762 364.871 90.3681C366.299 92.2844 368.052 93.8057 369.975 95.1939C369.02 96.7069 367.758 98.4531 366.73 99.9618L360.286 109.414C359.429 108.82 358.592 108.197 357.777 107.546C354.888 105.223 351.898 102.066 349.845 98.9589C351.375 97.8701 353.061 96.7613 354.621 95.6972L364.003 89.3006Z';

  static const String _vinylDetailBSvgData =
      'M404.886 43.9157C408.99 46.7163 412.533 50.2616 415.33 54.368C411.493 56.9085 407.608 59.6289 403.795 62.2276L401.116 64.0557C399.367 61.5266 397.643 60.0186 395.192 58.2318L404.886 43.9157Z';

  static const String _vinylCenterRingSvgData =
      'M381.032 59.343C390.593 58.4907 399.032 65.5576 399.872 75.12C400.711 84.6824 393.634 93.1123 384.071 93.9398C374.525 94.7658 366.114 87.7041 365.275 78.159C364.437 68.6142 371.488 60.1938 381.032 59.343ZM386.437 76.1646C386.164 74.0443 384.229 72.5432 382.108 72.8061C379.972 73.0708 378.459 75.0214 378.734 77.1556C379.008 79.2896 380.966 80.7939 383.099 80.5097C385.218 80.2269 386.71 78.2848 386.437 76.1646Z';

  static final RegExp _tokenPattern = RegExp(
    r'[MLHVCZ]|-?(?:\d+\.?\d*|\.\d+)',
  );

  static Path _parseSvgPath(
    String data, {
    PathFillType fillType = PathFillType.nonZero,
  }) {
    final path = Path()..fillType = fillType;
    final tokens = _tokenPattern
        .allMatches(data)
        .map((match) => match.group(0)!)
        .toList();

    var index = 0;
    var command = '';
    var currentX = 0.0;
    var currentY = 0.0;

    bool isCommand(String token) =>
        token.length == 1 && 'MLHVCZ'.contains(token);

    double nextNumber() => double.parse(tokens[index++]);

    while (index < tokens.length) {
      if (isCommand(tokens[index])) {
        command = tokens[index++];
        if (command == 'Z') {
          path.close();
          continue;
        }
      }
      switch (command) {
        case 'M':
          currentX = nextNumber();
          currentY = nextNumber();
          path.moveTo(currentX, currentY);
          command = 'L';
        case 'L':
          currentX = nextNumber();
          currentY = nextNumber();
          path.lineTo(currentX, currentY);
        case 'H':
          currentX = nextNumber();
          path.lineTo(currentX, currentY);
        case 'V':
          currentY = nextNumber();
          path.lineTo(currentX, currentY);
        case 'C':
          final x1 = nextNumber();
          final y1 = nextNumber();
          final x2 = nextNumber();
          final y2 = nextNumber();
          currentX = nextNumber();
          currentY = nextNumber();
          path.cubicTo(x1, y1, x2, y2, currentX, currentY);
      }
    }
    return path;
  }
}
