import 'global_search_models.dart';

abstract interface class GlobalSearchRepository {
  Future<GlobalSearchPage> search({
    required String query,
    required GlobalSearchCategory category,
    required int page,
    required int pageSize,
  });
}

class DemoGlobalSearchRepository implements GlobalSearchRepository {
  const DemoGlobalSearchRepository();

  static const _results = <GlobalSearchResult>[
    GlobalSearchResult(
      id: 'user-rock-1',
      category: GlobalSearchCategory.users,
      title: 'Rockero',
      subtitle: '@rockero',
      description: 'Fanático del rock clásico y alternativo.',
    ),
    GlobalSearchResult(
      id: 'user-rock-2',
      category: GlobalSearchCategory.users,
      title: 'Rock en español',
      subtitle: '@rockenespanol',
      description: 'Compartiendo bandas y novedades de rock.',
    ),
    GlobalSearchResult(
      id: 'user-vale',
      category: GlobalSearchCategory.users,
      title: 'Valentina',
      subtitle: '@vale.music',
      description: 'Música, conciertos y nuevas canciones.',
    ),
    GlobalSearchResult(
      id: 'post-rock-1',
      category: GlobalSearchCategory.posts,
      title: 'Mis discos favoritos de rock',
      subtitle: 'Publicado por @rockero',
      description: 'Un recorrido por algunos álbumes imprescindibles.',
    ),
    GlobalSearchResult(
      id: 'post-rock-2',
      category: GlobalSearchCategory.posts,
      title: 'Recomendaciones para descubrir rock',
      subtitle: 'Publicado por @rockenespanol',
      description: '¿Qué banda debería escuchar este fin de semana?',
    ),
    GlobalSearchResult(
      id: 'post-music',
      category: GlobalSearchCategory.posts,
      title: 'Playlist para el viaje',
      subtitle: 'Publicado por @vale.music',
      description: 'Un poco de rock y música para acompañar la ruta.',
    ),
    GlobalSearchResult(
      id: 'music-rock',
      category: GlobalSearchCategory.music,
      title: 'Rock',
      subtitle: 'Género musical',
      description: 'Explora artistas y canciones de rock.',
    ),
    GlobalSearchResult(
      id: 'music-rock-clasico',
      category: GlobalSearchCategory.music,
      title: 'Rock clásico',
      subtitle: 'Género musical',
      description: 'Éxitos y artistas fundamentales del rock.',
    ),
    GlobalSearchResult(
      id: 'music-rock-alternativo',
      category: GlobalSearchCategory.music,
      title: 'Rock alternativo',
      subtitle: 'Género musical',
      description: 'Sonidos alternativos y nuevas propuestas.',
    ),
    GlobalSearchResult(
      id: 'music-rock-indie',
      category: GlobalSearchCategory.music,
      title: 'Rock indie',
      subtitle: 'Género musical',
      description: 'Bandas independientes y canciones para descubrir.',
    ),
    GlobalSearchResult(
      id: 'music-rock-latino',
      category: GlobalSearchCategory.music,
      title: 'Rock latino',
      subtitle: 'Género musical',
      description: 'Artistas y canciones de rock en español.',
    ),
    GlobalSearchResult(
      id: 'music-rock-live',
      category: GlobalSearchCategory.music,
      title: 'Rock en vivo',
      subtitle: 'Lista de reproducción',
      description: 'Presentaciones y grabaciones en directo.',
    ),
    GlobalSearchResult(
      id: 'music-rock-90s',
      category: GlobalSearchCategory.music,
      title: 'Rock de los 90',
      subtitle: 'Lista de reproducción',
      description: 'Una selección de rock de la década de los noventa.',
    ),
    GlobalSearchResult(
      id: 'music-rock-acustico',
      category: GlobalSearchCategory.music,
      title: 'Rock acústico',
      subtitle: 'Lista de reproducción',
      description: 'Versiones acústicas de canciones de rock.',
    ),
    GlobalSearchResult(
      id: 'music-rock-nacional',
      category: GlobalSearchCategory.music,
      title: 'Rock nacional',
      subtitle: 'Género musical',
      description: 'Bandas de rock de la escena local.',
    ),
    GlobalSearchResult(
      id: 'music-rock-punk',
      category: GlobalSearchCategory.music,
      title: 'Rock punk',
      subtitle: 'Género musical',
      description: 'Punk rock y sus distintas escenas.',
    ),
    GlobalSearchResult(
      id: 'music-rock-progressive',
      category: GlobalSearchCategory.music,
      title: 'Rock progresivo',
      subtitle: 'Género musical',
      description: 'Composiciones y sonidos progresivos.',
    ),
  ];

  @override
  Future<GlobalSearchPage> search({
    required String query,
    required GlobalSearchCategory category,
    required int page,
    required int pageSize,
  }) async {
    final normalizedQuery = query.trim().toLowerCase();
    if (normalizedQuery.isEmpty) {
      return const GlobalSearchPage(results: [], hasMore: false);
    }

    final matchingResults = _results.where((result) {
      return result.category == category &&
          '${result.title} ${result.subtitle} ${result.description}'
              .toLowerCase()
              .contains(normalizedQuery);
    }).toList();

    final start = (page - 1) * pageSize;
    if (start >= matchingResults.length) {
      return const GlobalSearchPage(results: [], hasMore: false);
    }
    final end = (start + pageSize).clamp(0, matchingResults.length);
    return GlobalSearchPage(
      results: matchingResults.sublist(start, end),
      hasMore: end < matchingResults.length,
    );
  }
}
