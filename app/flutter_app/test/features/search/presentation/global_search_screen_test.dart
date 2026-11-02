import 'package:flutter/material.dart';
import 'package:flutter_app/features/search/presentation/global_search_screen.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('search updates when selecting Users, Posts, and Music', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(400, 800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(const MaterialApp(home: GlobalSearchScreen()));

    await tester.enterText(find.byType(TextField), 'rock');
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();
    expect(find.text('Rockero'), findsOneWidget);

    await tester.tap(find.text('Publicaciones'));
    await tester.pump();
    expect(find.text('Mis discos favoritos de rock'), findsOneWidget);

    await tester.tap(find.text('Música'));
    await tester.pump();
    expect(find.text('Rock'), findsOneWidget);
  });
}
