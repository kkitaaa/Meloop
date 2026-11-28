import 'package:flutter/material.dart';

import '../data/moderation_models.dart';
import '../data/moderation_repository.dart';

class ModerationQueueScreen extends StatefulWidget {
  const ModerationQueueScreen({super.key, required this.role, this.repository});

  final ModerationRole role;
  final ModerationRepository? repository;

  @override
  State<ModerationQueueScreen> createState() => _ModerationQueueScreenState();
}

class _ModerationQueueScreenState extends State<ModerationQueueScreen> {
  static const _pageSize = 10;

  late final ModerationRepository _repository =
      widget.repository ?? DemoModerationRepository();
  final ScrollController _scrollController = ScrollController();
  final List<ModerationReport> _reports = [];
  final Map<ModerationReportStatus, int> _counts = {};

  ModerationReportStatus _status = ModerationReportStatus.pending;
  ModerationReason? _reason;
  ModerationReport? _selectedReport;
  int _page = 1;
  bool _isLoading = false;
  bool _isRefreshing = false;
  bool _hasMore = true;
  String? _errorMessage;

  bool get _isAdmin => widget.role == ModerationRole.admin;
  bool get _canModerate =>
      widget.role == ModerationRole.admin ||
      widget.role == ModerationRole.moderator;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
    if (_canModerate) {
      _loadCounts();
      _loadReports(reset: true);
    }
  }

  @override
  void dispose() {
    _scrollController
      ..removeListener(_onScroll)
      ..dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.hasClients &&
        _scrollController.position.extentAfter < 300) {
      _loadReports();
    }
  }

  Future<void> _loadCounts() async {
    try {
      final entries = await Future.wait(
        ModerationReportStatus.values.map(
          (status) =>
              _repository.fetchReports(status: status, page: 1, pageSize: 100),
        ),
      );
      if (!mounted) return;
      setState(() {
        for (
          var index = 0;
          index < ModerationReportStatus.values.length;
          index++
        ) {
          _counts[ModerationReportStatus.values[index]] = entries[index].length;
        }
      });
    } catch (error) {
      if (!mounted) return;
      setState(
        () => _errorMessage = 'No se pudieron cargar los contadores: $error',
      );
    }
  }

  Future<void> _loadReports({bool reset = false}) async {
    if (_isLoading || (!_hasMore && !reset) || !_canModerate) return;
    if (reset) {
      _page = 1;
      _hasMore = true;
      _errorMessage = null;
      _reports.clear();
    }
    setState(() => _isLoading = true);
    try {
      final nextReports = await _repository.fetchReports(
        status: _status,
        reason: _reason,
        page: _page,
        pageSize: _pageSize,
      );
      if (!mounted) return;
      setState(() {
        _reports.addAll(nextReports);
        _page++;
        _hasMore = nextReports.length == _pageSize;
        _isLoading = false;
      });
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _errorMessage = 'No se pudieron cargar los reportes: $error';
        _isLoading = false;
      });
    }
  }

  Future<void> _refresh() async {
    setState(() => _isRefreshing = true);
    await Future.wait([_loadCounts(), _loadReports(reset: true)]);
    if (mounted) setState(() => _isRefreshing = false);
  }

  void _selectStatus(ModerationReportStatus status) {
    if (_status == status) return;
    setState(() {
      _status = status;
      _selectedReport = null;
    });
    _loadReports(reset: true);
  }

  void _selectReason(ModerationReason? reason) {
    if (_reason == reason) return;
    setState(() {
      _reason = reason;
      _selectedReport = null;
    });
    _loadReports(reset: true);
  }

  void _openReport(ModerationReport report) {
    if (!report.canAct) return;
    if (MediaQuery.sizeOf(context).width > 900) {
      setState(() => _selectedReport = report);
      return;
    }
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (context) => Padding(
        padding: const EdgeInsets.all(24),
        child: ModerationReportDetail(report: report),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (!_canModerate) {
      return Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(title: const Text('Moderación')),
        body: const Center(
          child: Text('Esta sección está disponible solo para moderadores.'),
        ),
      );
    }

    final isDesktop = MediaQuery.sizeOf(context).width > 900;
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: AppBar(
        title: const Text('Cola de moderación'),
        backgroundColor: const Color(0xFF1ABC9C),
        foregroundColor: Colors.white,
      ),
      body: Column(
        children: [
          _buildFilters(),
          Expanded(
            child: isDesktop
                ? Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      SizedBox(width: 430, child: _buildReportList()),
                      const VerticalDivider(width: 1),
                      Expanded(
                        child: _selectedReport == null
                            ? const Center(
                                child: Text(
                                  'Selecciona un reporte para revisar',
                                ),
                              )
                            : SingleChildScrollView(
                                padding: const EdgeInsets.all(32),
                                child: ModerationReportDetail(
                                  report: _selectedReport!,
                                ),
                              ),
                      ),
                    ],
                  )
                : _buildReportList(),
          ),
        ],
      ),
    );
  }

  Widget _buildFilters() {
    return Material(
      color: Colors.white,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(12, 12, 12, 8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: ModerationReportStatus.values.map((status) {
                  final selected = status == _status;
                  final isEscalated =
                      _isAdmin && status == ModerationReportStatus.escalated;
                  final count = _counts[status] ?? 0;
                  return Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: ChoiceChip(
                      label: Text('${status.label} ($count)'),
                      selected: selected,
                      onSelected: (_) => _selectStatus(status),
                      selectedColor: isEscalated
                          ? Colors.deepOrange.shade100
                          : const Color(0xFFD8F4EE),
                      labelStyle: TextStyle(
                        color: isEscalated
                            ? Colors.deepOrange.shade900
                            : Colors.black87,
                        fontWeight: isEscalated
                            ? FontWeight.bold
                            : FontWeight.normal,
                      ),
                      side: isEscalated
                          ? BorderSide(color: Colors.deepOrange.shade400)
                          : BorderSide.none,
                    ),
                  );
                }).toList(),
              ),
            ),
            DropdownButton<ModerationReason?>(
              value: _reason,
              isExpanded: true,
              hint: const Text('Filtrar por motivo'),
              items: [
                const DropdownMenuItem<ModerationReason?>(
                  value: null,
                  child: Text('Todos los motivos'),
                ),
                ...ModerationReason.values.map(
                  (reason) => DropdownMenuItem(
                    value: reason,
                    child: Text(reason.label),
                  ),
                ),
              ],
              onChanged: _selectReason,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildReportList() {
    return RefreshIndicator(
      onRefresh: _refresh,
      child: ListView.builder(
        controller: _scrollController,
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(12),
        itemCount:
            _reports.length +
            (_isLoading || _errorMessage != null || _reports.isEmpty ? 1 : 0),
        itemBuilder: (context, index) {
          if (index >= _reports.length) {
            if (_errorMessage != null) {
              return Padding(
                padding: const EdgeInsets.all(24),
                child: Text(_errorMessage!, textAlign: TextAlign.center),
              );
            }
            if (_isLoading || _isRefreshing) {
              return const Padding(
                padding: EdgeInsets.all(24),
                child: Center(child: CircularProgressIndicator()),
              );
            }
            return const Padding(
              padding: EdgeInsets.all(24),
              child: Center(child: Text('No hay reportes para estos filtros.')),
            );
          }
          final report = _reports[index];
          return _ModerationReportTile(
            report: report,
            selected: report.id == _selectedReport?.id,
            onTap: report.canAct ? () => _openReport(report) : null,
          );
        },
      ),
    );
  }
}

