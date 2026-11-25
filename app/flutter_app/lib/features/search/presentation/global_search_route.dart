import 'package:flutter/material.dart';

import '../../../core/navigation/desktop_navigation_shell.dart';
import 'global_search_screen.dart';

void openGlobalSearch(BuildContext context) {
  final shell = DesktopShellScope.maybeOf(context);
  if (shell != null) {
    shell.openSearch();
    return;
  }
  Navigator.of(context).push<void>(
    MaterialPageRoute<void>(builder: (_) => const GlobalSearchScreen()),
  );
}
