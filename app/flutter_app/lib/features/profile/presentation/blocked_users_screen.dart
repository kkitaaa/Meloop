import 'package:flutter/material.dart';

import '../../settings/presentation/settings_page_shell.dart';

class BlockedUsersScreen extends StatelessWidget {
  const BlockedUsersScreen({super.key});

  @override
  Widget build(BuildContext context) => SettingsPageShell(
    title: 'Usuarios bloqueados',
    icon: Icons.block_rounded,
    maxWidth: 720,
    child: ListView(
      padding: const EdgeInsets.all(20),
      children: const [
        SettingsSectionCard(
          title: 'Administrar bloqueos',
          icon: Icons.security_outlined,
          child: Padding(
            padding: EdgeInsets.symmetric(vertical: 8),
            child: Text(
              'La lista de bloqueados no está disponible porque el backend aún no tiene un endpoint para consultarla.',
              style: TextStyle(color: SettingsColors.text),
            ),
          ),
        ),
      ],
    ),
  );
}
