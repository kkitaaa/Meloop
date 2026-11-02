enum GlobalSearchCategory { users, posts, music }

class GlobalSearchResult {
  const GlobalSearchResult({
    required this.id,
    required this.category,
    required this.title,
    required this.subtitle,
    required this.description,
  });

  final String id;
  final GlobalSearchCategory category;
  final String title;
  final String subtitle;
  final String description;
}

class GlobalSearchPage {
  const GlobalSearchPage({required this.results, required this.hasMore});

  final List<GlobalSearchResult> results;
  final bool hasMore;
}
