import 'package:flutter/material.dart';

class CommentsScreen extends StatefulWidget {
  final Map<String, dynamic> postData;

  const CommentsScreen({super.key, required this.postData});

  @override
  State<CommentsScreen> createState() => _CommentsScreenState();
}

class _CommentsScreenState extends State<CommentsScreen> {
  final Color bgColor = const Color(0xFF0D5C5E);
  final Color tealAccent = const Color(0xFF1ABC9C);

  final TextEditingController _commentController = TextEditingController();
  final FocusNode _focusNode = FocusNode();

  String? _replyingToUser;
  bool _isSubmitting = false;

  // Lista de comentarios simulada
  final List<Map<String, dynamic>> _comments = [
    {
      "id": 1,
      "user": "Martín",
      "text": "Totalmente de acuerdo, la producción de esa época era distinta.",
      "time": "Hace 15 min",
      "likes": 5,
    },
    {
      "id": 2,
      "user": "Camila",
      "text": "Yo prefiero el sonido de ahora, más limpio.",
      "time": "Hace 5 min",
      "likes": 2,
    },
  ];

  @override
  void dispose() {
    _commentController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _startReply(String username) {
    setState(() {
      _replyingToUser = username;
    });
    _focusNode.requestFocus();
  }

  void _cancelReply() {
    setState(() {
      _replyingToUser = null;
    });
    _focusNode.unfocus();
  }

  Future<void> _submitComment() async {
    final text = _commentController.text.trim();
    if (text.isEmpty) return;

    setState(() => _isSubmitting = true);

    // Ocultar teclado
    FocusScope.of(context).unfocus();

    // Simular latencia HTTP al API Gateway
    await Future.delayed(const Duration(seconds: 1));

    // Simulación de Validación del Servidor (Ejemplo Regla RN-03: Moderación)
    if (text.toLowerCase().contains("insulto")) {
      setState(() => _isSubmitting = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              "Error: El comentario incumple las normas de la comunidad (RN-03).",
            ),
            backgroundColor: Colors.redAccent,
          ),
        );
      }
      return;
    }

    // Inserción optimista en la lista local
    setState(() {
      _comments.add({
        "id": DateTime.now().millisecondsSinceEpoch,
        "user": "MiUsuario", // Usuario logueado
        "text": _replyingToUser != null ? "@$_replyingToUser $text" : text,
        "time": "Ahora",
        "likes": 0,
      });
      _commentController.clear();
      _replyingToUser = null;
      _isSubmitting = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.of(context).size.width;
    final bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: bgColor,
      appBar: AppBar(
        backgroundColor: tealAccent,
        title: const Text(
          "Comentarios",
          style: TextStyle(color: Colors.white, fontSize: 16),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: BoxConstraints(
            maxWidth: isDesktop ? 650 : double.infinity,
          ),
          child: Container(
            color: const Color(0xFFFDFDFD),
            child: Column(
              children: [
                // Post Original (Contexto)
                _buildOriginalPost(),
                const Divider(height: 1, thickness: 4, color: Colors.black12),

                // Lista de Comentarios
                Expanded(
                  child: ListView.separated(
                    padding: const EdgeInsets.all(24.0),
                    itemCount: _comments.length,
                    separatorBuilder: (context, index) =>
                        const Divider(height: 32),
                    itemBuilder: (context, index) {
                      return _buildCommentTile(_comments[index]);
                    },
                  ),
                ),

                // Barra fija inferior para comentar
                _buildCommentInputBar(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildOriginalPost() {
    return Padding(
      padding: const EdgeInsets.all(24.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const CircleAvatar(
                radius: 16,
                backgroundColor: Colors.black87,
                child: Icon(Icons.person, color: Colors.white, size: 18),
              ),
              const SizedBox(width: 12),
              Text(
                widget.postData["user"] ?? "Usuario",
                style: const TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 14,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            widget.postData["content"] ?? "",
            style: const TextStyle(
              fontSize: 14,
              color: Colors.black87,
              height: 1.4,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCommentTile(Map<String, dynamic> comment) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const CircleAvatar(
          radius: 14,
          backgroundColor: Colors.grey,
          child: Icon(Icons.person, color: Colors.white, size: 16),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Text(
                    comment["user"],
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 13,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    comment["time"],
                    style: const TextStyle(fontSize: 10, color: Colors.grey),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                comment["text"],
                style: const TextStyle(fontSize: 13, color: Colors.black87),
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Text(
                    "${comment["likes"]} Likes",
                    style: const TextStyle(fontSize: 11, color: Colors.grey),
                  ),
                  const SizedBox(width: 16),
                  InkWell(
                    onTap: () => _startReply(comment["user"]),
                    child: const Text(
                      "Responder",
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: Colors.grey,
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildCommentInputBar() {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, -5),
          ),
        ],
      ),
      child: SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Indicador de "Respondiendo a..."
            if (_replyingToUser != null)
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 8,
                ),
                color: Colors.grey[100],
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      "Respondiendo a @$_replyingToUser",
                      style: TextStyle(
                        fontSize: 12,
                        color: tealAccent,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    InkWell(
                      onTap: _cancelReply,
                      child: const Icon(
                        Icons.close,
                        size: 16,
                        color: Colors.grey,
                      ),
                    ),
                  ],
                ),
              ),

            // Campo de texto y botón enviar
            Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: 16.0,
                vertical: 12.0,
              ),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _commentController,
                      focusNode: _focusNode,
                      decoration: InputDecoration(
                        hintText: "Escribe un comentario...",
                        hintStyle: const TextStyle(fontSize: 13),
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 12,
                        ),
                        filled: true,
                        fillColor: Colors.grey[100],
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(24),
                          borderSide: BorderSide.none,
                        ),
                      ),
                      onSubmitted: (_) => _submitComment(),
                    ),
                  ),
                  const SizedBox(width: 8),
                  _isSubmitting
                      ? const Padding(
                          padding: EdgeInsets.all(12.0),
                          child: SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        )
                      : IconButton(
                          onPressed: _submitComment,
                          icon: Icon(Icons.send, color: tealAccent),
                        ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
