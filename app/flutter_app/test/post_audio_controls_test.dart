import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_app/services/post_audio_playback.dart';
import 'package:flutter_app/widgets/post_audio_controls.dart';
import 'package:flutter_app/widgets/post_card.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:just_audio/just_audio.dart';
import 'package:mocktail/mocktail.dart';

class MockAudioPlayer extends Mock implements AudioPlayer {}

class FakeAudioSource extends Fake implements AudioSource {}

void main() {
  setUpAll(() {
    registerFallbackValue(FakeAudioSource());
  });

  testWidgets(
    'renders playback and progress controls for an audio attachment',
    (tester) async {
      final player = _createMockAudioPlayer();
      final controller = PostAudioPlaybackController.forTesting(player);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: PostAudioControls(
              postId: 'post-1',
              title: 'Canción de prueba',
              artist: 'Artista',
              audioBytes: Uint8List.fromList([1, 2, 3]),
              audioUrl: null,
              fileName: 'track.mp3',
              accentColor: Colors.teal,
              playbackController: controller,
            ),
          ),
        ),
      );

      expect(find.byTooltip('Reproducir audio'), findsOneWidget);
      expect(find.byType(Slider), findsOneWidget);
      expect(find.text('Canción de prueba'), findsOneWidget);

      await tester.pumpWidget(const SizedBox.shrink());
      await tester.runAsync(() => Future<void>.delayed(Duration.zero));

      verify(() => player.dispose()).called(1);
    },
  );

  testWidgets('PostCard connects attached audio to playback controls', (
    tester,
  ) async {
    final player = _createMockAudioPlayer();
    final controller = PostAudioPlaybackController.forTesting(player);

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PostCard(
            postData: {
              'id': 42,
              'user': 'Vicente',
              'content': 'Adjunto de audio',
              'song': 'Canción adjunta',
              'artist': 'Artista',
              'audio_name': 'track.mp3',
              'audio_bytes': Uint8List.fromList([1, 2, 3]),
            },
            tealAccent: Colors.teal,
            audioPlaybackController: controller,
          ),
        ),
      ),
    );

    expect(find.text('Canción adjunta'), findsOneWidget);
    expect(find.byTooltip('Reproducir audio'), findsOneWidget);
    expect(find.byType(Slider), findsOneWidget);

    await tester.pumpWidget(const SizedBox.shrink());
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    verify(() => player.dispose()).called(1);
  });

  testWidgets('switches tracks through the same shared player', (tester) async {
    final player = _createMockAudioPlayer();
    final controller = PostAudioPlaybackController.forTesting(player);
    final bytes = Uint8List.fromList([1, 2, 3]);

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              PostAudioControls(
                postId: 'post-1',
                title: 'Primera canción',
                artist: 'Artista',
                audioBytes: bytes,
                audioUrl: null,
                fileName: 'first.mp3',
                accentColor: Colors.teal,
                playbackController: controller,
              ),
              PostAudioControls(
                postId: 'post-2',
                title: 'Segunda canción',
                artist: 'Artista',
                audioBytes: bytes,
                audioUrl: null,
                fileName: 'second.mp3',
                accentColor: Colors.teal,
                playbackController: controller,
              ),
            ],
          ),
        ),
      ),
    );

    await tester.tap(find.byTooltip('Reproducir audio').first);
    await tester.pump();
    await tester.tap(find.byTooltip('Reproducir audio').last);
    await tester.pump();

    expect(controller.activePostId, 'post-2');
    verify(() => player.setAudioSource(any())).called(2);
    verify(() => player.stop()).called(2);

    await tester.pumpWidget(const SizedBox.shrink());
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    verify(() => player.dispose()).called(1);
  });

  testWidgets('releases the active source when its PostCard leaves the feed', (
    tester,
  ) async {
    final player = _createMockAudioPlayer();
    final controller = PostAudioPlaybackController.forTesting(player);
    final bytes = Uint8List.fromList([1, 2, 3]);
    var showFirstCard = true;
    late VoidCallback hideFirstCard;

    await tester.pumpWidget(
      MaterialApp(
        home: StatefulBuilder(
          builder: (context, setState) {
            hideFirstCard = () => setState(() => showFirstCard = false);
            return Scaffold(
              body: Column(
                children: [
                  if (showFirstCard)
                    PostAudioControls(
                      key: const ValueKey('post-1'),
                      postId: 'post-1',
                      title: 'Primera canción',
                      artist: 'Artista',
                      audioBytes: bytes,
                      audioUrl: null,
                      fileName: 'first.mp3',
                      accentColor: Colors.teal,
                      playbackController: controller,
                    ),
                  PostAudioControls(
                    key: const ValueKey('post-2'),
                    postId: 'post-2',
                    title: 'Segunda canción',
                    artist: 'Artista',
                    audioBytes: bytes,
                    audioUrl: null,
                    fileName: 'second.mp3',
                    accentColor: Colors.teal,
                    playbackController: controller,
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );

    await tester.tap(find.byTooltip('Reproducir audio').first);
    await tester.pump();
    hideFirstCard();
    await tester.pump();
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));

    expect(controller.activePostId, isNull);
    verify(() => player.clearAudioSources()).called(1);

    await tester.pumpWidget(const SizedBox.shrink());
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    verify(() => player.dispose()).called(1);
  });
}

MockAudioPlayer _createMockAudioPlayer() {
  final player = MockAudioPlayer();
  when(() => player.playerStateStream).thenAnswer((_) => const Stream.empty());
  when(() => player.positionStream).thenAnswer((_) => const Stream.empty());
  when(() => player.durationStream).thenAnswer((_) => const Stream.empty());
  when(() => player.playing).thenReturn(false);
  when(() => player.processingState).thenReturn(ProcessingState.idle);
  when(() => player.stop()).thenAnswer((_) async {});
  when(() => player.clearAudioSources()).thenAnswer((_) async {});
  when(() => player.setAudioSource(any()))
      .thenAnswer((_) async => const Duration(seconds: 30));
  when(() => player.setUrl(any())).thenAnswer((_) async => null);
  when(() => player.play()).thenAnswer((_) async {});
  when(() => player.pause()).thenAnswer((_) async {});
  when(() => player.seek(any())).thenAnswer((_) async {});
  when(() => player.dispose()).thenAnswer((_) async {});
  return player;
}
