import 'package:flutter/material.dart';

import '../data/admin_models.dart';
import '../data/admin_repository.dart';

class AdminModeratorsScreen extends StatefulWidget {
  const AdminModeratorsScreen({super.key, required this.repository});

  final AdminRepository repository;

  @override
  State<AdminModeratorsScreen> createState() => _AdminModeratorsScreenState();
}

class _AdminModeratorsScreenState extends State<AdminModeratorsScreen> {
  final _searchController = TextEditingController();
  List<AdminUser> _moderators = [];
  List<AdminUser> _matches = [];
  AdminUser? _selectedUser;
  bool _isLoading = true;
  bool _isSearching = false;
  int _searchVersion = 0;

  @override
  void initState() {
    super.initState();
    _loadModerators();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadModerators() async {
    setState(() => _isLoading = true);
    try {
      final moderators = <AdminUser>[];
      String? cursor;
      do {
        final page = await widget.repository.fetchUsers(
          role: AdminUserRole.moderator,
          cursor: cursor,
          pageSize: 100,
        );
        moderators.addAll(page.users);
        cursor = page.nextCursor;
      } while (cursor != null);
      if (!mounted) return;
      setState(() {
        _moderators = moderators;
        _isLoading = false;
      });
    } catch (error) {
      if (!mounted) return;
      setState(() => _isLoading = false);
      _showMessage('No se pudieron cargar los moderadores: $error');
    }
  }

  Future<void> _search(String query) async {
    final version = ++_searchVersion;
    if (query.trim().isEmpty) {
      setState(() {
        _matches = [];
        _selectedUser = null;
        _isSearching = false;
      });
      return;
    }
    setState(() {
      _isSearching = true;
      _selectedUser = null;
    });
    try {
      final users = await widget.repository.searchUsersByUsername(query);
      if (!mounted || version != _searchVersion) return;
      setState(() {
        _matches = users;
        _isSearching = false;
      });
    } catch (error) {
      if (!mounted || version != _searchVersion) return;
      setState(() => _isSearching = false);
      _showMessage('No se pudo buscar usuarios: $error');
    }
  }

  Future<void> _assignSelected() async {
    final user = _selectedUser;
    if (user == null) return;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Asignar rol de moderador'),
        content: Text(
          '¿Quieres asignar el rol de moderador a @${user.username}?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Asignar'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    try {
      await widget.repository.updateUserRole(user.id, AdminUserRole.moderator);
      if (!mounted) return;
      setState(() {
        _matches = _matches
            .where((match) => match.id != user.id)
            .toList(growable: false);
        _selectedUser = null;
      });
      await _loadModerators();
      if (mounted) {
        _showMessage(
          'Rol actualizado. La persona deberá volver a iniciar sesión para que el cambio se aplique.',
        );
      }
    } on AdminRepositoryException catch (error) {
      if (!mounted) return;
      _showMessage(
        error.code == 'USER_SUSPENDED'
            ? 'No se puede asignar a una cuenta suspendida.'
            : 'No se pudo asignar el rol (${error.code}).',
      );
    } catch (error) {
      if (mounted) _showMessage('No se pudo asignar el rol: $error');
    }
  }

  Future<void> _removeModerator(AdminUser user) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Retirar rol de moderador'),
        content: Text(
          '¿Retirar el rol de moderador a @${user.username}?\n\n'
          'Sus reportes en revisión volverán a la bandeja.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Retirar rol'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    try {
      await widget.repository.updateUserRole(user.id, AdminUserRole.user);
      if (!mounted) return;
      await _loadModerators();
      if (mounted) {
        _showMessage(
          'Rol actualizado. La persona deberá volver a iniciar sesión para que el cambio se aplique.',
        );
      }
    } on AdminRepositoryException catch (error) {
      if (mounted) _showMessage('No se pudo retirar el rol (${error.code}).');
    } catch (error) {
      if (mounted) _showMessage('No se pudo retirar el rol: $error');
    }
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        Text(
          'Moderadores actuales',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 12),
        if (_isLoading)
          const Center(child: CircularProgressIndicator())
        else if (_moderators.isEmpty)
          const Text('Todavía no hay moderadores asignados.')
        else
          ..._moderators.map(
            (user) => Card(
              child: ListTile(
                leading: CircleAvatar(
                  child: Text(user.username.characters.first.toUpperCase()),
                ),
                title: Text('@${user.username}'),
                subtitle: const _ModeratorTag(),
                trailing: IconButton(
                  tooltip: 'Retirar rol',
                  onPressed: () => _removeModerator(user),
                  icon: const Icon(Icons.person_remove_outlined),
                ),
              ),
            ),
          ),
        const SizedBox(height: 28),
        Text(
          'Asignar moderador',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 8),
        const Text('Busca por nombre de usuario y selecciona una cuenta.'),
        const SizedBox(height: 12),
        TextField(
          controller: _searchController,
          onChanged: _search,
          decoration: InputDecoration(
            labelText: 'Nombre de usuario',
            prefixIcon: const Icon(Icons.search),
            suffixIcon: _searchController.text.isEmpty
                ? null
                : IconButton(
                    tooltip: 'Limpiar',
                    onPressed: () {
                      _searchController.clear();
                      _search('');
                    },
                    icon: const Icon(Icons.close),
                  ),
            border: const OutlineInputBorder(),
          ),
        ),
        if (_isSearching)
          const Padding(
            padding: EdgeInsets.all(20),
            child: Center(child: CircularProgressIndicator()),
          )
        else if (_searchController.text.trim().isNotEmpty && _matches.isEmpty)
          const Padding(
            padding: EdgeInsets.all(16),
            child: Text('No se encontraron usuarios.'),
          )
        else
          RadioGroup<AdminUser>(
            groupValue: _selectedUser,
            onChanged: (selected) => setState(() => _selectedUser = selected),
            child: Column(
              children: [
                for (final user in _matches)
                  Card(
                    child: RadioListTile<AdminUser>(
                      value: user,
                      enabled: user.role != AdminUserRole.admin,
                      title: Text('@${user.username}'),
                      subtitle: Text(
                        '${user.role.label} · ${user.status.label}'
                        '${user.status == AdminUserStatus.suspended ? ' · No asignable' : ''}',
                      ),
                    ),
                  ),
              ],
            ),
          ),
        if (_matches.isNotEmpty) ...[
          const SizedBox(height: 12),
          Align(
            alignment: Alignment.centerRight,
            child: FilledButton.icon(
              onPressed: _selectedUser == null ? null : _assignSelected,
              icon: const Icon(Icons.shield_outlined),
              label: const Text('Asignar rol'),
            ),
          ),
        ],
      ],
    );
  }
}

class _ModeratorTag extends StatelessWidget {
  const _ModeratorTag();

  @override
  Widget build(BuildContext context) => DecoratedBox(
    decoration: BoxDecoration(
      color: const Color(0xFFE0F4EF),
      borderRadius: BorderRadius.circular(20),
    ),
    child: const Padding(
      padding: EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      child: Text(
        'Moderador',
        style: TextStyle(color: Color(0xFF087F6D), fontSize: 12),
      ),
    ),
  );
}
