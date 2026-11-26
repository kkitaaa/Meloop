import 'dart:convert';

import 'package:flutter/material.dart';

/// Estilos visuales globales de la aplicación desbloqueables por nivel.
enum MeloopAppStyle {
  paper(
    id: 'paper',
    label: 'Papel clásico',
    description: 'Hojas de papel apiladas con textura física y pines.',
    requiredLevel: 1,
  ),
  flatSquare(
    id: 'flat_square',
    label: 'Plano cuadrado',
    description: 'Tarjetas limpias de bordes rectos y estilo minimalista.',
    requiredLevel: 3,
  ),
  flatSquircle(
    id: 'flat_squircle',
    label: 'Squircle moderno',
    description: 'Superficies suaves con esquinas redondeadas tipo cápsula.',
    requiredLevel: 6,
  ),
  frutigerAero(
    id: 'frutiger_aero',
    label: 'Frutiger Aero',
    description: 'Cristal brillante, reflejos acuáticos y bordes luminosos.',
    requiredLevel: 10,
  ),
  neonNight(
    id: 'neon_night',
    label: 'Neón nocturno',
    description: 'Tarjetas oscuras elegantes con contorno luminoso.',
    requiredLevel: 14,
  );

  const MeloopAppStyle({
    required this.id,
    required this.label,
    required this.description,
    required this.requiredLevel,
  });

  final String id;
  final String label;
  final String description;
  final int requiredLevel;

  static MeloopAppStyle fromId(String? raw) {
    for (final value in values) {
      if (value.id == raw) return value;
    }
    return MeloopAppStyle.paper;
  }
}

/// Paletas de color globales de la aplicación desbloqueables por nivel.
enum MeloopColorTheme {
  meloopTeal(
    id: 'meloop_teal',
    label: 'Verde Meloop',
    requiredLevel: 1,
    headerTop: Color(0xFF59E5C9),
    headerBottom: Color(0xFF00AF92),
    primary: Color(0xFF00AF92),
    pillTrack: Color(0xFF16877F),
    softCardTint: Color(0xFFF2FAF8),
    xpCardTint: Color(0xFFE5F6FC),
    songCardTint: Color(0xFFDCFFE9),
    bgStart: Color(0xFF248686),
    bgEnd: Color(0xFF004562),
    gridBg: Color(0xFFF4FBF9),
    gridLine: Color(0xFFCDEBE5),
  ),
  sakuraPink(
    id: 'sakura_pink',
    label: 'Rosa Pastel',
    requiredLevel: 3,
    headerTop: Color(0xFFFFB6D9),
    headerBottom: Color(0xFFFF79B8),
    primary: Color(0xFFF062A4),
    pillTrack: Color(0xFFD95593),
    softCardTint: Color(0xFFFFF1F7),
    xpCardTint: Color(0xFFFFEBF4),
    songCardTint: Color(0xFFFFEDF6),
    bgStart: Color(0xFFFFB2D6),
    bgEnd: Color(0xFFB74F84),
    gridBg: Color(0xFFFFF8FB),
    gridLine: Color(0xFFFFD1E6),
  ),
  oceanAero(
    id: 'ocean_aero',
    label: 'Azul Aero',
    requiredLevel: 5,
    headerTop: Color(0xFF7BE0FF),
    headerBottom: Color(0xFF1E96E8),
    primary: Color(0xFF1D9BF0),
    pillTrack: Color(0xFF1472B8),
    softCardTint: Color(0xFFEDF8FF),
    xpCardTint: Color(0xFFE0F4FF),
    songCardTint: Color(0xFFDDF7F3),
    bgStart: Color(0xFF39B7E8),
    bgEnd: Color(0xFF0C4A8A),
    gridBg: Color(0xFFF3FAFF),
    gridLine: Color(0xFFC5E7FA),
  ),
  lavenderDream(
    id: 'lavender_dream',
    label: 'Lavanda',
    requiredLevel: 7,
    headerTop: Color(0xFFC5B4FF),
    headerBottom: Color(0xFF8B6CE8),
    primary: Color(0xFF8363E6),
    pillTrack: Color(0xFF6747C4),
    softCardTint: Color(0xFFF6F2FF),
    xpCardTint: Color(0xFFEEE8FF),
    songCardTint: Color(0xFFF3ECFF),
    bgStart: Color(0xFF8B6FE0),
    bgEnd: Color(0xFF3A256F),
    gridBg: Color(0xFFFAF8FF),
    gridLine: Color(0xFFE2D8FF),
  ),
  sunsetPeach(
    id: 'sunset_peach',
    label: 'Durazno Coral',
    requiredLevel: 9,
    headerTop: Color(0xFFFFBE94),
    headerBottom: Color(0xFFF27A54),
    primary: Color(0xFFE86C47),
    pillTrack: Color(0xFFC95432),
    softCardTint: Color(0xFFFFF5EF),
    xpCardTint: Color(0xFFFFEFE6),
    songCardTint: Color(0xFFFFF2DF),
    bgStart: Color(0xFFE8865C),
    bgEnd: Color(0xFF752F38),
    gridBg: Color(0xFFFFFAF6),
    gridLine: Color(0xFFFFDEC9),
  ),
  midnightViolet(
    id: 'midnight_violet',
    label: 'Medianoche',
    requiredLevel: 12,
    headerTop: Color(0xFF7D8BFF),
    headerBottom: Color(0xFF4F56D9),
    primary: Color(0xFF6366F1),
    pillTrack: Color(0xFF3B40A8),
    softCardTint: Color(0xFFF1F2FF),
    xpCardTint: Color(0xFFE8EAFF),
    songCardTint: Color(0xFFE6F0FF),
    bgStart: Color(0xFF282F68),
    bgEnd: Color(0xFF0E1128),
    gridBg: Color(0xFF14182B),
    gridLine: Color(0xFF272F52),
  );

