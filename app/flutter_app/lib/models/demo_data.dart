class DemoData {
  final String message;
  final String timestamp;

  const DemoData({required this.message, required this.timestamp});

  factory DemoData.fromJson(Object? json) {
    if (json is! Map<String, dynamic> ||
        json['message'] is! String ||
        json['timestamp'] is! String) {
      throw const FormatException('Los datos de prueba del API son inválidos.');
    }

    return DemoData(
      message: json['message'] as String,
      timestamp: json['timestamp'] as String,
    );
  }
}

class DemoPostData {
  final String receivedMessage;
  final String receivedAt;

  const DemoPostData({required this.receivedMessage, required this.receivedAt});

  factory DemoPostData.fromJson(Object? json) {
    if (json is! Map<String, dynamic> ||
        json['receivedMessage'] is! String ||
        json['receivedAt'] is! String) {
      throw const FormatException('La respuesta del POST es inválida.');
    }

    return DemoPostData(
      receivedMessage: json['receivedMessage'] as String,
      receivedAt: json['receivedAt'] as String,
    );
  }
}
