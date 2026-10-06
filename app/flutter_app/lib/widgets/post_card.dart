import 'package:flutter/material.dart';

import '../../features/comments/presentation/comments_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../features/profile/presentation/post_editor_dialog.dart';

class PostCard extends StatefulWidget {
  final Map<String, dynamic> postData;
  final Color tealAccent;
  final Color textColor;
  final VoidCallback?
  onDelete; // Callback para eliminar de la vista sin recargar

  const PostCard({
    super.key,
    required this.postData,
    required this.tealAccent,
    this.textColor = Colors.black87,
    this.onDelete,
  });

  @override
  State<PostCard> createState() => _PostCardState();
}

class _PostCardState extends State<PostCard> {
  void _showDeleteConfirmation(BuildContext context) {
    showDialog(
      context: context,
      builder: (BuildContext dialogContext) {
        bool isDeleting = false;

        return StatefulBuilder(
          builder: (context, setState) {
            return AlertDialog(
              backgroundColor: Colors.white,
              title: const Text(
                "Eliminar publicación",
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              content: const Text(
                "¿Estás seguro de que deseas eliminar esta publicación? Esta acción no se puede deshacer y se borrará del servidor.",
                style: TextStyle(fontSize: 14),
              ),
              actions: [
                TextButton(
                  onPressed: isDeleting
                      ? null
                      : () => Navigator.pop(dialogContext),
                  child: const Text(
                    "Cancelar",
                    style: TextStyle(color: Colors.grey),
                  ),
                ),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.redAccent,
                    elevation: 0,
                  ),
                  onPressed: isDeleting
                      ? null
                      : () async {
                          setState(() => isDeleting = true);

                          // Simulación de petición HTTP DELETE al API Gateway
                          await Future.delayed(const Duration(seconds: 1));

                          if (!context.mounted) return;
                          Navigator.pop(dialogContext); // Cierra el modal

                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text(
                                "Publicación eliminada correctamente.",
                              ),
                              backgroundColor: Colors.redAccent,
                              duration: Duration(seconds: 2),
                            ),
                          );

                          // Ejecuta el callback para quitar la card de la lista local
                          if (widget.onDelete != null) {
                            widget.onDelete!();
                          }
                        },
                  child: isDeleting
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : const Text(
                          "Eliminar",
                          style: TextStyle(color: Colors.white),
                        ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final String user = widget.postData["user"] ?? "Usuario";
    final String time = widget.postData["time"] ?? "Hace un momento";
    final String content = widget.postData["content"] ?? "";
    final String? imageUrl = widget.postData["image_url"];
    final String? audioTitle = widget.postData["song"];
    final String? artistName = widget.postData["artist"];
    final int likes = widget.postData["likes"] ?? 0;
    final int comments = widget.postData["comments"] ?? 0;

    // --- Variables de lógica de negocio (Roles y Permisos) ---
    final int postId = widget.postData["id"] ?? 0;
    final bool isAuthor = widget.postData["isAuthor"] ?? (postId % 2 == 0);
    final bool isEdited = widget.postData["isEdited"] ?? false;
    final bool within24Hours = widget.postData["within24Hours"] ?? true;

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
                  InkWell(
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => ProfileScreen(username: user),
                        ),
                      );
                    },
                    child: const CircleAvatar(
                      radius: 16,
                      backgroundColor: Colors.black87,
                      child: Icon(Icons.person, color: Colors.white, size: 18),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      InkWell(
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  ProfileScreen(username: user),
                            ),
                          );
                        },
                        child: Text(
                          user,
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 13,
                            color: widget.textColor,
                          ),
                        ),
                      ),
                      Text(
                        isEdited ? "$time (editado)" : time,
                        style: const TextStyle(
                          fontSize: 10,
                          color: Colors.grey,
                        ),
                      ),
                    ],
                  ),
                ],
              ),

              // Menú de opciones dinámico según el ROL
              PopupMenuButton<String>(
                icon: const Icon(
                  Icons.more_horiz,
                  color: Colors.grey,
                  size: 20,
                ),
                color: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
                onSelected: (value) async {
                  if (value == 'edit') {
                    // Esperamos el resultado del modal de edición
                    final String? updatedText = await showDialog<String>(
                      context: context,
                      builder: (context) => PostEditorDialog(
                        postData: widget.postData,
                        tealAccent: widget.tealAccent,
                      ),
                    );

                    // Si recibimos un texto nuevo, actualizamos la tarjeta
                    if (updatedText != null &&
                        updatedText != widget.postData["content"]) {
                      setState(() {
                        widget.postData["content"] = updatedText;
                        widget.postData["isEdited"] =
                            true; // Mostramos la etiqueta (editado)
                      });
                    }
                  } else if (value == 'delete') {
                    _showDeleteConfirmation(context);
                  } else if (value == 'save') {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: const Text(
                          "Publicación guardada en tu colección.",
                        ),
                        backgroundColor: widget.tealAccent,
                      ),
                    );
                  } else if (value == 'report') {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text(
                          "Publicación reportada a los administradores.",
                        ),
                        backgroundColor: Colors.orange,
                      ),
                    );
                  }
                },
                itemBuilder: (context) {
                  if (isAuthor) {
                    return [
                      if (within24Hours)
                        const PopupMenuItem(
                          value: 'edit',
                          child: Text(
                            "Editar publicación",
                            style: TextStyle(fontSize: 13),
                          ),
                        )
                      else
                        const PopupMenuItem(
                          value: 'expired',
                          enabled: false,
                          child: Text(
                            "Tiempo de edición agotado",
                            style: TextStyle(fontSize: 13, color: Colors.grey),
                          ),
                        ),
                      const PopupMenuItem(
                        value: 'delete',
                        child: Text(
                          "Eliminar",
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.redAccent,
                          ),
                        ),
                      ),
                    ];
                  } else {
                    return [
                      const PopupMenuItem(
                        value: 'save',
                        child: Text(
                          "Guardar publicación",
                          style: TextStyle(fontSize: 13),
                        ),
                      ),
                      const PopupMenuItem(
                        value: 'report',
                        child: Text(
                          "Reportar",
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.redAccent,
                          ),
                        ),
                      ),
                    ];
                  }
                },
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            content,
            style: TextStyle(
              fontSize: 13,
              color: widget.textColor,
              height: 1.4,
            ),
          ),
          const SizedBox(height: 12),

          if (imageUrl != null && imageUrl.isNotEmpty) ...[
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: Image.network(
                imageUrl,
                height: 200,
                width: double.infinity,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) => Container(
                  height: 150,
                  color: Colors.grey[200],
                  child: const Center(
                    child: Icon(Icons.broken_image, color: Colors.grey),
                  ),
                ),
              ),
            ),
            const SizedBox(height: 12),
          ],

          if (audioTitle != null && audioTitle.isNotEmpty) ...[
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: widget.tealAccent,
                borderRadius: BorderRadius.circular(6),
              ),
              child: Row(
                children: [
                  Container(
                    width: 36,
                    height: 36,
                    color: Colors.black87,
                    child: const Icon(
                      Icons.play_arrow,
                      color: Colors.white,
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          audioTitle,
                          style: const TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.bold,
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                        Text(
                          artistName ?? "Artista desconocido",
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 10,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Row(
                      children: [
                        Icon(
                          Icons.graphic_eq,
                          color: widget.tealAccent,
                          size: 12,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          "Escuchar",
                          style: TextStyle(
                            color: widget.tealAccent,
                            fontSize: 10,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],

          Row(
            children: [
              Row(
                children: [
                  const Icon(Icons.favorite, color: Colors.red, size: 16),
                  const SizedBox(width: 4),
                  Text(
                    "$likes Likes",
                    style: const TextStyle(color: Colors.grey, fontSize: 11),
                  ),
                ],
              ),
              const SizedBox(width: 16),
              InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) =>
                          CommentsScreen(postData: widget.postData),
                    ),
                  );
                },
                child: Row(
                  children: [
                    const Icon(
                      Icons.mode_comment_outlined,
                      color: Colors.grey,
                      size: 16,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      "$comments Comentarios",
                      style: const TextStyle(color: Colors.grey, fontSize: 11),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          const Divider(color: Colors.black12),
        ],
      ),
    );
  }
}
