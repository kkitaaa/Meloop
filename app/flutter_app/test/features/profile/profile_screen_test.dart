import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/profile/presentation/profile_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('shows persisted author posts instead of generated samples', (
    tester,
  ) async {
    final dio = Dio();
    dio.httpClientAdapter = _ProfilePostsAdapter((options) {
      expect(options.method, 'GET');
      expect(options.uri.path, '/api/v1/posts/author/author-1');
      return ResponseBody.fromString(
        jsonEncode([
          {
            'id': 'post-1',
            'author_id': 'author-1',
            'content': 'Publicación guardada en la base de datos',
            'created_at': '2026-11-24T20:00:00Z',
          },
        ]),
        200,
        headers: {
          Headers.contentTypeHeader: ['application/json'],
        },
      );
    });
    final client = ApiClient(dio: dio);
    addTearDown(client.close);

    await tester.pumpWidget(
      MaterialApp(
        home: ProfileScreen(userId: 'author-1', apiClient: client),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      find.text('Publicación guardada en la base de datos'),
      findsOneWidget,
    );
    expect(find.textContaining('Publicación número'), findsNothing);
  });
}

class _ProfilePostsAdapter implements HttpClientAdapter {
  _ProfilePostsAdapter(this.respond);

  final ResponseBody Function(RequestOptions options) respond;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return respond(options);
  }

  @override
  void close({bool force = false}) {}
}
