import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../../features/gamification/presentation/gamification_demo_controller.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/search/presentation/global_search_route.dart';
import '../../screens/friends_screen.dart';
import '../../theme/meloop_theme_controller.dart';
import '../../theme/meloop_typography.dart';
import '../../widgets/desktop_app_header.dart';
import '../../widgets/meloop_pill_tabs.dart';
import '../../widgets/meloop_toast_pill.dart';
import '../../widgets/mobile_app_header.dart';
import '../../widgets/paper/stacked_paper_card.dart';
import 'desktop_navigation_shell.dart';
import 'responsive_navigation_shell.dart';

enum _MusicFeedTab { recommended, trending, friendsArtists }

class ArtistsAndSongsScreen extends StatefulWidget {
  const ArtistsAndSongsScreen({super.key, this.initialQuery});

  final String? initialQuery;

  @override
  State<ArtistsAndSongsScreen> createState() => ArtistsAndSongsScreenState();
}

class ArtistsAndSongsScreenState extends State<ArtistsAndSongsScreen> {
  Color get _tealAccent => MeloopThemeController.instance.primaryColor;
  Color get _pillTrackColor => MeloopThemeController.instance.pillTrackColor;
  static const int _pageSize = 15; // 3 filas completas de 5 álbumes en desktop

  final TextEditingController _searchController = TextEditingController();
  final ScrollController _gridScrollController = ScrollController();

  _MusicFeedTab _activeTab = _MusicFeedTab.recommended;
  String _query = '';
  String? _selectedFriendFilter;
  int _loadedCount = _pageSize;
  bool _isLoadingMore = false;

  // TODO(backend): Persistir likes y álbumes favoritos del usuario en el
  // music-service / profile-service cuando existan los endpoints.
  final Set<String> _likedAlbumIds = <String>{};
  final Set<String> _favoriteAlbumIds = <String>{};

  // TODO(backend): Reemplazar este catálogo local por los endpoints reales de
  // recomendaciones y tendencias musicales (Spotify vía music-service) con
  // paginación infinita y búsqueda por álbum, artista y canción.
  static const List<_AlbumEntry> _baseCatalog = [
    _AlbumEntry(
      id: 'alb-1',
      title: 'Graduation',
      artist: 'Kanye West',
      year: '2007',
      genre: 'Hip Hop',
      songs: ['Flashing Lights', 'Stronger', 'Good Morning', 'Homecoming'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFFB556C9),
    ),
    _AlbumEntry(
      id: 'alb-2',
      title: 'Bangarang EP',
      artist: 'Skrillex',
      year: '2011',
      genre: 'Dubstep · Electrónica',
      songs: ['Bangarang', 'Kyoto', 'Summit', 'Right In'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF00AF92),
    ),
    _AlbumEntry(
      id: 'alb-3',
      title: 'Random Access Memories',
      artist: 'Daft Punk',
      year: '2013',
      genre: 'Electrónica · Funk',
      songs: ['Get Lucky', 'Instant Crush', 'Lose Yourself to Dance'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFFD4A037),
    ),
    _AlbumEntry(
      id: 'alb-4',
      title: 'Utopia',
      artist: 'Travis Scott',
      year: '2023',
      genre: 'Hip Hop · Trap',
      songs: ['FE!N', 'I KNOW ?', 'MY EYES', 'TELEKINESIS'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF3E6F8C),
    ),
    _AlbumEntry(
      id: 'alb-5',
      title: 'SCARING THE HOES',
      artist: 'JPEGMAFIA',
      year: '2023',
      genre: 'Hip Hop Experimental',
      songs: ['Lean Beef Patty', 'Steppa Pig', 'Garbage Pale Kids'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFFD9534F),
    ),
    _AlbumEntry(
      id: 'alb-6',
      title: 'Satirología Vol. 2',
      artist: 'Kidd Voodoo',
      year: '2024',
      genre: 'Urbano · Reggaetón',
      songs: ['Minnie <3', 'Enrolar', 'Ese Guille'],
      isRecommended: false,
      isTrending: true,
      vinylLabelColor: Color(0xFF288B83),
    ),
    _AlbumEntry(
      id: 'alb-7',
      title: 'Finding Urself',
      artist: 'DJ Gouz',
      year: '2024',
      genre: 'House · Melodic',
      songs: ['Finding Urself', 'Late Night Pulse', 'Neon Drift'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF0FB49A),
    ),
    _AlbumEntry(
      id: 'alb-8',
      title: 'Jet Set Radio Future OST',
      artist: 'Hideki Naganuma',
      year: '2002',
      genre: 'Funk · Breakbeat',
      songs: ['The Concept of Love', 'Funky Dealer', 'Fly Like a Butterfly'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFFE69D2E),
    ),
    _AlbumEntry(
      id: 'alb-9',
      title: 'Partyson',
      artist: 'Cris MJ',
      year: '2024',
      genre: 'Urbano Chileno',
      songs: ['Gata Only', 'Si No Es Contigo', 'Una Noche en Medellín'],
      isRecommended: false,
      isTrending: true,
      vinylLabelColor: Color(0xFF5A65D8),
    ),
    _AlbumEntry(
      id: 'alb-10',
      title: 'GOTTI A',
      artist: 'Tiago PZK',
      year: '2024',
      genre: 'Pop Urbano · R&B',
      songs: ['Piel', 'Alegria', 'Mi Corazón'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF2E9AA6),
    ),
    _AlbumEntry(
      id: 'alb-11',
      title: 'Currents',
      artist: 'Tame Impala',
      year: '2015',
      genre: 'Psychedelic Pop',
      songs: ['Let It Happen', 'The Less I Know the Better', 'Eventually'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF8C56C9),
    ),
    _AlbumEntry(
      id: 'alb-12',
      title: 'IGOR',
      artist: 'Tyler, The Creator',
      year: '2019',
      genre: 'Neo-Soul · Hip Hop',
      songs: ['EARFQUAKE', 'NEW MAGIC WAND', 'GONE, GONE / THANK YOU'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFFE07EA6),
    ),
    _AlbumEntry(
      id: 'alb-13',
      title: 'DeBÍ TiRAR MáS FOToS',
      artist: 'Bad Bunny',
      year: '2025',
      genre: 'Urbano · Salsa',
      songs: ['BAILE INoLVIDABLE', 'NUEVAYoL', 'DtMF'],
      isRecommended: false,
      isTrending: true,
      vinylLabelColor: Color(0xFF268B59),
    ),
    _AlbumEntry(
      id: 'alb-14',
      title: 'Plastic Hearts',
      artist: 'Miley Cyrus',
      year: '2020',
      genre: 'Pop Rock · Synth',
      songs: ['Midnight Sky', 'Prisoner', 'Angels Like You'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFFD93E8A),
    ),
    _AlbumEntry(
      id: 'alb-15',
      title: 'Fallen',
      artist: 'Evanescence',
      year: '2003',
      genre: 'Rock Alternativo',
      songs: ['Bring Me to Life', 'Going Under', 'My Immortal'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFF326AA8),
    ),
    _AlbumEntry(
      id: 'alb-16',
      title: 'Discovery',
      artist: 'Daft Punk',
      year: '2001',
      genre: 'French House',
      songs: ['One More Time', 'Digital Love', 'Harder, Better, Faster'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF1995AD),
    ),
    _AlbumEntry(
      id: 'alb-17',
      title: 'Blonde',
      artist: 'Frank Ocean',
      year: '2016',
      genre: 'R&B Alternativo',
      songs: ['Nights', 'Pink + White', 'Self Control', 'Ivy'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF4CA67B),
    ),
    _AlbumEntry(
      id: 'alb-18',
      title: 'Meltdown Core',
      artist: 'Nanahira',
      year: '2022',
      genre: 'J-Pop · Denpa · Breakcore',
      songs: ['Bassdrop Freaks', 'Monochrome', 'Candy Rush'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFFE86A92),
    ),
    _AlbumEntry(
      id: 'alb-19',
      title: 'After Hours',
      artist: 'The Weeknd',
      year: '2020',
      genre: 'Synth-Pop · R&B',
      songs: ['Blinding Lights', 'Save Your Tears', 'In Your Eyes'],
      isRecommended: false,
      isTrending: true,
      vinylLabelColor: Color(0xFFC93838),
    ),
    _AlbumEntry(
      id: 'alb-20',
      title: 'Quest for Fire',
      artist: 'Skrillex',
      year: '2023',
      genre: 'Bass · UK Garage',
      songs: ['Rumble', 'Leave Me Like This', 'XENA'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF009A84),
    ),
    _AlbumEntry(
      id: 'alb-21',
      title: 'DESCONTROL',
      artist: 'Sinaka & Easykid',
      year: '2024',
      genre: 'Urbano · Electrónica',
      songs: ['DESCONTROL', 'Noche Oscura', 'Vibe'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFF16A085),
    ),
    _AlbumEntry(
      id: 'alb-22',
      title: 'In Rainbows',
      artist: 'Radiohead',
      year: '2007',
      genre: 'Art Rock',
      songs: ['Weird Fishes/Arpeggi', '15 Step', 'Nude', 'Reckoner'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFFD97228),
    ),
    _AlbumEntry(
      id: 'alb-23',
      title: 'ASTROWORLD',
      artist: 'Travis Scott',
      year: '2018',
      genre: 'Psychedelic Trap',
      songs: ['SICKO MODE', 'STARGAZING', 'BUTTERFLY EFFECT'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFFC98B2B),
    ),
    _AlbumEntry(
      id: 'alb-24',
      title: 'Nurture',
      artist: 'Porter Robinson',
      year: '2021',
      genre: 'Electropop · Indie',
      songs: ['Get Your Wish', 'Look at the Sky', 'Something Comforting'],
      isRecommended: true,
      isTrending: false,
      vinylLabelColor: Color(0xFF56B89A),
    ),
    _AlbumEntry(
      id: 'alb-25',
      title: 'Yeezus',
      artist: 'Kanye West',
      year: '2013',
      genre: 'Industrial Hip Hop',
      songs: ['On Sight', 'Black Skinhead', 'Bound 2'],
      isRecommended: true,
      isTrending: true,
      vinylLabelColor: Color(0xFFE0483E),
    ),
  ];

