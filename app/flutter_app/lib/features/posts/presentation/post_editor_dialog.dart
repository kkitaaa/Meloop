import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';

const int maxPostImageSizeBytes = 5 * 1024 * 1024;
const int maxPostAudioSizeBytes = 15 * 1024 * 1024;

const List<String> supportedPostImageExtensions = [
  'jpg',
  'jpeg',
  'png',
  'webp',
];

const List<String> supportedPostAudioExtensions = [
  'mp3',
  'm4a',
  'aac',
  'wav',
  'ogg',
];

class PostEditorResult {
  final String content;
  final Uint8List? imageBytes;
  final String? imageName;
  final String? imageUrl;
  final Uint8List? audioBytes;
  final String? audioName;
  final String? songTitle;
  final String? artistName;

  const PostEditorResult({
    required this.content,
    this.imageBytes,
    this.imageName,
    this.imageUrl,
    this.audioBytes,
    this.audioName,
    this.songTitle,
    this.artistName,
  });
}

class PostAttachmentValidator {
  const PostAttachmentValidator._();

  static String? validateImage({required String name, required int sizeBytes}) {
    final extension = _extensionOf(name);
    if (!supportedPostImageExtensions.contains(extension)) {
      return 'Formato de imagen no permitido. Usa JPG, PNG o WEBP.';
    }
    if (sizeBytes <= 0) {
      return 'La imagen está vacía.';
    }
    if (sizeBytes > maxPostImageSizeBytes) {
      return 'La imagen supera el límite de 5 MB.';
    }
    return null;
  }

  static String? validateAudio({required String name, required int sizeBytes}) {
    final extension = _extensionOf(name);
    if (!supportedPostAudioExtensions.contains(extension)) {
      return 'Formato de audio no permitido. Usa MP3, M4A, AAC, WAV u OGG.';
    }
    if (sizeBytes <= 0) {
      return 'El archivo de audio está vacío.';
    }
    if (sizeBytes > maxPostAudioSizeBytes) {
      return 'El audio supera el límite de 15 MB.';
    }
    return null;
  }

  static String _extensionOf(String name) {
    final separator = name.lastIndexOf('.');
    return separator < 0 ? '' : name.substring(separator + 1).toLowerCase();
  }
}

class PostEditorDialog extends StatefulWidget {
  final Map<String, dynamic>? postData;
  final Color tealAccent;

  const PostEditorDialog({super.key, this.postData, required this.tealAccent});

  @override
  State<PostEditorDialog> createState() => _PostEditorDialogState();
}

class _PostEditorDialogState extends State<PostEditorDialog> {
  final TextEditingController _contentController = TextEditingController();
  Uint8List? _imageBytes;
  String? _imageName;
  String? _imageUrl;
  Uint8List? _audioBytes;
  String? _audioName;
  String? _songTitle;
  String? _artistName;
  bool _isSubmitting = false;

  bool get _hasContent =>
      _contentController.text.trim().isNotEmpty ||
      _imageBytes != null ||
      _imageUrl != null ||
      _songTitle != null;

  @override
  void initState() {
    super.initState();
    final post = widget.postData;
    if (post != null) {
      _contentController.text = post['content'] as String? ?? '';
      _imageBytes = post['image_bytes'] is Uint8List
          ? post['image_bytes'] as Uint8List
          : null;
      _imageName = post['image_name'] as String?;
      _imageUrl = post['image_url'] as String?;
      _audioBytes = post['audio_bytes'] is Uint8List
          ? post['audio_bytes'] as Uint8List
          : null;
      _audioName = post['audio_name'] as String?;
      _songTitle = post['song'] as String?;
      _artistName = post['artist'] as String?;
    }
    _contentController.addListener(_onContentChanged);
  }

  @override
  void dispose() {
    _contentController
      ..removeListener(_onContentChanged)
      ..dispose();
    super.dispose();
  }

  void _onContentChanged() {
    if (mounted) setState(() {});
  }

  Future<void> _pickImage() async {
    await _pickAttachment(
      allowedExtensions: supportedPostImageExtensions,
      validate: PostAttachmentValidator.validateImage,
      onSelected: (bytes, name) {
        setState(() {
          _imageBytes = bytes;
          _imageName = name;
          _imageUrl = null;
        });
      },
      failureMessage: 'No se pudo seleccionar la imagen.',
    );
  }

  Future<void> _pickAudio() async {
    await _pickAttachment(
      allowedExtensions: supportedPostAudioExtensions,
      validate: PostAttachmentValidator.validateAudio,
      onSelected: (bytes, name) {
        final title = name.replaceFirst(RegExp(r'\.[^.]+$'), '');
        setState(() {
          _audioBytes = bytes;
          _audioName = name;
          _songTitle = title;
          _artistName = null;
        });
      },
      failureMessage: 'No se pudo seleccionar el audio.',
    );
  }

