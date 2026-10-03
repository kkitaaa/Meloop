import 'package:flutter/material.dart';

class FriendsScreen extends StatefulWidget {
  const FriendsScreen({super.key});

  @override
  State<FriendsScreen> createState() => _FriendsScreenState();
}

class _FriendsScreenState extends State<FriendsScreen> {
  final Color bgColor = const Color(0xFF0D5C5E); 
  final Color tealAccent = const Color(0xFF1ABC9C);

  @override
  Widget build(BuildContext context) {
    double screenWidth = MediaQuery.of(context).size.width;
    bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: bgColor,
      body: Column(
        children: [
          _buildCustomTopBar(isDesktop, context),
          Expanded(
            child: isDesktop 
              ? _buildDesktopLayout() 
              : _buildMobileLayout(),
          ),
        ],
      ),
    );
  }

  // --- TOP BAR ---
  Widget _buildCustomTopBar(bool isDesktop, BuildContext context) {
    return Container(
      height: 55,
      color: tealAccent,
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Image.asset('assets/imagenes/meloop.png', height: 24, color: Colors.white, fit: BoxFit.contain),
          
          if (isDesktop) ...[
            Row(
              children: [
                _navLink('Inicio', isActive: false, onTap: () => Navigator.pop(context)),
                _navLink('Amigos', isActive: true, onTap: () {}),
                _navLink('Artistas y canciones', isActive: false, onTap: () {}),
              ],
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
              decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(20)),
              child: Row(
                children: const [
                  Icon(Icons.play_circle_fill, color: Colors.white, size: 18),
                  SizedBox(width: 8),
                  Text("Playing... DJ Gouz", style: TextStyle(color: Colors.white, fontSize: 12)),
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
              CircleAvatar(radius: 14, backgroundColor: Colors.white, child: Icon(Icons.person, color: Color(0xFF1ABC9C), size: 18)),
            ],
          )
        ],
      ),
    );
  }

  Widget _navLink(String text, {required bool isActive, required VoidCallback onTap}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 4.0),
        padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 6.0),
        decoration: BoxDecoration(
          color: isActive ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Text(
          text,
          style: TextStyle(color: isActive ? tealAccent : Colors.white, fontWeight: isActive ? FontWeight.bold : FontWeight.normal, fontSize: 12),
        ),
      ),
    );
  }

  // --- LAYOUTS ---
  Widget _buildDesktopLayout() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          SizedBox(width: 220, child: _buildPolaroidProfile()),
          const SizedBox(width: 24),
          ConstrainedBox(constraints: const BoxConstraints(maxWidth: 700), child: _buildFriendsPaper()),
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
          const SizedBox(height: 24),
          _buildFriendsPaper(),
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
            decoration: BoxDecoration(color: Colors.white, boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 8)]),
            child: Column(
              children: [
                Container(height: 140, width: double.infinity, color: Colors.grey[200], child: const Icon(Icons.pets, size: 50, color: Colors.grey)),
                const SizedBox(height: 16),
                const Text("Nombre de usuario", style: TextStyle(fontFamily: 'Comic Sans MS', fontWeight: FontWeight.bold, fontSize: 14)),
                const SizedBox(height: 4),
                const Text("245 Amigos", style: TextStyle(fontSize: 12, color: Color(0xFF1ABC9C), fontWeight: FontWeight.bold)),
              ],
            ),
          ),
        ),
        Positioned(top: 0, child: Transform.rotate(angle: -0.08, child: Container(width: 50, height: 18, color: Colors.white.withValues(alpha: 0.7)))),
        Positioned(top: 8, child: CircleAvatar(radius: 5, backgroundColor: Colors.red[700])),
      ],
    );
  }

  // --- CONTENEDOR CENTRAL DE AMIGOS ---
  Widget _buildFriendsPaper() {
    return Container(
      height: 650, 
      decoration: BoxDecoration(
        color: const Color(0xFFFDFDFD),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.15), blurRadius: 20)],
      ),
      child: DefaultTabController(
        length: 3,
        child: Column(
          children: [
            // Cabecera y Tabs
            Padding(
              padding: const EdgeInsets.only(left: 32, right: 32, top: 24, bottom: 8),
              child: Row(
                children: [
                  const Icon(Icons.people_alt, size: 28),
                  const SizedBox(width: 12),
                  const Text("Red de Amigos", style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold)),
                ],
              ),
            ),
            TabBar(
              indicatorColor: tealAccent,
              indicatorWeight: 3,
              labelColor: tealAccent,
              unselectedLabelColor: Colors.grey,
              labelStyle: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
              tabs: const [
                Tab(text: "Mis Amigos"),
                Tab(text: "Solicitudes (2)"),
                Tab(text: "Sugerencias"),
              ],
            ),
            const Divider(height: 1, thickness: 1, color: Colors.black12),
            
            // Vistas de las Tabs
            Expanded(
              child: TabBarView(
                children: [
                  _buildMyFriendsTab(),
                  _buildRequestsTab(),
                  _buildSuggestionsTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // 1. Pestaña: Mis Amigos
  Widget _buildMyFriendsTab() {
    return ListView.separated(
      padding: const EdgeInsets.all(24),
      itemCount: 4,
      separatorBuilder: (context, index) => const Divider(color: Colors.black12),
      itemBuilder: (context, index) {
        return ListTile(
          leading: const CircleAvatar(backgroundColor: Colors.black87, child: Icon(Icons.person, color: Colors.white)),
          title: Text("Amigo ${index + 1}", style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
          subtitle: const Text("Escuchando a Skrillex", style: TextStyle(fontSize: 12, color: Colors.grey)),
          trailing: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              IconButton(
                tooltip: "Bloquear",
                icon: const Icon(Icons.block, color: Colors.grey, size: 20),
                onPressed: () {},
              ),
              IconButton(
                tooltip: "Eliminar amigo",
                icon: const Icon(Icons.person_remove, color: Colors.redAccent, size: 20),
                onPressed: () {},
              ),
            ],
          ),
        );
      },
    );
  }

  // 2. Pestaña: Solicitudes
  Widget _buildRequestsTab() {
    return ListView.separated(
      padding: const EdgeInsets.all(24),
      itemCount: 2,
      separatorBuilder: (context, index) => const Divider(color: Colors.black12),
      itemBuilder: (context, index) {
        return ListTile(
          leading: const CircleAvatar(backgroundColor: Colors.indigo, child: Icon(Icons.person, color: Colors.white)),
          title: Text("NuevoUsuario_0${index + 1}", style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
          subtitle: const Text("Te ha enviado una solicitud de amistad", style: TextStyle(fontSize: 12, color: Colors.grey)),
          trailing: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              OutlinedButton(
                style: OutlinedButton.styleFrom(
                  side: const BorderSide(color: Colors.grey),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                ),
                onPressed: () {},
                child: const Text("Rechazar", style: TextStyle(color: Colors.grey, fontSize: 12)),
              ),
              const SizedBox(width: 8),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: tealAccent,
                  elevation: 0,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                ),
                onPressed: () {},
                child: const Text("Aceptar", style: TextStyle(color: Colors.white, fontSize: 12)),
              ),
            ],
          ),
        );
      },
    );
  }

  // 3. Pestaña: Sugerencias 
  Widget _buildSuggestionsTab() {
    // Datos simulados con porcentajes de compatibilidad
    final suggestions = [
      {"name": "Martín", "comp": 95, "desc": "Ambos escuchan a DJ Gouz y Skrillex"},
      {"name": "Camila", "comp": 82, "desc": "Tienen 12 artistas en común"},
      {"name": "Vicente", "comp": 60, "desc": "Ambos escuchan electrónica"},
    ];

    return ListView.separated(
      padding: const EdgeInsets.all(24),
      itemCount: suggestions.length,
      separatorBuilder: (context, index) => const SizedBox(height: 16),
      itemBuilder: (context, index) {
        final sugg = suggestions[index];
        final comp = sugg["comp"] as int;
        
        return Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            border: Border.all(color: Colors.grey.withValues(alpha:0.2)),
            borderRadius: BorderRadius.circular(12),
            boxShadow: [BoxShadow(color: Colors.black.withValues(alpha:0.02), blurRadius: 4, offset: const Offset(0, 2))],
          ),
          child: Row(
            children: [
              const CircleAvatar(radius: 24, backgroundColor: Colors.black87, child: Icon(Icons.person, color: Colors.white)),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(sugg["name"] as String, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
                    const SizedBox(height: 4),
                    Text(sugg["desc"] as String, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                    const SizedBox(height: 8),
                    // Etiqueta de Compatibilidad Musical
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: tealAccent.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: tealAccent.withValues(alpha:0.5)),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.graphic_eq, color: tealAccent, size: 12),
                          const SizedBox(width: 4),
                          Text("$comp% Compatibilidad", style: TextStyle(color: tealAccent, fontSize: 11, fontWeight: FontWeight.bold)),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: tealAccent,
                  elevation: 0,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                ),
                onPressed: () {},
                icon: const Icon(Icons.person_add, color: Colors.white, size: 16),
                label: const Text("Conectar", style: TextStyle(color: Colors.white, fontSize: 12)),
              ),
            ],
          ),
        );
      },
    );
  }
}