import 'package:flutter/material.dart';

import 'dart:async';

import '../../../screens/friends_screen.dart';
import '../../../widgets/post_card.dart';

class ProfileScreen extends StatefulWidget {
  final String? username;

  const ProfileScreen({super.key, this.username});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final Color _bgColor = const Color(0xFF0D5C5E);
  final Color _tealAccent = const Color(0xFF1ABC9C);

  final ScrollController _scrollController = ScrollController();
  final List<Map<String, dynamic>> _posts = [];

  bool _isInitialLoading = true;
  bool _isLoadingMore = false;
  int _currentPage = 1;

  // Estados del perfil: own, public, private, blocked
  late String _profileStatus;

  late String _displayUsername;
  String _biography =
      "Hola, gente soy Vicente. Me gusta mucho la música ambiental y la música rapidísima con mucho bajo, también el jazz fusion y otras cosas más.\nAl igual que a todos, me gustan los gatos (tengo 2).";

  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _bioController = TextEditingController();

  // --- VARIABLES DE PERSONALIZACIÓN (GAMIFICACIÓN) ---
  Color? _activeFrameColor;
  Color _activePinColor = Colors.red[700]!;
  Color _paperColor = const Color(0xFFFDFDFD);
  Color _paperTextColor = Colors.black87;

  final List<Map<String, dynamic>> _inventory = [
    {
      "id": "f1",
      "type": "frame",
      "name": "Marco Neón",
      "icon": Icons.crop_square,
      "color": Colors.cyanAccent,
      "isEquipped": false,
    },
    {
      "id": "f2",
      "type": "frame",
      "name": "Marco Dorado",
      "icon": Icons.crop_portrait,
      "color": Colors.amber,
      "isEquipped": false,
    },
    {
      "id": "p1",
      "type": "pin",
      "name": "Pin Zafiro",
      "icon": Icons.push_pin,
      "color": Colors.blueAccent,
      "isEquipped": false,
    },
    {
      "id": "bg1",
      "type": "bg",
      "name": "Modo Nocturno",
      "icon": Icons.format_paint,
      "color": const Color(0xFF2C3E50),
      "isEquipped": false,
    },
  ];

  @override
  void initState() {
    super.initState();
    _displayUsername = widget.username ?? "Nombre de usuario";
    _usernameController.text = _displayUsername;
    _bioController.text = _biography;

    _determineProfileStatus();

    if (_profileStatus == 'own' || _profileStatus == 'public') {
      _fetchInitialPosts();
    } else {
      _isInitialLoading = false;
    }

    _scrollController.addListener(() {
      if (_scrollController.position.pixels >=
          _scrollController.position.maxScrollExtent - 200) {
        if (!_isLoadingMore && _posts.isNotEmpty) {
          _fetchMorePosts();
        }
      }
    });
  }

