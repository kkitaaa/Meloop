import 'package:flutter/material.dart';
import 'package:just_audio/just_audio.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Meloop Local Audio',
      theme: ThemeData.dark(),
      home: const MainScreen(),
      debugShowCheckedModeBanner: false,
    );
  }
}

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  final AudioPlayer _player = AudioPlayer();
  int _currentIndex = 0;

  @override
  void initState() {
    super.initState();
    _loadLocalAudio();
  }

  Future<void> _loadLocalAudio() async {
    try {
      // Carga el archivo local declarado en el pubspec.yaml
      await _player.setAsset('assets/audio/Sinaka, Easykid - DESCONTROL.mp3');
      _player.play();
    } catch (e) {
      debugPrint('❌ Error cargando audio local: $e');
    }
  }

  @override
  void dispose() {
    _player.dispose();
    super.dispose();
  }

  final List<Widget> _screens = [
    const Center(
      child: Text(
        'Feed Principal\n(Navega al Perfil y la música no se cortará)',
        textAlign: TextAlign.center,
        style: TextStyle(fontSize: 18),
      ),
    ),
    const Center(
      child: Text(
        'Perfil de Usuario',
        style: TextStyle(fontSize: 24),
      ),
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Meloop'),
        backgroundColor: Colors.black87,
      ),
      body: Stack(
        children: [
          IndexedStack(
            index: _currentIndex,
            children: _screens,
          ),
          Align(
            alignment: Alignment.bottomCenter,
            child: _buildMiniPlayer(),
          ),
        ],
      ),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (index) => setState(() => _currentIndex = index),
        backgroundColor: Colors.black,
        selectedItemColor: Colors.deepPurpleAccent,
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.home),
            label: 'Feed',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.person),
            label: 'Perfil',
          ),
        ],
      ),
    );
  }

  Widget _buildMiniPlayer() {
    return Container(
      height: 75,
      decoration: BoxDecoration(
        color: Colors.grey[900],
        border: Border(
          top: BorderSide(
            color: Colors.grey[800]!,
            width: 1,
          ),
        ),
      ),
      child: Row(
        children: [
          const SizedBox(width: 16),
          const Icon(
            Icons.music_note,
            color: Colors.deepPurpleAccent,
            size: 40,
          ),
          const SizedBox(width: 16),
          const Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Pista Local',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
                Text(
                  'Reproduciendo desde assets',
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey,
                  ),
                ),
              ],
            ),
          ),
          StreamBuilder<PlayerState>(
            stream: _player.playerStateStream,
            builder: (context, snapshot) {
              final playerState = snapshot.data;
              final processingState = playerState?.processingState;
              final playing = playerState?.playing;

              if (processingState == ProcessingState.loading ||
                  processingState == ProcessingState.buffering) {
                return const Padding(
                  padding: EdgeInsets.all(16.0),
                  child: CircularProgressIndicator(
                    color: Colors.deepPurpleAccent,
                  ),
                );
              } else if (playing != true) {
                return IconButton(
                  icon: const Icon(Icons.play_arrow, size: 36),
                  onPressed: _player.play,
                );
              } else {
                return IconButton(
                  icon: const Icon(Icons.pause, size: 36),
                  onPressed: _player.pause,
                );
              }
            },
          ),
          const SizedBox(width: 8),
        ],
      ),
    );
  }
}