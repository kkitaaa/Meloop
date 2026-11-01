import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/post_publisher.dart';

import 'music_search_screen.dart';

const int maxPostImageSizeBytes = 5 * 1024 * 1024;
const int maxPostAudioSizeBytes = 15 * 1024 * 1024;
const int maxPostImages = 4;
const int maxPostContentLength = 2000;

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
  final List<PostImageUpload> images;
  final Uint8List? imageBytes;
  final String? imageName;
  final String? imageUrl;
  final Uint8List? audioBytes;
  final String? audioName;
  final String? songTitle;
  final String? artistName;
  final Map<String, dynamic>? publishedPost;
  final int? xpAwarded;

  const PostEditorResult({
    required this.content,
    this.images = const [],
    this.imageBytes,
    this.imageName,
    this.imageUrl,
    this.audioBytes,
    this.audioName,
    this.songTitle,
    this.artistName,
    this.publishedPost,
    this.xpAwarded,
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

  static String? validateImageCount(int currentCount, int selectedCount) {
    if (selectedCount < 1) return 'Selecciona al menos una imagen.';
    if (currentCount + selectedCount > maxPostImages) {
      return 'Puedes adjuntar hasta $maxPostImages imágenes.';
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
  final ApiClient? apiClient;
  final PostPublisher? publisher;

  const PostEditorDialog({
    super.key,
    this.postData,
    required this.tealAccent,
    this.apiClient,
    this.publisher,
  });

  @override
  State<PostEditorDialog> createState() => _PostEditorDialogState();
}

class _PostEditorDialogState extends State<PostEditorDialog> {
  final TextEditingController _contentController = TextEditingController();
  final List<PostImageUpload> _images = [];
  String? _imageUrl;
  Uint8List? _audioBytes;
  String? _audioName;
  MusicTrack? _selectedSong;
  String? _legacySongTitle;
  String? _legacyArtistName;
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  late final PostPublisher _publisher;
  bool _isSubmitting = false;
  double? _uploadProgress;
  CancelToken? _uploadCancelToken;
  String? _submissionError;

  bool get _hasContent =>
      _contentController.text.trim().isNotEmpty ||
      _images.isNotEmpty ||
      _imageUrl != null ||
      _audioBytes != null ||
      _selectedSong != null ||
      _legacySongTitle != null;

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;
    _publisher = widget.publisher ?? _apiClient;
    final post = widget.postData;
    if (post != null) {
      _contentController.text = post['content'] as String? ?? '';
      if (post['image_bytes'] is Uint8List) {
        _images.add(
          PostImageUpload(
            bytes: post['image_bytes'] as Uint8List,
            name: post['image_name'] as String? ?? 'imagen.jpg',
          ),
        );
      }
      _imageUrl = post['image_url'] as String?;
      _audioBytes = post['audio_bytes'] is Uint8List
          ? post['audio_bytes'] as Uint8List
          : null;
      _audioName = post['audio_name'] as String?;
      _legacySongTitle = post['song'] as String?;
      _legacyArtistName = post['artist'] as String?;
    }
    _contentController.addListener(_onContentChanged);
  }

  @override
  void dispose() {
    _contentController
      ..removeListener(_onContentChanged)
      ..dispose();
    _uploadCancelToken?.cancel('El formulario se cerró.');
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  void _onContentChanged() {
    if (mounted) setState(() {});
  }

  Future<void> _pickImage() async {
    try {
      final selection = await FilePicker.pickFiles(
        type: FileType.custom,
        allowedExtensions: supportedPostImageExtensions,
      );
      if (!mounted || selection.isEmpty) return;

      final countError = PostAttachmentValidator.validateImageCount(
        _images.length,
        selection.length,
      );
      if (countError != null) {
        _showMessage(countError);
        return;
      }

      final selectedImages = <PostImageUpload>[];
      for (final file in selection) {
        final fileSize = await file.length();
        if (fileSize == null) {
          _showMessage('No se pudo determinar el tamaño de ${file.name}.');
          return;
        }
        var validationError = PostAttachmentValidator.validateImage(
          name: file.name,
          sizeBytes: fileSize,
        );
        if (validationError != null) {
          _showMessage('${file.name}: $validationError');
          return;
        }
        final bytes = await file.readAsBytes();
        validationError = PostAttachmentValidator.validateImage(
          name: file.name,
          sizeBytes: bytes.length,
        );
        if (validationError != null) {
          _showMessage('${file.name}: $validationError');
          return;
        }
        selectedImages.add(PostImageUpload(bytes: bytes, name: file.name));
      }
      if (!mounted) return;
      setState(() {
        _images.addAll(selectedImages);
        _imageUrl = null;
      });
    } catch (error) {
      if (mounted) {
        _showMessage('No se pudieron seleccionar las imágenes: $error');
      }
    }
  }

  Future<void> _pickAudio() async {
    await _pickAttachment(
      allowedExtensions: supportedPostAudioExtensions,
      validate: PostAttachmentValidator.validateAudio,
      onSelected: (bytes, name) {
        setState(() {
          _audioBytes = bytes;
          _audioName = name;
        });
      },
      failureMessage: 'No se pudo seleccionar el audio.',
    );
  }

  Future<void> _chooseSong() async {
    final song = await Navigator.of(context).push<MusicTrack>(
      MaterialPageRoute(
        builder: (_) =>
            MusicSearchScreen(pickForPost: true, apiClient: _apiClient),
      ),
    );
    if (!mounted || song == null) return;
    setState(() {
      _selectedSong = song;
      _legacySongTitle = null;
      _legacyArtistName = null;
    });
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

    FocusScope.of(context).unfocus();
    final isEdit = widget.postData != null;
    final within24Hours = widget.postData?['within24Hours'] as bool? ?? true;
    if (isEdit && !within24Hours) {
      _showMessage(
        'Error (RN-11): El plazo de 24 horas para editar ha expirado.',
      );
      return;
    }

    if (isEdit) {
      Navigator.pop(context, _currentResult());
      return;
    }

    final cancelToken = CancelToken();
    setState(() {
      _isSubmitting = true;
      _uploadProgress = 0;
      _uploadCancelToken = cancelToken;
      _submissionError = null;
    });

    PostPublishResult? published;
    try {
      published = await _publisher.publishPost(
        content: _contentController.text.trim(),
        images: List.unmodifiable(_images),
        audioBytes: _audioBytes,
        audioName: _audioName,
        song: _selectedSong,
        cancelToken: cancelToken,
        onSendProgress: (sent, total) {
          if (!mounted || total <= 0) return;
          setState(() => _uploadProgress = (sent / total).clamp(0, 1));
        },
      );
    } on ApiException catch (error) {
      if (mounted) setState(() => _submissionError = error.message);
    } catch (error) {
      if (mounted) {
        setState(() => _submissionError = 'No se pudo publicar: $error');
      }
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
          _uploadProgress = null;
          _uploadCancelToken = null;
        });
      }
    }

    if (!mounted || published == null) return;
    Navigator.pop(
      context,
      _currentResult(
        publishedPost: published.post,
        xpAwarded: published.xpAwarded,
      ),
    );
  }

  PostEditorResult _currentResult({
    Map<String, dynamic>? publishedPost,
    int? xpAwarded,
  }) {
    final firstImage = _images.isEmpty ? null : _images.first;
    return PostEditorResult(
      content: _contentController.text.trim(),
      images: List.unmodifiable(_images),
      imageBytes: firstImage?.bytes,
      imageName: firstImage?.name,
      imageUrl: _imageUrl,
      audioBytes: _audioBytes,
      audioName: _audioName,
      songTitle: _selectedSong?.title ?? _legacySongTitle,
      artistName: _selectedSong?.artist ?? _legacyArtistName,
      publishedPost: publishedPost,
      xpAwarded: xpAwarded,
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
                maxLength: maxPostContentLength,
                maxLengthEnforcement: MaxLengthEnforcement.enforced,
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
                    onPressed: _isSubmitting || _images.length >= maxPostImages
                        ? null
                        : _pickImage,
                    icon: const Icon(Icons.image_outlined),
                    label: Text('Imágenes (${_images.length}/$maxPostImages)'),
                  ),
                  OutlinedButton.icon(
                    onPressed: _isSubmitting || _audioBytes != null
                        ? null
                        : _pickAudio,
                    icon: const Icon(Icons.audiotrack),
                    label: const Text('Adjuntar audio'),
                  ),
                  OutlinedButton.icon(
                    onPressed: _isSubmitting ? null : _chooseSong,
                    icon: const Icon(Icons.search),
                    label: const Text('Elegir canción'),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              const Text(
                'Hasta 4 imágenes JPG, PNG o WEBP (5 MB c/u). '
                'Audio: MP3, M4A, AAC, WAV u OGG (máx. 15 MB).',
                style: TextStyle(fontSize: 11, color: Colors.grey),
              ),
              if (_images.isNotEmpty || _imageUrl != null) ...[
                const SizedBox(height: 16),
                _buildImagePreview(),
              ],
              if (_audioName != null) ...[
                const SizedBox(height: 12),
                _buildAudioPreview(),
              ],
              if (_selectedSong != null || _legacySongTitle != null) ...[
                const SizedBox(height: 12),
                _buildSongPreview(),
              ],
              if (_isSubmitting) ...[
                const SizedBox(height: 16),
                LinearProgressIndicator(value: _uploadProgress),
                const SizedBox(height: 6),
                Text(
                  _uploadProgress == null
                      ? 'Subiendo publicación…'
                      : 'Subiendo ${(100 * _uploadProgress!).round()}%',
                  style: const TextStyle(fontSize: 12, color: Colors.black54),
                ),
              ],
              if (_submissionError != null) ...[
                const SizedBox(height: 12),
                Text(
                  _submissionError!,
                  style: TextStyle(
                    color: Theme.of(context).colorScheme.error,
                    fontSize: 13,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isSubmitting
              ? () => _uploadCancelToken?.cancel()
              : () => Navigator.pop(context),
          child: Text(
            _isSubmitting ? 'Cancelar subida' : 'Cancelar',
            style: const TextStyle(color: Colors.grey),
          ),
        ),
        ElevatedButton(
          style: ElevatedButton.styleFrom(backgroundColor: widget.tealAccent),
          onPressed:
              _isSubmitting ||
                  !_hasContent ||
                  _contentController.text.length > maxPostContentLength
              ? null
              : _submit,
          child: _isSubmitting
              ? const Text('Subiendo…', style: TextStyle(color: Colors.white))
              : Text(
                  isEdit ? 'Guardar cambios' : 'Publicar',
                  style: const TextStyle(color: Colors.white),
                ),
        ),
      ],
    );
  }

  Widget _buildImagePreview() {
    if (_images.isEmpty && _imageUrl != null) {
      return Stack(
        alignment: Alignment.topRight,
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: Image.network(
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
                : () => setState(() => _imageUrl = null),
            icon: const Icon(Icons.close),
            tooltip: 'Quitar imagen',
          ),
        ],
      );
    }

    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: List.generate(_images.length, (index) {
        return SizedBox(
          width: 92,
          height: 92,
          child: Stack(
            alignment: Alignment.topRight,
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: Image.memory(
                  _images[index].bytes,
                  width: 92,
                  height: 92,
                  fit: BoxFit.cover,
                  errorBuilder: (context, error, stackTrace) =>
                      _invalidImagePreview(),
                ),
              ),
              IconButton.filled(
                visualDensity: VisualDensity.compact,
                onPressed: _isSubmitting
                    ? null
                    : () => setState(() => _images.removeAt(index)),
                icon: const Icon(Icons.close, size: 16),
                tooltip: 'Quitar imagen ${index + 1}',
              ),
            ],
          ),
        );
      }),
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
    final title = _selectedSong?.title ?? _legacySongTitle!;
    final artist = _selectedSong?.artist ?? _legacyArtistName;
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
                  title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
                if (artist != null)
                  Text(
                    artist,
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
                    _selectedSong = null;
                    _legacySongTitle = null;
                    _legacyArtistName = null;
                  }),
            icon: const Icon(Icons.close),
            tooltip: 'Quitar canción',
          ),
        ],
      ),
    );
  }

  Widget _buildAudioPreview() {
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 10),
      leading: const Icon(Icons.audiotrack),
      title: Text(_audioName!, maxLines: 1, overflow: TextOverflow.ellipsis),
      subtitle: const Text('Audio adjunto'),
      trailing: IconButton(
        onPressed: _isSubmitting
            ? null
            : () => setState(() {
                _audioBytes = null;
                _audioName = null;
              }),
        icon: const Icon(Icons.close),
        tooltip: 'Quitar audio',
      ),
    );
  }
}
