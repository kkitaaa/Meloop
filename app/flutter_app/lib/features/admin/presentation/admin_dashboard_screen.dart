import 'package:flutter/material.dart';

import '../../../core/auth/role_guard.dart';
import '../../moderation/presentation/moderation_queue_screen.dart';
import '../../notifications/presentation/notification_scope.dart';
import '../../../screens/api_test_screen.dart';
import '../../../services/api_client.dart';
import '../../../services/auth_session.dart';
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
  const AdminDashboardScreen({
    super.key,
    this.session,
    this.repository,
    this.apiClient,
  });

  final AuthSession? session;
  final AdminRepository? repository;
  final ApiClient? apiClient;

  @override
  State<AdminDashboardScreen> createState() => _AdminDashboardScreenState();
}

class _AdminDashboardScreenState extends State<AdminDashboardScreen> {
  late final AuthSession _session = widget.session ?? AuthSession.shared;
  late final AdminRepository _repository =
      widget.repository ?? DemoAdminRepository();
  _AdminSection _section = _AdminSection.users;

  @override
  Widget build(BuildContext context) {
    return RoleGuard(
      permission: AppPermission.administration,
      session: _session,
      child: _buildDashboard(context),
    );
  }

  Widget _buildDashboard(BuildContext context) {
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
          backgroundColor: Colors.transparent,
          foregroundColor: Colors.white,
          shape: const RoundedRectangleBorder(
            borderRadius: BorderRadius.vertical(bottom: Radius.circular(24)),
          ),
          flexibleSpace: const DecoratedBox(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [Color(0xFF59E5C9), Color(0xFF00AF92)],
              ),
            ),
          ),
          actions: [
            TextButton.icon(
              key: const ValueKey('admin-api-test'),
              style: TextButton.styleFrom(foregroundColor: Colors.white),
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => ApiTestScreen(apiClient: widget.apiClient),
                ),
              ),
              icon: const Icon(Icons.cloud_outlined, size: 18),
              label: const Text('Probar API'),
            ),
            if (!wide)
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

  Future<void> _logout() async {
    NotificationScope.maybeOf(context)?.stop();
    try {
      if (widget.apiClient == null) {
        await _session.clear();
      } else {
        await widget.apiClient!.logout();
      }
    } on ApiException catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Se cerró la sesión localmente: ${error.message}'),
        ),
      );
      await _session.clear();
    }
  }

  Widget _buildSection() => switch (_section) {
    _AdminSection.users => AdminUsersScreen(repository: _repository),
    _AdminSection.moderators => AdminModeratorsScreen(repository: _repository),
    _AdminSection.rewards => AdminRewardsScreen(repository: _repository),
    _AdminSection.moderation => ModerationQueueScreen(session: _session),
  };
}
