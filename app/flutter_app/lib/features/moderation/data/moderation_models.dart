enum ModerationTargetType {
  post('publicacion', 'Publicación'),
  comment('comentario', 'Comentario'),
  user('usuario', 'Perfil');

  const ModerationTargetType(this.backendCode, this.label);

  final String backendCode;
  final String label;
}

enum ModerationReason {
  spam('spam', 'Spam'),
  harassment('acoso', 'Acoso'),
  hateSpeech('odio', 'Discurso de odio'),
  sexualContent('contenido_sexual', 'Contenido sexual'),
  violence('violencia', 'Violencia'),
  misinformation('desinformacion', 'Desinformación'),
  impersonation('suplantacion', 'Suplantación de identidad'),
  other('otro', 'Otro');

  const ModerationReason(this.backendCode, this.label);

  final String backendCode;
  final String label;
}

enum ModerationReportOutcome {
  success,
  alreadyReported,
  contentUnavailable,
  rateLimited;

  String get message => switch (this) {
    success => 'Gracias. Revisaremos tu reporte.',
    alreadyReported => 'Ya reportaste este contenido.',
    contentUnavailable => 'Este contenido ya no está disponible.',
    rateLimited => 'Espera un momento antes de enviar otro reporte.',
  };
}

enum ModerationReportStatus {
  pending('Pendientes'),
  inReview('En revisión'),
  resolved('Resueltos'),
  escalated('Escalados');

  const ModerationReportStatus(this.label);

  final String label;
}

enum ModerationRole { user, moderator, admin }

class ModerationReport {
  const ModerationReport({
    required this.id,
    required this.targetType,
    required this.reason,
    required this.preview,
    required this.dueAt,
    required this.status,
    this.overdue = false,
    this.canAct = true,
  });

  final String id;
  final ModerationTargetType targetType;
  final ModerationReason reason;
  final String preview;
  final DateTime dueAt;
  final ModerationReportStatus status;
  final bool overdue;
  final bool canAct;
}
