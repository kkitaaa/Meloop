import 'package:flutter/material.dart';

class FeedScreen extends StatelessWidget {
  const FeedScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF1E3D59),
      appBar: AppBar(
        backgroundColor: const Color(0xFF11B49F),
        title: const Text("Meloop", style: TextStyle(fontWeight: FontWeight.bold, color: Colors.white)),
        centerTitle: true,
      ),
      body: LayoutBuilder(
        builder: (context, constraints) {
          if (constraints.maxWidth > 800) {
            // Diseño para Windows (Dos columnas)
            return Padding(
              padding: const EdgeInsets.all(32.0),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(flex: 1, child: _buildProfileCard()),
                  const SizedBox(width: 24),
                  Expanded(flex: 3, child: _buildFeed()),
                ],
              ),
            );
          } else {
            // Diseño para Android (Una columna apilada)
            return SingleChildScrollView(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                children: [
                  _buildProfileCard(),
                  const SizedBox(height: 16),
                  SizedBox(height: 600, child: _buildFeed()), // Altura fija temporal para la lista
                ],
              ),
            );
          }
        },
      ),
    );
  }

  Widget _buildProfileCard() {
    return Card(
      color: const Color(0xFFF0EBE1),
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: const [
            Icon(Icons.person, size: 80),
            Text("Vicente Flores", style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
            SizedBox(height: 8),
            Text("Me siento Feliz"),
            Text("Escuchando: DJ GOUZ", style: TextStyle(fontStyle: FontStyle.italic)),
          ],
        ),
      ),
    );
  }

  Widget _buildFeed() {
    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.only(topRight: Radius.circular(20), topLeft: Radius.circular(20)),
      ),
      padding: const EdgeInsets.all(24.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text("Últimos Blogs:", style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          const Divider(),
          Expanded(
            child: ListView.builder(
              itemCount: 5,
              itemBuilder: (context, index) => ListTile(
                leading: const CircleAvatar(child: Icon(Icons.music_note)),
                title: Text("Usuario0${index + 1} - Nueva mezcla disponible"),
                subtitle: const Text("Toca para reproducir"),
                trailing: IconButton(
                  icon: const Icon(Icons.favorite_border),
                  onPressed: () {}, // Aquí irá la lógica del Like
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}