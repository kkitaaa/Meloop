import 'package:flutter/material.dart';

import 'dart:async';

import 'friends_screen.dart';
import '../../features/comments/presentation/comments_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../widgets/post_card.dart';
import '../widgets/desktop_chat.dart';
import '../widgets/notifications_menu.dart';
import '../widgets/level_progress/level_progress.dart';
import '../features/gamification/presentation/gamification_demo_controller.dart';

class FeedScreen extends StatefulWidget {
  const FeedScreen({super.key});

  @override
  State<FeedScreen> createState() => _FeedScreenState();
}

class _LevelUpBanner extends StatefulWidget {
  const _LevelUpBanner({required this.reward, required this.onClose, required this.onEquip});

  final LevelUpReward reward;
  final VoidCallback onClose;
  final VoidCallback onEquip;

  @override
  State<_LevelUpBanner> createState() => _LevelUpBannerState();
}

class _LevelUpBannerState extends State<_LevelUpBanner> with SingleTickerProviderStateMixin {
  late final AnimationController _timer;

  @override
  void initState() {
    super.initState();
    _timer = AnimationController(vsync: this, duration: const Duration(seconds: 10))
      ..forward()
      ..addStatusListener((status) {
        if (status == AnimationStatus.completed) widget.onClose();
      });
  }

  @override
  void dispose() {
    _timer.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SlideTransition(
        position: Tween<Offset>(begin: const Offset(0, -1.2), end: Offset.zero).animate(
          CurvedAnimation(parent: _timer, curve: const Interval(0, .08, curve: Curves.easeOutBack)),
        ),
        child: Container(
          width: 370,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: .22), blurRadius: 18)],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [Icon(Icons.circle, color: Colors.amber, size: 9), Icon(Icons.square, color: Color(0xFFFF6262), size: 10), Icon(Icons.circle, color: Color(0xFF00AF92), size: 8), Icon(Icons.star, color: Colors.purple, size: 13)],
              ),
              const SizedBox(height: 8),
              Text('¡Subiste a Nivel ${widget.reward.level}!', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF1E7E70))),
              const SizedBox(height: 6),
              const Text('Recompensa desbloqueada', style: TextStyle(color: Colors.black54, fontSize: 12)),
              const SizedBox(height: 10),
              Row(mainAxisAlignment: MainAxisAlignment.center, children: [const Icon(Icons.gradient, color: Colors.purple), const SizedBox(width: 7), Text(widget.reward.name, style: const TextStyle(fontWeight: FontWeight.bold))]),
              const SizedBox(height: 12),
              ElevatedButton(onPressed: widget.onEquip, style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF00AF92), foregroundColor: Colors.white), child: const Text('Equipar ahora')),
              TextButton(onPressed: widget.onClose, child: const Text('Cerrar')),
              AnimatedBuilder(
                animation: _timer,
                builder: (context, _) => LinearProgressIndicator(value: 1 - _timer.value, minHeight: 3, color: const Color(0xFF00AF92), backgroundColor: Colors.grey[200]),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _CompactLevelUpBanner extends StatefulWidget {
  const _CompactLevelUpBanner({required this.reward, required this.onClose, required this.onEquip});
  final LevelUpReward reward;
  final VoidCallback onClose;
  final VoidCallback onEquip;
  @override
  State<_CompactLevelUpBanner> createState() => _CompactLevelUpBannerState();
}

class _CompactLevelUpBannerState extends State<_CompactLevelUpBanner>
    with TickerProviderStateMixin {
  late final AnimationController _timer;
  late final AnimationController _confetti;
  @override
  void initState() {
    super.initState();
    _timer = AnimationController(vsync: this, duration: const Duration(seconds: 10))
      ..forward()
      ..addStatusListener((status) { if (status == AnimationStatus.completed) widget.onClose(); });
    _confetti = AnimationController(vsync: this, duration: const Duration(milliseconds: 1500))..repeat();
  }
  @override
  void dispose() { _timer.dispose(); _confetti.dispose(); super.dispose(); }
  @override
  Widget build(BuildContext context) => Center(
    child: SlideTransition(
      position: Tween<Offset>(begin: const Offset(0, -1), end: Offset.zero).animate(CurvedAnimation(parent: _timer, curve: const Interval(0, .08, curve: Curves.easeOutBack))),
      child: Container(
        width: 520, padding: const EdgeInsets.fromLTRB(18, 12, 18, 8),
        decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: .22), blurRadius: 18)]),
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          SizedBox(height: 57, child: Stack(children: [
            AnimatedBuilder(animation: _confetti, builder: (_, __) => Stack(children: [_confettiPiece(.08, Colors.amber, Icons.circle), _confettiPiece(.31, const Color(0xFFFF6262), Icons.square), _confettiPiece(.68, const Color(0xFF00AF92), Icons.circle), _confettiPiece(.9, Colors.purple, Icons.star)])),
            Center(child: Row(mainAxisAlignment: MainAxisAlignment.center, children: [const Icon(Icons.emoji_events, color: Colors.amber, size: 36), const SizedBox(width: 10), Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [Text('Nivel ${widget.reward.level}', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF1E7E70))), Text(widget.reward.name, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600))])]))
          ])),
          Row(mainAxisAlignment: MainAxisAlignment.center, children: [TextButton(onPressed: widget.onClose, child: const Text('Cerrar')), const SizedBox(width: 10), ElevatedButton(onPressed: widget.onEquip, style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF00AF92), foregroundColor: Colors.white), child: const Text('Equipar ahora'))]),
          AnimatedBuilder(animation: _timer, builder: (_, __) => LinearProgressIndicator(value: 1 - _timer.value, minHeight: 3, color: const Color(0xFF00AF92), backgroundColor: Colors.grey[200])),
        ]),
      ),
    ),
  );
  Widget _confettiPiece(double x, Color color, IconData icon) {
    final fall = (_confetti.value + x) % 1;
    return Positioned(left: x * 480, top: fall * 50, child: Transform.rotate(angle: fall * 6.28, child: Icon(icon, color: color, size: 10)));
  }
}

