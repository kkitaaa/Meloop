import 'package:flutter/material.dart';

class PostEditorDialog extends StatefulWidget {
  final Map<String, dynamic>?
  postData; // Si es null, es Creación. Si trae datos, es Edición.
  final Color tealAccent;

  const PostEditorDialog({super.key, this.postData, required this.tealAccent});

  @override
  State<PostEditorDialog> createState() => _PostEditorDialogState();
}

class _PostEditorDialogState extends State<PostEditorDialog> {
  final TextEditingController _contentController = TextEditingController();
  bool _isSubmitting = false;

  @override
  void initState() {
    super.initState();
    if (widget.postData != null) {
      _contentController.text = widget.postData!["content"] ?? "";
    }
  }

  @override
  void dispose() {
    _contentController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final String text = _contentController.text.trim();
    if (text.isEmpty) return;

    setState(() => _isSubmitting = true);
    FocusScope.of(context).unfocus();

    // Simula la latencia de la petición HTTP al API Gateway en Go
    await Future.delayed(const Duration(seconds: 1));

    if (!mounted) return;

    // Validación del Servidor (RN-11): Edición fuera de las 24 horas
    final bool isEdit = widget.postData != null;
    final bool within24Hours = widget.postData?["within24Hours"] ?? true;

    if (isEdit && !within24Hours) {
      setState(() => _isSubmitting = false);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            "Error (RN-11): El plazo de 24 horas para editar ha expirado.",
          ),
          backgroundColor: Colors.redAccent,
        ),
      );
      return;
    }

    // Éxito
    Navigator.pop(
      context,
      text,
    ); // Retornamos el nuevo texto para actualizar la UI localmente
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          isEdit
              ? "Publicación editada con éxito."
              : "Publicación creada con éxito.",
        ),
        backgroundColor: widget.tealAccent,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final bool isEdit = widget.postData != null;

    return AlertDialog(
      backgroundColor: Colors.white,
      title: Text(
        isEdit ? "Editar Publicación" : "Crear nueva publicación",
        style: const TextStyle(fontWeight: FontWeight.bold),
      ),
      content: SizedBox(
        width: 500,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _contentController,
              maxLines: 4,
              decoration: InputDecoration(
                hintText: "¿Qué estás escuchando o pensando?",
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                IconButton(
                  onPressed: () {},
                  icon: const Icon(Icons.image, color: Colors.grey),
                  tooltip: "Adjuntar imagen",
                ),
                IconButton(
                  onPressed: () {},
                  icon: const Icon(Icons.audiotrack, color: Colors.grey),
                  tooltip: "Vincular canción",
                ),
                const Text(
                  "Adjuntar multimedia",
                  style: TextStyle(fontSize: 12, color: Colors.grey),
                ),
              ],
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isSubmitting ? null : () => Navigator.pop(context),
          child: const Text("Cancelar", style: TextStyle(color: Colors.grey)),
        ),
        ElevatedButton(
          style: ElevatedButton.styleFrom(backgroundColor: widget.tealAccent),
          onPressed: _isSubmitting ? null : _submit,
          child: _isSubmitting
              ? const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                    color: Colors.white,
                    strokeWidth: 2,
                  ),
                )
              : Text(
                  isEdit ? "Guardar cambios" : "Publicar",
                  style: const TextStyle(color: Colors.white),
                ),
        ),
      ],
    );
  }
}