  // TODO(backend): Reemplazar por la actividad real de escucha de los amigos
  // del usuario (p. ej. `GET /v1/social/friends/listening-artists`).
  static const List<_FriendListeningArtist> _friendArtists = [
    _FriendListeningArtist(
      artist: 'Skrillex',
      genres: 'Dubstep · Bass · Electrónica',
      topTrack: 'Bangarang',
      listeners: ['Vicente', 'Felipe'],
      isLiveNow: true,
    ),
    _FriendListeningArtist(
      artist: 'Kanye West',
      genres: 'Hip Hop · Soul',
      topTrack: 'Flashing Lights',
      listeners: ['Vicente', 'Andrea'],
      isLiveNow: true,
    ),
    _FriendListeningArtist(
      artist: 'Kidd Voodoo',
      genres: 'Urbano · Reggaetón',
      topTrack: 'Enrolar',
      listeners: ['Andrea', 'Valentina'],
      isLiveNow: false,
    ),
    _FriendListeningArtist(
      artist: 'Daft Punk',
      genres: 'Electrónica · Funk',
      topTrack: 'Instant Crush',
      listeners: ['Felipe', 'Valentina'],
      isLiveNow: true,
    ),
    _FriendListeningArtist(
      artist: 'Travis Scott',
      genres: 'Hip Hop · Trap',
      topTrack: 'MY EYES',
      listeners: ['Vicente'],
      isLiveNow: false,
    ),
    _FriendListeningArtist(
      artist: 'Tame Impala',
      genres: 'Psychedelic Pop · Indie',
      topTrack: 'Let It Happen',
      listeners: ['Andrea'],
      isLiveNow: false,
    ),
    _FriendListeningArtist(
      artist: 'DJ Gouz',
      genres: 'House · Melodic',
      topTrack: 'Finding Urself',
      listeners: ['Felipe'],
      isLiveNow: true,
    ),
    _FriendListeningArtist(
      artist: 'JPEGMAFIA',
      genres: 'Hip Hop Experimental',
      topTrack: 'Lean Beef Patty',
      listeners: ['Valentina'],
      isLiveNow: false,
    ),
  ];

  static const List<String> _friendNames = [
    'Vicente',
    'Andrea',
    'Felipe',
    'Valentina',
  ];

  @override
  void initState() {
    super.initState();
    _gridScrollController.addListener(_onGridScroll);
    final initial = widget.initialQuery?.trim();
    if (initial != null && initial.isNotEmpty) {
      _searchController.text = initial;
      _query = initial;
    }
  }

  void applySearchQuery(String query) {
    final trimmed = query.trim();
    setState(() {
      _activeTab = _MusicFeedTab.recommended;
      _selectedFriendFilter = null;
      _searchController.text = trimmed;
      _query = trimmed;
      _loadedCount = _pageSize;
    });
    if (_gridScrollController.hasClients) {
      _gridScrollController.jumpTo(0);
    }
  }