class _FullScreenConfetti extends StatefulWidget {
  const _FullScreenConfetti();
  @override
  State<_FullScreenConfetti> createState() => _FullScreenConfettiState();
}

class _FullScreenConfettiState extends State<_FullScreenConfetti> with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  @override
  void initState() { super.initState(); _controller = AnimationController(vsync: this, duration: const Duration(milliseconds: 1250))..repeat(); }
  @override
  void dispose() { _controller.dispose(); super.dispose(); }
  @override
  Widget build(BuildContext context) => IgnorePointer(child: LayoutBuilder(builder: (context, size) => AnimatedBuilder(
    animation: _controller,
    builder: (_, __) => Stack(children: List.generate(10, (i) {
      const colors = [Colors.amber, Color(0xFFFF6262), Color(0xFF00AF92), Colors.purple, Colors.orange];
      final x = ((i * .173) % 1);
      final fall = (_controller.value + i * .19) % 1;
      return Positioned(left: x * size.maxWidth, top: -18 + fall * (size.maxHeight + 36), child: Transform.rotate(angle: fall * 10, child: Icon(i.isEven ? Icons.square : Icons.circle, color: colors[i % colors.length], size: 10)));
    })),
  )));
}

class _RewardLevelUpBanner extends StatefulWidget {
  const _RewardLevelUpBanner({required this.reward, required this.onClose, required this.onEquip});
  final LevelUpReward reward;
  final VoidCallback onClose;
  final VoidCallback onEquip;
  @override
  State<_RewardLevelUpBanner> createState() => _RewardLevelUpBannerState();
}

