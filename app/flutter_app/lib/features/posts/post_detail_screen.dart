import 'package:flutter/material.dart';

import 'dart:async';

import '../../../widgets/post_card.dart';

class PostDetailScreen extends StatefulWidget {
  final int postId;

  const PostDetailScreen({super.key, required this.postId});

  @override
  State<PostDetailScreen> createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  final Color _bgColor = const Color(0xFF0D5C5E);
  final Color _tealAccent = const Color(0xFF1ABC9C);

  bool _isLoading = true;
  bool _isError = false;
  Map<String, dynamic>? _postData;

  @override
  void initState() {
    super.initState();
    _fetchPostDetail();
  }

  // Simulación de petición HTTP GET al API Gateway
  Future<void> _fetchPostDetail() async {
    setState(() {
      _isLoading = true;
      _isError = false;
    });

    await Future.delayed(const Duration(seconds: 1)); // Simula latencia de red

    if (!mounted) return;

    // Simulación CU-08 E3: Publicación eliminada o no disponible
    // Si el ID es 999, forzamos el error para probar la vista vacía.
    if (widget.postId == 999) {
      setState(() {
        _isLoading = false;
        _isError = true;
      });
      return;
    }

    // Respuesta exitosa simulada desde el backend
    setState(() {
      _postData = {
        "id": widget.postId,
        "user": "Autor_0${widget.postId}",
        "time": "Hace 2 horas",
        "content":
            "Visualizando el detalle completo de la publicación #${widget.postId}. Aquí el texto completo, multimedia y toda la información asociada a esta entrada (RF-19).",
        "song": "Track Detail #${widget.postId}",
        "artist": "Artista Principal",
        "likes": 42,
        "comments": 5,
        // "image_url": "https://via.placeholder.com/600x400" // Descomentar si hay imagen
      };
      _isLoading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.of(context).size.width;
    final bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: _bgColor,
      appBar: AppBar(
        backgroundColor: _tealAccent,
        title: const Text(
          "Detalle de Publicación",
          style: TextStyle(color: Colors.white, fontSize: 16),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
        centerTitle: true,
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: BoxConstraints(
            maxWidth: isDesktop ? 650 : double.infinity,
          ),
          child: Container(
            height: double.infinity,
            color: const Color(0xFFFDFDFD),
            child: _buildBodyContent(),
          ),
        ),
      ),
    );
  }

  Widget _buildBodyContent() {
    if (_isLoading) {
      return Center(child: CircularProgressIndicator(color: _tealAccent));
    }

    if (_isError || _postData == null) {
      return _buildErrorState();
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.all(32.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Renderiza el componente reutilizable de la tarjeta
          PostCard(postData: _postData!, tealAccent: _tealAccent),

          const SizedBox(height: 16),
          const Text(
            "Sección de comentarios",
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
          ),
          const Divider(height: 32, thickness: 1, color: Colors.black12),

          // Punto de entrada visual para los comentarios
          Center(
            child: OutlinedButton.icon(
              style: OutlinedButton.styleFrom(
                side: BorderSide(color: _tealAccent),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(20),
                ),
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 12,
                ),
              ),
              onPressed: () {
                // Aquí podrías abrir la hoja de comentarios o navegar a la CommentsScreen
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: const Text("Abriendo comentarios..."),
                    backgroundColor: _tealAccent,
                  ),
                );
              },
              icon: Icon(Icons.comment, color: _tealAccent, size: 16),
              label: Text(
                "Escribir un comentario",
                style: TextStyle(
                  color: _tealAccent,
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.sentiment_dissatisfied,
              size: 64,
              color: Colors.grey.withValues(alpha: 0.5),
            ),
            const SizedBox(height: 16),
            const Text(
              "Publicación no encontrada",
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: Colors.black87,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              "Esta publicación ya no está disponible o fue eliminada por el autor.",
              textAlign: TextAlign.center,
              style: TextStyle(fontSize: 13, color: Colors.grey),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: _tealAccent,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(20),
                ),
              ),
              onPressed: () => Navigator.pop(context),
              child: const Text(
                "Volver al inicio",
                style: TextStyle(color: Colors.white, fontSize: 12),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
