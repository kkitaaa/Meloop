// ignore_for_file: experimental_member_use

import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/scheduler.dart';
import 'package:just_audio/just_audio.dart';

class PostAudioPlaybackController extends ChangeNotifier {
  static final PostAudioPlaybackController shared =
      PostAudioPlaybackController._();

  PostAudioPlaybackController._({AudioPlayer Function()? playerFactory})
    : _playerFactory = playerFactory ?? AudioPlayer.new;

  @visibleForTesting
  PostAudioPlaybackController.forTesting(AudioPlayer player)
    : this._(playerFactory: () => player);

  final AudioPlayer Function() _playerFactory;
  AudioPlayer? _player;
  StreamSubscription<PlayerState>? _playerStateSubscription;
  StreamSubscription<Duration>? _positionSubscription;
  StreamSubscription<Duration?>? _durationSubscription;

  String? _activePostId;
  String? _activeSourceKey;
  String? _errorMessage;
  Duration _position = Duration.zero;
  Duration? _duration;
  int _attachedCards = 0;
  int _playRequest = 0;

  String? get activePostId => _activePostId;
  String? get errorMessage => _errorMessage;
  Duration get position => _position;
  Duration? get duration => _duration;

  bool get isPlaying => _activePostId != null && (_player?.playing ?? false);

  bool get isLoading {
    final player = _player;
    return _activePostId != null &&
        (player == null ||
            player.processingState == ProcessingState.loading ||
            player.processingState == ProcessingState.buffering);
  }

  void attachCard() {
    _attachedCards++;
    _ensurePlayer();
  }

  Future<void> toggle({
    required String postId,
    required Uint8List? bytes,
    required String? audioUrl,
    required String sourceKey,
    required String fileName,
  }) async {
    final player = _ensurePlayer();
    final isCurrentSource =
        _activePostId == postId && _activeSourceKey == sourceKey;
    _errorMessage = null;

    if (isCurrentSource) {
      try {
        if (player.playing) {
          await player.pause();
        } else {
          if (player.processingState == ProcessingState.completed) {
            await player.seek(Duration.zero);
          }
          _playRequest++;
          _notifyListeners();
          unawaited(_startPlayback(player, _playRequest));
        }
      } catch (error) {
        _setPlaybackError();
      }
      _notifyListeners();
      return;
    }

    final request = ++_playRequest;
    _activePostId = postId;
    _activeSourceKey = sourceKey;
    _position = Duration.zero;
    _duration = null;
    _notifyListeners();

    try {
      await player.stop();
      if (request != _playRequest) return;
      if (bytes != null) {
        await player.setAudioSource(
          _PostBytesAudioSource(bytes, _contentType(fileName)),
        );
      } else if (audioUrl != null) {
        await player.setUrl(audioUrl);
      } else {
        throw StateError('La publicación no tiene un archivo de audio.');
      }
      if (request != _playRequest) return;
      _notifyListeners();
      unawaited(_startPlayback(player, request));
    } catch (error) {
      if (request != _playRequest) return;
      _errorMessage = 'No se pudo reproducir este audio.';
      _notifyListeners();
    }
  }

  Future<void> seek(Duration position) async {
    final player = _player;
    if (player == null || _activePostId == null) return;

    final duration = _duration;
    final clampedPosition = duration == null
        ? position
        : Duration(
            milliseconds: position.inMilliseconds.clamp(
              0,
              duration.inMilliseconds,
            ),
          );
    try {
      await player.seek(clampedPosition);
    } catch (error) {
      _errorMessage = 'No se pudo cambiar la posición del audio.';
      _notifyListeners();
    }
  }

  void releaseCard(String postId) {
    if (_attachedCards > 0) _attachedCards--;
    if (_activePostId == postId) {
      _playRequest++;
      _activePostId = null;
      _activeSourceKey = null;
      _errorMessage = null;
      _position = Duration.zero;
      _duration = null;
      final player = _player;
      if (player != null && _attachedCards > 0) {
        unawaited(_stopAndClearSource(player, _playRequest));
      }
      _notifyAfterFrame();
    }

    if (_attachedCards == 0) _disposePlayer();
  }

  AudioPlayer _ensurePlayer() {
    final existingPlayer = _player;
    if (existingPlayer != null) return existingPlayer;

    final player = _playerFactory();
    _player = player;
    _playerStateSubscription = player.playerStateStream.listen(
      (_) => _notifyListeners(),
      onError: (Object error) => _setPlaybackError(),
    );
    _positionSubscription = player.positionStream.listen((position) {
      _position = position;
      _notifyListeners();
    }, onError: (Object error) => _setPlaybackError());
    _durationSubscription = player.durationStream.listen((duration) {
      _duration = duration;
      _notifyListeners();
    }, onError: (Object error) => _setPlaybackError());
    return player;
  }

  Future<void> _startPlayback(AudioPlayer player, int request) async {
    try {
      await player.play();
    } catch (error) {
      if (request != _playRequest) return;
      _setPlaybackError();
    }
  }

  Future<void> _stopAndClearSource(AudioPlayer player, int request) async {
    try {
      await player.stop();
      if (request != _playRequest || _activePostId != null) return;
      await player.clearAudioSources();
    } catch (error) {
      FlutterError.reportError(
        FlutterErrorDetails(
          exception: error,
          library: 'post audio playback',
          context: ErrorDescription('while releasing a detached audio track'),
        ),
      );
    }
  }

  void _setPlaybackError() {
    _errorMessage = 'No se pudo reproducir este audio.';
    _notifyListeners();
  }

  void _notifyListeners() {
    notifyListeners();
  }

  void _notifyAfterFrame() {
    SchedulerBinding.instance.addPostFrameCallback((_) => _notifyListeners());
  }

  Future<void> _disposePlayer() async {
    final player = _player;
    _player = null;
    if (player == null) return;

    await _playerStateSubscription?.cancel();
    await _positionSubscription?.cancel();
    await _durationSubscription?.cancel();
    _playerStateSubscription = null;
    _positionSubscription = null;
    _durationSubscription = null;
    await player.dispose();
  }
}

class _PostBytesAudioSource extends StreamAudioSource {
  final Uint8List _bytes;
  final String _contentType;

  _PostBytesAudioSource(this._bytes, this._contentType);

  @override
  Future<StreamAudioResponse> request([int? start, int? end]) async {
    final from = (start ?? 0).clamp(0, _bytes.length);
    final to = (end ?? _bytes.length).clamp(from, _bytes.length);
    final chunk = _bytes.sublist(from, to);
    return StreamAudioResponse(
      sourceLength: _bytes.length,
      contentLength: chunk.length,
      offset: from,
      stream: Stream.value(chunk),
      contentType: _contentType,
    );
  }
}

String _contentType(String fileName) {
  final extension = fileName.split('.').last.toLowerCase();
  return switch (extension) {
    'mp3' => 'audio/mpeg',
    'm4a' => 'audio/mp4',
    'aac' => 'audio/aac',
    'wav' => 'audio/wav',
    'ogg' => 'audio/ogg',
    _ => 'application/octet-stream',
  };
}
