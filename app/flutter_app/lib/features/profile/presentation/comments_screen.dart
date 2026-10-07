import 'package:flutter/material.dart';
import '../../../widgets/states/loading_state.dart';

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

  bool _isLoading = true;
  bool _isSubmitting = false;

  // Variables para anidación de respuestas
  int? _replyingToCommentId;
  String? _replyingToUser;

  // Estructura de árbol simulando respuesta del GET /v1/posts/{id}/comments
  List<Map<String, dynamic>> _comments = [];

  @override
  void initState() {
    super.initState();
    _fetchInitialComments();
  }

  @override
  void dispose() {
    _commentController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  // --- SIMULACIÓN DE PETICIONES HTTP AL API GATEWAY ---

  Future<void> _fetchInitialComments() async {
    setState(() => _isLoading = true);
    await Future.delayed(const Duration(seconds: 1)); // Simula latencia red

    if (!mounted) return;

    setState(() {
      _comments = [
        {
          "id": 1,
          "user": "Martín",
          "text": "Totalmente de acuerdo, la producción de esa época era distinta.",
          "time": "Hace 15 min",
          "likes": 5,
          "isLiked": false,
          "isLiking": false,
          "replies": <Map<String, dynamic>>[],
          "replyCount": 2, // El backend dice que hay 2 respuestas sin cargar
          "isLoadingReplies": false,
        },
        {
          "id": 2,
          "user": "Camila",
          "text": "Yo prefiero el sonido de ahora, más limpio.",
          "time": "Hace 5 min",
          "likes": 2,
          "isLiked": true,
          "isLiking": false,
          "replies": <Map<String, dynamic>>[],
          "replyCount": 0, // No hay respuestas ocultas
          "isLoadingReplies": false,
        },
      ];
      _isLoading = false;
    });
  }

  // Paginación simulada para hilos de comentarios
  Future<void> _fetchReplies(int parentId) async {
    final parent = _findCommentById(_comments, parentId);
    if (parent == null) return;

    setState(() => parent["isLoadingReplies"] = true);
    await Future.delayed(const Duration(milliseconds: 800));

    if (!mounted) return;

    setState(() {
      parent["isLoadingReplies"] = false;
      int newId = DateTime.now().millisecondsSinceEpoch;
      
      parent["replies"].addAll([
        {
          "id": newId,
          "user": "Usuario_Respuesta",
          "text": "¡Exacto! Tienes mucha razón en eso.",
          "time": "Hace un momento",
          "likes": 1,
          "isLiked": false,
          "isLiking": false,
          "replies": <Map<String, dynamic>>[],
          "replyCount": 0,
          "isLoadingReplies": false,
        },
        {
          "id": newId + 1,
          "user": "CriticoMusical",
          "text": "Aunque depende del álbum la verdad...",
          "time": "Hace un momento",
          "likes": 0,
          "isLiked": false,
          "isLiking": false,
          "replies": <Map<String, dynamic>>[],
          "replyCount": 0,
          "isLoadingReplies": false,
        }
      ]);
    });
  }

  // --- LÓGICA DE ÁRBOL Y OPTIMISTIC UI ---

  // Búsqueda recursiva para encontrar cualquier comentario sin importar su anidación
  Map<String, dynamic>? _findCommentById(List<Map<String, dynamic>> list, int id) {
    for (var comment in list) {
      if (comment["id"] == id) return comment;
      if (comment["replies"] != null) {
        final found = _findCommentById(comment["replies"], id);
        if (found != null) return found;
      }
    }
    return null;
  }

  void _startReply(int commentId, String username) {
    setState(() {
      _replyingToCommentId = commentId;
      _replyingToUser = username;
    });
    _focusNode.requestFocus();
  }

  void _cancelReply() {
    setState(() {
      _replyingToCommentId = null;
      _replyingToUser = null;
    });
    _focusNode.unfocus();
  }

  Future<void> _toggleCommentLike(int id) async {
    final comment = _findCommentById(_comments, id);
    if (comment == null || comment["isLiking"] == true) return;

    final bool wasLiked = comment["isLiked"] ?? false;
    final int currentLikes = comment["likes"] as int;

    // Actualización Optimista
    setState(() {
      comment["isLiking"] = true;
      comment["isLiked"] = !wasLiked;
      comment["likes"] = wasLiked ? currentLikes - 1 : currentLikes + 1;
    });

    await Future.delayed(const Duration(milliseconds: 600));
    if (!mounted) return;

    // Simulación de respuesta exitosa del servidor
    bool httpSuccess = DateTime.now().second % 2 == 0;

    if (httpSuccess) {
      setState(() => comment["isLiking"] = false);
    } else {
      // Rollback
      setState(() {
        comment["isLiking"] = false;
        comment["isLiked"] = wasLiked;
        comment["likes"] = currentLikes;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("Error de conexión."), backgroundColor: Colors.redAccent),
      );
    }
  }

  Future<void> _togglePostLike() async {
    if (widget.postData["isLiking"] == true) return;

    final bool wasLiked = widget.postData["isLiked"] ?? false;
    final int currentLikes = widget.postData["likes"] ?? 0;

    setState(() {
      widget.postData["isLiking"] = true;
      widget.postData["isLiked"] = !wasLiked;
      widget.postData["likes"] = wasLiked ? currentLikes - 1 : currentLikes + 1;
    });

    await Future.delayed(const Duration(milliseconds: 600));
    if (!mounted) return;

    setState(() => widget.postData["isLiking"] = false);
  }

  Future<void> _submitComment() async {
    final text = _commentController.text.trim();
    if (text.isEmpty) return;

    setState(() => _isSubmitting = true);
    FocusScope.of(context).unfocus();

    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;

    // Simulación de error (RN-03)
    if (text.toLowerCase().contains("insulto")) {
      setState(() => _isSubmitting = false);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text("Error: El comentario incumple las normas de la comunidad."),
          backgroundColor: Colors.redAccent,
        ),
      );
      return;
    }

    setState(() {
      final newComment = {
        "id": DateTime.now().millisecondsSinceEpoch,
        "user": "MiUsuario",
        "text": _replyingToUser != null ? "@$_replyingToUser $text" : text,
        "time": "Ahora",
        "likes": 0,
        "isLiked": false,
        "isLiking": false,
        "replies": <Map<String, dynamic>>[],
        "replyCount": 0,
        "isLoadingReplies": false,
      };

      if (_replyingToCommentId != null) {
        // Añadir como respuesta anidada
        final parent = _findCommentById(_comments, _replyingToCommentId!);
        if (parent != null) {
          parent["replies"].add(newComment);
        }
      } else {
        // Añadir a la raíz
        _comments.add(newComment);
      }

      // Actualizamos el contador global de la tarjeta
      widget.postData["comments"] = (widget.postData["comments"] ?? 0) + 1;

      _commentController.clear();
      _cancelReply();
      _isSubmitting = false;
    });
  }

  // --- RENDERIZADO DE LA INTERFAZ ---

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.of(context).size.width;
    final bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: _bgColor,
      appBar: AppBar(
        backgroundColor: _tealAccent,
        title: const Text("Comentarios", style: TextStyle(color: Colors.white, fontSize: 16)),
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: BoxConstraints(maxWidth: isDesktop ? 650 : double.infinity),
          child: Container(
            color: const Color(0xFFFDFDFD),
            child: Column(
              children: [
                _buildOriginalPost(),
                const Divider(height: 1, thickness: 4, color: Colors.black12),
                Expanded(
                  child: _isLoading
                      ? const LoadingState()
                      : ListView.separated(
                          padding: const EdgeInsets.all(24.0),
                          itemCount: _comments.length,
                          separatorBuilder: (context, index) => const Divider(height: 32),
                          itemBuilder: (context, index) => _buildCommentNode(_comments[index], 0),
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
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            widget.postData["content"] ?? "",
            style: const TextStyle(fontSize: 14, color: Colors.black87, height: 1.4),
          ),
          const SizedBox(height: 16),
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

  // Widget recursivo para renderizar el árbol de comentarios
  Widget _buildCommentNode(Map<String, dynamic> comment, int depth) {
    final List<Map<String, dynamic>> replies = comment["replies"] ?? [];
    final int replyCount = comment["replyCount"] ?? 0;
    
    // Sangría visual por nivel (máximo nivel visual recomendado: 3)
    final double leftPadding = depth * 32.0;

    return Padding(
      padding: EdgeInsets.only(top: depth == 0 ? 0 : 16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Comentario Actual
          Padding(
            padding: EdgeInsets.only(left: leftPadding),
            child: _buildSingleComment(comment),
          ),
          
          // Renderiza respuestas ya cargadas de forma recursiva
          for (var reply in replies) 
            _buildCommentNode(reply, depth + 1),

          // Botón "Ver más respuestas" (Si hay respuestas en el servidor que no hemos cargado)
          if (replyCount > replies.length)
            Padding(
              padding: EdgeInsets.only(left: leftPadding + 40.0, top: 12.0),
              child: InkWell(
                onTap: () => _fetchReplies(comment["id"]),
                child: comment["isLoadingReplies"] == true
                    ? const SizedBox(width: 14, height: 14, child: CircularProgressIndicator(strokeWidth: 2))
                    : Text(
                        "Ver más respuestas (${replyCount - replies.length})",
                        style: TextStyle(
                          color: _tealAccent,
                          fontWeight: FontWeight.bold,
                          fontSize: 11,
                        ),
                      ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildSingleComment(Map<String, dynamic> comment) {
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
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
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
                    onTap: () => _toggleCommentLike(comment["id"]),
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
                    onTap: () => _startReply(comment["id"], comment["user"]),
                    child: const Text(
                      "Responder",
                      style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.grey),
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
                padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 8),
                color: Colors.grey[100],
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      "Respondiendo a @$_replyingToUser",
                      style: TextStyle(fontSize: 12, color: _tealAccent, fontWeight: FontWeight.bold),
                    ),
                    InkWell(
                      onTap: _cancelReply,
                      child: const Icon(Icons.close, size: 16, color: Colors.grey),
                    ),
                  ],
                ),
              ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 12.0),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _commentController,
                      focusNode: _focusNode,
                      decoration: InputDecoration(
                        hintText: "Escribe un comentario...",
                        hintStyle: const TextStyle(fontSize: 13),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
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