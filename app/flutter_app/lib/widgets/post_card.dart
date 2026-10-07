import 'package:flutter/material.dart';

import '../../features/comments/presentation/comments_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../features/profile/presentation/post_editor_dialog.dart';

class PostCard extends StatelessWidget {
  final Map<String, dynamic> postData;
  final Color tealAccent;
  final Color textColor;

  const PostCard({
    super.key,
    required this.postData,
    required this.tealAccent,
    this.textColor = Colors.black87,
  });

  @override
  Widget build(BuildContext context) {
    final String user = postData["user"] ?? "Usuario";
    final String time = postData["time"] ?? "Hace un momento";
    final String content = postData["content"] ?? "";
    final String? imageUrl = postData["image_url"];
    final String? audioTitle = postData["song"];
    final String? artistName = postData["artist"];
    final int likes = postData["likes"] ?? 0;
    final int comments = postData["comments"] ?? 0;

    // --- Variables de lógica de negocio (RN-11, RF-17) ---
    // En producción esto viene del backend (JSON)
    final bool isAuthor =
        postData["isAuthor"] ?? true; // Simulado a true para probar
    final bool isEdited = postData["isEdited"] ?? false;
    final bool within24Hours = postData["within24Hours"] ?? true;

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
                            color: textColor,
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

              // Menú de opciones (Visible solo para el autor)
              if (isAuthor)
                PopupMenuButton<String>(
                  icon: const Icon(
                    Icons.more_horiz,
                    color: Colors.grey,
                    size: 20,
                  ),
                  color: Colors.white,
                  onSelected: (value) async {
                    if (value == 'edit') {
                      // Abre el modal reutilizable en modo "Edición"
                      await showDialog(
                        context: context,
                        builder: (context) => PostEditorDialog(
                          postData: postData,
                          tealAccent: tealAccent,
                        ),
                      );
                    }
                  },
                  itemBuilder: (context) {
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
                  },
                )
              else
                const Icon(Icons.more_horiz, color: Colors.grey, size: 20),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            content,
            style: TextStyle(fontSize: 13, color: textColor, height: 1.4),
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
                color: tealAccent,
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
                        Icon(Icons.graphic_eq, color: tealAccent, size: 12),
                        const SizedBox(width: 4),
                        Text(
                          "Escuchar",
                          style: TextStyle(
                            color: tealAccent,
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
                      builder: (context) => CommentsScreen(postData: postData),
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
