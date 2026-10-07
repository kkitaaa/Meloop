import 'package:flutter/material.dart';

import '../features/posts/post_detail_screen.dart';
import '../screens/friends_screen.dart';

class NotificationsMenu extends StatefulWidget {
  const NotificationsMenu({super.key});

  @override
  State<NotificationsMenu> createState() => _NotificationsMenuState();
}

class _NotificationsMenuState extends State<NotificationsMenu> {
  final MenuController _menuController = MenuController();
  bool _showSettings = false;
  final _preferences = <String, bool>{'like': true, 'comment': true, 'friend': true};
  final _items = <_NotificationItem>[
    _NotificationItem(type: 'like', title: 'A Camila le gustó tu publicación', detail: 'Hace 2 min', postId: 1),
    _NotificationItem(type: 'comment', title: 'Felipe comentó tu publicación', detail: '“Qué buena canción”', postId: 2),
    _NotificationItem(type: 'friend', title: 'Valentina te envió una solicitud', detail: 'Hace 1 hora'),
  ];

  @override
  Widget build(BuildContext context) {
    final unread = _items.where((item) => !item.read && _preferences[item.type]!).length;
    return MenuAnchor(
      controller: _menuController,
      style: const MenuStyle(
        padding: WidgetStatePropertyAll(EdgeInsets.zero),
        elevation: WidgetStatePropertyAll(12),
      ),
      menuChildren: [
        SizedBox(width: 365, child: _buildPanel(context)),
      ],
      builder: (context, controller, child) => Stack(
        clipBehavior: Clip.none,
        children: [
          IconButton(
            tooltip: 'Notificaciones',
            icon: const Icon(Icons.notifications, color: Colors.white, size: 20),
            onPressed: () => controller.isOpen ? controller.close() : controller.open(),
          ),
          if (unread > 0)
            Positioned(
              top: 6,
              right: 6,
              child: Container(
                width: 14,
                height: 14,
                alignment: Alignment.center,
                decoration: const BoxDecoration(color: Color(0xFFFF6262), shape: BoxShape.circle),
                child: Text('$unread', style: const TextStyle(color: Colors.white, fontSize: 8, fontWeight: FontWeight.bold)),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildPanel(BuildContext context) {
    final items = _items.where((item) => _preferences[item.type]!).toList();
    return SizedBox(
      height: _showSettings ? 270 : 360,
      child: Material(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 8, 8),
              child: Row(children: [
                Text(_showSettings ? 'Preferencias' : 'Notificaciones', style: const TextStyle(fontSize: 17, fontWeight: FontWeight.bold)),
                const Spacer(),
                IconButton(
                  tooltip: _showSettings ? 'Ver notificaciones' : 'Configurar notificaciones',
                  icon: Icon(_showSettings ? Icons.arrow_back : Icons.tune, color: const Color(0xFF1E7E70)),
                  onPressed: () => setState(() => _showSettings = !_showSettings),
                ),
              ]),
            ),
            const Divider(height: 1),
            if (_showSettings)
              _settingsPanel()
            else
              Expanded(child: _notificationsList(context, items)),
          ],
        ),
      ),
    );
  }

  Widget _settingsPanel() => Padding(
    padding: const EdgeInsets.all(10),
    child: Column(mainAxisSize: MainAxisSize.min, children: [
      const Padding(padding: EdgeInsets.all(6), child: Text('Elige qué alertas quieres recibir.', style: TextStyle(color: Colors.black54, fontSize: 12))),
      _preferenceSwitch('like', 'Likes en publicaciones', Icons.favorite),
      _preferenceSwitch('comment', 'Comentarios', Icons.mode_comment_outlined),
      _preferenceSwitch('friend', 'Solicitudes de amistad', Icons.person_add_alt_1),
    ]),
  );

  Widget _preferenceSwitch(String key, String label, IconData icon) => SwitchListTile(
    value: _preferences[key]!,
    activeColor: const Color(0xFF00AF92),
    onChanged: (value) => setState(() => _preferences[key] = value),
    secondary: Icon(icon, color: const Color(0xFF1E7E70), size: 20),
    title: Text(label, style: const TextStyle(fontSize: 13)),
  );

  Widget _notificationsList(BuildContext context, List<_NotificationItem> items) {
    if (items.isEmpty) return const Padding(padding: EdgeInsets.all(32), child: Text('No hay notificaciones para mostrar.'));
    return ListView.separated(
      shrinkWrap: true,
      itemCount: items.length,
      separatorBuilder: (_, _) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final item = items[index];
        return InkWell(
          onTap: () => _openNotification(context, item),
          child: Container(
            color: item.read ? Colors.white : const Color(0xFFE9F8F5),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            child: Row(children: [
              CircleAvatar(radius: 18, backgroundColor: _iconColor(item.type).withValues(alpha: .15), child: Icon(_icon(item.type), size: 18, color: _iconColor(item.type))),
              const SizedBox(width: 10),
              Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [Text(item.title, style: TextStyle(fontSize: 13, fontWeight: item.read ? FontWeight.normal : FontWeight.bold)), const SizedBox(height: 3), Text(item.detail, style: const TextStyle(fontSize: 11, color: Colors.black54))])),
              if (!item.read) const Icon(Icons.circle, size: 8, color: Color(0xFF00AF92)),
            ]),
          ),
        );
      },
    );
  }

  void _openNotification(BuildContext context, _NotificationItem item) {
    setState(() => item.read = true);
    _menuController.close();
    if (item.postId != null) {
      Navigator.push(context, MaterialPageRoute(builder: (_) => PostDetailScreen(postId: item.postId!)));
    } else {
      Navigator.push(context, MaterialPageRoute(builder: (_) => const FriendsScreen()));
    }
  }

  IconData _icon(String type) => switch (type) { 'like' => Icons.favorite, 'comment' => Icons.mode_comment, _ => Icons.person_add_alt_1 };
  Color _iconColor(String type) => switch (type) { 'like' => const Color(0xFFFF6262), 'comment' => const Color(0xFF00AF92), _ => const Color(0xFF5B7FFF) };
}

class _NotificationItem {
  _NotificationItem({required this.type, required this.title, required this.detail, this.postId});
  final String type;
  final String title;
  final String detail;
  final int? postId;
  bool read = false;
}
