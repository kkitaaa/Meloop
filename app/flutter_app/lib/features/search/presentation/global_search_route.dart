import 'package:flutter/material.dart';

import 'global_search_screen.dart';

void openGlobalSearch(BuildContext context) {
  Navigator.of(context).push<void>(
    MaterialPageRoute<void>(builder: (_) => const GlobalSearchScreen()),
  );
}