  Future<void> _pickAttachment({
    required List<String> allowedExtensions,
    required String? Function({required String name, required int sizeBytes})
    validate,
    required void Function(Uint8List bytes, String name) onSelected,
    required String failureMessage,
  }) async {
    try {
      final files = await FilePicker.pickFiles(
        type: FileType.custom,
        allowedExtensions: allowedExtensions,
      );
      if (!mounted || files.isEmpty) return;

      final file = files.first;
      final size = await file.length();
      if (!mounted) return;
      if (size == null) {
        _showMessage('No se pudo determinar el tamaño del archivo.');
        return;
      }

      final validationError = validate(name: file.name, sizeBytes: size);
      if (validationError != null) {
        _showMessage(validationError);
        return;
      }

      final bytes = await file.readAsBytes();
      if (!mounted) return;
      final byteValidationError = validate(
        name: file.name,
        sizeBytes: bytes.length,
      );
      if (byteValidationError != null) {
        _showMessage(byteValidationError);
        return;
      }
      onSelected(bytes, file.name);
    } catch (error) {
      if (mounted) _showMessage('$failureMessage $error');
    }
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _submit() async {
    if (!_hasContent || _isSubmitting) return;

    setState(() => _isSubmitting = true);
    FocusScope.of(context).unfocus();
    await Future.delayed(const Duration(seconds: 1));

    if (!mounted) return;

    final isEdit = widget.postData != null;
    final within24Hours = widget.postData?['within24Hours'] as bool? ?? true;
    if (isEdit && !within24Hours) {
      setState(() => _isSubmitting = false);
      _showMessage(
        'Error (RN-11): El plazo de 24 horas para editar ha expirado.',
      );
      return;
    }

    Navigator.pop(
      context,
      PostEditorResult(
        content: _contentController.text.trim(),
        imageBytes: _imageBytes,
        imageName: _imageName,
        imageUrl: _imageUrl,
        audioBytes: _audioBytes,
        audioName: _audioName,
        songTitle: _songTitle,
        artistName: _artistName,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isEdit = widget.postData != null;
    final dialogWidth = (MediaQuery.sizeOf(context).width - 48)
        .clamp(0, 500)
        .toDouble();

    return AlertDialog(
      backgroundColor: Colors.white,
      title: Text(
        isEdit ? 'Editar Publicación' : 'Crear nueva publicación',
        style: const TextStyle(fontWeight: FontWeight.bold),
      ),
      content: SizedBox(
        width: dialogWidth,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              TextField(
                controller: _contentController,
                maxLines: 4,
                decoration: InputDecoration(
                  hintText: '¿Qué estás escuchando o pensando?',
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                children: [
                  OutlinedButton.icon(
                    onPressed: _isSubmitting ? null : _pickImage,
                    icon: const Icon(Icons.image_outlined),
                    label: const Text('Adjuntar imagen'),
                  ),
                  OutlinedButton.icon(
                    onPressed: _isSubmitting ? null : _pickAudio,
                    icon: const Icon(Icons.audiotrack),
                    label: const Text('Seleccionar canción'),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              const Text(
                'Imagen: JPG, PNG o WEBP (máx. 5 MB). '
                'Audio: MP3, M4A, AAC, WAV u OGG (máx. 15 MB).',
                style: TextStyle(fontSize: 11, color: Colors.grey),
              ),
              if (_imageBytes != null || _imageUrl != null) ...[
                const SizedBox(height: 16),
                _buildImagePreview(),
              ],
              if (_songTitle != null) ...[
                const SizedBox(height: 12),
                _buildSongPreview(),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isSubmitting ? null : () => Navigator.pop(context),
          child: const Text('Cancelar', style: TextStyle(color: Colors.grey)),
        ),
        ElevatedButton(
          style: ElevatedButton.styleFrom(backgroundColor: widget.tealAccent),
          onPressed: _isSubmitting || !_hasContent ? null : _submit,
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
                  isEdit ? 'Guardar cambios' : 'Publicar',
                  style: const TextStyle(color: Colors.white),
                ),
        ),
      ],
    );
  }

  Widget _buildImagePreview() {
    return Stack(
      alignment: Alignment.topRight,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: _imageBytes != null
              ? Image.memory(
                  _imageBytes!,
                  height: 150,
                  width: double.infinity,
                  fit: BoxFit.cover,
                  errorBuilder: (context, error, stackTrace) =>
                      _invalidImagePreview(),
                )
              : Image.network(
                  _imageUrl!,
                  height: 150,
                  width: double.infinity,
                  fit: BoxFit.cover,
                  errorBuilder: (context, error, stackTrace) =>
                      _invalidImagePreview(),
                ),
        ),
        IconButton.filled(
          onPressed: _isSubmitting
              ? null
              : () => setState(() {
                  _imageBytes = null;
                  _imageName = null;
                  _imageUrl = null;
                }),
          icon: const Icon(Icons.close),
          tooltip: 'Quitar imagen',
        ),
      ],
    );
  }

  Widget _invalidImagePreview() {
    return Container(
      height: 100,
      color: Colors.grey.shade200,
      alignment: Alignment.center,
      child: const Text('No se pudo mostrar la imagen'),
    );
  }

  Widget _buildSongPreview() {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: widget.tealAccent.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          Icon(Icons.music_note, color: widget.tealAccent),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _songTitle!,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
                if (_audioName != null)
                  Text(
                    _audioName!,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontSize: 11, color: Colors.grey),
                  ),
              ],
            ),
          ),
          IconButton(
            onPressed: _isSubmitting
                ? null
                : () => setState(() {
                    _audioBytes = null;
                    _audioName = null;
                    _songTitle = null;
                    _artistName = null;
                  }),
            icon: const Icon(Icons.close),
            tooltip: 'Quitar canción',
          ),
        ],
      ),
    );
  }
}