  const MeloopColorTheme({
    required this.id,
    required this.label,
    required this.requiredLevel,
    required this.headerTop,
    required this.headerBottom,
    required this.primary,
    required this.pillTrack,
    required this.softCardTint,
    required this.xpCardTint,
    required this.songCardTint,
    required this.bgStart,
    required this.bgEnd,
    required this.gridBg,
    required this.gridLine,
  });

  final String id;
  final String label;
  final int requiredLevel;
  final Color headerTop;
  final Color headerBottom;
  final Color primary;
  final Color pillTrack;
  final Color softCardTint;
  final Color xpCardTint;
  final Color songCardTint;
  final Color bgStart;
  final Color bgEnd;
  final Color gridBg;
  final Color gridLine;

  static MeloopColorTheme fromId(String? raw) {
    for (final value in values) {
      if (value.id == raw) return value;
    }
    return MeloopColorTheme.meloopTeal;
  }
}

/// Patrones de fondo globales de la aplicación desbloqueables por nivel.
enum MeloopBackgroundPattern {
  tealRadial(
    id: 'radial_classic',
    label: 'Gradiente clásico',
    requiredLevel: 1,
  ),
  pastelGrid(
    id: 'pastel_grid',
    label: 'Cuadrícula pastel',
    requiredLevel: 3,
  ),
  aeroBubbles(
    id: 'aero_bubbles',
    label: 'Burbujas Aero',
    requiredLevel: 6,
  ),
  dotNotebook(
    id: 'dot_notebook',
    label: 'Libreta de puntos',
    requiredLevel: 8,
  ),
  starlightNight(
    id: 'starlight_night',
    label: 'Noche estrellada',
    requiredLevel: 11,
  );

  const MeloopBackgroundPattern({
    required this.id,
    required this.label,
    required this.requiredLevel,
  });

  final String id;
  final String label;
  final int requiredLevel;

  static MeloopBackgroundPattern fromId(String? raw) {
    for (final value in values) {
      if (value.id == raw) return value;
    }
    return MeloopBackgroundPattern.tealRadial;
  }
}

