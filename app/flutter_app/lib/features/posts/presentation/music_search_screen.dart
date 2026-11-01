import 'package:flutter/material.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';

class MusicSearchScreen extends StatefulWidget {
  final bool pickForPost;
  final ApiClient? apiClient;

  const MusicSearchScreen({
    super.key,
    this.pickForPost = false,
    this.apiClient,
  });

  @override
  State<MusicSearchScreen> createState() => _MusicSearchScreenState();
}

class _MusicSearchScreenState extends State<MusicSearchScreen> {
  final TextEditingController _queryController = TextEditingController();
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  List<MusicTrack> _tracks = const [];
  bool _isLoading = false;
  String? _errorMessage;
  bool _hasSearched = false;

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;
  }

  @override
  void dispose() {
    _queryController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _search() async {
    final query = _queryController.text.trim();
    if (query.isEmpty || _isLoading) return;

    setState(() {
      _isLoading = true;
      _errorMessage = null;
      _hasSearched = true;
    });
    try {
      final response = await _apiClient.searchMusicTracks(query);
      final page = response.data;
      if (page == null) {
        throw const ApiException(
          message: 'El servicio de música no devolvió resultados válidos.',
        );
      }
      if (!mounted) return;
      setState(() {
        _tracks = page.items;
        _isLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _errorMessage = error.message;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.pickForPost ? 'Elegir canción' : 'Buscar música'),
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _queryController,
                    textInputAction: TextInputAction.search,
                    onSubmitted: (_) => _search(),
                    decoration: const InputDecoration(
                      labelText: 'Canción o artista',
                      prefixIcon: Icon(Icons.search),
                      border: OutlineInputBorder(),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                IconButton.filled(
                  onPressed: _isLoading ? null : _search,
                  tooltip: 'Buscar canciones',
                  icon: const Icon(Icons.search),
                ),
              ],
            ),
          ),
          if (_isLoading) const LinearProgressIndicator(),
          if (_errorMessage != null)
            Padding(
              padding: const EdgeInsets.all(24),
              child: Text(
                _errorMessage!,
                textAlign: TextAlign.center,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            )
          else if (!_isLoading && _hasSearched && _tracks.isEmpty)
            const Padding(
              padding: EdgeInsets.all(24),
              child: Text('No encontramos canciones con esa búsqueda.'),
            )
          else if (!_hasSearched)
            const Padding(
              padding: EdgeInsets.all(24),
              child: Text('Busca una canción para añadirla a tu publicación.'),
            )
          else
            Expanded(
              child: ListView.builder(
                itemCount: _tracks.length,
                itemBuilder: (context, index) {
                  final track = _tracks[index];
                  return ListTile(
                    leading: track.artworkUrl == null
                        ? const CircleAvatar(
                            child: Icon(Icons.music_note_outlined),
                          )
                        : CircleAvatar(
                            backgroundImage: NetworkImage(track.artworkUrl!),
                          ),
                    title: Text(track.title),
                    subtitle: Text(track.artist),
                    trailing: widget.pickForPost
                        ? const Icon(Icons.add_circle_outline)
                        : null,
                    onTap: widget.pickForPost
                        ? () => Navigator.pop(context, track)
                        : null,
                  );
                },
              ),
            ),
        ],
      ),
    );
  }
}
