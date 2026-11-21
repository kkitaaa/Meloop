import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/features/posts/presentation/post_editor_dialog.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/post_publisher.dart';

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

    test('limita el total de imágenes a cuatro', () {
      expect(PostAttachmentValidator.validateImageCount(2, 2), isNull);
      expect(
        PostAttachmentValidator.validateImageCount(3, 2),
        contains('hasta 4'),
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
      expect(
        tester.widget<TextField>(find.byType(TextField)).maxLength,
        maxPostContentLength,
      );
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
      await tester.ensureVisible(find.byTooltip('Quitar canción'));
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
                    builder: (context) => PostEditorDialog(
                      tealAccent: const Color(0xFF1ABC9C),
                      publisher: _FakePostPublisher(),
                    ),
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
      expect(find.byType(LinearProgressIndicator), findsOneWidget);
      expect(
        tester
            .widget<ElevatedButton>(find.byType(ElevatedButton).last)
            .onPressed,
        isNull,
      );

      await tester.pump(const Duration(seconds: 1));
      await tester.pumpAndSettle();
      expect(result?.content, 'Una publicación con texto');
      expect(result?.publishedPost?['id'], 42);
    });

    testWidgets('conserva el borrador cuando falla la publicación', (
      tester,
    ) async {
      final publisher = _FakePostPublisher(failure: true);
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => Scaffold(
              body: TextButton(
                onPressed: () => showDialog<void>(
                  context: context,
                  builder: (_) => PostEditorDialog(
                    tealAccent: const Color(0xFF1ABC9C),
                    publisher: publisher,
                  ),
                ),
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
        'Texto que debe conservarse',
      );
      await tester.pump();
      await tester.ensureVisible(find.text('Publicar'));
      expect(
        tester
            .widget<ElevatedButton>(
              find.widgetWithText(ElevatedButton, 'Publicar'),
            )
            .onPressed,
        isNotNull,
      );
      await tester.tap(find.text('Publicar'));
      await tester.pump();
      await tester.pump(const Duration(seconds: 1));
      await tester.pump(const Duration(milliseconds: 300));

      expect(publisher.calls, 1);
      expect(find.text('Texto que debe conservarse'), findsOneWidget);
      expect(
        find.text('No encontramos esa canción en Spotify.'),
        findsOneWidget,
      );
      expect(find.text('Crear nueva publicación'), findsOneWidget);
    });
  });
}

class _FakePostPublisher implements PostPublisher {
  final bool failure;
  int calls = 0;

  _FakePostPublisher({this.failure = false});

  @override
  Future<PostPublishResult> publishPost({
    required String content,
    required List<PostImageUpload> images,
    Uint8List? audioBytes,
    String? audioName,
    MusicTrack? song,
    required CancelToken cancelToken,
    required void Function(int sent, int total) onSendProgress,
  }) async {
    calls++;
    onSendProgress(50, 100);
    if (failure) {
      throw const ApiException(
        statusCode: 422,
        errorCode: 'MUSIC_NOT_FOUND',
        message: 'No encontramos esa canción en Spotify.',
      );
    }
    await Future<void>.delayed(const Duration(seconds: 1));
    return const PostPublishResult(post: {'id': 42}, xpAwarded: 10);
  }
}