/// Marcos de foto de perfil (Polaroid) desbloqueables por nivel.
enum ProfileBorderStyle {
  none(
    id: 'none',
    label: 'Sin marco',
    requiredLevel: 1,
    previewColor: Color(0xFFD6E0DE),
  ),
  pastelRibbon(
    id: 'pastel_ribbon',
    label: 'Marco Pastel',
    requiredLevel: 2,
    previewColor: Color(0xFFFF8EC4),
  ),
  goldenFrame(
    id: 'golden_frame',
    label: 'Marco Dorado',
    requiredLevel: 4,
    previewColor: Color(0xFFE5B32E),
  ),
  rainbowRgb(
    id: 'rainbow_rgb',
    label: 'Marco RGB',
    requiredLevel: 4,
    previewColor: Color(0xFF00E5FF),
  ),
  neonCyan(
    id: 'neon_cyan',
    label: 'Marco Neón',
    requiredLevel: 7,
    previewColor: Color(0xFF26F7DF),
  ),
  aeroGlass(
    id: 'aero_glass',
    label: 'Marco Cristal Aero',
    requiredLevel: 10,
    previewColor: Color(0xFF53C6FF),
  );

  const ProfileBorderStyle({
    required this.id,
    required this.label,
    required this.requiredLevel,
    required this.previewColor,
  });

  final String id;
  final String label;
  final int requiredLevel;
  final Color previewColor;

  static ProfileBorderStyle fromId(String? raw) {
    for (final value in values) {
      if (value.id == raw) return value;
    }
    return ProfileBorderStyle.none;
  }
}

/// Pines personalizables de la polaroid y paneles.
class ProfilePinOption {
  const ProfilePinOption({
    required this.id,
    required this.label,
    required this.requiredLevel,
    required this.color,
  });

  final String id;
  final String label;
  final int requiredLevel;
  final Color? color; // null = rojo clásico del asset original
}

const List<ProfilePinOption> kProfilePinCatalog = [
  ProfilePinOption(
    id: 'pin_classic_red',
    label: 'Rojo clásico',
    requiredLevel: 1,
    color: null,
  ),
  ProfilePinOption(
    id: 'pin_sakura',
    label: 'Rosa kawaii',
    requiredLevel: 2,
    color: Color(0xFFFF5C9F),
  ),
  ProfilePinOption(
    id: 'pin_sapphire',
    label: 'Pin Zafiro',
    requiredLevel: 3,
    color: Color(0xFF2B7FFF),
  ),
  ProfilePinOption(
    id: 'pin_gold',
    label: 'Pin Dorado',
    requiredLevel: 5,
    color: Color(0xFFF5B72E),
  ),
  ProfilePinOption(
    id: 'pin_emerald',
    label: 'Pin Esmeralda',
    requiredLevel: 7,
    color: Color(0xFF10B981),
  ),
  ProfilePinOption(
    id: 'pin_amethyst',
    label: 'Pin Amatista',
    requiredLevel: 9,
    color: Color(0xFF9B5CF6),
  ),
  ProfilePinOption(
    id: 'pin_obsidian',
    label: 'Pin Obsidiana',
    requiredLevel: 12,
    color: Color(0xFF28303D),
  ),
];

/// Presets seguros de distribución del perfil (permiten personalizar el layout
/// sin romper la legibilidad ni desordenar la interfaz).
enum ProfileLayoutPreset {
  classic(
    id: 'classic',
    label: 'Clásico equilibrado',
    description: 'Polaroid a la izquierda, Bio y Artistas al centro, Canción y Amigos a la derecha.',
  ),
  showcase(
    id: 'showcase',
    label: 'Vitrina amplia',
    description: 'Da más espacio a tu biografía, artistas y stickers con Canción arriba a la derecha.',
  ),
  compactFocus(
    id: 'compact_focus',
    label: 'Enfoque blogs',
    description: 'Agrupa tu tarjeta y música a la izquierda para destacar tus publicaciones.',
  );

  const ProfileLayoutPreset({
    required this.id,
    required this.label,
    required this.description,
  });

  final String id;
  final String label;
  final String description;

