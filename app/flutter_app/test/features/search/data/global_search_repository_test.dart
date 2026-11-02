import 'package:flutter_app/features/search/data/global_search_models.dart';
import 'package:flutter_app/features/search/data/global_search_repository.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const repository = DemoGlobalSearchRepository();

  test('search filters by category and paginates results', () async {
    final firstPage = await repository.search(
      query: 'rock',
      category: GlobalSearchCategory.music,
      page: 1,
      pageSize: 8,
    );
    final secondPage = await repository.search(
      query: 'rock',
      category: GlobalSearchCategory.music,
      page: 2,
      pageSize: 8,
    );

    expect(firstPage.results, hasLength(8));
    expect(firstPage.hasMore, isTrue);
    expect(secondPage.results, hasLength(3));
    expect(secondPage.hasMore, isFalse);
    expect(
      firstPage.results.every(
        (result) => result.category == GlobalSearchCategory.music,
      ),
      isTrue,
    );
  });

  test('empty query returns no demo results', () async {
    final page = await repository.search(
      query: '  ',
      category: GlobalSearchCategory.users,
      page: 1,
      pageSize: 8,
    );

    expect(page.results, isEmpty);
    expect(page.hasMore, isFalse);
  });
}
