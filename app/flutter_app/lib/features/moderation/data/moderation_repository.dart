import 'moderation_models.dart';

abstract interface class ModerationRepository {
  Future<ModerationReportOutcome> submitReport({
    required ModerationTargetType targetType,
    required String targetId,
    required ModerationReason reason,
    required String details,
  });

  Future<void> blockUser(String username);

  Future<List<ModerationReport>> fetchReports({
    required ModerationReportStatus status,
    ModerationReason? reason,
    required int page,
    required int pageSize,
  });
}

class DemoModerationRepository implements ModerationRepository {
  DemoModerationRepository({
    this.reportOutcome = ModerationReportOutcome.success,
    DateTime Function()? clock,
  }) : _clock = clock ?? DateTime.now;

  final ModerationReportOutcome reportOutcome;
  final DateTime Function() _clock;

  @override
  Future<ModerationReportOutcome> submitReport({
    required ModerationTargetType targetType,
    required String targetId,
    required ModerationReason reason,
    required String details,
  }) async {
    await Future<void>.delayed(const Duration(milliseconds: 250));
    if (targetId.isEmpty) {
      throw ArgumentError.value(targetId, 'targetId', 'No puede estar vacío.');
    }
    return reportOutcome;
  }

  @override
  Future<void> blockUser(String username) async {
    await Future<void>.delayed(const Duration(milliseconds: 200));
    if (username.isEmpty) {
      throw ArgumentError.value(username, 'username', 'No puede estar vacío.');
    }
  }

  @override
  Future<List<ModerationReport>> fetchReports({
    required ModerationReportStatus status,
    ModerationReason? reason,
    required int page,
    required int pageSize,
  }) async {
    await Future<void>.delayed(const Duration(milliseconds: 250));
    final now = _clock();
    final allReports = [
      ModerationReport(
        id: 'report-1',
        targetType: ModerationTargetType.post,
        reason: ModerationReason.harassment,
        preview: 'Publicación con mensajes dirigidos a acosar a otra persona.',
        dueAt: now.add(const Duration(hours: 5)),
        status: ModerationReportStatus.pending,
      ),
      ModerationReport(
        id: 'report-2',
        targetType: ModerationTargetType.comment,
        reason: ModerationReason.hateSpeech,
        preview: 'Comentario marcado por discurso de odio.',
        dueAt: now.subtract(const Duration(minutes: 25)),
        overdue: true,
        status: ModerationReportStatus.pending,
      ),
      ModerationReport(
        id: 'report-3',
        targetType: ModerationTargetType.user,
        reason: ModerationReason.impersonation,
        preview: 'Perfil que podría estar suplantando a otra persona.',
        dueAt: now.add(const Duration(hours: 1)),
        canAct: false,
        status: ModerationReportStatus.pending,
      ),
      ModerationReport(
        id: 'report-4',
        targetType: ModerationTargetType.post,
        reason: ModerationReason.spam,
        preview: 'Contenido repetido con enlaces promocionales.',
        dueAt: now.add(const Duration(hours: 12)),
        status: ModerationReportStatus.inReview,
      ),
      ModerationReport(
        id: 'report-5',
        targetType: ModerationTargetType.comment,
        reason: ModerationReason.misinformation,
        preview: 'Comentario con información que requiere verificación.',
        dueAt: now.add(const Duration(hours: 20)),
        status: ModerationReportStatus.escalated,
      ),
      ModerationReport(
        id: 'report-6',
        targetType: ModerationTargetType.user,
        reason: ModerationReason.other,
        preview: 'Reporte revisado y resuelto por el equipo.',
        dueAt: now.subtract(const Duration(hours: 2)),
        status: ModerationReportStatus.resolved,
      ),
    ];
    final filtered =
        allReports
            .where(
              (report) =>
                  report.status == status &&
                  (reason == null || report.reason == reason),
            )
            .toList()
          ..sort((a, b) => a.dueAt.compareTo(b.dueAt));
    final start = (page - 1) * pageSize;
    if (start >= filtered.length) return const [];
    return filtered.skip(start).take(pageSize).toList(growable: false);
  }
}
