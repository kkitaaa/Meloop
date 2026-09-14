import 'package:flutter_app/main.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('Initial structure smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const MeloopApp());
    expect(find.text('Meloop'), findsOneWidget);
  });
}
