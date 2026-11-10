import 'package:flutter/foundation.dart';

/// Estado temporal de gamificación. Será sustituido por el progreso entregado
/// por el backend cuando esté disponible.
class GamificationDemoController extends ChangeNotifier {
  GamificationDemoController._();

  static final instance = GamificationDemoController._();

  int level = 3;
  int xp = 70;
  static const xpNeeded = 100;
  bool rgbFrameUnlocked = false;
  bool rgbFrameEquipped = false;
  LevelUpReward? _pendingLevelUp;

  /// Vicente: keep this demo progression and its connected UI. Replace the
  /// values with API progress when endpoints are ready, without removing the
  /// level/XP animations or level-up reward presentation.
  int get experienceNeeded => xpNeeded;

  LevelUpReward? takePendingLevelUp() {
    final reward = _pendingLevelUp;
    _pendingLevelUp = null;
    return reward;
  }

  void grantLikeXp() {
    xp += 10;
    while (xp >= xpNeeded) {
      xp -= xpNeeded;
      level++;
      rgbFrameUnlocked = true;
      _pendingLevelUp = LevelUpReward(
        level: level,
        name: 'Marco RGB',
        description: 'Un marco animado multicolor para tu perfil.',
      );
    }
    notifyListeners();
  }

  void equipRgbFrame() {
    if (!rgbFrameUnlocked) return;
    rgbFrameEquipped = true;
    notifyListeners();
  }
}

class LevelUpReward {
  const LevelUpReward({
    required this.level,
    required this.name,
    required this.description,
  });

  final int level;
  final String name;
  final String description;
}
