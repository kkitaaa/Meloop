import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';

/// Carátulas de álbumes flotantes con efecto paralaje para el fondo del inicio
/// de sesión:
/// - 3 planos de profundidad (más grandes = más cerca = más rápidos), con tope
///   estricto de 360px para mantener bajo consumo de CPU/GPU.
/// - Carriles horizontales desfasados entre planos + selección inteligente de
///   carril para repartir las portadas de forma pareja y evitar que una tape
///   completamente a otra.
/// - Se desplazan hacia arriba: nacen por debajo del borde inferior y se
///   destruyen (despawn) al salir completamente por la parte superior.
/// - El orden de aparición sigue la numeración de los archivos placeholder
///   (`image.png`, `image copy.png`, `image copy 2.png` ... `image copy 28.png`).
///
/// TODO(backend): En el futuro reemplazar `_placeholderAlbumCovers` por URLs de
/// carátulas obtenidas desde Spotify (vía music-service / recomendaciones).
class ParallaxAlbumCovers extends StatefulWidget {
  const ParallaxAlbumCovers({super.key, required this.animate});

  final bool animate;

  static const int decodeSizePx = 280;

  // Orden exacto según número de copia en assets/textures/PLACEHOLDER_ALBUMCOVER/
  static const List<String> placeholderAlbumCovers = [
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 2.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 3.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 4.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 5.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 6.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 7.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 8.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 9.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 10.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 12.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 14.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 15.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 16.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 17.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 18.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 19.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 21.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 22.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 23.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 24.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 25.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 26.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 27.png',
    'assets/textures/PLACEHOLDER_ALBUMCOVER/image copy 28.png',
  ];

  /// Precarga las carátulas en lotes pequeños (de 4 en 4) para evitar picos
  /// altos de memoria nativa al descomprimir los PNG grandes en paralelo.
  static Future<void> preload(BuildContext context) async {
    const batchSize = 4;
    for (var i = 0; i < placeholderAlbumCovers.length; i += batchSize) {
      if (!context.mounted) return;
      final end = math.min(i + batchSize, placeholderAlbumCovers.length);
      await Future.wait(
        placeholderAlbumCovers
            .sublist(i, end)
            .map(
              (assetPath) => precacheImage(
                ResizeImage.resizeIfNeeded(
                  decodeSizePx,
                  decodeSizePx,
                  AssetImage(assetPath),
                ),
                context,
                onError: (_, _) {},
              ),
            ),
      );
    }
  }

  /// Libera de `ImageCache` las 26 carátulas del login al entrar a la app para
  /// no retener RAM innecesaria durante el resto de la sesión.
  static void evictFromCache() {
    for (final assetPath in placeholderAlbumCovers) {
      ResizeImage.resizeIfNeeded(
        decodeSizePx,
        decodeSizePx,
        AssetImage(assetPath),
      ).evict();
    }
    PaintingBinding.instance.imageCache.clearLiveImages();
  }

  @override
  State<ParallaxAlbumCovers> createState() => _ParallaxAlbumCoversState();
}

