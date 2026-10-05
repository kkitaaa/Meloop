import 'package:flutter/material.dart';

class CommentsScreen extends StatefulWidget {
  final Map<String, dynamic> postData;

  const CommentsScreen({super.key, required this.postData});

  @override
  State<CommentsScreen> createState() => _CommentsScreenState();
}

class _CommentsScreenState extends State<CommentsScreen> {
  final Color _bgColor = const Color(0xFF0D5C5E);
  final Color _tealAccent = const Color(0xFF1ABC9C);

  final TextEditingController _commentController = TextEditingController();
  final FocusNode _focusNode = FocusNode();

  String? _replyingToUser;
  bool _isSubmitting = false;

  // Lista local simulando la base de datos (con soporte para Likes)
  final List<Map<String, dynamic>> _comments = [
    {
      "id": 1,
      "user": "Martín",
      "text": "Totalmente de acuerdo, la producción de esa época era distinta.",
      "time": "Hace 15 min",
      "likes": 5,
      "isLiked": false,
      "isLiking": false,
    },
    {
      "id": 2,
      "user": "Camila",
      "text": "Yo prefiero el sonido de ahora, más limpio.",
      "time": "Hace 5 min",
      "likes": 2,
      "isLiked": true,
      "isLiking": false,
    },
  ];

  @override
  void dispose() {
    _commentController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _startReply(String username) {
    setState(() => _replyingToUser = username);
    _focusNode.requestFocus();
  }

  void _cancelReply() {
    setState(() => _replyingToUser = null);
    _focusNode.unfocus();
  }

  // Optimistic UI para dar Like a un comentario
  Future<void> _toggleCommentLike(int index) async {
    final comment = _comments[index];
    
    // Evitar spam de toques mientras se procesa la petición HTTP
    if (comment["isLiking"] == true) return;

    final bool wasLiked = comment["isLiked"] ?? false;
    final int currentLikes = comment["likes"] as int;

    // 1. Actualización Optimista inmediata
    setState(() {
      comment["isLiking"] = true;
      comment["isLiked"] = !wasLiked;
      comment["likes"] = wasLiked ? currentLikes - 1 : currentLikes + 1;
    });

    // 2. Simulamos latencia del API Gateway
    await Future.delayed(const Duration(milliseconds: 800));
    if (!mounted) return;

    // 3. Resultado del servidor (simulamos éxito aleatorio para quitar el Dead Code)
    // Ahora Dart no sabe qué pasará, así que quita la alerta amarilla.
    bool httpSuccess = DateTime.now().second % 2 == 0; 

    if (httpSuccess) {
      setState(() {
        comment["isLiking"] = false;
      });
    } else {
      // 4. Rollback visual si falla el servidor
      setState(() {
        comment["isLiking"] = false;
        comment["isLiked"] = wasLiked;
        comment["likes"] = currentLikes;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text("Error al conectar con el servidor. Se revirtió tu like."),
          backgroundColor: Colors.redAccent,
        ),
      );
    }
  }

  // Optimistic UI para dar Like a la Publicación Original desde esta vista
  Future<void> _togglePostLike() async {
    if (widget.postData["isLiking"] == true) return;

    final bool wasLiked = widget.postData["isLiked"] ?? false;
    final int currentLikes = widget.postData["likes"] ?? 0;

    setState(() {
      widget.postData["isLiking"] = true;
      widget.postData["isLiked"] = !wasLiked;
      widget.postData["likes"] = wasLiked ? currentLikes - 1 : currentLikes + 1;
    });

    await Future.delayed(const Duration(milliseconds: 800));
    if (!mounted) return;

    setState(() {
      widget.postData["isLiking"] = false;
    });
  }

  Future<void> _submitComment() async {
    final text = _commentController.text.trim();
    if (text.isEmpty) return;

    setState(() => _isSubmitting = true);
    FocusScope.of(context).unfocus();

    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;

    if (text.toLowerCase().contains("insulto")) {
      setState(() => _isSubmitting = false);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text("Error: El comentario incumple las normas de la comunidad (RN-03)."),
          backgroundColor: Colors.redAccent,
        ),
      );
      return;
    }

    setState(() {
      _comments.add({
        "id": DateTime.now().millisecondsSinceEpoch,
        "user": "MiUsuario",
        "text": _replyingToUser != null ? "@$_replyingToUser $text" : text,
        "time": "Ahora",
        "likes": 0,
        "isLiked": false,
        "isLiking": false,
      });

      // Actualizamos el contador de comentarios en la PostCard Original
      widget.postData["comments"] = (widget.postData["comments"] ?? 0) + 1;
      
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
      backgroundColor: _bgColor,
      appBar: AppBar(
        backgroundColor: _tealAccent,
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
                _buildOriginalPost(),
                const Divider(height: 1, thickness: 4, color: Colors.black12),
                Expanded(
                  child: ListView.separated(
                    padding: const EdgeInsets.all(24.0),
                    itemCount: _comments.length,
                    separatorBuilder: (context, index) => const Divider(height: 32),
                    itemBuilder: (context, index) => _buildCommentTile(index),
                  ),
                ),
                _buildCommentInputBar(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildOriginalPost() {
    final bool isPostLiked = widget.postData["isLiked"] ?? false;

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
          const SizedBox(height: 16),
          // Contadores sincronizados con la PostCard
          Row(
            children: [
              InkWell(
                onTap: _togglePostLike,
                borderRadius: BorderRadius.circular(4),
                child: Row(
                  children: [
                    Icon(
                      isPostLiked ? Icons.favorite : Icons.favorite_border,
                      color: isPostLiked ? Colors.red : Colors.grey,
                      size: 16,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      "${widget.postData["likes"] ?? 0} Likes",
                      style: TextStyle(
                        color: isPostLiked ? Colors.red : Colors.grey,
                        fontSize: 11,
                        fontWeight: isPostLiked ? FontWeight.bold : FontWeight.normal,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Row(
                children: [
                  const Icon(Icons.mode_comment_outlined, color: Colors.grey, size: 16),
                  const SizedBox(width: 4),
                  Text(
                    "${widget.postData["comments"] ?? 0} Comentarios",
                    style: const TextStyle(color: Colors.grey, fontSize: 11),
                  ),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCommentTile(int index) {
    final comment = _comments[index];
    final bool isLiked = comment["isLiked"] ?? false;

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
                  InkWell(
                    onTap: () => _toggleCommentLike(index),
                    borderRadius: BorderRadius.circular(4),
                    child: Padding(
                      padding: const EdgeInsets.only(right: 8.0, top: 4.0, bottom: 4.0),
                      child: Row(
                        children: [
                          Icon(
                            isLiked ? Icons.favorite : Icons.favorite_border,
                            color: isLiked ? Colors.red : Colors.grey,
                            size: 14,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            "${comment["likes"]} Likes",
                            style: TextStyle(
                              fontSize: 11,
                              color: isLiked ? Colors.red : Colors.grey,
                              fontWeight: isLiked ? FontWeight.bold : FontWeight.normal,
                            ),
                          ),
                        ],
                      ),
                    ),
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
                        color: _tealAccent,
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
                          icon: Icon(Icons.send, color: _tealAccent),
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