  void _determineProfileStatus() {
    if (widget.username == "UsuarioPrivado") {
      _profileStatus = 'private';
    } else if (widget.username == "UsuarioBloqueado") {
      _profileStatus = 'blocked';
    } else if (widget.username == null || widget.username == "MiUsuario") {
      _profileStatus = 'own';
    } else {
      _profileStatus = 'public';
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    _usernameController.dispose();
    _bioController.dispose();
    super.dispose();
  }

  Future<void> _fetchInitialPosts() async {
    setState(() => _isInitialLoading = true);
    await Future.delayed(const Duration(seconds: 1));

    if (!mounted) return;

    setState(() {
      _posts.addAll(_generateMockPosts(1, 3));
      _isInitialLoading = false;
    });
  }

  Future<void> _fetchMorePosts() async {
    setState(() => _isLoadingMore = true);
    await Future.delayed(const Duration(seconds: 1));

    if (!mounted) return;

    setState(() {
      _currentPage++;
      _posts.addAll(_generateMockPosts(_currentPage, 2));
      _isLoadingMore = false;
    });
  }

  List<Map<String, dynamic>> _generateMockPosts(int page, int count) {
    return List.generate(count, (index) {
      final id = (page - 1) * count + index + 1;
      return {
        "id": id,
        "user": _displayUsername,
        "time": "Hace ${id * 2} horas",
        "content":
            "Publicación número $id del perfil de $_displayUsername. Consumiendo ruta paginada del post-service.",
        "song": "Track del Perfil $id",
        "artist": "Artista Afín",
        "likes": 15 * id,
        "comments": id * 2,
        "isAuthor": _profileStatus == 'own', // Si es tu perfil, eres autor
        "isEdited": false,
        "within24Hours": true,
      };
    });
  }

  void _toggleEquipItem(int index) {
    final item = _inventory[index];
    final bool isEquipping = !(item['isEquipped'] as bool);

    setState(() {
      if (isEquipping) {
        for (var i in _inventory) {
          if (i['type'] == item['type']) i['isEquipped'] = false;
        }
        item['isEquipped'] = true;

        if (item['type'] == 'frame') _activeFrameColor = item['color'] as Color;
        if (item['type'] == 'pin') _activePinColor = item['color'] as Color;
        if (item['type'] == 'bg') {
          _paperColor = item['color'] as Color;
          _paperTextColor = Colors.white;
        }
      } else {
        item['isEquipped'] = false;

        if (item['type'] == 'frame') _activeFrameColor = null;
        if (item['type'] == 'pin') _activePinColor = Colors.red[700]!;
        if (item['type'] == 'bg') {
          _paperColor = const Color(0xFFFDFDFD);
          _paperTextColor = Colors.black87;
        }
      }
    });

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          isEquipping
              ? "Equipado: ${item['name']}"
              : "Desequipado: ${item['name']}",
        ),
        backgroundColor: _tealAccent,
        duration: const Duration(milliseconds: 1500),
      ),
    );
  }

  void _showEditProfileDialog() {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          backgroundColor: Colors.white,
          title: const Text(
            "Editar Perfil",
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
          content: SizedBox(
            width: 400,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: _usernameController,
                  decoration: const InputDecoration(
                    labelText: "Nombre de usuario",
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _bioController,
                  maxLines: 4,
                  decoration: const InputDecoration(
                    labelText: "Sobre mí (Biografía)",
                    border: OutlineInputBorder(),
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text(
                "Cancelar",
                style: TextStyle(color: Colors.grey),
              ),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: _tealAccent),
              onPressed: () {
                setState(() {
                  _displayUsername = _usernameController.text;
                  _biography = _bioController.text;
                });
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: const Text("Perfil actualizado correctamente."),
                    backgroundColor: _tealAccent,
                  ),
                );
              },
              child: const Text(
                "Guardar",
                style: TextStyle(color: Colors.white),
              ),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.of(context).size.width;
    final bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: _bgColor,
      body: Column(
        children: [
          _buildTopBar(isDesktop, context),
          Expanded(
            child: isDesktop ? _buildDesktopLayout() : _buildMobileLayout(),
          ),
        ],
      ),
    );
  }

  Widget _buildTopBar(bool isDesktop, BuildContext context) {
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
            fit: BoxFit.contain,
          ),
          if (isDesktop) ...[
            Row(
              children: [
                _navLink(
                  'Inicio',
                  isActive: false,
                  onTap: () => Navigator.pop(context),
                ),
                _navLink(
                  'Amigos',
                  isActive: false,
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const FriendsScreen(),
                      ),
                    );
                  },
                ),
                _navLink('Artistas y canciones', isActive: false, onTap: () {}),
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
            children: const [
              Icon(Icons.notifications, color: Colors.white, size: 20),
              SizedBox(width: 16),
              CircleAvatar(
                radius: 14,
                backgroundColor: Colors.white,
                child: Icon(Icons.person, color: Color(0xFF1ABC9C), size: 18),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _navLink(
    String text, {
    required bool isActive,
    required VoidCallback onTap,
  }) {
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
                  if (_profileStatus != 'blocked') _buildLevelCard(),
                ],
              ),
            ),
          ),
          const SizedBox(width: 24),
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 800),
            child: _profileStatus == 'blocked'
                ? _buildBlockedState()
                : SingleChildScrollView(
                    controller: _scrollController,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildStatsAndActionsRow(),
                        const SizedBox(height: 16),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Expanded(flex: 3, child: _buildAboutMe()),
                            const SizedBox(width: 16),
                            if (_profileStatus != 'private')
                              Expanded(
                                flex: 2,
                                child: Column(
                                  children: [
                                    _buildFavoriteArtists(),
                                    const SizedBox(height: 16),
                                    _buildFavoriteSong(),
                                  ],
                                ),
                              ),
                          ],
                        ),
                        const SizedBox(height: 24),
                        _profileStatus == 'private'
                            ? _buildPrivateState()
                            : _buildProfilePaperTabs(isMobile: false),
                      ],
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildMobileLayout() {
    return SingleChildScrollView(
      controller: _scrollController,
      padding: const EdgeInsets.all(16.0),
      child: Column(
        children: [
          _buildPolaroidProfile(),
          const SizedBox(height: 16),
          if (_profileStatus == 'blocked')
            _buildBlockedState()
          else ...[
            _buildLevelCard(),
            const SizedBox(height: 16),
            _buildStatsAndActionsRow(isMobile: true),
            const SizedBox(height: 16),
            _buildAboutMe(),
            const SizedBox(height: 16),
            if (_profileStatus == 'private')
              _buildPrivateState()
            else ...[
              _buildFavoriteSong(),
              const SizedBox(height: 16),
              _buildFavoriteArtists(),
              const SizedBox(height: 24),
              _buildProfilePaperTabs(isMobile: true),
            ],
          ],
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
              borderRadius: BorderRadius.circular(4),
              border: _activeFrameColor != null
                  ? Border.all(color: _activeFrameColor!, width: 3)
                  : null,
              boxShadow: [
                BoxShadow(
                  color: _activeFrameColor != null
                      ? _activeFrameColor!.withValues(alpha: 0.5)
                      : Colors.black.withValues(alpha: 0.1),
                  blurRadius: _activeFrameColor != null ? 15 : 8,
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
                Text(
                  _displayUsername,
                  style: const TextStyle(
                    fontFamily: 'Comic Sans MS',
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 4),
                if (_profileStatus != 'blocked')
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
          child: CircleAvatar(
            radius: 6,
            backgroundColor: _activePinColor,
            child: _activePinColor != Colors.red[700]
                ? const Icon(Icons.star, size: 8, color: Colors.white)
                : null,
          ),
        ),
      ],
    );
  }

  Widget _buildLevelCard() {
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
              const Text(
                "Nivel 5 -> Buen progreso",
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
            value: 0.8,
            backgroundColor: Colors.grey[200],
            valueColor: AlwaysStoppedAnimation<Color>(_tealAccent),
            minHeight: 6,
            borderRadius: BorderRadius.circular(3),
          ),
        ],
      ),
    );
  }

  Widget _buildStatsAndActionsRow({bool isMobile = false}) {
    final int equippedCount = _inventory
        .where((i) => i['isEquipped'] as bool)
        .length;
    final statsRow = Row(
      mainAxisAlignment: isMobile
          ? MainAxisAlignment.spaceEvenly
          : MainAxisAlignment.start,
      children: [
        _pillStat("Amigos 5"),
        if (!isMobile) const SizedBox(width: 12),
        _pillStat("Artistas 13"),
        if (_profileStatus == 'own') ...[
          if (!isMobile) const SizedBox(width: 12),
          _pillStat("Decoraciones $equippedCount"),
        ],
      ],
    );

    if (isMobile) {
      return Column(
        children: [statsRow, const SizedBox(height: 12), _buildActionButtons()],
      );
    }

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Expanded(child: statsRow),
        _buildActionButtons(),
      ],
    );
  }

  Widget _pillStat(String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 4),
        ],
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: Colors.black87,
          fontSize: 12,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }

  Widget _buildActionButtons() {
    if (_profileStatus == 'own') {
      return OutlinedButton(
        style: OutlinedButton.styleFrom(
          side: BorderSide(color: _tealAccent),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          backgroundColor: Colors.white,
        ),
        onPressed: _showEditProfileDialog,
        child: Text(
          "Editar perfil",
          style: TextStyle(
            color: _tealAccent,
            fontSize: 12,
            fontWeight: FontWeight.bold,
          ),
        ),
      );
    } else {
      return Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          OutlinedButton.icon(
            style: OutlinedButton.styleFrom(
              side: BorderSide(color: _tealAccent),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20),
              ),
              backgroundColor: Colors.white,
            ),
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text("Solicitud enviada a $_displayUsername"),
                  backgroundColor: _tealAccent,
                ),
              );
            },
            icon: Icon(Icons.person_add, color: _tealAccent, size: 16),
            label: Text(
              "Añadir amigo",
              style: TextStyle(
                color: _tealAccent,
                fontSize: 12,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      );
    }
  }

  Widget _buildAboutMe() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: const Color(0xFFFDFDFD),
        borderRadius: BorderRadius.circular(8),
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 10),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            "Sobre mi:",
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 12),
          Text(
            _biography,
            style: const TextStyle(
              fontSize: 13,
              color: Colors.black87,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFavoriteArtists() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 10,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            "Artistas que me gustan:",
            style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
          ),
          const Divider(),
          _artistListTile("Cris MJ"),
          _artistListTile("Arcangel"),
          _artistListTile("Tito el Bambino"),
          _artistListTile("Kidd Voodoo"),
          Center(
            child: TextButton(
              onPressed: () {},
              child: Text(
                "Ver mas",
                style: TextStyle(color: _tealAccent, fontSize: 11),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _artistListTile(String name) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0),
      child: Row(
        children: [
          const CircleAvatar(
            radius: 12,
            backgroundColor: Colors.black12,
            child: Icon(Icons.person, size: 14, color: Colors.black54),
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

  Widget _buildFavoriteSong() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFFDFDFD),
        borderRadius: BorderRadius.circular(8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 10,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          const Text(
            "Cancion Favorita:",
            style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: _tealAccent,
              borderRadius: BorderRadius.circular(8),
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
                    children: const [
                      Text(
                        "Flashing Lights",
                        style: TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                          fontSize: 11,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        "Kanye West",
                        style: TextStyle(color: Colors.white70, fontSize: 9),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // --- SECCIÓN DE ESTADOS ESPECIALES (RN-09, RF-55) ---
  Widget _buildBlockedState() {
    return Container(
      padding: const EdgeInsets.all(48.0),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.no_accounts, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 24),
          const Text(
            "Esta cuenta no está disponible",
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: Colors.black87,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            "El perfil que intentas buscar no existe o ha restringido tu acceso.",
            textAlign: TextAlign.center,
            style: TextStyle(fontSize: 13, color: Colors.grey),
          ),
        ],
      ),
    );
  }

  Widget _buildPrivateState() {
    return Container(
      padding: const EdgeInsets.all(48.0),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.lock, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 24),
          const Text(
            "Este perfil es privado",
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: Colors.black87,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            "Añade a este usuario como amigo para ver sus publicaciones, artistas favoritos y más detalles.",
            textAlign: TextAlign.center,
            style: TextStyle(fontSize: 13, color: Colors.grey),
          ),
        ],
      ),
    );
  }

  // --- SECCIÓN DE TABS (BLOGS E INVENTARIO) ---
  Widget _buildProfilePaperTabs({required bool isMobile}) {
    final bool isOwn = _profileStatus == 'own';

    return Container(
      height: 700,
      decoration: BoxDecoration(
        color: _paperColor,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.15),
            blurRadius: 20,
          ),
        ],
      ),
      child: DefaultTabController(
        length: isOwn ? 2 : 1,
        child: Column(
          children: [
            TabBar(
              indicatorColor: _tealAccent,
              indicatorWeight: 3,
              labelColor: _tealAccent,
              unselectedLabelColor: Colors.grey,
              tabs: [
                const Tab(text: "Tus Blogs"),
                if (isOwn) const Tab(text: "Inventario"),
              ],
            ),
            const Divider(height: 1, thickness: 1, color: Colors.black12),
            Expanded(
              child: TabBarView(
                children: [
                  _buildBlogsTab(isMobile: isMobile),
                  if (isOwn) _buildInventoryTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBlogsTab({required bool isMobile}) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  _filterChip("recientes", isActive: true),
                  _filterChip("tendencias", isActive: false),
                ],
              ),
              if (_profileStatus == 'own')
                OutlinedButton.icon(
                  style: OutlinedButton.styleFrom(
                    side: BorderSide(color: _tealAccent),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                    ),
                  ),
                  onPressed: () {},
                  icon: Icon(Icons.edit, color: _tealAccent, size: 14),
                  label: Text(
                    "Tus reacciones",
                    style: TextStyle(
                      color: _tealAccent,
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 24),
          Expanded(
            child: _isInitialLoading
                ? Center(child: CircularProgressIndicator(color: _tealAccent))
                : _posts.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.post_add, size: 48, color: Colors.grey[400]),
                        const SizedBox(height: 12),
                        Text(
                          "Aún no tiene publicaciones",
                          style: TextStyle(
                            color: _paperTextColor.withValues(alpha: 0.6),
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  )
                : ListView.builder(
                    controller: isMobile ? null : ScrollController(),
                    itemCount: _posts.length + (_isLoadingMore ? 1 : 0),
                    itemBuilder: (context, index) {
                      if (index == _posts.length) {
                        return Padding(
                          padding: const EdgeInsets.symmetric(vertical: 16.0),
                          child: Center(
                            child: CircularProgressIndicator(
                              color: _tealAccent,
                            ),
                          ),
                        );
                      }
                      // Renderiza usando la Tarjeta Reutilizable PostCard
                      return PostCard(
                        postData: _posts[index],
                        tealAccent: _tealAccent,
                        textColor: _paperTextColor,
                        onDelete: () {
                          setState(() {
                            _posts.removeAt(index);
                          });
                        },
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildInventoryTab() {
    final double screenWidth = MediaQuery.of(context).size.width;
    final int crossAxisCount = screenWidth > 600 ? 3 : 2;

    return GridView.builder(
      padding: const EdgeInsets.all(32.0),
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: crossAxisCount,
        crossAxisSpacing: 16,
        mainAxisSpacing: 16,
        childAspectRatio: 0.85,
      ),
      itemCount: _inventory.length,
      itemBuilder: (context, index) {
        final item = _inventory[index];
        final bool isEquipped = item['isEquipped'] as bool;

        return Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: isEquipped
                ? Border.all(color: _tealAccent, width: 2)
                : Border.all(color: Colors.grey.withValues(alpha: 0.2)),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.05),
                blurRadius: 8,
              ),
            ],
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                item['icon'] as IconData,
                size: 40,
                color: item['color'] as Color,
              ),
              const SizedBox(height: 12),
              Text(
                item['name'] as String,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 13,
                  color: Colors.black87,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                item['type'] == 'frame'
                    ? "Marco"
                    : item['type'] == 'pin'
                    ? "Pin"
                    : "Fondo",
                style: const TextStyle(fontSize: 10, color: Colors.grey),
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: isEquipped ? Colors.white : _tealAccent,
                  foregroundColor: isEquipped ? Colors.redAccent : Colors.white,
                  side: isEquipped
                      ? const BorderSide(color: Colors.redAccent)
                      : null,
                  elevation: 0,
                ),
                onPressed: () => _toggleEquipItem(index),
                child: Text(
                  isEquipped ? "Desequipar" : "Equipar",
                  style: const TextStyle(fontSize: 11),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _filterChip(String label, {required bool isActive}) {
    return Container(
      margin: const EdgeInsets.only(right: 8),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: isActive ? Colors.white : Colors.transparent,
        border: isActive
            ? Border.all(color: _tealAccent, width: 1)
            : Border.all(color: Colors.grey),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: isActive ? _tealAccent : Colors.grey,
          fontSize: 10,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}