class _ParallaxAlbumCoversState extends State<ParallaxAlbumCovers>
    with SingleTickerProviderStateMixin {

  // Planos de paralaje con carriles horizontales desfasados entre sí para que
  // las portadas se repartan de forma uniforme y solo solapen bordes/esquinas.
  static const _depthSpecs = <({
    double minSize,
    double maxSize,
    double speedPxPerSec,
    double opacity,
    double shadowBlur,
    List<double> laneCenters,
  })>[
    // Plano 0: Fondo (pequeño-mediano, deriva ambiental muy suave)
    (
      minSize: 170,
      maxSize: 210,
      speedPxPerSec: 11,
      opacity: 0.78,
      shadowBlur: 12,
      laneCenters: [0.06, 0.27, 0.48, 0.70, 0.91],
    ),
    // Plano 1: Medio (mediano-grande, velocidad media calma)
    (
      minSize: 240,
      maxSize: 285,
      speedPxPerSec: 19,
      opacity: 0.92,
      shadowBlur: 18,
      laneCenters: [0.16, 0.39, 0.62, 0.85],
    ),
    // Plano 2: Primer plano (grande, movimiento fluido sin distraer a 60+ fps)
    (
      minSize: 315,
      maxSize: 360,
      speedPxPerSec: 30,
      opacity: 1.0,
      shadowBlur: 26,
      laneCenters: [0.09, 0.35, 0.57, 0.80],
    ),
  ];

  final List<_ParallaxAlbumInstance> _activeAlbums = [];
  final math.Random _random = math.Random(42);
  Ticker? _ticker;
  Duration? _lastTick;
  Size _viewportSize = Size.zero;
  int _nextAssetIndex = 0;
  int _nextInstanceId = 0;

  @override
  void initState() {
    super.initState();
    if (widget.animate) {
      _ticker = createTicker(_onTick)..start();
    }
  }

  @override
  void didUpdateWidget(covariant ParallaxAlbumCovers oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.animate == oldWidget.animate) return;
    if (widget.animate) {
      _lastTick = null;
      _ticker ??= createTicker(_onTick);
      if (!_ticker!.isActive) {
        _ticker!.start();
      }
    } else {
      _ticker?.stop();
      _lastTick = null;
    }
  }

  @override
  void dispose() {
    _ticker?.dispose();
    ParallaxAlbumCovers.evictFromCache();
    super.dispose();
  }

  void _ensureInitialized(Size size) {
    if (size.width <= 0 || size.height <= 0) return;
    _viewportSize = size;
    if (_activeAlbums.isNotEmpty) return;

    // 22 portadas iniciales distribuidas de manera pareja por todo el lienzo
    // (sin taparse completamente unas a otras) en orden secuencial de copia.
    const initialSlots = <({int depth, int lane, double yFraction})>[
      // Plano 0 (8 portadas)
      (depth: 0, lane: 0, yFraction: -0.02),
      (depth: 0, lane: 2, yFraction: 0.05),
      (depth: 0, lane: 4, yFraction: 0.02),
      (depth: 0, lane: 1, yFraction: 0.36),
      (depth: 0, lane: 3, yFraction: 0.42),
      (depth: 0, lane: 0, yFraction: 0.72),
      (depth: 0, lane: 2, yFraction: 0.78),
      (depth: 0, lane: 4, yFraction: 0.68),

      // Plano 1 (8 portadas)
      (depth: 1, lane: 0, yFraction: 0.06),
      (depth: 1, lane: 2, yFraction: 0.14),
      (depth: 1, lane: 1, yFraction: 0.46),
      (depth: 1, lane: 3, yFraction: 0.34),
      (depth: 1, lane: 0, yFraction: 0.82),
      (depth: 1, lane: 2, yFraction: 0.86),
      (depth: 1, lane: 1, yFraction: 1.04),
      (depth: 1, lane: 3, yFraction: 0.96),

      // Plano 2 (6 portadas grandes hasta 360px)
      (depth: 2, lane: 1, yFraction: -0.04),
      (depth: 2, lane: 3, yFraction: 0.10),
      (depth: 2, lane: 0, yFraction: 0.40),
      (depth: 2, lane: 2, yFraction: 0.54),
      (depth: 2, lane: 1, yFraction: 0.84),
      (depth: 2, lane: 3, yFraction: 0.92),
    ];

    for (final slot in initialSlots) {
      _spawnAlbum(
        depth: slot.depth,
        y: size.height * slot.yFraction,
        viewportWidth: size.width,
        forcedLane: slot.lane,
      );
    }
    _sortByDepth();
  }

  void _spawnAlbum({
    required int depth,
    required double y,
    required double viewportWidth,
    int? forcedLane,
  }) {
    final spec = _depthSpecs[depth];
    final assetPath =
        ParallaxAlbumCovers.placeholderAlbumCovers[_nextAssetIndex %
            ParallaxAlbumCovers.placeholderAlbumCovers.length];
    _nextAssetIndex++;

    // Cuantizamos el tamaño a pasos discretos (múltiplos de 24px) para que Skia
    // reutilice las mismas máscaras de sombra en GPU en lugar de acumular
    // cientos de texturas únicas por cada portada que reaparece.
    final rawSize =
        spec.minSize + _random.nextDouble() * (spec.maxSize - spec.minSize);
    final double size = ((rawSize / 24.0).round() * 24.0)
        .clamp(spec.minSize, 360.0)
        .toDouble();

    final laneIdx =
        forcedLane ??
        _pickBestLane(depth: depth, y: y, size: size, width: viewportWidth);

    final centerFraction = spec.laneCenters[laneIdx % spec.laneCenters.length];
    final jitter =
        (_random.nextDouble() - 0.5) * math.min(48.0, viewportWidth * 0.03);
    // Permite sangrado sutil en los bordes izquierdo y derecho como en el mockup
    final minX = -size * 0.18;
    final maxX = viewportWidth - size * 0.82;
    final double x = (centerFraction * viewportWidth - size / 2 + jitter)
        .clamp(minX, math.max(minX, maxX))
        .toDouble();

    final sizeRatio =
        (size - spec.minSize) / math.max(1.0, spec.maxSize - spec.minSize);
    final speed = spec.speedPxPerSec * (0.93 + sizeRatio * 0.15);

    _activeAlbums.add(
      _ParallaxAlbumInstance(
        id: _nextInstanceId++,
        assetPath: assetPath,
        depth: depth,
        lane: laneIdx,
        x: x,
        y: y,
        size: size,
        speed: speed,
        opacity: spec.opacity,
        shadowBlur: spec.shadowBlur,
      ),
    );
  }

  /// Elige el carril más despejado en la parte inferior para repartir las
  /// portadas uniformemente y evitar que una portada tape completamente a otra.
  int _pickBestLane({
    required int depth,
    required double y,
    required double size,
    required double width,
  }) {
    final spec = _depthSpecs[depth];
    var bestLane = 0;
    var bestScore = double.infinity;

    for (var lane = 0; lane < spec.laneCenters.length; lane++) {
      final candidateCenterX = spec.laneCenters[lane] * width;
      final candidateCenterY = y + size / 2;
      var penalty = 0.0;

      for (final other in _activeAlbums) {
        final otherCenterX = other.x + other.size / 2;
        final otherCenterY = other.y + other.size / 2;
        final dx = (candidateCenterX - otherCenterX).abs();
        final dy = (candidateCenterY - otherCenterY).abs();

        // Evita amontonamiento vertical dentro del mismo plano y carril
        if (other.depth == depth && other.lane == lane) {
          final verticalGap = y - (other.y + other.size);
          if (verticalGap < size * 0.65) {
            penalty += (size * 0.65 - verticalGap) * 4.0 + 400.0;
          } else {
            // Prefiere carriles cuyo último álbum ya subió más arriba
            penalty += math.max(0.0, 320.0 - verticalGap * 0.35);
          }
        }

        // Evita que un álbum nazca centrado justo sobre otro de distinto plano
        final minAllowedDx = (size + other.size) * 0.28;
        final minAllowedDy = (size + other.size) * 0.42;
        if (dx < minAllowedDx && dy < minAllowedDy) {
          penalty += (minAllowedDx - dx) * 2.5 + 180.0;
        }
      }

      // Pequeña variación aleatoria para que no se vuelva robótico
      penalty += _random.nextDouble() * 18.0;
      if (penalty < bestScore) {
        bestScore = penalty;
        bestLane = lane;
      }
    }

    return bestLane;
  }

  void _sortByDepth() {
    _activeAlbums.sort((a, b) {
      final depthCmp = a.depth.compareTo(b.depth);
      if (depthCmp != 0) return depthCmp;
      return a.size.compareTo(b.size);
    });
  }

  void _onTick(Duration elapsed) {
    final previous = _lastTick;
    _lastTick = elapsed;
    if (previous == null || _viewportSize.isEmpty || _activeAlbums.isEmpty) {
      return;
    }

    final dt = ((elapsed - previous).inMicroseconds / 1000000.0).clamp(
      0.0,
      0.05,
    );
    if (dt <= 0) return;

    final respawnDepths = <int>[];
    for (var i = _activeAlbums.length - 1; i >= 0; i--) {
      final album = _activeAlbums[i];
      album.y -= album.speed * dt;
      if (album.y + album.size < -36.0) {
        respawnDepths.add(album.depth);
        _activeAlbums.removeAt(i);
      }
    }

    if (respawnDepths.isNotEmpty) {
      for (final depth in respawnDepths) {
        final spawnY = _viewportSize.height + 24.0 + _random.nextDouble() * 44.0;
        _spawnAlbum(
          depth: depth,
          y: spawnY,
          viewportWidth: _viewportSize.width,
        );
      }
      _sortByDepth();
    }

    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final size = Size(constraints.maxWidth, constraints.maxHeight);
        _ensureInitialized(size);

        return ClipRect(
          child: IgnorePointer(
            child: Stack(
              fit: StackFit.expand,
              children: [
                for (final album in _activeAlbums)
                  Positioned(
                    key: ValueKey<int>(album.id),
                    left: 0,
                    top: 0,
                    child: Transform.translate(
                      offset: Offset(album.x, album.y),
                      child: Opacity(
                        opacity: album.opacity,
                        child: Container(
                          width: album.size,
                          height: album.size,
                          decoration: BoxDecoration(
                            color: const Color(0xFF0E3D4B),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withValues(alpha: 0.34),
                                blurRadius: album.shadowBlur,
                                offset: const Offset(0, 8),
                              ),
                            ],
                          ),
                          child: Image.asset(
                            album.assetPath,
                            width: album.size,
                            height: album.size,
                            cacheWidth: ParallaxAlbumCovers.decodeSizePx,
                            cacheHeight: ParallaxAlbumCovers.decodeSizePx,
                            fit: BoxFit.cover,
                            gaplessPlayback: true,
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
            ),
          ),
        );
      },
    );
  }
}

class _ParallaxAlbumInstance {
  _ParallaxAlbumInstance({
    required this.id,
    required this.assetPath,
    required this.depth,
    required this.lane,
    required this.x,
    required this.y,
    required this.size,
    required this.speed,
    required this.opacity,
    required this.shadowBlur,
  });

  final int id;
  final String assetPath;
  final int depth;
  final int lane;
  final double x;
  double y;
  final double size;
  final double speed;
  final double opacity;
  final double shadowBlur;
}
