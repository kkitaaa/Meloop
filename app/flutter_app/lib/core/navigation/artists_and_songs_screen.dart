import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';

import '../../features/profile/presentation/profile_screen.dart';
import '../../features/search/presentation/global_search_route.dart';
import '../../screens/friends_screen.dart';
import '../../widgets/desktop_app_header.dart';
import '../../widgets/mobile_app_header.dart';

class ArtistsAndSongsScreen extends StatefulWidget {
  const ArtistsAndSongsScreen({super.key});

  @override
  State<ArtistsAndSongsScreen> createState() => _ArtistsAndSongsScreenState();
}

class _ArtistsAndSongsScreenState extends State<ArtistsAndSongsScreen> {
  static const _accentColor = Color(0xFF1ABC9C);
  static const _catalog = <({String title, String subtitle, IconData icon})>[
    (title: 'Kidd Voodoo', subtitle: 'Artista', icon: Icons.person_outline),
    (
      title: 'Finding Urself',
      subtitle: 'DJ Gouz · Canción',
      icon: Icons.music_note,
    ),
    (title: 'Tiago PZK', subtitle: 'Artista', icon: Icons.person_outline),
    (title: 'Skrillex', subtitle: 'Artista', icon: Icons.person_outline),
    (title: 'Cris MJ', subtitle: 'Artista', icon: Icons.person_outline),
  ];

  String _query = '';

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.sizeOf(context).width;
    final isDesktop = screenWidth > 900;
    final isAndroidMobile =
        !isDesktop && defaultTargetPlatform == TargetPlatform.android;
    final results = _catalog
        .where(
          (item) => item.title.toLowerCase().contains(_query.toLowerCase()),
        )
        .toList();

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Column(
        children: [
          if (isAndroidMobile)
            MobileAppHeader(onSearchTap: () => openGlobalSearch(context))
          else if (isDesktop)
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
          else
            Container(
              height: 55,
              color: _accentColor,
              padding: const EdgeInsets.symmetric(horizontal: 20),
              alignment: Alignment.centerLeft,
              child: const Text(
                'Artistas y canciones',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.fromLTRB(16, 20, 16, 24),
              children: [
                TextField(
                  onChanged: (value) => setState(() => _query = value.trim()),
                  decoration: InputDecoration(
                    hintText: 'Buscar artistas o canciones',
                    prefixIcon: const Icon(Icons.search),
                    filled: true,
                    fillColor: Colors.white,
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(24),
                      borderSide: BorderSide.none,
                    ),
                  ),
                ),
                const SizedBox(height: 20),
                ...results.map(
                  (item) => Card(
                    margin: const EdgeInsets.only(bottom: 10),
                    child: ListTile(
                      leading: CircleAvatar(
                        backgroundColor: _accentColor.withValues(alpha: 0.15),
                        child: Icon(item.icon, color: _accentColor),
                      ),
                      title: Text(item.title),
                      subtitle: Text(item.subtitle),
                      trailing: const Icon(Icons.chevron_right),
                    ),
                  ),
                ),
                if (results.isEmpty)
                  const Padding(
                    padding: EdgeInsets.only(top: 32),
                    child: Text(
                      'No se encontraron resultados.',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.white70),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
