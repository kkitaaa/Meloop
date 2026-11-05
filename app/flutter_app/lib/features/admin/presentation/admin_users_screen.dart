import 'package:flutter/material.dart';

import '../data/admin_models.dart';
import '../data/admin_repository.dart';

class AdminUsersScreen extends StatefulWidget {
  const AdminUsersScreen({super.key, required this.repository});

  final AdminRepository repository;

  @override
  State<AdminUsersScreen> createState() => _AdminUsersScreenState();
}

class _AdminUsersScreenState extends State<AdminUsersScreen> {
  final _queryController = TextEditingController();
  final List<AdminUser> _users = [];
  AdminUserStatus? _status;
  AdminUserRole? _role;
  String? _nextCursor;
  bool _isLoading = true;
  bool _isLoadingMore = false;
  int _requestId = 0;

  @override
  void initState() {
    super.initState();
    _loadUsers(reset: true);
  }

  @override
  void dispose() {
    _queryController.dispose();
    super.dispose();
  }

  Future<void> _loadUsers({required bool reset}) async {
    if (_isLoadingMore && !reset) return;
    final requestId = ++_requestId;
    setState(() {
      if (reset) {
        _isLoading = true;
        _users.clear();
        _nextCursor = null;
      } else {
        _isLoadingMore = true;
      }
    });
    try {
      final page = await widget.repository.fetchUsers(
        status: _status,
        role: _role,
        usernameQuery: _queryController.text,
        cursor: reset ? null : _nextCursor,
      );
      if (!mounted || requestId != _requestId) return;
      setState(() {
        _users.addAll(page.users);
        _nextCursor = page.nextCursor;
        _isLoading = false;
        _isLoadingMore = false;
      });
    } catch (error) {
      if (!mounted || requestId != _requestId) return;
      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
      });
      _showMessage('No se pudieron cargar los usuarios: $error');
    }
  }

  Future<void> _changeStatus(AdminUser user, AdminUserStatus nextStatus) async {
    String? reason;
    if (nextStatus == AdminUserStatus.suspended) {
      reason = await _askSuspensionReason(user);
      if (reason == null) return;
    } else {
      final confirmed = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Reactivar cuenta'),
          content: Text('¿Quieres reactivar la cuenta @${user.username}?'),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancelar'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Reactivar'),
            ),
          ],
        ),
      );
      if (confirmed != true) return;
    }

    try {
      await widget.repository.updateUserStatus(
        user.id,
        nextStatus,
        reason: reason,
      );
      if (!mounted) return;
      _showMessage(
        nextStatus == AdminUserStatus.suspended
            ? 'Cuenta suspendida.'
            : 'Cuenta reactivada.',
      );
      await _loadUsers(reset: true);
    } on AdminRepositoryException catch (error) {
      if (!mounted) return;
      _showMessage(switch (error.code) {
        'ALREADY_IN_STATE' => 'La cuenta ya se encuentra en ese estado.',
        'FORBIDDEN' => 'No tienes permisos para cambiar esta cuenta.',
        'INVALID_REASON' => 'El motivo debe tener entre 5 y 300 caracteres.',
        _ => 'No se pudo actualizar la cuenta (${error.code}).',
      });
    } catch (error) {
      if (mounted) _showMessage('No se pudo actualizar la cuenta: $error');
    }
  }

  Future<String?> _askSuspensionReason(AdminUser user) {
    return showDialog<String>(
      context: context,
      builder: (context) => _SuspensionReasonDialog(username: user.username),
    );
  }

  Future<void> _showHistory(AdminUser user) async {
    try {
      final entries = await widget.repository.fetchUserHistory(user.id);
      if (!mounted) return;
      await showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          title: Text('Historial de @${user.username}'),
          content: SizedBox(
            width: 460,
            child: entries.isEmpty
                ? const Text('No hay eventos en el historial.')
                : ListView(
                    shrinkWrap: true,
                    children: [
                      for (final entry in entries)
                        ListTile(
                          contentPadding: EdgeInsets.zero,
                          leading: const Icon(Icons.circle, size: 12),
                          title: Text(entry.action),
                          subtitle: Text(
                            '${entry.details}\n${_formatDate(entry.createdAt)}',
                          ),
                        ),
                    ],
                  ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cerrar'),
            ),
          ],
        ),
      );
    } catch (error) {
      if (mounted) _showMessage('No se pudo cargar el historial: $error');
    }
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  String _formatDate(DateTime date) =>
      '${date.day.toString().padLeft(2, '0')}/'
      '${date.month.toString().padLeft(2, '0')}/${date.year} '
      '${date.hour.toString().padLeft(2, '0')}:'
      '${date.minute.toString().padLeft(2, '0')}';

  @override
  Widget build(BuildContext context) {
    final wide = MediaQuery.sizeOf(context).width >= 900;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(16),
          child: Wrap(
            spacing: 12,
            runSpacing: 12,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              SizedBox(
                width: 280,
                child: TextField(
                  controller: _queryController,
                  onChanged: (_) => _loadUsers(reset: true),
                  decoration: const InputDecoration(
                    labelText: 'Buscar usuario',
                    prefixIcon: Icon(Icons.search),
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
              SizedBox(
                width: 210,
                child: DropdownButtonFormField<AdminUserStatus?>(
                  key: ValueKey(_status),
                  initialValue: _status,
                  isExpanded: true,
                  decoration: const InputDecoration(
                    labelText: 'Estado',
                    border: OutlineInputBorder(),
                  ),
                  items: [
                    const DropdownMenuItem<AdminUserStatus?>(
                      child: Text('Todos', overflow: TextOverflow.ellipsis),
                    ),
                    for (final status in AdminUserStatus.values)
                      DropdownMenuItem(
                        value: status,
                        child: Text(
                          status.label,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                  ],
                  onChanged: (value) {
                    _status = value;
                    _loadUsers(reset: true);
                  },
                ),
              ),
              SizedBox(
                width: 210,
                child: DropdownButtonFormField<AdminUserRole?>(
                  key: ValueKey(_role),
                  initialValue: _role,
                  isExpanded: true,
                  decoration: const InputDecoration(
                    labelText: 'Rol',
                    border: OutlineInputBorder(),
                  ),
                  items: [
                    const DropdownMenuItem<AdminUserRole?>(
                      child: Text('Todos', overflow: TextOverflow.ellipsis),
                    ),
                    for (final role in AdminUserRole.values)
                      DropdownMenuItem(
                        value: role,
                        child: Text(
                          role.label,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                  ],
                  onChanged: (value) {
                    _role = value;
                    _loadUsers(reset: true);
                  },
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: _isLoading
              ? const Center(child: CircularProgressIndicator())
              : _users.isEmpty
              ? const Center(child: Text('No se encontraron usuarios.'))
              : wide
              ? _buildTable()
              : _buildCards(),
        ),
        if (_nextCursor != null)
          Padding(
            padding: const EdgeInsets.all(12),
            child: OutlinedButton(
              onPressed: _isLoadingMore ? null : () => _loadUsers(reset: false),
              child: _isLoadingMore
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Text('Cargar más'),
            ),
          ),
      ],
    );
  }

  Widget _buildTable() => SingleChildScrollView(
    padding: const EdgeInsets.symmetric(horizontal: 16),
    child: SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: DataTable(
        columns: const [
          DataColumn(label: Text('Usuario')),
          DataColumn(label: Text('Rol')),
          DataColumn(label: Text('Estado')),
          DataColumn(label: Text('Nivel')),
          DataColumn(label: Text('Última actividad')),
          DataColumn(label: SizedBox(width: 144, child: Text('Acciones'))),
        ],
        rows: _users.map((user) => _userRow(user)).toList(),
      ),
    ),
  );

  DataRow _userRow(AdminUser user) => DataRow(
    cells: [
      DataCell(_userIdentity(user)),
      DataCell(Text(user.role.label)),
      DataCell(_StatusTag(status: user.status)),
      DataCell(Text('${user.level}')),
      DataCell(Text(_formatDate(user.lastActivity))),
      DataCell(SizedBox(width: 144, child: _actions(user))),
    ],
  );

  Widget _buildCards() => ListView.builder(
    padding: const EdgeInsets.symmetric(horizontal: 12),
    itemCount: _users.length,
    itemBuilder: (context, index) {
      final user = _users[index];
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(child: _userIdentity(user)),
                  _StatusTag(status: user.status),
                ],
              ),
              const SizedBox(height: 8),
              Text('${user.role.label} · Nivel ${user.level}'),
              Text('Última actividad: ${_formatDate(user.lastActivity)}'),
              const Divider(),
              Align(
                alignment: Alignment.centerRight,
                child: SizedBox(width: 144, child: _actions(user)),
              ),
            ],
          ),
        ),
      );
    },
  );

  Widget _userIdentity(AdminUser user) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      CircleAvatar(
        radius: 17,
        backgroundImage: user.avatarUrl == null
            ? null
            : NetworkImage(user.avatarUrl!),
        child: user.avatarUrl == null
            ? Text(user.username.characters.first.toUpperCase())
            : null,
      ),
      const SizedBox(width: 10),
      Text('@${user.username}'),
    ],
  );

  Widget _actions(AdminUser user) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      _actionButton(
        tooltip: 'Ver historial',
        icon: Icons.history,
        onPressed: () => _showHistory(user),
      ),
      if (user.role != AdminUserRole.admin &&
          user.status != AdminUserStatus.suspended)
        _actionButton(
          tooltip: 'Suspender',
          icon: Icons.block_outlined,
          color: Colors.deepOrange,
          onPressed: () => _changeStatus(user, AdminUserStatus.suspended),
        ),
      if (user.status == AdminUserStatus.suspended ||
          user.status == AdminUserStatus.inactive)
        _actionButton(
          tooltip: 'Reactivar',
          icon: Icons.check_circle_outline,
          color: Colors.green,
          onPressed: () => _changeStatus(user, AdminUserStatus.active),
        ),
    ],
  );

  Widget _actionButton({
    required String tooltip,
    required IconData icon,
    required VoidCallback onPressed,
    Color? color,
  }) => IconButton(
    tooltip: tooltip,
    onPressed: onPressed,
    constraints: const BoxConstraints.tightFor(width: 40, height: 40),
    padding: EdgeInsets.zero,
    visualDensity: VisualDensity.compact,
    icon: Icon(icon, color: color),
  );
}