class _ModerationReportTile extends StatelessWidget {
  const _ModerationReportTile({
    required this.report,
    required this.selected,
    required this.onTap,
  });

  final ModerationReport report;
  final bool selected;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final isOverdue = report.overdue || report.dueAt.isBefore(DateTime.now());
    final remaining = report.dueAt.difference(DateTime.now());
    final hoursRemaining = ((remaining.inSeconds + 3599) ~/ 3600).clamp(1, 999);
    final timeLabel = isOverdue ? 'Vencido' : 'Vence en $hoursRemaining h';
    final icon = switch (report.targetType) {
      ModerationTargetType.post => Icons.article_outlined,
      ModerationTargetType.comment => Icons.mode_comment_outlined,
      ModerationTargetType.user => Icons.person_outline,
    };

    return Card(
      color: !report.canAct
          ? Colors.grey.shade100
          : selected
          ? const Color(0xFFE7F7F2)
          : Colors.white,
      child: ListTile(
        enabled: report.canAct,
        onTap: onTap,
        leading: Icon(
          icon,
          color: report.canAct ? const Color(0xFF0D5C5E) : Colors.grey,
        ),
        title: Row(
          children: [
            Expanded(child: Text(report.reason.label)),
            Text(
              timeLabel,
              style: TextStyle(
                color: isOverdue ? Colors.red : Colors.black54,
                fontSize: 12,
                fontWeight: isOverdue ? FontWeight.bold : FontWeight.normal,
              ),
            ),
          ],
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(report.preview, maxLines: 2, overflow: TextOverflow.ellipsis),
            if (!report.canAct)
              const Text(
                'Contenido propio',
                style: TextStyle(color: Colors.grey, fontSize: 12),
              ),
          ],
        ),
      ),
    );
  }
}

class ModerationReportDetail extends StatelessWidget {
  const ModerationReportDetail({super.key, required this.report});

  final ModerationReport report;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    mainAxisSize: MainAxisSize.min,
    children: [
      Text(
        'Reporte ${report.id}',
        style: Theme.of(context).textTheme.headlineSmall,
      ),
      const SizedBox(height: 16),
      Text('Contenido: ${report.targetType.label}'),
      const SizedBox(height: 8),
      Text('Motivo: ${report.reason.label}'),
      const SizedBox(height: 16),
      Text(report.preview),
      const SizedBox(height: 24),
      if (!report.canAct)
        const Chip(label: Text('Contenido propio'))
      else
        const Text('Acciones de moderación disponibles próximamente.'),
    ],
  );
}
