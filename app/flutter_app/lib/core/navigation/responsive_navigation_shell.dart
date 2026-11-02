import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../features/moderation/data/moderation_models.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/profile/presentation/post_editor_dialog.dart';
import '../../screens/feed_screen.dart';
import '../../screens/friends_screen.dart';
import '../../widgets/desktop_chat.dart';
import 'artists_and_songs_screen.dart';

class ResponsiveNavigationShell extends StatefulWidget {
  const ResponsiveNavigationShell({
    super.key,
    this.userRole = ModerationRole.user,
  });

  final ModerationRole userRole;

  @override
  State<ResponsiveNavigationShell> createState() =>
      _ResponsiveNavigationShellState();
}

class _ResponsiveNavigationShellState extends State<ResponsiveNavigationShell> {
  static const _accentColor = Color(0xFF00AF92);
  int _selectedIndex = 0;
  bool _isChatConversationOpen = false;

  late final List<Widget> _screens = [
    FeedScreen(userRole: widget.userRole),
    FriendsScreen(),
    ArtistsAndSongsScreen(),
    DesktopChat(
      fullScreen: true,
      onConversationChanged: (isOpen) {
        if (_isChatConversationOpen != isOpen) {
          setState(() => _isChatConversationOpen = isOpen);
        }
      },
    ),
    const ProfileScreen(),
  ];

  Future<void> _createBlog() async {
    await showPostEditorDialog(context, tealAccent: const Color(0xFF1ABC9C));
  }

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.viewPaddingOf(context).bottom;
    final hideNavigation = _selectedIndex == 3 && _isChatConversationOpen;

    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: const SystemUiOverlayStyle(
        statusBarColor: Color(0xFF248686),
        statusBarIconBrightness: Brightness.light,
      ),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        extendBody: true,
        body: SafeArea(
          bottom: false,
          child: IndexedStack(index: _selectedIndex, children: _screens),
        ),
        bottomNavigationBar: hideNavigation
            ? null
            : Container(
                height: bottomInset + 132,
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [Color(0x00000000), Color(0x88000000)],
                  ),
                ),
                child: Stack(
                  clipBehavior: Clip.none,
                  children: [
                    Positioned(
                      left: 14,
                      right: 14,
                      bottom: bottomInset + 10,
                      height: 52,
                      child: _buildNavigationPill(),
                    ),
                    Positioned(
                      right: 24,
                      bottom: bottomInset + 72,
                      child: Material(
                        color: const Color(0xFF50DDC7),
                        shape: const CircleBorder(),
                        elevation: 8,
                        shadowColor: Colors.black.withValues(alpha: 0.3),
                        child: IconButton(
                          tooltip: 'Crear blog',
                          onPressed: _createBlog,
                          icon: const Icon(
                            Icons.edit_note,
                            color: Colors.white,
                          ),
                          iconSize: 26,
                          constraints: const BoxConstraints.tightFor(
                            width: 56,
                            height: 56,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
      ),
    );
  }

  Widget _buildNavigationPill() {
    const items = <({IconData icon, String label})>[
      (icon: Icons.home_filled, label: 'Inicio'),
      (icon: Icons.people_outline, label: 'Amigos'),
      (icon: Icons.graphic_eq, label: 'Artistas y canciones'),
      (icon: Icons.send_outlined, label: 'Chat'),
      (icon: Icons.person_outline, label: 'Perfil'),
    ];

    return Container(
      padding: const EdgeInsets.all(5),
      decoration: BoxDecoration(
        color: _accentColor,
        borderRadius: BorderRadius.circular(32),
        border: Border.all(color: Colors.white.withValues(alpha: 0.16)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.3),
            blurRadius: 18,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Row(
        children: List.generate(items.length, (index) {
          final item = items[index];
          final isSelected = index == _selectedIndex;

          return Expanded(
            child: Semantics(
              button: true,
              selected: isSelected,
              label: item.label,
              child: InkWell(
                borderRadius: BorderRadius.circular(26),
                onTap: () => setState(() => _selectedIndex = index),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 180),
                  curve: Curves.easeOut,
                  height: 42,
                  decoration: BoxDecoration(
                    color: isSelected
                        ? const Color(0xFF50DDC7)
                        : Colors.transparent,
                    borderRadius: BorderRadius.circular(26),
                  ),
                  child: Center(
                    child: index == 4
                        ? const CircleAvatar(
                            radius: 15,
                            backgroundColor: Colors.white,
                            child: Icon(
                              Icons.person,
                              color: Color(0xFF1ABC9C),
                              size: 19,
                            ),
                          )
                        : Icon(item.icon, color: Colors.white, size: 23),
                  ),
                ),
              ),
            ),
          );
        }),
      ),
    );
  }
}
