import 'package:flutter_app/main.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('Initial structure smoke test', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    await tester.pumpWidget(const MyApp());

    // Verify that the initial text is displayed.
    expect(
      find.text('Feed Principal\n(Navega al Perfil y la música no se cortará)'),
      findsOneWidget,
    );
  });
}