  @override
  void dispose() {
    _gridScrollController.removeListener(_onGridScroll);
    _gridScrollController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  void _onGridScroll() {
    if (!_gridScrollController.hasClients) return;
    if (_gridScrollController.position.pixels >=
        _gridScrollController.position.maxScrollExtent - 260) {
      _loadMoreAlbums();
    }
  }

  void _loadMoreAlbums() {
    if (_isLoadingMore) return;
    final pool = _filteredAlbumPool;
    if (pool.isEmpty) return;

    // En modo búsqueda mostramos todos los resultados que coinciden;
    // sin búsqueda activa generamos páginas adicionales en scroll infinito.
    if (_query.trim().isNotEmpty && _loadedCount >= pool.length) return;

    setState(() => _isLoadingMore = true);
    Future<void>.delayed(const Duration(milliseconds: 220), () {
      if (!mounted) return;
      setState(() {
        _loadedCount += _pageSize;
        _isLoadingMore = false;
      });
    });
  }

  List<_AlbumEntry> get _filteredAlbumPool {
    final normalizedQuery = _query.trim().toLowerCase();
    final tabFiltered = _baseCatalog.where((album) {
      if (_activeTab == _MusicFeedTab.trending) return album.isTrending;
      return album.isRecommended;
    });

    final source = normalizedQuery.isEmpty ? tabFiltered : _baseCatalog;
    if (normalizedQuery.isEmpty) return source.toList();

    return source.where((album) {
      if (album.title.toLowerCase().contains(normalizedQuery)) return true;
      if (album.artist.toLowerCase().contains(normalizedQuery)) return true;
      return album.songs.any(
        (song) => song.toLowerCase().contains(normalizedQuery),
      );
    }).toList();
  }

  List<_AlbumEntry> get _visibleAlbums {
    final pool = _filteredAlbumPool;
    if (pool.isEmpty) return const [];
    if (_query.trim().isNotEmpty) return pool;

    return List<_AlbumEntry>.generate(
      _loadedCount,
      (index) => pool[index % pool.length],
    );
  }

  String? _matchingSongFor(_AlbumEntry album) {
    final q = _query.trim().toLowerCase();
    if (q.isEmpty) return null;
    for (final song in album.songs) {
      if (song.toLowerCase().contains(q)) return song;
    }
    return null;
  }

  List<_FriendListeningArtist> get _visibleFriendArtists {
    if (_selectedFriendFilter == null) return _friendArtists;
    return _friendArtists
        .where((entry) => entry.listeners.contains(_selectedFriendFilter))
        .toList();
  }

  void _selectTab(_MusicFeedTab tab) {
    if (_activeTab == tab) return;
    setState(() {
      _activeTab = tab;
      _loadedCount = _pageSize;
    });
    if (_gridScrollController.hasClients) {
      _gridScrollController.jumpTo(0);
    }
  }

  void _filterByArtistFromSidebar(String artistName) {
    setState(() {
      _searchController.text = artistName;
      _query = artistName;
      if (_activeTab == _MusicFeedTab.friendsArtists) {
        _activeTab = _MusicFeedTab.recommended;
      }
    });
    if (_gridScrollController.hasClients) {
      _gridScrollController.jumpTo(0);
    }
  }

  bool _toggleAlbumLike(_AlbumEntry album) {
    final nowLiked = !_likedAlbumIds.contains(album.id);
    setState(() {
      if (nowLiked) {
        _likedAlbumIds.add(album.id);
      } else {
        _likedAlbumIds.remove(album.id);
      }
    });
    showMeloopToastPill(
      context,
      message: nowLiked
          ? 'Te gusta "${album.title}"'
          : 'Quitaste el me gusta de "${album.title}"',
      icon: nowLiked ? Icons.favorite_rounded : Icons.favorite_border_rounded,
    );
    return nowLiked;
  }

  bool _toggleAlbumFavorite(_AlbumEntry album) {
    final nowFavorite = !_favoriteAlbumIds.contains(album.id);
    setState(() {
      if (nowFavorite) {
        _favoriteAlbumIds.add(album.id);
      } else {
        _favoriteAlbumIds.remove(album.id);
      }
    });
    showMeloopToastPill(
      context,
      message: nowFavorite
          ? '"${album.title}" añadido a favoritos'
          : '"${album.title}" quitado de favoritos',
      icon: nowFavorite ? Icons.star_rounded : Icons.star_border_rounded,
    );
    return nowFavorite;
  }

  void _openAlbumPreview(_AlbumEntry album) {
    showGeneralDialog<void>(
      context: context,
      barrierDismissible: true,
      barrierLabel: 'Cerrar detalle del álbum',
      barrierColor: Colors.black.withValues(alpha: 0.54),
      transitionDuration: const Duration(milliseconds: 340),
      pageBuilder: (dialogContext, animation, secondaryAnimation) {
        return Center(
          child: _AlbumDetailPaperDialog(
            album: album,
            initiallyLiked: _likedAlbumIds.contains(album.id),
            initiallyFavorite: _favoriteAlbumIds.contains(album.id),
            onToggleLike: () => _toggleAlbumLike(album),
            onToggleFavorite: () => _toggleAlbumFavorite(album),
            onClose: () => Navigator.of(dialogContext).pop(),
          ),
        );
      },
      transitionBuilder: (dialogContext, animation, secondaryAnimation, child) {
        // Al cerrar: desvanecimiento limpio (solo fade).
        if (animation.status == AnimationStatus.reverse) {
          final fadeOut = CurvedAnimation(
            parent: animation,
            curve: Curves.easeOut,
          );
          return FadeTransition(opacity: fadeOut, child: child);
        }

        // Al abrir: la hoja entra lanzada desde el borde derecho con leve rotación
        // y aterriza suavemente en el centro de la pantalla.
        final curved = CurvedAnimation(
          parent: animation,
          curve: Curves.easeOutCubic,
        );
        final slideOffset = Tween<Offset>(
          begin: const Offset(1.12, -0.06),
          end: Offset.zero,
        ).evaluate(curved);
        final angle = Tween<double>(begin: 0.10, end: 0.0).evaluate(curved);
        final scale = Tween<double>(begin: 1.04, end: 1.0).evaluate(curved);

        return FadeTransition(
          opacity: Tween<double>(begin: 0.35, end: 1.0).animate(curved),
          child: FractionalTranslation(
            translation: slideOffset,
            child: Transform.rotate(
              angle: angle,
              child: Transform.scale(scale: scale, child: child),
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.sizeOf(context).width;
    final isDesktop = screenWidth > 900;
    final isAndroidMobile =
        !isDesktop && defaultTargetPlatform == TargetPlatform.android;
    final insideDesktopShell = DesktopShellScope.maybeOf(context) != null;
    final insideMobileShell = MobileShellScope.maybeOf(context) != null;
    final hideTopBar = insideDesktopShell || insideMobileShell;

    if (isDesktop && _activeTab == _MusicFeedTab.friendsArtists) {
      _activeTab = _MusicFeedTab.recommended;
    }

    return AnimatedBuilder(
      animation: MeloopThemeController.instance,
      builder: (context, _) => Scaffold(
        backgroundColor: Colors.transparent,
        body: Column(
          children: [
            if (!hideTopBar && isAndroidMobile)
              MobileAppHeader(onSearchTap: () => openGlobalSearch(context))
            else if (!hideTopBar && isDesktop)
              DesktopAppHeader(
                activeSection: DesktopNavigationSection.artists,
                onHomeTap: () => navigateToDesktopHome(context),
                onFriendsTap: () => navigateToDesktopSection(
                  context,
                  (_) => const FriendsScreen(),
                ),
                onArtistsTap: () {},
                onProfileTap: () => navigateToDesktopSection(
                  context,
                  (_) => const ProfileScreen(),
                ),
                onSearchTap: () => openGlobalSearch(context),
              )
            else if (!hideTopBar)
              Container(
                height: 55,
                color: _tealAccent,
                padding: const EdgeInsets.symmetric(horizontal: 20),
                alignment: Alignment.centerLeft,
                child: Text(
                  'Artistas y canciones',
                  style: MeloopTypography.lexend(
                    color: Colors.white,
                    fontSize: 18,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            Expanded(
              child: isDesktop ? _buildDesktopLayout() : _buildMobileLayout(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDesktopLayout() {
    final gamification = GamificationDemoController.instance;

    return LayoutBuilder(
      builder: (context, constraints) {
        final double horizontalPadding = (constraints.maxWidth * 0.028).clamp(
          22.0,
          52.0,
        );
        final double contentWidth = math.min(
          constraints.maxWidth - horizontalPadding * 2,
          1640.0,
        );
        final double columnGap = (contentWidth * 0.024).clamp(22.0, 38.0);
        final double rightPanelWidth = (contentWidth * 0.27).clamp(
          290.0,
          395.0,
        );

        return Center(
          child: SizedBox(
            width: contentWidth,
            child: Padding(
              padding: const EdgeInsets.only(top: 32, bottom: 22),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Hoja izquierda principal: Álbumes en grilla infinita (5 por fila)
                  Expanded(
                    child: TapeCornerOverlay(
                      leftWidth: 162,
                      rightWidth: 140,
                      child: StackedPaperCard(
                        textureAsset: 'assets/textures/paper.png',
                        padding: const EdgeInsets.fromLTRB(32, 28, 32, 18),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            _buildDesktopLeftHeader(),
                            const SizedBox(height: 16),
                            _buildSearchBar(isMobile: false),
                            const SizedBox(height: 18),
                            Expanded(child: _buildAlbumGrid(isMobile: false)),
                          ],
                        ),
                      ),
                    ),
                  ),
                  SizedBox(width: columnGap),
                  // Hoja derecha con pin: Artistas que escuchan tus amigos
                  SizedBox(
                    width: rightPanelWidth,
                    child: AnimatedBuilder(
                      animation: gamification,
                      builder: (context, _) => Stack(
                        alignment: Alignment.topCenter,
                        clipBehavior: Clip.none,
                        children: [
                          Positioned.fill(
                            child: StackedPaperCard(
                              textureAsset: 'assets/textures/paper.png',
                              tintColor: MeloopThemeController
                                  .instance
                                  .colorTheme
                                  .softCardTint,
                              padding: const EdgeInsets.fromLTRB(
                                20,
                                28,
                                20,
                                18,
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.stretch,
                                children: [
                                  Text(
                                    'Lo que escuchan tus amigos',
                                    textAlign: TextAlign.center,
                                    style: MeloopTypography.lexend(
                                      fontSize: 17.5,
                                      fontWeight: FontWeight.w700,
                                      color: MeloopThemeController
                                          .instance
                                          .cardTitleColor,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    'Artistas en tendencia dentro de tu red',
                                    textAlign: TextAlign.center,
                                    style: MeloopTypography.lexend(
                                      fontSize: 11.5,
                                      color: MeloopThemeController
                                          .instance
                                          .cardMutedColor,
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                  _buildFriendFilterChips(),
                                  const SizedBox(height: 12),
                                  Divider(
                                    height: 1,
                                    thickness: 1,
                                    color: MeloopThemeController
                                        .instance
                                        .cardDividerColor,
                                  ),
                                  const SizedBox(height: 12),
                                  Expanded(
                                    child: _buildFriendArtistsList(
                                      isMobile: false,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                          Positioned(
                            top: -20,
                            child: CustomizablePin(
                              color: gamification.activePinColor,
                              size: 48,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildDesktopLeftHeader() {
    final titleBlock = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Transform.rotate(
          angle: -2.0 * math.pi / 180.0,
          child: Text(
            'Artistas y canciones',
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: MeloopTypography.momoDisplay(
              fontSize: 28,
              color: MeloopThemeController.instance.cardTitleColor,
            ),
          ),
        ),
        const SizedBox(height: 2),
        Text(
          'Explora por álbum, artista o canción',
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: MeloopTypography.lexend(
            fontSize: 12,
            color: MeloopThemeController.instance.cardMutedColor,
          ),
        ),
      ],
    );

    final pillTabs = MeloopPillTabs<_MusicFeedTab>(
      items: const [
        MeloopPillItem(value: _MusicFeedTab.recommended, label: 'Recomendados'),
        MeloopPillItem(value: _MusicFeedTab.trending, label: 'Tendencia'),
      ],
      selectedValue: _activeTab == _MusicFeedTab.friendsArtists
          ? _MusicFeedTab.recommended
          : _activeTab,
      onSelected: _selectTab,
      height: 44,
      padding: 4,
      fontSize: 13.5,
      horizontalItemPadding: 16,
      fixedButtonWidth: 148,
      pillColor: _pillTrackColor,
      equalWidths: true,
    );

    return LayoutBuilder(
      builder: (context, constraints) {
        if (constraints.maxWidth < 680) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              titleBlock,
              const SizedBox(height: 12),
              Center(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: pillTabs,
                ),
              ),
            ],
          );
        }

        return Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Expanded(child: titleBlock),
            const SizedBox(width: 16),
            pillTabs,
          ],
        );
      },
    );
  }

  Widget _buildMobileLayout() {
    final isAndroid = defaultTargetPlatform == TargetPlatform.android;

    return Padding(
      padding: const EdgeInsets.fromLTRB(10, 12, 10, 0),
      child: TapeCornerOverlay(
        leftWidth: 96,
        rightWidth: 84,
        child: StackedPaperCard(
          textureAsset: 'assets/textures/paper.png',
          padding: const EdgeInsets.fromLTRB(14, 16, 14, 0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Align(
                alignment: Alignment.centerLeft,
                child: Transform.rotate(
                  angle: -1.8 * math.pi / 180.0,
                  child: Text(
                    'Artistas y canciones',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: MeloopTypography.momoDisplay(
                      fontSize: 21,
                      color: MeloopThemeController.instance.cardTitleColor,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 12),
              LayoutBuilder(
                builder: (context, tabConstraints) {
                  final buttonWidth = ((tabConstraints.maxWidth - 8) / 3).clamp(
                    84.0,
                    140.0,
                  );
                  return Center(
                    child: MeloopPillTabs<_MusicFeedTab>(
                      items: const [
                        MeloopPillItem(
                          value: _MusicFeedTab.recommended,
                          label: 'Recomendados',
                        ),
                        MeloopPillItem(
                          value: _MusicFeedTab.trending,
                          label: 'Tendencia',
                        ),
                        MeloopPillItem(
                          value: _MusicFeedTab.friendsArtists,
                          label: 'De amigos',
                        ),
                      ],
                      selectedValue: _activeTab,
                      onSelected: _selectTab,
                      height: 38,
                      padding: 4,
                      fontSize: 11.5,
                      horizontalItemPadding: 6,
                      fixedButtonWidth: buttonWidth,
                      pillColor: _pillTrackColor,
                      equalWidths: true,
                    ),
                  );
                },
              ),
              const SizedBox(height: 12),
              if (_activeTab != _MusicFeedTab.friendsArtists) ...[
                _buildSearchBar(isMobile: true),
                const SizedBox(height: 12),
              ] else ...[
                _buildFriendFilterChips(),
                const SizedBox(height: 10),
              ],
              Expanded(
                child: _activeTab == _MusicFeedTab.friendsArtists
                    ? _buildFriendArtistsList(
                        isMobile: true,
                        bottomPadding: isAndroid ? 142 : 24,
                      )
                    : _buildAlbumGrid(
                        isMobile: true,
                        bottomPadding: isAndroid ? 142 : 24,
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSearchBar({required bool isMobile}) {
    final themeCtrl = MeloopThemeController.instance;
    return Container(
      height: isMobile ? 40 : 42,
      decoration: BoxDecoration(
        color: themeCtrl.inputFillColor,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: themeCtrl.inputBorderColor),
        boxShadow: themeCtrl.isDarkCardStyle
            ? null
            : [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.03),
                  blurRadius: 6,
                  offset: const Offset(0, 2),
                ),
              ],
      ),
      child: TextField(
        controller: _searchController,
        onChanged: (value) {
          setState(() {
            _query = value;
            _loadedCount = _pageSize;
          });
        },
        style: MeloopTypography.lexend(
          fontSize: isMobile ? 12.5 : 13.5,
          color: themeCtrl.cardTitleColor,
        ),
        decoration: InputDecoration(
          hintText: 'Buscar por álbum, artista o canción...',
          hintStyle: MeloopTypography.lexend(
            fontSize: isMobile ? 12.5 : 13.5,
            color: themeCtrl.cardSubtleMutedColor,
          ),
          prefixIcon: Icon(
            Icons.search_rounded,
            size: 20,
            color: themeCtrl.cardMutedColor,
          ),
          suffixIcon: _query.isNotEmpty
              ? IconButton(
                  tooltip: 'Limpiar búsqueda',
                  padding: EdgeInsets.zero,
                  icon: Icon(
                    Icons.close_rounded,
                    size: 18,
                    color: themeCtrl.cardMutedColor,
                  ),
                  onPressed: () {
                    _searchController.clear();
                    setState(() {
                      _query = '';
                      _loadedCount = _pageSize;
                    });
                  },
                )
              : null,
          border: InputBorder.none,
          contentPadding: EdgeInsets.symmetric(
            vertical: isMobile ? 10 : 11,
            horizontal: 8,
          ),
          isDense: true,
        ),
      ),
    );
  }

  Widget _buildAlbumGrid({required bool isMobile, double bottomPadding = 20}) {
    final albums = _visibleAlbums;
    if (albums.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.album_outlined,
                size: 48,
                color: MeloopThemeController.instance.cardSubtleMutedColor,
              ),
              const SizedBox(height: 12),
              Text(
                'No encontramos coincidencias',
                textAlign: TextAlign.center,
                style: MeloopTypography.lexend(
                  fontSize: 15.5,
                  fontWeight: FontWeight.w700,
                  color: MeloopThemeController.instance.cardTitleColor,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Prueba buscando otro nombre de álbum, artista o canción.',
                textAlign: TextAlign.center,
                style: MeloopTypography.lexend(
                  fontSize: 12.5,
                  color: MeloopThemeController.instance.cardMutedColor,
                ),
              ),
            ],
          ),
        ),
      );
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        final int columns;
        if (isMobile) {
          columns = constraints.maxWidth >= 560 ? 3 : 2;
        } else if (constraints.maxWidth >= 780) {
          // 5 álbumes por sección/fila en formato widescreen
          columns = 5;
        } else if (constraints.maxWidth >= 580) {
          columns = 4;
        } else {
          columns = 3;
        }

        final horizontalSpacing = isMobile ? 14.0 : 14.0;
        final verticalSpacing = isMobile ? 16.0 : 22.0;

        return CustomScrollView(
          controller: _gridScrollController,
          slivers: [
            SliverPadding(
              padding: EdgeInsets.fromLTRB(
                isMobile ? 4 : 8,
                6,
                isMobile ? 4 : 12,
                bottomPadding,
              ),
              sliver: SliverGrid(
                gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: columns,
                  crossAxisSpacing: horizontalSpacing,
                  mainAxisSpacing: verticalSpacing,
                  childAspectRatio: isMobile ? 0.72 : 0.92,
                ),
                delegate: SliverChildBuilderDelegate((context, index) {
                  final album = albums[index];
                  return _AlbumVinylCard(
                    key: ValueKey<String>('${album.id}-$index'),
                    album: album,
                    matchedSong: _matchingSongFor(album),
                    enableVinylHover: !isMobile,
                    isLiked: _likedAlbumIds.contains(album.id),
                    isFavorite: _favoriteAlbumIds.contains(album.id),
                    onTap: () => _openAlbumPreview(album),
                  );
                }, childCount: albums.length),
              ),
            ),
            if (_isLoadingMore)
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.only(bottom: 24),
                  child: Center(
                    child: SizedBox(
                      width: 22,
                      height: 22,
                      child: CircularProgressIndicator(
                        strokeWidth: 2.4,
                        color: _tealAccent,
                      ),
                    ),
                  ),
                ),
              ),
          ],
        );
      },
    );
  }

  Widget _buildFriendFilterChips() {
    return SizedBox(
      height: 32,
      child: ListView(
        scrollDirection: Axis.horizontal,
        children: [
          _friendChip(
            label: 'Todos',
            selected: _selectedFriendFilter == null,
            onTap: () {
              setState(() => _selectedFriendFilter = null);
            },
          ),
          for (final friend in _friendNames)
            _friendChip(
              label: friend,
              selected: _selectedFriendFilter == friend,
              onTap: () {
                setState(() {
                  _selectedFriendFilter = _selectedFriendFilter == friend
                      ? null
                      : friend;
                });
              },
            ),
        ],
      ),
    );
  }

  Widget _friendChip({
    required String label,
    required bool selected,
    required VoidCallback onTap,
  }) {
    final themeCtrl = MeloopThemeController.instance;
    return Padding(
      padding: const EdgeInsets.only(right: 6),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 6),
          decoration: BoxDecoration(
            color: selected ? _tealAccent : themeCtrl.inputFillColor,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: selected ? _tealAccent : themeCtrl.inputBorderColor,
            ),
          ),
          child: Text(
            label,
            style: MeloopTypography.lexend(
              fontSize: 11.5,
              fontWeight: FontWeight.w600,
              color: selected ? Colors.white : themeCtrl.cardBodyColor,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFriendArtistsList({
    required bool isMobile,
    double bottomPadding = 12,
  }) {
    final themeCtrl = MeloopThemeController.instance;
    final isDark = themeCtrl.isDarkCardStyle;
    final artists = _visibleFriendArtists;
    if (artists.isEmpty) {
      return Center(
        child: Text(
          'Sin actividad reciente de este amigo.',
          textAlign: TextAlign.center,
          style: MeloopTypography.lexend(
            fontSize: 12.5,
            color: themeCtrl.cardMutedColor,
          ),
        ),
      );
    }

    return ListView.builder(
      padding: EdgeInsets.only(bottom: bottomPadding),
      itemCount: artists.length,
      itemBuilder: (context, index) {
        final item = artists[index];
        final listenersText = item.listeners.length == 1
            ? '${item.listeners.first} escucha esto'
            : '${item.listeners.join(' y ')} escuchan esto';

        return Container(
          margin: const EdgeInsets.only(bottom: 10),
          decoration: BoxDecoration(
            color: themeCtrl.innerTileColor,
            border: Border.all(color: themeCtrl.innerTileBorderColor),
            boxShadow: isDark
                ? null
                : [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.05),
                      blurRadius: 6,
                      offset: const Offset(0, 2),
                    ),
                  ],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: () => _filterByArtistFromSidebar(item.artist),
              child: Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 11,
                  vertical: 10,
                ),
                child: Row(
                  children: [
                    Container(
                      width: 46,
                      height: 46,
                      padding: const EdgeInsets.all(3),
                      decoration: BoxDecoration(
                        color: isDark
                            ? Colors.white.withValues(alpha: 0.14)
                            : Colors.white,
                        boxShadow: isDark
                            ? null
                            : [
                                BoxShadow(
                                  color: Colors.black.withValues(alpha: 0.13),
                                  blurRadius: 4,
                                  offset: const Offset(0, 1),
                                ),
                              ],
                      ),
                      child: MeloopImagePlaceholder(
                        iconSize: 20,
                        backgroundColor: themeCtrl.placeholderBgColor,
                        iconColor: themeCtrl.placeholderIconColor,
                      ),
                    ),
                    const SizedBox(width: 11),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Row(
                            children: [
                              Flexible(
                                child: Text(
                                  item.artist,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: MeloopTypography.lexend(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w700,
                                    color: themeCtrl.cardTitleColor,
                                  ),
                                ),
                              ),
                              if (item.isLiveNow) ...[
                                const SizedBox(width: 6),
                                Icon(
                                  Icons.graphic_eq_rounded,
                                  size: 14,
                                  color: _tealAccent,
                                ),
                              ],
                            ],
                          ),
                          const SizedBox(height: 2),
                          Text(
                            '♪ ${item.topTrack} · ${item.genres}',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: 11,
                              color: themeCtrl.cardMutedColor,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Row(
                            children: [
                              Container(
                                width: 7,
                                height: 7,
                                decoration: BoxDecoration(
                                  color: item.isLiveNow
                                      ? const Color(0xFF16A34A)
                                      : const Color(0xFF9BB2B0),
                                  shape: BoxShape.circle,
                                ),
                              ),
                              const SizedBox(width: 5),
                              Expanded(
                                child: Text(
                                  listenersText,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: MeloopTypography.lexend(
                                    fontSize: 10.5,
                                    fontWeight: FontWeight.w500,
                                    color: themeCtrl.accentLinkColor,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 4),
                    Icon(
                      Icons.chevron_right_rounded,
                      size: 18,
                      color: themeCtrl.cardSubtleMutedColor,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

/// Tarjeta de álbum con efecto de vinilo deslizándose hacia la derecha al
/// pasar el cursor por encima (simulando revisar discos de vinilo).
/// El ancho de la portada (74% de la celda) reserva un espacio proporcional
/// garantizado para que el vinilo salga completo sin quedar detrás del álbum
/// vecino en ninguna resolución.
class _AlbumVinylCard extends StatefulWidget {
  const _AlbumVinylCard({
    super.key,
    required this.album,
    required this.enableVinylHover,
    required this.onTap,
    this.isLiked = false,
    this.isFavorite = false,
    this.matchedSong,
  });

  final _AlbumEntry album;
  final String? matchedSong;
  final bool enableVinylHover;
  final bool isLiked;
  final bool isFavorite;
  final VoidCallback onTap;

  @override
  State<_AlbumVinylCard> createState() => _AlbumVinylCardState();
}

class _AlbumVinylCardState extends State<_AlbumVinylCard> {
  bool _hovered = false;

  @override
  Widget build(BuildContext context) {
    final showVinyl = widget.enableVinylHover && _hovered;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: widget.enableVinylHover
          ? (_) => setState(() => _hovered = true)
          : null,
      onExit: widget.enableVinylHover
          ? (_) => setState(() => _hovered = false)
          : null,
      child: GestureDetector(
        onTap: widget.onTap,
        child: LayoutBuilder(
          builder: (context, cellConstraints) {
            final cellWidth = cellConstraints.maxWidth;
            final coverSize = widget.enableVinylHover
                ? cellWidth * 0.74
                : cellWidth;
            final vinylDiameter = coverSize * 0.90;
            // (0.95 + 0.39) * 0.74 = 0.9916 * cellWidth -> 100% dentro de su celda
            final peekOffset = coverSize * 0.39;

            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                SizedBox(
                  width: cellWidth,
                  height: coverSize,
                  child: Stack(
                    clipBehavior: Clip.none,
                    alignment: Alignment.centerLeft,
                    children: [
                      // Disco de vinilo que emerge completo hacia la derecha al hacer hover
                      Positioned(
                        left: (coverSize - vinylDiameter) / 2,
                        child: TweenAnimationBuilder<double>(
                          key: ValueKey<String>('vinyl-${widget.album.id}'),
                          tween: Tween<double>(end: showVinyl ? 1.0 : 0.0),
                          duration: const Duration(milliseconds: 240),
                          curve: Curves.easeOutCubic,
                          builder: (context, progress, child) {
                            if (progress <= 0.01) {
                              return const SizedBox.shrink();
                            }
                            return Transform.translate(
                              offset: Offset(peekOffset * progress, 0),
                              child: Transform.rotate(
                                angle: 0.45 * progress,
                                child: Opacity(
                                  opacity: progress.clamp(0.0, 1.0),
                                  child: child,
                                ),
                              ),
                            );
                          },
                          child: _VinylDiscGraphic(
                            diameter: vinylDiameter,
                            labelColor: widget.album.vinylLabelColor,
                          ),
                        ),
                      ),

                      // Portada cuadrada sin borde blanco (placeholder con símbolo de imagen)
                      AnimatedContainer(
                        duration: const Duration(milliseconds: 220),
                        curve: Curves.easeOutCubic,
                        transform: Matrix4.translationValues(
                          showVinyl ? -3.0 : 0.0,
                          showVinyl ? -2.0 : 0.0,
                          0.0,
                        ),
                        decoration: BoxDecoration(
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withValues(
                                alpha: showVinyl ? 0.24 : 0.12,
                              ),
                              blurRadius: showVinyl ? 14 : 7,
                              offset: Offset(0, showVinyl ? 6 : 3),
                            ),
                          ],
                        ),
                        child: Stack(
                          children: [
                            _AlbumSleevePlaceholder(size: coverSize),
                            if (widget.isFavorite || widget.isLiked)
                              Positioned(
                                top: 6,
                                left: 6,
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    if (widget.isFavorite)
                                      _SleeveBadge(
                                        icon: Icons.star_rounded,
                                        color: const Color(0xFFF5B52E),
                                      ),
                                    if (widget.isFavorite && widget.isLiked)
                                      const SizedBox(width: 4),
                                    if (widget.isLiked)
                                      _SleeveBadge(
                                        icon: Icons.favorite_rounded,
                                        color: const Color(0xFFFF6262),
                                      ),
                                  ],
                                ),
                              ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                SizedBox(height: widget.matchedSong != null ? 5 : 8),
                SizedBox(
                  width: coverSize + (widget.enableVinylHover ? 16 : 0),
                  child: Text(
                    widget.album.title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: MeloopTypography.lexend(
                      fontSize: 13,
                      fontWeight: FontWeight.w700,
                      color: MeloopThemeController.instance.cardTitleColor,
                      height: 1.15,
                    ),
                  ),
                ),
                const SizedBox(height: 1),
                SizedBox(
                  width: coverSize + (widget.enableVinylHover ? 16 : 0),
                  child: Text(
                    widget.album.artist,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: MeloopTypography.lexend(
                      fontSize: 11.5,
                      fontWeight: FontWeight.w500,
                      color: MeloopThemeController.instance.cardMutedColor,
                      height: 1.15,
                    ),
                  ),
                ),
                if (widget.matchedSong != null) ...[
                  const SizedBox(height: 1),
                  Flexible(
                    child: SizedBox(
                      width: coverSize + (widget.enableVinylHover ? 16 : 0),
                      child: Text(
                        '♪ ${widget.matchedSong}',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: 10.5,
                          fontWeight: FontWeight.w600,
                          color: MeloopThemeController.instance.accentLinkColor,
                          height: 1.15,
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            );
          },
        ),
      ),
    );
  }
}

class _SleeveBadge extends StatelessWidget {
  const _SleeveBadge({required this.icon, required this.color});

  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 22,
      height: 22,
      decoration: BoxDecoration(
        color: Colors.black.withValues(alpha: 0.56),
        shape: BoxShape.circle,
      ),
      child: Icon(icon, size: 13, color: color),
    );
  }
}

class _AlbumSleevePlaceholder extends StatelessWidget {
  const _AlbumSleevePlaceholder({required this.size});

  final double size;

  @override
  Widget build(BuildContext context) {
    final themeCtrl = MeloopThemeController.instance;
    return MeloopImagePlaceholder(
      width: size,
      height: size,
      iconSize: (size * 0.28).clamp(22.0, 46.0),
      backgroundColor: themeCtrl.placeholderBgColor,
      iconColor: themeCtrl.placeholderIconColor,
    );
  }
}

/// Ventana ampliada de detalle de álbum estilo Spotify sobre textura de papel
/// apilado (`assets/textures/paper.png`) con cintas en las esquinas, lista de
/// canciones interactiva y botones de Me gusta y Favorito.
class _AlbumDetailPaperDialog extends StatefulWidget {
  const _AlbumDetailPaperDialog({
    required this.album,
    required this.initiallyLiked,
    required this.initiallyFavorite,
    required this.onToggleLike,
    required this.onToggleFavorite,
    required this.onClose,
  });

  final _AlbumEntry album;
  final bool initiallyLiked;
  final bool initiallyFavorite;
  final bool Function() onToggleLike;
  final bool Function() onToggleFavorite;
  final VoidCallback onClose;

  @override
  State<_AlbumDetailPaperDialog> createState() =>
      _AlbumDetailPaperDialogState();
}

class _AlbumDetailPaperDialogState extends State<_AlbumDetailPaperDialog> {
  Color get _tealAccent => MeloopThemeController.instance.primaryColor;

  late bool _isLiked = widget.initiallyLiked;
  late bool _isFavorite = widget.initiallyFavorite;
  int? _playingTrackIndex;

  static const List<String> _sampleDurations = [
    '3:57',
    '3:24',
    '4:12',
    '3:46',
    '4:05',
  ];

  String _trackDuration(int index) {
    final seed = (widget.album.id.hashCode + index * 3).abs();
    return _sampleDurations[seed % _sampleDurations.length];
  }

  String _totalDurationLabel() {
    var totalSeconds = 0;
    for (var i = 0; i < widget.album.songs.length; i++) {
      final parts = _trackDuration(i).split(':');
      totalSeconds += int.parse(parts[0]) * 60 + int.parse(parts[1]);
    }
    final mins = totalSeconds ~/ 60;
    final secs = totalSeconds % 60;
    return '$mins min $secs s';
  }

  void _togglePlayAlbum() {
    setState(() {
      _playingTrackIndex = _playingTrackIndex == null ? 0 : null;
    });
  }

  @override
  Widget build(BuildContext context) {
    final screenSize = MediaQuery.sizeOf(context);
    final isCompact = screenSize.width < 680;
    final dialogWidth = math.min(screenSize.width - 32, 780.0);
    final dialogMaxHeight = math.min(screenSize.height - 48, 660.0);
    final coverSize = isCompact ? 104.0 : 168.0;
    final vinylDiameter = coverSize * 0.90;

    return Material(
      color: Colors.transparent,
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: dialogWidth,
          maxHeight: dialogMaxHeight,
        ),
        child: StackedPaperCard(
          textureAsset: 'assets/textures/photopaper.png',
          padding: EdgeInsets.fromLTRB(
            isCompact ? 18 : 34,
            isCompact ? 18 : 28,
            isCompact ? 18 : 34,
            isCompact ? 16 : 24,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Cabecera estilo Spotify sobre papel
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  SizedBox(
                    width: coverSize + vinylDiameter * 0.42,
                    height: coverSize,
                    child: Stack(
                      clipBehavior: Clip.none,
                      alignment: Alignment.centerLeft,
                      children: [
                        Positioned(
                          right: 0,
                          child: _VinylDiscGraphic(
                            diameter: vinylDiameter,
                            labelColor: widget.album.vinylLabelColor,
                          ),
                        ),
                        Positioned(
                          left: 0,
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withValues(alpha: 0.22),
                                  blurRadius: 16,
                                  offset: const Offset(0, 6),
                                ),
                              ],
                            ),
                            child: _AlbumSleevePlaceholder(size: coverSize),
                          ),
                        ),
                      ],
                    ),
                  ),
                  SizedBox(width: isCompact ? 14 : 24),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 9,
                                vertical: 3,
                              ),
                              decoration: BoxDecoration(
                                color: MeloopThemeController
                                    .instance
                                    .softBadgeBgColor,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                'ÁLBUM',
                                style: MeloopTypography.lexend(
                                  fontSize: 10.5,
                                  fontWeight: FontWeight.w700,
                                  color: MeloopThemeController
                                      .instance
                                      .softBadgeFgColor,
                                  letterSpacing: 1.1,
                                ),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Flexible(
                              child: Text(
                                widget.album.genre,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: MeloopTypography.lexend(
                                  fontSize: 11.5,
                                  color: MeloopThemeController
                                      .instance
                                      .cardMutedColor,
                                ),
                              ),
                            ),
                          ],
                        ),
                        SizedBox(height: isCompact ? 6 : 10),
                        Text(
                          widget.album.title,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: MeloopTypography.lexend(
                            fontSize: isCompact ? 20 : 28,
                            fontWeight: FontWeight.w700,
                            color:
                                MeloopThemeController.instance.cardTitleColor,
                            height: 1.15,
                          ),
                        ),
                        const SizedBox(height: 6),
                        Wrap(
                          spacing: 6,
                          runSpacing: 4,
                          crossAxisAlignment: WrapCrossAlignment.center,
                          children: [
                            Text(
                              widget.album.artist,
                              style: MeloopTypography.lexend(
                                fontSize: isCompact ? 13.5 : 15,
                                fontWeight: FontWeight.w700,
                                color: MeloopThemeController
                                    .instance
                                    .accentLinkColor,
                              ),
                            ),
                            Text(
                              '· ${widget.album.year} · ${widget.album.songs.length} canciones · ${_totalDurationLabel()}',
                              style: MeloopTypography.lexend(
                                fontSize: isCompact ? 12 : 13,
                                color: MeloopThemeController
                                    .instance
                                    .cardMutedColor,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    tooltip: 'Cerrar',
                    onPressed: widget.onClose,
                    icon: Icon(
                      Icons.close_rounded,
                      color: MeloopThemeController.instance.cardMutedColor,
                      size: 22,
                    ),
                  ),
                ],
              ),

              SizedBox(height: isCompact ? 16 : 22),

              // Barra de acciones: Play, Me gusta y Favorito
              Wrap(
                spacing: 10,
                runSpacing: 10,
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  Material(
                    color: _tealAccent,
                    shape: const CircleBorder(),
                    elevation: 3,
                    shadowColor: _tealAccent.withValues(alpha: 0.35),
                    child: InkWell(
                      customBorder: const CircleBorder(),
                      onTap: _togglePlayAlbum,
                      child: SizedBox(
                        width: 48,
                        height: 48,
                        child: Icon(
                          _playingTrackIndex != null
                              ? Icons.pause_rounded
                              : Icons.play_arrow_rounded,
                          color: Colors.white,
                          size: 28,
                        ),
                      ),
                    ),
                  ),
                  _AlbumActionPillButton(
                    icon: _isLiked
                        ? Icons.favorite_rounded
                        : Icons.favorite_border_rounded,
                    label: _isLiked ? 'Te gusta' : 'Me gusta',
                    active: _isLiked,
                    activeColor: const Color(0xFFFF5C5C),
                    onTap: () {
                      final liked = widget.onToggleLike();
                      setState(() => _isLiked = liked);
                    },
                  ),
                  _AlbumActionPillButton(
                    icon: _isFavorite
                        ? Icons.star_rounded
                        : Icons.star_border_rounded,
                    label: _isFavorite
                        ? 'En tus favoritos'
                        : 'Marcar como favorito',
                    active: _isFavorite,
                    activeColor: const Color(0xFFE69F17),
                    onTap: () {
                      final fav = widget.onToggleFavorite();
                      setState(() => _isFavorite = fav);
                    },
                  ),
                ],
              ),

              SizedBox(height: isCompact ? 14 : 20),

              // Encabezado de tabla de pistas estilo Spotify
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                child: Row(
                  children: [
                    SizedBox(
                      width: 28,
                      child: Text(
                        '#',
                        style: MeloopTypography.lexend(
                          fontSize: 11.5,
                          fontWeight: FontWeight.w600,
                          color: const Color(0xFF7B8B8B),
                        ),
                      ),
                    ),
                    Expanded(
                      child: Text(
                        'TÍTULO',
                        style: MeloopTypography.lexend(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          letterSpacing: 0.8,
                          color: const Color(0xFF7B8B8B),
                        ),
                      ),
                    ),
                    const Icon(
                      Icons.schedule_rounded,
                      size: 15,
                      color: Color(0xFF7B8B8B),
                    ),
                  ],
                ),
              ),
              Divider(
                height: 1,
                thickness: 1,
                color: Colors.black.withValues(alpha: 0.10),
              ),

              // Lista de canciones
              Flexible(
                child: ListView.separated(
                  shrinkWrap: true,
                  padding: const EdgeInsets.symmetric(vertical: 6),
                  itemCount: widget.album.songs.length,
                  separatorBuilder: (context, index) => Divider(
                    height: 1,
                    color: Colors.black.withValues(alpha: 0.05),
                  ),
                  itemBuilder: (context, index) {
                    final isPlaying = _playingTrackIndex == index;
                    return Material(
                      color: isPlaying
                          ? MeloopThemeController.instance.softBadgeBgColor
                          : Colors.transparent,
                      child: InkWell(
                        onTap: () {
                          setState(() {
                            _playingTrackIndex = isPlaying ? null : index;
                          });
                        },
                        child: Padding(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 11,
                          ),
                          child: Row(
                            children: [
                              SizedBox(
                                width: 28,
                                child: isPlaying
                                    ? Icon(
                                        Icons.graphic_eq_rounded,
                                        size: 17,
                                        color: _tealAccent,
                                      )
                                    : Text(
                                        '${index + 1}',
                                        style: MeloopTypography.lexend(
                                          fontSize: 13,
                                          color: MeloopThemeController
                                              .instance
                                              .cardMutedColor,
                                        ),
                                      ),
                              ),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      widget.album.songs[index],
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: MeloopTypography.lexend(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w600,
                                        color: isPlaying
                                            ? MeloopThemeController
                                                  .instance
                                                  .accentLinkColor
                                            : MeloopThemeController
                                                  .instance
                                                  .cardTitleColor,
                                      ),
                                    ),
                                    const SizedBox(height: 1),
                                    Text(
                                      widget.album.artist,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: MeloopTypography.lexend(
                                        fontSize: 11.5,
                                        color: MeloopThemeController
                                            .instance
                                            .cardMutedColor,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 12),
                              Text(
                                _trackDuration(index),
                                style: MeloopTypography.musicPlayer(
                                  fontSize: 13.5,
                                  color: MeloopThemeController
                                      .instance
                                      .cardMutedColor,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),

              const SizedBox(height: 10),
              Divider(height: 1, color: Colors.black.withValues(alpha: 0.08)),
              const SizedBox(height: 10),
              Row(
                children: [
                  Expanded(
                    child: Text(
                      'Lanzamiento: ${widget.album.year} · ${widget.album.genre}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: MeloopTypography.lexend(
                        fontSize: 11.5,
                        color: const Color(0xFF7D8C8C),
                      ),
                    ),
                  ),
                  if (!isCompact) ...[
                    const SizedBox(width: 12),
                    Flexible(
                      child: Text(
                        'Integración con Spotify próximamente',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        textAlign: TextAlign.end,
                        style: MeloopTypography.lexend(
                          fontSize: 11,
                          color: const Color(0xFF96A5A5),
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _AlbumActionPillButton extends StatelessWidget {
  const _AlbumActionPillButton({
    required this.icon,
    required this.label,
    required this.active,
    required this.activeColor,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final bool active;
  final Color activeColor;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final themeCtrl = MeloopThemeController.instance;
    return Material(
      color: active
          ? activeColor.withValues(alpha: 0.16)
          : themeCtrl.innerTileColor,
      borderRadius: BorderRadius.circular(24),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(24),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(24),
            border: Border.all(
              color: active
                  ? activeColor.withValues(alpha: 0.65)
                  : themeCtrl.inputBorderColor,
              width: 1.2,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                icon,
                size: 19,
                color: active ? activeColor : themeCtrl.cardMutedColor,
              ),
              const SizedBox(width: 7),
              Text(
                label,
                style: MeloopTypography.lexend(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: active ? activeColor : themeCtrl.cardTitleColor,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VinylDiscGraphic extends StatelessWidget {
  const _VinylDiscGraphic({required this.diameter, required this.labelColor});

  final double diameter;
  final Color labelColor;

  @override
  Widget build(BuildContext context) {
    return RepaintBoundary(
      child: CustomPaint(
        size: Size.square(diameter),
        painter: _VinylDiscPainter(labelColor: labelColor),
      ),
    );
  }
}

class _VinylDiscPainter extends CustomPainter {
  const _VinylDiscPainter({required this.labelColor});

  final Color labelColor;

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = size.width / 2;

    // Sombra suave del vinilo
    canvas.drawCircle(
      center.translate(-2, 3),
      radius,
      Paint()
        ..color = Colors.black.withValues(alpha: 0.28)
        ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 5),
    );

    // Base negra mate del disco
    canvas.drawCircle(center, radius, Paint()..color = const Color(0xFF181E20));

    // Reflejo diagonal sutil de luz sobre los surcos
    final sheenPaint = Paint()
      ..shader = SweepGradient(
        colors: [
          Colors.white.withValues(alpha: 0.0),
          Colors.white.withValues(alpha: 0.14),
          Colors.white.withValues(alpha: 0.0),
          Colors.white.withValues(alpha: 0.10),
          Colors.white.withValues(alpha: 0.0),
        ],
        stops: const [0.0, 0.18, 0.38, 0.68, 1.0],
      ).createShader(Rect.fromCircle(center: center, radius: radius));
    canvas.drawCircle(center, radius * 0.96, sheenPaint);

    // Surcos concéntricos
    final groovePaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 0.85
      ..color = Colors.white.withValues(alpha: 0.10);
    for (final factor in const [0.88, 0.78, 0.68, 0.58, 0.48]) {
      canvas.drawCircle(center, radius * factor, groovePaint);
    }

    // Etiqueta central de color
    canvas.drawCircle(center, radius * 0.33, Paint()..color = labelColor);
    canvas.drawCircle(
      center,
      radius * 0.30,
      Paint()
        ..style = PaintingStyle.stroke
        ..strokeWidth = 1.0
        ..color = Colors.white.withValues(alpha: 0.35),
    );

    // Orificio central del vinilo
    canvas.drawCircle(
      center,
      radius * 0.09,
      Paint()..color = const Color(0xFFF3F6F6),
    );
  }

  @override
  bool shouldRepaint(covariant _VinylDiscPainter oldDelegate) =>
      oldDelegate.labelColor != labelColor;
}

class _AlbumEntry {
  const _AlbumEntry({
    required this.id,
    required this.title,
    required this.artist,
    required this.year,
    required this.genre,
    required this.songs,
    required this.isRecommended,
    required this.isTrending,
    required this.vinylLabelColor,
  });

  final String id;
  final String title;
  final String artist;
  final String year;
  final String genre;
  final List<String> songs;
  final bool isRecommended;
  final bool isTrending;
  final Color vinylLabelColor;
}

class _FriendListeningArtist {
  const _FriendListeningArtist({
    required this.artist,
    required this.genres,
    required this.topTrack,
    required this.listeners,
    required this.isLiveNow,
  });

  final String artist;
  final String genres;
  final String topTrack;
  final List<String> listeners;
  final bool isLiveNow;
}
