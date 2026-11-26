import 'package:flutter/material.dart';

import '../features/gamification/presentation/gamification_demo_controller.dart';
import '../features/profile/models/profile_customization.dart';

/// Controlador global de apariencia de Meloop.
///
/// Centraliza el estilo (`Papel`, `Plano cuadrado`, `Squircle`, `Frutiger Aero`,
/// `Neón nocturno`), la paleta de colores y el fondo elegidos por el usuario en
/// su perfil para que toda la aplicación refleje los cambios en vivo.
///
/// Además soporta un `viewingOverride` temporal cuando se visita el perfil de
/// otro usuario: mientras ese perfil está abierto se muestra el tema de ese
/// usuario (o el tema por defecto si aún no tiene uno en el servidor), y al
/// salir se restaura automáticamente el tema personal del usuario actual.
///
/// TODO(backend): Al restaurar sesión (`GET /users/me/profile/customization`),
/// llamar a `applyServerPayload(json)` y, al guardar cambios desde el editor,
/// enviar `ownConfig.toJson()` a `PUT /users/me/profile/customization`.
class MeloopThemeController extends ChangeNotifier {
  MeloopThemeController._();

  static final MeloopThemeController instance = MeloopThemeController._();

  ProfileCustomizationConfig _ownConfig = const ProfileCustomizationConfig();
  Object? _overrideOwner;
  ProfileCustomizationConfig? _overrideConfig;

  /// Configuración visual efectiva en pantalla (respeta override temporal al
  /// visitar el perfil de otra persona).
  ProfileCustomizationConfig get config => _overrideConfig ?? _ownConfig;

  /// Configuración propia del usuario autenticado.
  ProfileCustomizationConfig get ownConfig => _ownConfig;

  bool get isViewingOtherProfile => _overrideConfig != null;

  MeloopAppStyle get appStyle => config.appStyle;
  MeloopColorTheme get colorTheme => config.colorTheme;
  MeloopBackgroundPattern get backgroundPattern => config.backgroundPattern;

  bool get isDarkCardStyle => config.appStyle == MeloopAppStyle.neonNight;
  bool get isDarkTheme =>
      config.appStyle == MeloopAppStyle.neonNight ||
      config.colorTheme == MeloopColorTheme.midnightViolet;

  // Tokens semánticos para superficies y tipografía dentro de tarjetas:
  // garantizan contraste alto tanto en estilos claros como en "Neón nocturno".
  Color get cardTitleColor =>
      isDarkCardStyle ? Colors.white : const Color(0xFF141919);
  Color get cardBodyColor => isDarkCardStyle
      ? Colors.white.withValues(alpha: 0.90)
      : const Color(0xFF283434);
  Color get cardMutedColor => isDarkCardStyle
      ? Colors.white.withValues(alpha: 0.68)
      : const Color(0xFF6B7777);
  Color get cardSubtleMutedColor => isDarkCardStyle
      ? Colors.white.withValues(alpha: 0.56)
      : const Color(0xFF93A1A1);
  Color get cardDividerColor => isDarkCardStyle
      ? Colors.white.withValues(alpha: 0.14)
      : Colors.black.withValues(alpha: 0.10);
  Color get innerTileColor =>
      isDarkCardStyle ? Colors.white.withValues(alpha: 0.08) : Colors.white;
  Color get innerTileBorderColor => isDarkCardStyle
      ? Colors.white.withValues(alpha: 0.14)
      : const Color(0xFFE4ECEB);
  Color get inputFillColor =>
      isDarkCardStyle ? Colors.white.withValues(alpha: 0.10) : Colors.white;
  Color get inputBorderColor => isDarkCardStyle
      ? Colors.white.withValues(alpha: 0.20)
      : const Color(0xFFD5E1DF);
  Color get placeholderBgColor => isDarkCardStyle
      ? const Color(0xFF283142)
      : const Color(0xFFE7ECEB);
  Color get placeholderIconColor => isDarkCardStyle
      ? const Color(0xFF9AA7BD)
      : const Color(0xFF7B8E8C);
  Color get placeholderBorderColor => isDarkCardStyle
      ? const Color(0xFF3D4B63)
      : const Color(0xFFD0DADA);
  Color get softBadgeBgColor => isDarkCardStyle
      ? colorTheme.primary.withValues(alpha: 0.22)
      : colorTheme.softCardTint;
  Color get softBadgeFgColor =>
      isDarkCardStyle ? colorTheme.headerTop : colorTheme.pillTrack;
  Color get accentLinkColor =>
      isDarkCardStyle ? colorTheme.headerTop : colorTheme.pillTrack;

  Color get primaryColor => config.colorTheme.primary;
  Color get pillTrackColor => config.colorTheme.pillTrack;
  LinearGradient get headerGradient => LinearGradient(
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
    colors: [config.colorTheme.headerTop, config.colorTheme.headerBottom],
  );

  void updateConfig(ProfileCustomizationConfig next) {
    _ownConfig = next;
    GamificationDemoController.instance.setActivePinColor(
      next.activePinOption.color,
    );
    notifyListeners();
  }

  /// Aplica temporalmente el tema de otro perfil visitado sin sobrescribir la
  /// personalización propia del usuario.
  void setViewingOverride(
    Object owner,
    ProfileCustomizationConfig overrideConfig,
  ) {
    _overrideOwner = owner;
    _overrideConfig = overrideConfig;
    notifyListeners();
  }

  /// Restaura el tema propio al cerrar un perfil ajeno.
  void clearViewingOverride(Object owner) {
    if (!identical(_overrideOwner, owner)) return;
    _overrideOwner = null;
    _overrideConfig = null;
    notifyListeners();
  }

  void applyServerPayload(Map<String, dynamic> json, {int? userLevel}) {
    final parsed = ProfileCustomizationConfig.fromJson(json);
    updateConfig(
      userLevel != null ? parsed.sanitizedForLevel(userLevel) : parsed,
    );
  }

  void resetToDefault() {
    _overrideOwner = null;
    _overrideConfig = null;
    updateConfig(const ProfileCustomizationConfig());
  }
}
