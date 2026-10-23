class Reward {
  const Reward({
    required this.id,
    required this.type,
    required this.levelId,
    this.equipped = false,
  });

  final String id;
  final String type;
  final int levelId;
  final bool equipped;

  factory Reward.fromJson(Map<String, dynamic> json) {
    final id = _stringValue(
      json['id_recompensa'] ?? json['reward_id'] ?? json['id'],
    );
    final type = _stringValue(
      json['tipo'] ?? json['reward_type'] ?? json['type'],
    );
    final levelId = _intValue(
      json['id_nivel'] ?? json['level_id'] ?? json['level'],
    );

    if (id == null || type == null || levelId == null) {
      throw const FormatException(
        'Cada recompensa necesita id_recompensa, tipo e id_nivel.',
      );
    }

    final equippedValue = json['equipada'] ?? json['equipped'];
    if (equippedValue != null && equippedValue is! bool) {
      throw const FormatException('El estado equipada debe ser booleano.');
    }

    return Reward(
      id: id,
      type: type,
      levelId: levelId,
      equipped: equippedValue as bool? ?? false,
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
