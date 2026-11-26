import 'package:flutter_app/models/feed_data.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('keeps the persisted like state and count in post-card data', () {
    final post = FeedPost.fromJson({
      'id': 'post-1',
      'user': 'Ana',
      'created_at': '2026-11-24T20:00:00Z',
      'content': 'Publicación con like',
      'likes_count': 3,
      'comments_count': 0,
      'is_liked': true,
    });

    expect(post.likes, 3);
    expect(post.toPostCardData()['is_liked'], isTrue);
  });
}