  static ProfileLayoutPreset fromId(String? raw) {
    for (final value in values) {
      if (value.id == raw) return value;
    }
    return ProfileLayoutPreset.classic;
  }
}

/// Sticker colocado libremente sobre el lienzo del perfil (coordenadas
/// normalizadas 0.0..1.0 para adaptarse a distintos tamaños de pantalla).
class ProfileStickerPlacement {
  const ProfileStickerPlacement({
    required this.id,
    required this.stickerId,
    required this.x,
    required this.y,
    this.scale = 1.0,
    this.rotation = 0.0,
  });

  final String id;
  final String stickerId;
  final double x;
  final double y;
  final double scale;
  final double rotation;

  ProfileStickerPlacement copyWith({
    double? x,
    double? y,
    double? scale,
    double? rotation,
  }) {
    return ProfileStickerPlacement(
      id: id,
      stickerId: stickerId,
      x: (x ?? this.x).clamp(0.02, 0.94),
      y: (y ?? this.y).clamp(0.02, 0.92),
      scale: (scale ?? this.scale).clamp(0.65, 1.55),
      rotation: (rotation ?? this.rotation).clamp(-0.78, 0.78),
    );
  }

  Map<String, Object?> toJson() => {
    'id': id,
    'sticker_id': stickerId,
    'x': double.parse(x.toStringAsFixed(4)),
    'y': double.parse(y.toStringAsFixed(4)),
    'scale': double.parse(scale.toStringAsFixed(2)),
    'rotation': double.parse(rotation.toStringAsFixed(3)),
  };

  factory ProfileStickerPlacement.fromJson(Map<String, dynamic> json) {
    double asDouble(Object? value, double fallback) {
      if (value is num) return value.toDouble();
      return fallback;
    }

    return ProfileStickerPlacement(
      id: (json['id'] as String?)?.trim().isNotEmpty == true
          ? json['id'] as String
          : 'stk-${DateTime.now().microsecondsSinceEpoch}',
      stickerId: (json['sticker_id'] as String?) ?? 'cat_shy',
      x: asDouble(json['x'], 0.25).clamp(0.02, 0.94),
      y: asDouble(json['y'], 0.25).clamp(0.02, 0.92),
      scale: asDouble(json['scale'], 1.0).clamp(0.65, 1.55),
      rotation: asDouble(json['rotation'], 0.0).clamp(-0.78, 0.78),
    );
  }
}

/// Reglas acotadas del layout del perfil (visibilidad y orden seguro).
class ProfileLayoutConfig {
  const ProfileLayoutConfig({
    this.preset = ProfileLayoutPreset.classic,
    this.showFriends = true,
    this.showFavoriteSong = true,
    this.showFavoriteArtists = true,
    this.showMood = true,
    this.artistsBeforeBio = false,
  });

  final ProfileLayoutPreset preset;
  final bool showFriends;
  final bool showFavoriteSong;
  final bool showFavoriteArtists;
  final bool showMood;
  final bool artistsBeforeBio;

  ProfileLayoutConfig copyWith({
    ProfileLayoutPreset? preset,
    bool? showFriends,
    bool? showFavoriteSong,
    bool? showFavoriteArtists,
    bool? showMood,
    bool? artistsBeforeBio,
  }) {
    return ProfileLayoutConfig(
      preset: preset ?? this.preset,
      showFriends: showFriends ?? this.showFriends,
      showFavoriteSong: showFavoriteSong ?? this.showFavoriteSong,
      showFavoriteArtists: showFavoriteArtists ?? this.showFavoriteArtists,
      showMood: showMood ?? this.showMood,
      artistsBeforeBio: artistsBeforeBio ?? this.artistsBeforeBio,
    );
  }

  Map<String, Object?> toJson() => {
    'preset': preset.id,
    'show_friends': showFriends,
    'show_favorite_song': showFavoriteSong,
    'show_favorite_artists': showFavoriteArtists,
    'show_mood': showMood,
    'artists_before_bio': artistsBeforeBio,
  };

