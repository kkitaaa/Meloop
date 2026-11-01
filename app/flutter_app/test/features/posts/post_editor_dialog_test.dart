import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/features/posts/presentation/post_editor_dialog.dart';

void main() {
  group('PostAttachmentValidator', () {
    test('acepta formatos y tamaños de imagen permitidos', () {
      expect(
        PostAttachmentValidator.validateImage(
          name: 'cover.PNG',
          sizeBytes: maxPostImageSizeBytes,
        ),
        isNull,
      );
    });

    test('rechaza formatos de imagen no permitidos y excedentes', () {
      expect(
        PostAttachmentValidator.validateImage(
          name: 'cover.gif',
          sizeBytes: 100,
        ),
        contains('Formato'),
      );
      expect(
        PostAttachmentValidator.validateImage(
          name: 'cover.jpg',
          sizeBytes: maxPostImageSizeBytes + 1,
        ),
        contains('5 MB'),
      );
    });

    test('valida formato y límite de tamaño de audio', () {
      expect(
        PostAttachmentValidator.validateAudio(
          name: 'track.m4a',
          sizeBytes: maxPostAudioSizeBytes,
        ),
        isNull,
      );
      expect(
        PostAttachmentValidator.validateAudio(name: 'track.txt', sizeBytes: 1),
        contains('Formato'),
      );
      expect(
        PostAttachmentValidator.validateAudio(
          name: 'track.mp3',
          sizeBytes: maxPostAudioSizeBytes + 1,
        ),
        contains('15 MB'),
      );
    });
  });

  group('PostEditorDialog', () {
    testWidgets('desactiva Publicar con contenido vacío', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(body: PostEditorDialog(tealAccent: Color(0xFF1ABC9C))),
        ),
      );

      final publishButton = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Publicar'),
      );
      expect(publishButton.onPressed, isNull);
    });

    testWidgets('permite quitar la canción seleccionada antes de guardar', (
      tester,
    ) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: PostEditorDialog(
              postData: {'song': 'Canción de prueba'},
              tealAccent: const Color(0xFF1ABC9C),
            ),
          ),
        ),
      );

      expect(find.text('Canción de prueba'), findsOneWidget);
      await tester.tap(find.byTooltip('Quitar canción'));
      await tester.pump();

      final saveButton = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Guardar cambios'),
      );
      expect(saveButton.onPressed, isNull);
    });

    testWidgets('permite publicar texto y bloquea mientras está enviando', (
      tester,
    ) async {
      PostEditorResult? result;
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => Scaffold(
              body: TextButton(
                onPressed: () async {
                  result = await showDialog<PostEditorResult>(
                    context: context,
                    builder: (context) =>
                        const PostEditorDialog(tealAccent: Color(0xFF1ABC9C)),
                  );
                },
                child: const Text('Abrir editor'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Abrir editor'));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byType(TextField),
        'Una publicación con texto',
      );
      await tester.pump();

      var publishButton = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Publicar'),
      );
      expect(publishButton.onPressed, isNotNull);

      await tester.tap(find.text('Publicar'));
      await tester.pump();
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(
        tester
            .widget<ElevatedButton>(find.byType(ElevatedButton).last)
            .onPressed,
        isNull,
      );

      await tester.pump(const Duration(seconds: 1));
      await tester.pumpAndSettle();
      expect(result?.content, 'Una publicación con texto');
    });
  });
}
