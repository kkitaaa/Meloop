import 'package:flutter/material.dart';
import 'package:flutter_app/screens/api_test_screen.dart';
import 'package:flutter_app/screens/feed_screen.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('GET and POST communicate with the running API Gateway', (
    tester,
  ) async {
    await tester.pumpWidget(const MaterialApp(home: ApiTestScreen()));
    await tester.pumpAndSettle();

    expect(
      find.text('Conexión con el API Gateway funcionando.'),
      findsOneWidget,
    );

    await tester.tap(find.text('Enviar POST'));
    await tester.pumpAndSettle();

    expect(
      find.text('Respuesta del servidor: Hola desde Flutter'),
      findsOneWidget,
    );
    expect(find.textContaining('No se pudo conectar'), findsNothing);
  });

  testWidgets('feed loads posts from the running API Gateway', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: FeedScreen()));
    await tester.pumpAndSettle(const Duration(seconds: 1));

    expect(
      find.textContaining('Contenido de demostración servido'),
      findsWidgets,
    );
    expect(find.text('Gateway no disponible.'), findsNothing);
  });
}