  factory ProfileLayoutConfig.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const ProfileLayoutConfig();
    return ProfileLayoutConfig(
      preset: ProfileLayoutPreset.fromId(json['preset'] as String?),
      showFriends: json['show_friends'] is bool
          ? json['show_friends'] as bool
          : true,
      showFavoriteSong: json['show_favorite_song'] is bool
          ? json['show_favorite_song'] as bool
          : true,
      showFavoriteArtists: json['show_favorite_artists'] is bool
          ? json['show_favorite_artists'] as bool
          : true,
      showMood: json['show_mood'] is bool ? json['show_mood'] as bool : true,
      artistsBeforeBio: json['artists_before_bio'] is bool
          ? json['artists_before_bio'] as bool
          : false,
    );
  }
}

/// Contrato serializable completo de personalización del usuario para enviar al
/// backend (`PUT /users/me/profile/customization`) y reconstruir el perfil de
/// cualquier usuario al visitarlo.
class ProfileCustomizationConfig {
  const ProfileCustomizationConfig({
    this.schemaVersion = currentSchemaVersion,
    this.appStyle = MeloopAppStyle.paper,
    this.colorTheme = MeloopColorTheme.meloopTeal,
    this.backgroundPattern = MeloopBackgroundPattern.tealRadial,
    this.borderStyle = ProfileBorderStyle.none,
    this.pinId = 'pin_classic_red',
    this.mood = 'Feliz',
    this.favoriteSongTitle = 'Flashing Lights',
    this.favoriteSongAlbum = 'Graduation',
    this.favoriteSongArtist = 'Kanye West',
    this.layout = const ProfileLayoutConfig(),
    this.stickers = const [],
  });

  static const int currentSchemaVersion = 1;

  final int schemaVersion;
  final MeloopAppStyle appStyle;
  final MeloopColorTheme colorTheme;
  final MeloopBackgroundPattern backgroundPattern;
  final ProfileBorderStyle borderStyle;
  final String pinId;
  final String mood;
  final String favoriteSongTitle;
  final String favoriteSongAlbum;
  final String favoriteSongArtist;
  final ProfileLayoutConfig layout;
  final List<ProfileStickerPlacement> stickers;

  /// Cantidad máxima de stickers permitidos según el nivel del usuario.
  static int maxStickersForLevel(int level) {
    if (level >= 15) return 5;
    if (level >= 10) return 3;
    if (level >= 5) return 2;
    return 1;
  }

  /// Cantidad de decoraciones activas (estilo distinto al base, color, fondo,
  /// marco o pin personalizado).
  int get equippedDecorationsCount {
    var count = 0;
    if (appStyle != MeloopAppStyle.paper) count++;
    if (colorTheme != MeloopColorTheme.meloopTeal) count++;
    if (backgroundPattern != MeloopBackgroundPattern.tealRadial) count++;
    if (borderStyle != ProfileBorderStyle.none) count++;
    if (pinId != 'pin_classic_red') count++;
    return count;
  }

  ProfilePinOption get activePinOption {
    for (final option in kProfilePinCatalog) {
      if (option.id == pinId) return option;
    }
    return kProfilePinCatalog.first;
  }

  ProfileCustomizationConfig copyWith({
    MeloopAppStyle? appStyle,
    MeloopColorTheme? colorTheme,
    MeloopBackgroundPattern? backgroundPattern,
    ProfileBorderStyle? borderStyle,
    String? pinId,
    String? mood,
    String? favoriteSongTitle,
    String? favoriteSongAlbum,
    String? favoriteSongArtist,
    ProfileLayoutConfig? layout,
    List<ProfileStickerPlacement>? stickers,
  }) {
    return ProfileCustomizationConfig(
      schemaVersion: schemaVersion,
      appStyle: appStyle ?? this.appStyle,
      colorTheme: colorTheme ?? this.colorTheme,
      backgroundPattern: backgroundPattern ?? this.backgroundPattern,
      borderStyle: borderStyle ?? this.borderStyle,
      pinId: pinId ?? this.pinId,
      mood: mood ?? this.mood,
      favoriteSongTitle: favoriteSongTitle ?? this.favoriteSongTitle,
      favoriteSongAlbum: favoriteSongAlbum ?? this.favoriteSongAlbum,
      favoriteSongArtist: favoriteSongArtist ?? this.favoriteSongArtist,
      layout: layout ?? this.layout,
      stickers: stickers ?? this.stickers,
    );
  }

