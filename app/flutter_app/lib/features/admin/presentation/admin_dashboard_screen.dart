import 'package:flutter/material.dart';

import '../../moderation/data/moderation_models.dart';
import '../../moderation/presentation/moderation_queue_screen.dart';
import '../../notifications/presentation/notification_scope.dart';
import '../../../screens/auth_screen.dart';
import '../data/admin_repository.dart';
import 'admin_moderators_screen.dart';
import 'admin_rewards_screen.dart';
import 'admin_users_screen.dart';

enum _AdminSection {
  users('Usuarios', Icons.people_outline),
  moderators('Moderadores', Icons.shield_outlined),
  rewards('Recompensas', Icons.card_giftcard_outlined),
  moderation('Cola de moderación', Icons.report_outlined);

  const _AdminSection(this.label, this.icon);

  final String label;
  final IconData icon;
}

class AdminDashboardScreen extends StatefulWidget {
  const AdminDashboardScreen({super.key, required this.role, this.repository});

  final ModerationRole role;
  final AdminRepository? repository;

  @override
  State<AdminDashboardScreen> createState() => _AdminDashboardScreenState();
}

class _AdminDashboardScreenState extends State<AdminDashboardScreen> {
  late final AdminRepository _repository =
      widget.repository ?? DemoAdminRepository();
  _AdminSection _section = _AdminSection.users;

  @override
  Widget build(BuildContext context) {
    if (widget.role != ModerationRole.admin) {
      return Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(title: const Text('Administración')),
        body: const Center(
          child: Text(
            'Esta sección está disponible solo para administradores.',
          ),
        ),
      );
    }

    final wide = MediaQuery.sizeOf(context).width >= 900;
    return Theme(
      data: Theme.of(context).copyWith(
        inputDecorationTheme: const InputDecorationTheme(
          filled: true,
          fillColor: Colors.white,
          border: OutlineInputBorder(
            borderRadius: BorderRadius.all(Radius.circular(8)),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.all(Radius.circular(8)),
            borderSide: BorderSide(color: Color(0xFFB8C7C4)),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.all(Radius.circular(8)),
            borderSide: BorderSide(color: Color(0xFF16877F), width: 2),
          ),
        ),
      ),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: const Text('Administración'),
          backgroundColor: const Color(0xFF1ABC9C),
          foregroundColor: Colors.white,
          actions: wide
              ? null
              : [
                  IconButton(
                    tooltip: 'Cerrar sesión',
                    onPressed: _logout,
                    icon: const Icon(Icons.logout),
                  ),
                ],
        ),
        body: Stack(
          fit: StackFit.expand,
          children: [
            Image.asset(
              'assets/imagenes/FondoAdmin.png',
              fit: BoxFit.cover,
              alignment: Alignment.center,
            ),
            if (wide)
              Row(
                children: [
                  NavigationRail(
                    backgroundColor: Colors.transparent,
                    groupAlignment: -1,
                    trailing: Padding(
                      padding: const EdgeInsets.only(top: 16),
                      child: TextButton.icon(
                        key: const ValueKey('admin-logout'),
                        onPressed: _logout,
                        icon: const Icon(Icons.logout),
                        label: const Text('Cerrar sesión'),
                      ),
                    ),
                    selectedIndex: _section.index,
                    onDestinationSelected: (index) =>
                        setState(() => _section = _AdminSection.values[index]),
                    labelType: NavigationRailLabelType.all,
                    destinations: [
                      for (final section in _AdminSection.values)
                        NavigationRailDestination(
                          icon: Icon(section.icon),
                          label: Text(section.label),
                        ),
                    ],
                  ),
                  const VerticalDivider(width: 1),
                  Expanded(child: _buildSection()),
                ],
              )
            else
              _buildSection(),
          ],
        ),
        bottomNavigationBar: wide
            ? null
            : NavigationBar(
                backgroundColor: Colors.transparent,
                elevation: 0,
                selectedIndex: _section.index,
                onDestinationSelected: (index) =>
                    setState(() => _section = _AdminSection.values[index]),
                destinations: [
                  for (final section in _AdminSection.values)
                    NavigationDestination(
                      icon: Icon(section.icon),
                      label: section.label,
                    ),
                ],
              ),
      ),
    );
  }

  void _logout() {
    NotificationScope.maybeOf(context)?.stop();
    Navigator.of(context).pushAndRemoveUntil<void>(
      MaterialPageRoute<void>(builder: (_) => const AuthScreen()),
      (route) => false,
    );
  }

  Widget _buildSection() => switch (_section) {
    _AdminSection.users => AdminUsersScreen(repository: _repository),
    _AdminSection.moderators => AdminModeratorsScreen(repository: _repository),
    _AdminSection.rewards => AdminRewardsScreen(repository: _repository),
    _AdminSection.moderation => const ModerationQueueScreen(
      role: ModerationRole.admin,
    ),
  };
}