class _RewardLevelUpBannerState extends State<_RewardLevelUpBanner> with SingleTickerProviderStateMixin {
  late final AnimationController _timer;
  @override
  void initState() { super.initState(); _timer = AnimationController(vsync: this, duration: const Duration(seconds: 10))..forward()..addStatusListener((s) { if (s == AnimationStatus.completed) widget.onClose(); }); }
  @override
  void dispose() { _timer.dispose(); super.dispose(); }
  @override
  Widget build(BuildContext context) => Center(child: SlideTransition(
    position: Tween<Offset>(begin: const Offset(0, -1), end: Offset.zero).animate(CurvedAnimation(parent: _timer, curve: const Interval(0, .08, curve: Curves.easeOutBack))),
    child: Container(width: 520, padding: const EdgeInsets.fromLTRB(18, 14, 18, 8), decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: .22), blurRadius: 18)]), child: Column(mainAxisSize: MainAxisSize.min, children: [
      SizedBox(height: 58, child: AnimatedBuilder(animation: _timer, builder: (_, __) {
        if (_timer.value < .3) {
          const phrase = '¡Subiste de nivel!';
          final progress = (_timer.value / .3).clamp(0, 1).toDouble();
          final fade = (1 - (progress - .8).clamp(0, .2) * 5).toDouble();
          return Center(child: Opacity(opacity: fade, child: Text(phrase.substring(0, (phrase.length * progress).floor()), style: const TextStyle(fontSize: 23, fontWeight: FontWeight.bold, color: Color(0xFF1E7E70)))));
        }
        return Row(children: [const Icon(Icons.gradient, color: Colors.purple, size: 40), const SizedBox(width: 12), Expanded(child: Column(mainAxisAlignment: MainAxisAlignment.center, crossAxisAlignment: CrossAxisAlignment.start, children: [Text(widget.reward.name, style: const TextStyle(fontSize: 17, fontWeight: FontWeight.bold)), const Text('Marco de perfil', style: TextStyle(color: Colors.black54, fontSize: 12))])), ElevatedButton(onPressed: widget.onEquip, style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF00AF92), foregroundColor: Colors.white), child: const Text('Equipar'))]);
      })),
      AnimatedBuilder(animation: _timer, builder: (_, __) => LinearProgressIndicator(value: 1 - _timer.value, minHeight: 3, color: const Color(0xFF00AF92), backgroundColor: Colors.grey[200])),
    ])),
  ));
}

class _FeedScreenState extends State<FeedScreen> {
  final Color _bgColor = const Color(0xFF0D5C5E);
  final Color _tealAccent = const Color(0xFF1ABC9C);

  final ScrollController _scrollController = ScrollController();
  final List<Map<String, dynamic>> _posts = [];
  bool _isInitialLoading = true;
  bool _isLoadingMore = false;
  int _currentPage = 1;
  LevelUpReward? _levelUpReward;

  @override
  void initState() {
    super.initState();
    GamificationDemoController.instance.addListener(_onGamificationChanged);
    _fetchInitialPosts();

    _scrollController.addListener(() {
      if (_scrollController.position.pixels >=
          _scrollController.position.maxScrollExtent - 200) {
        if (!_isLoadingMore) {
          _fetchMorePosts();
        }
      }
    });
  }

  @override
  void dispose() {
    GamificationDemoController.instance.removeListener(_onGamificationChanged);
    _scrollController.dispose();
    super.dispose();
  }

  void _onGamificationChanged() {
    final reward = GamificationDemoController.instance.takePendingLevelUp();
    if (reward != null && mounted) setState(() => _levelUpReward = reward);
  }

  Future<void> _fetchInitialPosts() async {
    setState(() => _isInitialLoading = true);
    await Future.delayed(const Duration(seconds: 2));

    if (!mounted) return;

    setState(() {
      _posts.addAll(_generateMockPosts(1, 5));
      _isInitialLoading = false;
    });
  }

  Future<void> _fetchMorePosts() async {
    setState(() => _isLoadingMore = true);
    await Future.delayed(const Duration(seconds: 2));

    if (!mounted) return;

    setState(() {
      _currentPage++;
      _posts.addAll(_generateMockPosts(_currentPage, 3));
      _isLoadingMore = false;
    });
  }

