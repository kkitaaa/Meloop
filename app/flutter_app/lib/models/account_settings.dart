class AccountDetails {
  const AccountDetails({
    required this.id,
    required this.username,
    required this.email,
  });

  final String id;
  final String username;
  final String email;

  factory AccountDetails.fromJson(Object? value) {
    if (value is! Map<String, dynamic>) {
      throw const FormatException('El servidor devolvió una cuenta inválida.');
    }

    final id = value['id_usuario'];
    final username = value['username'];
    final email = value['correo'];
    if (id is! String || username is! String || email is! String) {
      throw const FormatException(
        'El servidor devolvió una cuenta incompleta.',
      );
    }

    return AccountDetails(id: id, username: username, email: email);
  }
}

class PrivacySettings {
  const PrivacySettings({
    required this.profileVisibility,
    required this.postsVisibility,
    required this.interactionsVisibility,
    required this.messageReception,
    required this.friendRequestReception,
  });

  final String profileVisibility;
  final String postsVisibility;
  final String interactionsVisibility;
  final String messageReception;
  final String friendRequestReception;

  factory PrivacySettings.fromJson(Object? value) {
    if (value is! Map<String, dynamic>) {
      throw const FormatException(
        'El servidor devolvió una configuración de privacidad inválida.',
      );
    }

    String readString(String key) {
      final item = value[key];
      if (item is! String) {
        throw FormatException('Falta el valor "$key" en la configuración.');
      }
      return item;
    }

    return PrivacySettings(
      profileVisibility: readString('visibilidad_perfil'),
      postsVisibility: readString('visibilidad_publicaciones'),
      interactionsVisibility: readString('visibilidad_interacciones'),
      messageReception: readString('recepcion_mensajes'),
      friendRequestReception: readString('recepcion_solicitudes_amistad'),
    );
  }
}

class PrivacySettingsPatch {
  const PrivacySettingsPatch({
    this.profileVisibility,
    this.postsVisibility,
    this.interactionsVisibility,
    this.messageReception,
    this.friendRequestReception,
  });

  final String? profileVisibility;
  final String? postsVisibility;
  final String? interactionsVisibility;
  final String? messageReception;
  final String? friendRequestReception;

  Map<String, String> toJson() => {
    'visibilidad_perfil': ?profileVisibility,
    'visibilidad_publicaciones': ?postsVisibility,
    'visibilidad_interacciones': ?interactionsVisibility,
    'recepcion_mensajes': ?messageReception,
    'recepcion_solicitudes_amistad': ?friendRequestReception,
  };
}

class NotificationPreference {
  const NotificationPreference({required this.type, required this.enabled});

  final String type;
  final bool enabled;

  factory NotificationPreference.fromJson(Object? value) {
    if (value is! Map<String, dynamic>) {
      throw const FormatException(
        'El servidor devolvió una preferencia de notificación inválida.',
      );
    }

    final type = value['tipo_notificacion'];
    final enabled = value['habilitada'];
    if (type is! String || enabled is! bool) {
      throw const FormatException(
        'El servidor devolvió una preferencia de notificación incompleta.',
      );
    }

    return NotificationPreference(type: type, enabled: enabled);
  }

  static List<NotificationPreference> listFromJson(Object? value) {
    if (value is! List) {
      throw const FormatException(
        'El servidor devolvió una lista de notificaciones inválida.',
      );
    }

    return List<NotificationPreference>.unmodifiable(
      value.map(NotificationPreference.fromJson),
    );
  }
}
