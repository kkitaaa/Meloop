import 'reward.dart';

class UserProgress {
  const UserProgress({
    required this.userId,
    required this.levelId,
    required this.levelNumber,
    required this.experience,
    this.rewards = const [],
  });

  final String userId;
  final int levelId;
  final int levelNumber;
  final int experience;
  final List<Reward> rewards;

  factory UserProgress.fromJson(Map<String, dynamic> json) {
    final userId = _stringValue(json['user_id'] ?? json['userId']);
    final levelId = _intValue(json['id_nivel'] ?? json['level_id']);
    final levelNumber = _intValue(
      json['level_number'] ?? json['numero_nivel'] ?? json['levelNumber'],
    );
    final experience = _intValue(
      json['experiencia'] ?? json['experience'] ?? json['xp'],
    );

    if (userId == null ||
        levelId == null ||
        levelNumber == null ||
        experience == null) {
      throw const FormatException(
        'El progreso requiere user_id, id_nivel, level_number y experiencia.',
      );
    }

    final rawRewards =
        json['rewards'] ?? json['recompensas'] ?? json['unlocked_rewards'];
    if (rawRewards != null && rawRewards is! List) {
      throw const FormatException(
        'La lista de recompensas tiene formato inválido.',
      );
    }

    final rewards = rawRewards == null
        ? const <Reward>[]
        : rawRewards
              .map((entry) {
                if (entry is! Map) {
                  throw const FormatException(
                    'Recompensa con formato inválido.',
                  );
                }
                return Reward.fromJson(Map<String, dynamic>.from(entry));
              })
              .toList(growable: false);

    return UserProgress(
      userId: userId,
      levelId: levelId,
      levelNumber: levelNumber,
      experience: experience,
      rewards: List.unmodifiable(rewards),
    );
  }

  static String? _stringValue(Object? value) =>
      value is String && value.isNotEmpty ? value : null;

  static int? _intValue(Object? value) => switch (value) {
    int number => number,
    String text => int.tryParse(text),
    _ => null,
  };
}
