import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:flutter_app/widgets/post_card.dart';

// Mock para simular la función onDelete
class MockDeleteCallback extends Mock {
  void call();
}

void main() {
  const Color tealAccent = Color(0xFF1ABC9C);

  Widget createWidgetUnderTest(
    Map<String, dynamic> postData, {
    VoidCallback? onDelete,
  }) {
    return MaterialApp(
      home: Scaffold(
        body: SingleChildScrollView(
          child: PostCard(
            postData: postData,
            tealAccent: tealAccent,
            onDelete: onDelete,
          ),
        ),
      ),
    );
  }

  group('PostCard Widget Tests', () {
    testWidgets(
      'Renderiza contenido básico (texto, autor, likes y comentarios)',
      (WidgetTester tester) async {
        final postData = {
          "id": 1,
          "user": "VicenteTest",
          "time": "Hace 5 min",
          "content": "Contenido de prueba automatizada",
          "likes": 42,
          "comments": 5,
          "isAuthor": true,
        };

        await tester.pumpWidget(createWidgetUnderTest(postData));

        expect(find.text("VicenteTest"), findsOneWidget);
        expect(find.text("Contenido de prueba automatizada"), findsOneWidget);
        expect(find.text("42 Likes"), findsOneWidget);
        expect(find.text("5 Comentarios"), findsOneWidget);
      },
    );

    testWidgets('Muestra reproductor musical e imagen si vienen en los datos', (
      WidgetTester tester,
    ) async {
      final postData = {
        "id": 2,
        "user": "Camila",
        "content": "Escuchando mi canción favorita",
        "song": "Bohemian Rhapsody",
        "artist": "Queen",
        "image_url": "https://via.placeholder.com/150",
        "isAuthor": false,
      };

      await tester.pumpWidget(createWidgetUnderTest(postData));

      // Verifica componente musical
      expect(find.text("Bohemian Rhapsody"), findsOneWidget);
      expect(find.text("Queen"), findsOneWidget);
      expect(find.byIcon(Icons.graphic_eq), findsOneWidget);

      // Verifica que el widget de Imagen se intenta renderizar
      expect(find.byType(Image), findsOneWidget);
    });

    testWidgets(
      'Muestra opciones correctas en el menú según el ROL (No Autor = Guardar/Reportar)',
      (WidgetTester tester) async {
        final postData = {
          "id": 3,
          "user": "OtroUsuario",
          "content": "Post ajeno",
          "isAuthor": false,
        };

        await tester.pumpWidget(createWidgetUnderTest(postData));

        // Abrir menú
        await tester.tap(find.byIcon(Icons.more_horiz));
        await tester.pumpAndSettle();

        expect(find.text("Guardar publicación"), findsOneWidget);
        expect(find.text("Reportar"), findsOneWidget);
        expect(find.text("Eliminar"), findsNothing); // No debe poder eliminar
      },
    );

    testWidgets('Elimina publicación exitosamente mediante el callback', (
      WidgetTester tester,
    ) async {
      final mockOnDelete = MockDeleteCallback();
      final postData = {
        "id": 4,
        "user": "Autor_Post",
        "content": "Publicación a eliminar",
        "isAuthor": true,
      };

      await tester.pumpWidget(
        createWidgetUnderTest(postData, onDelete: mockOnDelete),
      );

      await tester.tap(find.byIcon(Icons.more_horiz));
      await tester.pumpAndSettle();

      // Tocar eliminar
      await tester.tap(find.text("Eliminar"));
      await tester.pumpAndSettle();

      // Confirmar en el AlertDialog
      await tester.tap(find.widgetWithText(ElevatedButton, "Eliminar"));
      await tester.pump(); // Dispara estado de carga
      await tester.pump(const Duration(seconds: 1)); // Pasa la simulación HTTP
      await tester.pumpAndSettle();

      // Verificar que se llamó al callback y salió el Snackbar
      verify(() => mockOnDelete.call()).called(1);
      expect(find.text("Publicación eliminada correctamente."), findsOneWidget);
    });
  });
}
