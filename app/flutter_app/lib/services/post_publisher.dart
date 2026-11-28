import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_app/models/music_track.dart';

class PostImageUpload {
  final Uint8List bytes;
  final String name;

  const PostImageUpload({required this.bytes, required this.name});
}

class PostPublishResult {
  final Map<String, dynamic> post;
  final int? xpAwarded;

  const PostPublishResult({required this.post, this.xpAwarded});
}

abstract interface class PostPublisher {
  Future<PostPublishResult> publishPost({
    required String content,
    required List<PostImageUpload> images,
    Uint8List? audioBytes,
    String? audioName,
    MusicTrack? song,
    required CancelToken cancelToken,
    required void Function(int sent, int total) onSendProgress,
  });
}
