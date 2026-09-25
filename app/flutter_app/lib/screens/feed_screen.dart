import 'package:flutter/material.dart';

class FeedScreen extends StatefulWidget {
  const FeedScreen({super.key});

  @override
  State<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends State<FeedScreen> {
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
          _buildCustomTopBar(isDesktop),
          Expanded(
            child: isDesktop 
              ? _buildDesktopLayout() 
              : _buildMobileLayout(),
          ),
        ],
      ),
    );
  }

  // --- 1. TOP BAR CORREGIDA ---
  Widget _buildCustomTopBar(bool isDesktop) {
    return Container(
      height: 55, // Más delgada como en Figma
      color: tealAccent,
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Logo ajustado
          Image.asset(
            'assets/imagenes/meloop.png', 
            height: 24, // Mucho más pequeño
            color: Colors.white, 
            fit: BoxFit.contain, // Evita que se estire o se vea gordo
          ),
          
          if (isDesktop) ...[
            // Menú central
            Row(
              children: [
                _navLink('Inicio', isActive: true),
                _navLink('Amigos'),
                _navLink('Artistas y canciones'),
              ],
            ),
            // Reproductor superior
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
                  Text("Playing... DJ Gouz", style: TextStyle(color: Colors.white, fontSize: 12)),
                  SizedBox(width: 8),
                  Icon(Icons.graphic_eq, color: Colors.white, size: 16),
                ],
              ),
            ),
          ],
          
          // Iconos derechos
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
          )
        ],
      ),
    );
  }

  // Pastilla de navegación exacta a Figma
  Widget _navLink(String text, {bool isActive = false}) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 4.0),
      padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 6.0),
      decoration: BoxDecoration(
        color: isActive ? Colors.white : Colors.transparent,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: isActive ? tealAccent : Colors.white,
          fontWeight: isActive ? FontWeight.bold : FontWeight.normal,
          fontSize: 12,
        ),
      ),
    );
  }

  // --- LAYOUT ESCRITORIO ---
  Widget _buildDesktopLayout() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center, // Centra el contenido
        children: [
          // Sidebar Izquierdo
          SizedBox(
            width: 220, // Un poco más angosto
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
          // Feed Central
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 650), // Ancho máximo para el papel
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
          _buildFeedPaper(),
        ],
      ),
    );
  }

  // --- 2. TARJETAS LATERALES BLANCAS COMO EN FIGMA ---

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
              color: Colors.white, // Blanco puro
              boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 8)],
            ),
            child: Column(
              children: [
                Container(
                  height: 140,
                  width: double.infinity,
                  color: Colors.grey[200], // Fondo foto
                  child: const Icon(Icons.pets, size: 50, color: Colors.grey),
                ),
                const SizedBox(height: 16),
                const Text("Nombre de usuario", style: TextStyle(fontFamily: 'Comic Sans MS', fontWeight: FontWeight.bold, fontSize: 14)),
                const SizedBox(height: 4),
                const Text("Me siento Feliz", style: TextStyle(fontSize: 11, color: Colors.black54)),
              ],
            ),
          ),
        ),
        // Cinta adhesiva superior
        Positioned(
          top: 0,
          child: Transform.rotate(
            angle: -0.08,
            child: Container(
              width: 50,
              height: 18,
              color: Colors.white.withValues(alpha: 0.7), // Efecto masking tape
            ),
          ),
        ),
        // Pin rojo
        Positioned(top: 8, child: CircleAvatar(radius: 5, backgroundColor: Colors.red[700])),
      ],
    );
  }

  Widget _buildListeningTo() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white, // Corregido: Era blanco en Figma
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          Container(
            width: 40, height: 40, 
            decoration: BoxDecoration(color: Colors.purple[800], borderRadius: BorderRadius.circular(4)), 
            child: const Icon(Icons.music_note, color: Colors.white)
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: const [
              Text("Escuchando:", style: TextStyle(color: Colors.black54, fontSize: 10)),
              Text("Finding Urself", style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold, fontSize: 12)),
              Text("DJ GOUZ", style: TextStyle(color: Colors.black54, fontSize: 10)),
            ],
          )
        ],
      ),
    );
  }

  Widget _buildLevelCard() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white, // Corregido: Era blanco en Figma
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          Row(
            children: [
              Icon(Icons.star, color: tealAccent, size: 14),
              const SizedBox(width: 4),
              const Text("Nivel 3 -> Buen progreso", style: TextStyle(color: Colors.black87, fontSize: 11, fontWeight: FontWeight.bold)),
            ],
          ),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: 0.6,
            backgroundColor: Colors.grey[200],
            valueColor: AlwaysStoppedAnimation<Color>(tealAccent),
            minHeight: 6,
            borderRadius: BorderRadius.circular(3),
          )
        ],
      ),
    );
  }

  Widget _buildSuggestions() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(8)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text("Artistas que te podrían gustar", style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.black87)),
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
          const CircleAvatar(radius: 10, backgroundColor: Colors.black12, child: Icon(Icons.person, size: 12, color: Colors.black54)),
          const SizedBox(width: 8),
          Text(name, style: const TextStyle(fontSize: 11, color: Colors.black87)),
        ],
      ),
    );
  }

  // --- 3. MURO DE PUBLICACIONES ---
  Widget _buildFeedPaper() {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFDFDFD), // Blanco papel
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.15), blurRadius: 20)],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
        child: Column(
          children: [
            // Cabecera del Feed
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    const Text("Ultimos Blogs:", style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                    const SizedBox(width: 16),
                    _filterChip("recientes", isActive: true),
                    _filterChip("tendencias", isActive: false),
                    _filterChip("amigos", isActive: false),
                  ],
                ),
                // Botón Outlined (Borde Verde, texto verde)
                OutlinedButton.icon(
                  style: OutlinedButton.styleFrom(
                    side: BorderSide(color: tealAccent),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12)
                  ),
                  onPressed: () {},
                  icon: Icon(Icons.edit, color: tealAccent, size: 14),
                  label: Text("escribir blog", style: TextStyle(color: tealAccent, fontSize: 12, fontWeight: FontWeight.bold)),
                )
              ],
            ),
            const Divider(height: 32, thickness: 1, color: Colors.black12),
            
            // Lista de Publicaciones
            Expanded(
              child: ListView.builder(
                itemCount: 2,
                itemBuilder: (context, index) {
                  return _buildPostCard();
                },
              ),
            ),
            // Botón inferior
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: tealAccent,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                elevation: 0,
              ),
              onPressed: () {},
              child: const Text("Cargar más blogs", style: TextStyle(color: Colors.white, fontSize: 12)),
            )
          ],
        ),
      ),
    );
  }

  // Chips corregidos
  Widget _filterChip(String label, {required bool isActive}) {
    return Container(
      margin: const EdgeInsets.only(right: 8),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: isActive ? Colors.white : Colors.grey[200],
        border: isActive ? Border.all(color: tealAccent, width: 1) : Border.all(color: Colors.transparent),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(label, style: TextStyle(color: isActive ? tealAccent : Colors.grey[600], fontSize: 10, fontWeight: FontWeight.bold)),
    );
  }

  Widget _buildPostCard() {
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
                  const CircleAvatar(radius: 16, backgroundColor: Colors.black87, child: Icon(Icons.person, color: Colors.white, size: 18)),
                  const SizedBox(width: 12),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: const [
                      Text("UsuarioEjemplo", style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                      Text("Hace 20 minutos", style: TextStyle(fontSize: 10, color: Colors.grey)),
                    ],
                  ),
                ],
              ),
              const Icon(Icons.more_horiz, color: Colors.grey, size: 20),
            ],
          ),
          const SizedBox(height: 12),
          const Text(
            "No les parece raro como es que el sonido cambió después del 2011? Es un tema de discusión muy interesante en la producción musical actual.", 
            style: TextStyle(fontSize: 13, color: Colors.black87, height: 1.4)
          ),
          const SizedBox(height: 12),
          
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(color: tealAccent, borderRadius: BorderRadius.circular(6)),
            child: Row(
              children: [
                Container(width: 36, height: 36, color: Colors.black87, child: const Icon(Icons.play_arrow, color: Colors.white, size: 20)),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: const [
                      Text("Bangarang (feat. Sirah)", style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 12)),
                      Text("Skrillex, Sirah", style: TextStyle(color: Colors.white70, fontSize: 10)),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                  decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
                  child: Row(
                    children: [
                      Icon(Icons.graphic_eq, color: tealAccent, size: 12),
                      const SizedBox(width: 4),
                      Text("Escuchar", style: TextStyle(color: tealAccent, fontSize: 10, fontWeight: FontWeight.bold)),
                    ],
                  ),
                )
              ],
            ),
          ),
          const SizedBox(height: 12),
          
          Row(
            children: [
              Row(children: const [Icon(Icons.favorite, color: Colors.red, size: 16), SizedBox(width: 4), Text("21 Likes", style: TextStyle(color: Colors.grey, fontSize: 11))]),
              const SizedBox(width: 16),
              Row(children: const [Icon(Icons.mode_comment_outlined, color: Colors.grey, size: 16), SizedBox(width: 4), Text("3 Comentarios", style: TextStyle(color: Colors.grey, fontSize: 11))]),
            ],
          ),
          const SizedBox(height: 16),
          const Divider(color: Colors.black12),
        ],
      ),
    );
  }
}