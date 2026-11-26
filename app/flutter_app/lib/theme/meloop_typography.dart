import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

/// Tipografía centralizada de Meloop:
/// - Interfaz general: Lexend (Google Fonts)
/// - Títulos de sección estilo logo ("Ultimos Blogs:"): Momo Trust Display (Google Fonts)
/// - Nombres de usuario (estilo Figma Hand): Kalam (equivalente manuscrito de
///   Google Fonts; Figma Hand es propietaria de Figma y no está en Google Fonts).
/// - Reproductores de música: Roboto Condensed Regular (empaquetada localmente
///   en assets/fonts/RobotoCondensed-Regular.ttf).
class MeloopTypography {
  const MeloopTypography._();

  static bool get _isTestBinding {
    try {
      return WidgetsBinding.instance.runtimeType.toString().contains('Test');
    } catch (_) {
      return false;
    }
  }

  static TextTheme lexendTextTheme([TextTheme? base]) {
    final textTheme = base ?? ThemeData.light().textTheme;
    return TextTheme(
      displayLarge: lexend(textStyle: textTheme.displayLarge),
      displayMedium: lexend(textStyle: textTheme.displayMedium),
      displaySmall: lexend(textStyle: textTheme.displaySmall),
      headlineLarge: lexend(textStyle: textTheme.headlineLarge),
      headlineMedium: lexend(textStyle: textTheme.headlineMedium),
      headlineSmall: lexend(textStyle: textTheme.headlineSmall),
      titleLarge: lexend(textStyle: textTheme.titleLarge),
      titleMedium: lexend(textStyle: textTheme.titleMedium),
      titleSmall: lexend(textStyle: textTheme.titleSmall),
      bodyLarge: lexend(textStyle: textTheme.bodyLarge),
      bodyMedium: lexend(textStyle: textTheme.bodyMedium),
      bodySmall: lexend(textStyle: textTheme.bodySmall),
      labelLarge: lexend(textStyle: textTheme.labelLarge),
      labelMedium: lexend(textStyle: textTheme.labelMedium),
      labelSmall: lexend(textStyle: textTheme.labelSmall),
    );
  }

  static TextStyle lexend({
    TextStyle? textStyle,
    double? fontSize,
    FontWeight? fontWeight,
    Color? color,
    double? height,
    double? letterSpacing,
  }) {
    if (_isTestBinding) {
      return (textStyle ?? const TextStyle()).copyWith(
        fontFamily: 'Lexend',
        fontSize: fontSize,
        fontWeight: fontWeight,
        color: color,
        height: height,
        letterSpacing: letterSpacing,
      );
    }
    return GoogleFonts.lexend(
      textStyle: textStyle,
      fontSize: fontSize,
      fontWeight: fontWeight,
      color: color,
      height: height,
      letterSpacing: letterSpacing,
    );
  }

  static TextStyle momoDisplay({
    double? fontSize,
    Color? color,
    double? height,
    double? letterSpacing,
  }) {
    if (_isTestBinding) {
      return TextStyle(
        fontFamily: 'MomoTrustDisplay',
        fontSize: fontSize,
        fontWeight: FontWeight.w400,
        color: color,
        height: height,
        letterSpacing: letterSpacing,
      );
    }
    return GoogleFonts.momoTrustDisplay(
      fontSize: fontSize,
      fontWeight: FontWeight.w400,
      color: color,
      height: height,
      letterSpacing: letterSpacing,
    );
  }

  /// Estilo manuscrito para nombres de usuario (equivalente a Figma Hand).
  static TextStyle username({
    double? fontSize,
    FontWeight? fontWeight = FontWeight.w700,
    Color? color,
    double? height,
    double? letterSpacing,
  }) {
    if (_isTestBinding) {
      return TextStyle(
        fontFamily: 'Kalam',
        fontSize: fontSize,
        fontWeight: fontWeight,
        color: color,
        height: height,
        letterSpacing: letterSpacing,
      );
    }
    return GoogleFonts.kalam(
      fontSize: fontSize,
      fontWeight: fontWeight,
      color: color,
      height: height,
      letterSpacing: letterSpacing,
    );
  }

  /// Tipografía de reproductores de música: Roboto Condensed Regular.
  static TextStyle musicPlayer({
    double? fontSize,
    Color? color,
    double? height,
    double? letterSpacing,
  }) => TextStyle(
    fontFamily: 'RobotoCondensed',
    fontSize: fontSize,
    fontWeight: FontWeight.w400,
    fontVariations: const [FontVariation('wght', 400)],
    color: color,
    height: height,
    letterSpacing: letterSpacing,
  );
}