  /// Devuelve una copia garantizada dentro de los desbloqueos del [level].
  ProfileCustomizationConfig sanitizedForLevel(int level) {
    final maxStickers = maxStickersForLevel(level);
    final safeStyle = appStyle.requiredLevel <= level
        ? appStyle
        : MeloopAppStyle.paper;
    final safeColor = colorTheme.requiredLevel <= level
        ? colorTheme
        : MeloopColorTheme.meloopTeal;
    final safeBg = backgroundPattern.requiredLevel <= level
        ? backgroundPattern
        : MeloopBackgroundPattern.tealRadial;
    final safeBorder = borderStyle.requiredLevel <= level
        ? borderStyle
        : ProfileBorderStyle.none;
    final pinOpt = activePinOption;
    final safePinId = pinOpt.requiredLevel <= level
        ? pinOpt.id
        : 'pin_classic_red';

    return copyWith(
      appStyle: safeStyle,
      colorTheme: safeColor,
      backgroundPattern: safeBg,
      borderStyle: safeBorder,
      pinId: safePinId,
      stickers: stickers.take(maxStickers).toList(growable: false),
    );
  }

  /// Payload listo para enviar al backend (`PUT /users/me/profile/customization`).
  Map<String, Object?> toJson() => {
    'schema_version': schemaVersion,
    'theme': {
      'app_style': appStyle.id,
      'color_theme': colorTheme.id,
      'background_pattern': backgroundPattern.id,
    },
    'profile_decorations': {
      'border_style': borderStyle.id,
      'pin_id': pinId,
      'mood': mood,
      'favorite_song': {
        'title': favoriteSongTitle,
        'album': favoriteSongAlbum,
        'artist': favoriteSongArtist,
      },
    },
    'layout': layout.toJson(),
    'stickers': stickers.map((s) => s.toJson()).toList(growable: false),
  };

  String toPrettyJson() => const JsonEncoder.withIndent('  ').convert(toJson());

  factory ProfileCustomizationConfig.fromJson(Map<String, dynamic> json) {
    final theme = json['theme'] as Map<String, dynamic>?;
    final decorations = json['profile_decorations'] as Map<String, dynamic>?;
    final favSong = decorations?['favorite_song'] as Map<String, dynamic>?;
    final rawStickers = json['stickers'];
    final parsedStickers = <ProfileStickerPlacement>[];
    if (rawStickers is List) {
      for (final item in rawStickers) {
        if (item is Map<String, dynamic>) {
          parsedStickers.add(ProfileStickerPlacement.fromJson(item));
        }
      }
    }

    return ProfileCustomizationConfig(
      schemaVersion: json['schema_version'] is int
          ? json['schema_version'] as int
          : currentSchemaVersion,
      appStyle: MeloopAppStyle.fromId(theme?['app_style'] as String?),
      colorTheme: MeloopColorTheme.fromId(theme?['color_theme'] as String?),
      backgroundPattern: MeloopBackgroundPattern.fromId(
        theme?['background_pattern'] as String?,
      ),
      borderStyle: ProfileBorderStyle.fromId(
        decorations?['border_style'] as String?,
      ),
      pinId: (decorations?['pin_id'] as String?) ?? 'pin_classic_red',
      mood: (decorations?['mood'] as String?) ?? 'Feliz',
      favoriteSongTitle: (favSong?['title'] as String?) ?? 'Flashing Lights',
      favoriteSongAlbum: (favSong?['album'] as String?) ?? 'Graduation',
      favoriteSongArtist: (favSong?['artist'] as String?) ?? 'Kanye West',
      layout: ProfileLayoutConfig.fromJson(
        json['layout'] as Map<String, dynamic>?,
      ),
      stickers: parsedStickers,
    );
  }
}