class _SuspensionReasonDialog extends StatefulWidget {
  const _SuspensionReasonDialog({required this.username});

  final String username;

  @override
  State<_SuspensionReasonDialog> createState() =>
      _SuspensionReasonDialogState();
}

class _SuspensionReasonDialogState extends State<_SuspensionReasonDialog> {
  final _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text('Suspender @${widget.username}'),
    content: Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Indica el motivo de la suspensión.'),
        const SizedBox(height: 12),
        TextField(
          controller: _controller,
          minLines: 3,
          maxLines: 5,
          maxLength: 300,
          decoration: const InputDecoration(
            hintText: 'Motivo (5 a 300 caracteres)',
            border: OutlineInputBorder(),
          ),
        ),
      ],
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Cancelar'),
      ),
      FilledButton(
        onPressed: () {
          final reason = _controller.text.trim();
          if (reason.length < 5 || reason.length > 300) return;
          Navigator.pop(context, reason);
        },
        child: const Text('Suspender'),
      ),
    ],
  );
}

class _StatusTag extends StatelessWidget {
  const _StatusTag({required this.status});

  final AdminUserStatus status;

  @override
  Widget build(BuildContext context) {
    final color = switch (status) {
      AdminUserStatus.active => Colors.green,
      AdminUserStatus.suspended => Colors.red,
      AdminUserStatus.inactive => Colors.grey,
      AdminUserStatus.pendingDeletion => Colors.orange,
    };
    return DecoratedBox(
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        child: Text(
          status.label,
          style: TextStyle(color: color.shade700, fontSize: 12),
        ),
      ),
    );
  }
}