  List<Map<String, dynamic>> _generateMockPosts(int page, int count) {
    return List.generate(count, (index) {
      int id = (page - 1) * count + index + 1;
      return {
        "id": id,
        "user": "Usuario_0$id",
        "time": "Hace ${id * 5} minutos",
        "content":
            "Esta es la publicación dinámica número $id cargada desde el servidor simulado. Probando el scroll infinito y la paginación.",
        "song": "Canción $id",
        "artist": "Artista Generado",
        "likes": 10 * id,
        "comments": id,
      };
    });
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.of(context).size.width;
    final bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: _bgColor,
      body: Stack(
        children: [
          Column(
            children: [
              _buildTopBar(isDesktop),
              Expanded(
                child: isDesktop ? _buildDesktopLayout() : _buildMobileLayout(),
              ),
            ],
          ),
          if (isDesktop) const Positioned(right: 24, bottom: 0, child: DesktopChat()),
          if (_levelUpReward != null)
            const Positioned.fill(child: _FullScreenConfetti()),
          if (_levelUpReward != null)
            Positioned(
              top: 18,
              left: 0,
              right: 0,
              child: _RewardLevelUpBanner(
                reward: _levelUpReward!,
                onClose: () => setState(() => _levelUpReward = null),
                onEquip: () {
                  setState(() => _levelUpReward = null);
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => const ProfileScreen(openInventory: true),
                    ),
                  );
                },
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildTopBar(bool isDesktop) {
    return Container(
      height: 55,
      color: _tealAccent,
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Image.asset(
            'assets/imagenes/meloop.png',
            height: 24,
            color: Colors.white,
            fit: BoxFit.contain,
          ),
          if (isDesktop) ...[
            Row(
              children: [
                _navLink('Inicio', isActive: true),
                _navLink(
                  'Amigos',
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const FriendsScreen(),
                      ),
                    );
                  },
                ),
                _navLink(
                  'Perfil',
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const ProfileScreen(),
                      ),
                    );
                  },
                ),
              ],
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.2),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Row(
                children: const [
                  Icon(Icons.play_circle_fill, color: Colors.white, size: 18),
                  SizedBox(width: 8),
                  Text(
                    "Playing... DJ Gouz",
                    style: TextStyle(color: Colors.white, fontSize: 12),
                  ),
                  SizedBox(width: 8),
                  Icon(Icons.graphic_eq, color: Colors.white, size: 16),
                ],
              ),
            ),
          ],
          Row(
            children: [
              const NotificationsMenu(),
              const SizedBox(width: 16),
              InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ProfileScreen(),
                    ),
                  );
                },
                child: const CircleAvatar(
                  radius: 14,
                  backgroundColor: Colors.white,
                  child: Icon(Icons.person, color: Color(0xFF1ABC9C), size: 18),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _navLink(String text, {bool isActive = false, VoidCallback? onTap}) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 4.0),
        padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 6.0),
        decoration: BoxDecoration(
          color: isActive ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Text(
          text,
          style: TextStyle(
            color: isActive ? _tealAccent : Colors.white,
            fontWeight: isActive ? FontWeight.bold : FontWeight.normal,
            fontSize: 12,
          ),
        ),
      ),
    );
  }

  Widget _buildDesktopLayout() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          SizedBox(
            width: 220,
            child: SingleChildScrollView(
              child: Column(
                children: [
                  _buildPolaroidProfile(),
                  const SizedBox(height: 16),
                  _buildListeningTo(),
                  const SizedBox(height: 16),
                  _buildLevelCard(),
                  const SizedBox(height: 16),
                  _buildSuggestions(),
                ],
              ),
            ),
          ),
          const SizedBox(width: 24),
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 650),
            child: _buildFeedPaper(),
          ),
        ],
      ),
    );
  }

  Widget _buildMobileLayout() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        children: [
          _buildPolaroidProfile(),
          const SizedBox(height: 16),
          _buildListeningTo(),
          const SizedBox(height: 24),
          SizedBox(height: 700, child: _buildFeedPaper()),
        ],
      ),
    );
  }

  Widget _buildPolaroidProfile() {
    return Stack(
      alignment: Alignment.topCenter,
      clipBehavior: Clip.none,
      children: [
        Transform.rotate(
          angle: -0.02,
          child: Container(
            margin: const EdgeInsets.only(top: 10),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.1),
                  blurRadius: 8,
                ),
              ],
            ),
            child: Column(
              children: [
                Container(
                  height: 140,
                  width: double.infinity,
                  color: Colors.grey[200],
                  child: const Icon(Icons.pets, size: 50, color: Colors.grey),
                ),
                const SizedBox(height: 16),
                const Text(
                  "Nombre de usuario",
                  style: TextStyle(
                    fontFamily: 'Comic Sans MS',
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 4),
                const Text(
                  "Me siento Feliz",
                  style: TextStyle(fontSize: 11, color: Colors.black54),
                ),
              ],
            ),
          ),
        ),
        Positioned(
          top: 0,
          child: Transform.rotate(
            angle: -0.08,
            child: Container(
              width: 50,
              height: 18,
              color: Colors.white.withValues(alpha: 0.7),
            ),
          ),
        ),
        Positioned(
          top: 8,
          child: CircleAvatar(radius: 5, backgroundColor: Colors.red[700]),
        ),
      ],
    );
  }

  Widget _buildListeningTo() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: Colors.purple[800],
              borderRadius: BorderRadius.circular(4),
            ),
            child: const Icon(Icons.music_note, color: Colors.white),
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: const [
              Text(
                "Escuchando:",
                style: TextStyle(color: Colors.black54, fontSize: 10),
              ),
              Text(
                "Finding Urself",
                style: TextStyle(
                  color: Colors.black,
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
              Text(
                "DJ GOUZ",
                style: TextStyle(color: Colors.black54, fontSize: 10),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildLevelCard() {
    return AnimatedBuilder(
      animation: GamificationDemoController.instance,
      builder: (context, _) {
        final game = GamificationDemoController.instance;
        return LevelProgress(
          level: game.level,
          xp: game.xp,
          xpNeeded: GamificationDemoController.xpNeeded,
          compact: true,
          accentColor: _tealAccent,
        );
      },
    );
  }

  Widget _buildLegacyLevelCard() {
    return AnimatedBuilder(
      animation: GamificationDemoController.instance,
      builder: (context, _) {
        final game = GamificationDemoController.instance;
        final percentage = (game.xp / GamificationDemoController.xpNeeded * 100).round();
        return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          Row(
            children: [
              Icon(Icons.star, color: _tealAccent, size: 14),
              const SizedBox(width: 4),
              Text(
                "Nivel ${game.level} · (${game.xp} XP / 100 XP) $percentage%",
                style: TextStyle(
                  color: Colors.black87,
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: game.xp / GamificationDemoController.xpNeeded,
            backgroundColor: Colors.grey[200],
            valueColor: AlwaysStoppedAnimation<Color>(_tealAccent),
            minHeight: 6,
            borderRadius: BorderRadius.circular(3),
          ),
        ],
      ),
        );
      },
    );
  }

  Widget _buildSuggestions() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            "Artistas que te podrían gustar",
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.bold,
              color: Colors.black87,
            ),
          ),
          const Divider(),
          _artistListTile("Kidd Voodoo"),
          _artistListTile("Tiago PZK"),
          _artistListTile("Cris MJ"),
        ],
      ),
    );
  }

  Widget _artistListTile(String name) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10.0),
      child: Row(
        children: [
          const CircleAvatar(
            radius: 10,
            backgroundColor: Colors.black12,
            child: Icon(Icons.person, size: 12, color: Colors.black54),
          ),
          const SizedBox(width: 8),
          Text(
            name,
            style: const TextStyle(fontSize: 11, color: Colors.black87),
          ),
        ],
      ),
    );
  }

  Widget _buildFeedPaper() {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFDFDFD),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.15),
            blurRadius: 20,
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
        child: Column(
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    const Text(
                      "Ultimos Blogs:",
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(width: 16),
                    _filterChip("recientes", isActive: true),
                    _filterChip("tendencias", isActive: false),
                    _filterChip("amigos", isActive: false),
                  ],
                ),
                OutlinedButton.icon(
                  style: OutlinedButton.styleFrom(
                    side: BorderSide(color: _tealAccent),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                    ),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 12,
                    ),
                  ),
                  onPressed: () {},
                  icon: Icon(Icons.edit, color: _tealAccent, size: 14),
                  label: Text(
                    "escribir blog",
                    style: TextStyle(
                      color: _tealAccent,
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            const Divider(height: 32, thickness: 1, color: Colors.black12),
            Expanded(
              child: _isInitialLoading
                  ? Center(child: CircularProgressIndicator(color: _tealAccent))
                  : ListView.builder(
                      controller: _scrollController,
                      itemCount: _posts.length + (_isLoadingMore ? 1 : 0),
                      itemBuilder: (context, index) {
                        if (index == _posts.length) {
                          return Padding(
                            padding: const EdgeInsets.symmetric(vertical: 24.0),
                            child: Center(
                              child: CircularProgressIndicator(
                                color: _tealAccent,
                              ),
                            ),
                          );
                        }
                        return PostCard(
                          postData: _posts[index],
                          tealAccent: _tealAccent,
                        );
                      },
                    ),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: _tealAccent,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(20),
                ),
                elevation: 0,
              ),
              onPressed: () {},
              child: const Text(
                "Cargar más blogs",
                style: TextStyle(color: Colors.white, fontSize: 12),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _filterChip(String label, {required bool isActive}) {
    return Container(
      margin: const EdgeInsets.only(right: 8),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: isActive ? Colors.white : Colors.grey[200],
        border: isActive
            ? Border.all(color: _tealAccent, width: 1)
            : Border.all(color: Colors.transparent),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: isActive ? _tealAccent : Colors.grey[600],
          fontSize: 10,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }

  Widget _buildPostCard(Map<String, dynamic> postData) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 24.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  InkWell(
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) =>
                              ProfileScreen(username: postData["user"]),
                        ),
                      );
                    },
                    child: const CircleAvatar(
                      radius: 16,
                      backgroundColor: Colors.black87,
                      child: Icon(Icons.person, color: Colors.white, size: 18),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      InkWell(
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  ProfileScreen(username: postData["user"]),
                            ),
                          );
                        },
                        child: Text(
                          postData["user"],
                          style: const TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 13,
                          ),
                        ),
                      ),
                      Text(
                        postData["time"],
                        style: const TextStyle(
                          fontSize: 10,
                          color: Colors.grey,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              const Icon(Icons.more_horiz, color: Colors.grey, size: 20),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            postData["content"],
            style: const TextStyle(
              fontSize: 13,
              color: Colors.black87,
              height: 1.4,
            ),
          ),
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: _tealAccent,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Row(
              children: [
                Container(
                  width: 36,
                  height: 36,
                  color: Colors.black87,
                  child: const Icon(
                    Icons.play_arrow,
                    color: Colors.white,
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        postData["song"],
                        style: const TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                      ),
                      Text(
                        postData["artist"],
                        style: const TextStyle(
                          color: Colors.white70,
                          fontSize: 10,
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.graphic_eq, color: _tealAccent, size: 12),
                      const SizedBox(width: 4),
                      Text(
                        "Escuchar",
                        style: TextStyle(
                          color: _tealAccent,
                          fontSize: 10,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Row(
                children: [
                  const Icon(Icons.favorite, color: Colors.red, size: 16),
                  const SizedBox(width: 4),
                  Text(
                    "${postData["likes"]} Likes",
                    style: const TextStyle(color: Colors.grey, fontSize: 11),
                  ),
                ],
              ),
              const SizedBox(width: 16),
              InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => CommentsScreen(postData: postData),
                    ),
                  );
                },
                child: Row(
                  children: [
                    const Icon(
                      Icons.mode_comment_outlined,
                      color: Colors.grey,
                      size: 16,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      "${postData["comments"]} Comentarios",
                      style: const TextStyle(color: Colors.grey, fontSize: 11),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          const Divider(color: Colors.black12),
        ],
      ),
    );
  }
}
