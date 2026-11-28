import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/features/comments/presentation/comments_screen.dart';

void main() {
  final Map<String, dynamic> mockPostData = {
    "id": 1,
    "user": "CreadorDelPost",
    "content": "Este es el post principal que estamos comentando",
    "likes": 10,
    "comments": 2,
    "isLiked": false,
  };

  Widget createWidgetUnderTest() {
    return MaterialApp(home: CommentsScreen(postData: mockPostData));
  }

  group('CommentsScreen Widget Tests', () {
    testWidgets(
      'Renderiza la publicación original y la lista base de comentarios',
      (WidgetTester tester) async {
        await tester.pumpWidget(createWidgetUnderTest());

        // Publicación original
        expect(find.text("CreadorDelPost"), findsOneWidget);
        expect(
          find.text("Este es el post principal que estamos comentando"),
          findsOneWidget,
        );

        // Comentarios precargados
        expect(find.text("Martín"), findsOneWidget);
        expect(find.text("Camila"), findsOneWidget);
      },
    );

    testWidgets(
      'Responder a un comentario activa la anidación en la barra inferior',
      (WidgetTester tester) async {
        await tester.pumpWidget(createWidgetUnderTest());

        // Buscar los botones "Responder"
        final responderBtns = find.text("Responder");
        expect(responderBtns, findsWidgets);

        // Tocar el primer botón (Responder a Martín)
        await tester.tap(responderBtns.first);
        await tester.pumpAndSettle();

        // Validar estado visual de anidación
        expect(find.text("Respondiendo a @Martín"), findsOneWidget);
      },
    );

    testWidgets(
      'Simulación de error del servidor (RN-03) al usar palabras prohibidas',
      (WidgetTester tester) async {
        await tester.pumpWidget(createWidgetUnderTest());

        // Ingresar texto que detona la regla de negocio
        await tester.enterText(find.byType(TextField), "esto es un insulto");
        await tester.tap(find.byIcon(Icons.send));

        await tester.pump(); // Inicia la carga
        await tester.pump(const Duration(seconds: 1)); // Simula la espera HTTP
        await tester.pumpAndSettle();

        // Verifica el estado de error sin añadir el comentario
        expect(find.textContaining("incumple las normas"), findsOneWidget);
      },
    );

    testWidgets(
      'Crear un comentario válido lo añade inmediatamente a la lista',
      (WidgetTester tester) async {
        await tester.pumpWidget(createWidgetUnderTest());

        await tester.enterText(
          find.byType(TextField),
          "¡Excelente publicación, me encantó!",
        );
        await tester.tap(find.byIcon(Icons.send));

        await tester.pump();
        await tester.pump(const Duration(seconds: 1));
        await tester.pumpAndSettle();

        // Verifica que el comentario se añadió a la vista
        expect(
          find.text("¡Excelente publicación, me encantó!"),
          findsOneWidget,
        );
        expect(
          find.text("MiUsuario"),
          findsOneWidget,
        ); // Usuario simulado en sesión

        // Verifica que el input se limpió tras enviar
        final TextField input = tester.widget(find.byType(TextField));
        expect(input.controller?.text, isEmpty);
      },
    );
  });
}
