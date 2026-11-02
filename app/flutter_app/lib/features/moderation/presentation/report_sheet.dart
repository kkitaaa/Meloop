import 'package:flutter/material.dart';

import '../data/moderation_models.dart';
import '../data/moderation_repository.dart';

Future<void> showReportSheet(
  BuildContext context, {
  required ModerationTargetType targetType,
  required String targetId,
  String? targetUsername,
  ModerationRepository? repository,
}) {
  final isDesktop = MediaQuery.sizeOf(context).width > 900;
  final reportForm = ReportSheet(
    targetType: targetType,
    targetId: targetId,
    targetUsername: targetUsername,
    repository: repository ?? DemoModerationRepository(),
  );
  if (isDesktop) {
    return showDialog<void>(
      context: context,
      builder: (context) => Dialog(
        clipBehavior: Clip.antiAlias,
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520, maxHeight: 680),
          child: reportForm,
        ),
      ),
    );
  }
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    backgroundColor: Colors.transparent,
    builder: (context) => reportForm,
  );
}

class ReportSheet extends StatefulWidget {
  const ReportSheet({
    super.key,
    required this.targetType,
    required this.targetId,
    this.targetUsername,
    required this.repository,
  });

  final ModerationTargetType targetType;
  final String targetId;
  final String? targetUsername;
  final ModerationRepository repository;

  @override
  State<ReportSheet> createState() => _ReportSheetState();
}

class _ReportSheetState extends State<ReportSheet> {
  final TextEditingController _detailsController = TextEditingController();
  ModerationReason? _selectedReason;
  bool _isSubmitting = false;
  bool _isBlocking = false;
  ModerationReportOutcome? _outcome;
  String? _errorMessage;

  bool get _detailsRequired =>
      _selectedReason == ModerationReason.other &&
      _detailsController.text.trim().isEmpty;

  @override
  void dispose() {
    _detailsController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final reason = _selectedReason;
    if (reason == null || _detailsRequired || _isSubmitting) return;
    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
    });

    try {
      final outcome = await widget.repository.submitReport(
        targetType: widget.targetType,
        targetId: widget.targetId,
        reason: reason,
        details: _detailsController.text.trim(),
      );
      if (!mounted) return;
      setState(() {
        _outcome = outcome;
        _isSubmitting = false;
      });
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _errorMessage = 'No se pudo enviar el reporte: $error';
        _isSubmitting = false;
      });
    }
  }

  Future<void> _blockUser() async {
    final username = widget.targetUsername;
    if (username == null || _isBlocking) return;
    setState(() {
      _isBlocking = true;
      _errorMessage = null;
    });
    try {
      await widget.repository.blockUser(username);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text('@$username fue bloqueado.')));
      Navigator.of(context).pop();
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _errorMessage = 'No se pudo bloquear a @$username: $error';
        _isBlocking = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final outcome = _outcome;
    return Material(
      color: Colors.white,
      borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: EdgeInsets.fromLTRB(
            24,
            20,
            24,
            24 + MediaQuery.viewInsetsOf(context).bottom,
          ),
          child: outcome == null ? _buildForm() : _buildOutcome(outcome),
        ),
      ),
    );
  }

  Widget _buildForm() {
    final canSubmit =
        _selectedReason != null && !_detailsRequired && !_isSubmitting;
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Center(
          child: Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: Colors.black26,
              borderRadius: BorderRadius.circular(4),
            ),
          ),
        ),
        const SizedBox(height: 18),
        Text(
          'Reportar ${widget.targetType.label.toLowerCase()}',
          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 12),
        Flexible(
          child: SingleChildScrollView(
            child: Column(
              children: [
                for (final reason in ModerationReason.values)
                  RadioListTile<ModerationReason>(
                    dense: true,
                    contentPadding: EdgeInsets.zero,
                    value: reason,
                    groupValue: _selectedReason,
                    title: Text(reason.label),
                    onChanged: _isSubmitting
                        ? null
                        : (value) => setState(() {
                            _selectedReason = value;
                            _errorMessage = null;
                          }),
                  ),
                TextField(
                  controller: _detailsController,
                  maxLength: 500,
                  minLines: 2,
                  maxLines: 4,
                  onChanged: (_) => setState(() => _errorMessage = null),
                  decoration: InputDecoration(
                    labelText: _selectedReason == ModerationReason.other
                        ? 'Detalles (obligatorio)'
                        : 'Detalles (opcional)',
                    hintText: 'Cuéntanos más sobre el reporte',
                    border: const OutlineInputBorder(),
                    counterText: '${_detailsController.text.length}/500',
                  ),
                ),
                if (_errorMessage != null) ...[
                  const SizedBox(height: 8),
                  Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      _errorMessage!,
                      style: const TextStyle(color: Colors.red),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          width: double.infinity,
          child: FilledButton(
            onPressed: canSubmit ? _submit : null,
            child: _isSubmitting
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Enviar reporte'),
          ),
        ),
      ],
    );
  }

  Widget _buildOutcome(ModerationReportOutcome outcome) {
    final success = outcome == ModerationReportOutcome.success;
    final username = widget.targetUsername;
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          success ? 'Reporte enviado' : 'No se pudo completar',
          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 12),
        Text(outcome.message, key: const Key('report-outcome-message')),
        if (success && username != null && username.isNotEmpty) ...[
          const SizedBox(height: 20),
          OutlinedButton.icon(
            key: const Key('block-reported-user'),
            onPressed: _isBlocking ? null : _blockUser,
            icon: _isBlocking
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.block),
            label: Text('Bloquear también a @$username'),
          ),
        ],
        if (_errorMessage != null) ...[
          const SizedBox(height: 8),
          Text(_errorMessage!, style: const TextStyle(color: Colors.red)),
        ],
        const SizedBox(height: 12),
        Align(
          alignment: Alignment.centerRight,
          child: TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cerrar'),
          ),
        ),
      ],
    );
  }
}
