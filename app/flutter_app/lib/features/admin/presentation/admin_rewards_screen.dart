import 'package:flutter/material.dart';

import '../data/admin_models.dart';
import '../data/admin_repository.dart';

class AdminRewardsScreen extends StatefulWidget {
  const AdminRewardsScreen({super.key, required this.repository});

  final AdminRepository repository;

  @override
  State<AdminRewardsScreen> createState() => _AdminRewardsScreenState();
}

class _AdminRewardsScreenState extends State<AdminRewardsScreen> {
  List<AdminReward> _rewards = [];
  AdminRewardType? _type;
  bool? _enabled;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadRewards();
  }

  Future<void> _loadRewards() async {
    setState(() => _isLoading = true);
    try {
      final rewards = await widget.repository.fetchRewards(
        type: _type,
        enabled: _enabled,
      );
      if (!mounted) return;
      setState(() {
        _rewards = rewards;
        _isLoading = false;
      });
    } catch (error) {
      if (!mounted) return;
      setState(() => _isLoading = false);
      _showMessage('No se pudieron cargar las recompensas: $error');
    }
  }

  Future<void> _editReward([AdminReward? reward]) async {
    final saved = await showDialog<AdminReward>(
      context: context,
      builder: (context) =>
          _RewardFormDialog(reward: reward, repository: widget.repository),
    );
    if (saved == null || !mounted) return;
    await _loadRewards();
    if (!mounted) return;
    if (reward != null && reward.requiredLevel != saved.requiredLevel) {
      _showMessage(
        'Nivel requerido actualizado. Quienes ya tienen esta recompensa no la pierden.',
      );
    } else {
      _showMessage(
        reward == null ? 'Recompensa creada.' : 'Recompensa actualizada.',
      );
    }
  }

  Future<void> _toggleReward(AdminReward reward) async {
    final nextEnabled = !reward.enabled;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(
          nextEnabled ? 'Habilitar recompensa' : 'Deshabilitar recompensa',
        ),
        content: Text(
          '${nextEnabled ? 'Habilitar' : 'Deshabilitar'} "${reward.name}".\n\n'
          'Quienes ya la tienen la conservan; deja de poder desbloquearse para el resto.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(nextEnabled ? 'Habilitar' : 'Deshabilitar'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    try {
      await widget.repository.setRewardEnabled(reward.id, nextEnabled);
      if (!mounted) return;
      await _loadRewards();
      if (mounted) {
        _showMessage(
          nextEnabled ? 'Recompensa habilitada.' : 'Recompensa deshabilitada.',
        );
      }
    } on AdminRepositoryException catch (error) {
      if (mounted) {
        _showMessage('No se pudo cambiar el estado (${error.code}).');
      }
    } catch (error) {
      if (mounted) _showMessage('No se pudo cambiar el estado: $error');
    }
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) => Column(
    children: [
      Padding(
        padding: const EdgeInsets.all(16),
        child: Wrap(
          spacing: 12,
          runSpacing: 12,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            SizedBox(
              width: 200,
              child: DropdownButtonFormField<AdminRewardType?>(
                key: ValueKey(_type),
                initialValue: _type,
                isExpanded: true,
                decoration: const InputDecoration(
                  labelText: 'Tipo',
                  border: OutlineInputBorder(),
                ),
                items: [
                  const DropdownMenuItem<AdminRewardType?>(
                    child: Text('Todos', overflow: TextOverflow.ellipsis),
                  ),
                  for (final type in AdminRewardType.values)
                    DropdownMenuItem(
                      value: type,
                      child: Text(
                        type.label,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                ],
                onChanged: (value) {
                  _type = value;
                  _loadRewards();
                },
              ),
            ),
            SizedBox(
              width: 220,
              child: DropdownButtonFormField<bool?>(
                key: ValueKey(_enabled),
                initialValue: _enabled,
                isExpanded: true,
                decoration: const InputDecoration(
                  labelText: 'Disponibilidad',
                  border: OutlineInputBorder(),
                ),
                items: const [
                  DropdownMenuItem<bool?>(
                    child: Text('Todas', overflow: TextOverflow.ellipsis),
                  ),
                  DropdownMenuItem(
                    value: true,
                    child: Text('Habilitadas', overflow: TextOverflow.ellipsis),
                  ),
                  DropdownMenuItem(
                    value: false,
                    child: Text(
                      'Deshabilitadas',
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
                onChanged: (value) {
                  _enabled = value;
                  _loadRewards();
                },
              ),
            ),
            FilledButton.icon(
              onPressed: () => _editReward(),
              icon: const Icon(Icons.add),
              label: const Text('Crear recompensa'),
            ),
          ],
        ),
      ),
      const Padding(
        padding: EdgeInsets.symmetric(horizontal: 16),
        child: Align(
          alignment: Alignment.centerLeft,
          child: Text(
            'Niveles 1–10 y recompensas de ejemplo: datos temporales hasta que el backend exponga el catálogo real.',
          ),
        ),
      ),
      const SizedBox(height: 8),
      Expanded(
        child: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : _rewards.isEmpty
            ? const Center(
                child: Text('No hay recompensas para estos filtros.'),
              )
            : ListView.builder(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                itemCount: _rewards.length,
                itemBuilder: (context, index) {
                  final reward = _rewards[index];
                  return Card(
                    child: ListTile(
                      leading: _RewardThumbnail(reward: reward),
                      title: Text(reward.name),
                      subtitle: Text(
                        '${reward.type.label} · Nivel ${reward.requiredLevel}'
                        '${reward.enabled ? '' : ' · Deshabilitada'}',
                      ),
                      trailing: Wrap(
                        spacing: 0,
                        children: [
                          IconButton(
                            tooltip: 'Editar recompensa',
                            onPressed: () => _editReward(reward),
                            icon: const Icon(Icons.edit_outlined),
                          ),
                          TextButton(
                            onPressed: () => _toggleReward(reward),
                            child: Text(
                              reward.enabled ? 'Deshabilitar' : 'Habilitar',
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
      ),
    ],
  );
}

class _RewardThumbnail extends StatelessWidget {
  const _RewardThumbnail({required this.reward});

  final AdminReward reward;

  @override
  Widget build(BuildContext context) => ClipRRect(
    borderRadius: BorderRadius.circular(8),
    child: SizedBox(
      width: 54,
      height: 54,
      child: reward.imageUrl == null || reward.imageUrl!.isEmpty
          ? ColoredBox(
              color: const Color(0xFFE0F4EF),
              child: Icon(
                _iconFor(reward.type),
                color: const Color(0xFF087F6D),
              ),
            )
          : Image.network(
              reward.imageUrl!,
              fit: BoxFit.cover,
              errorBuilder: (context, error, stackTrace) => const ColoredBox(
                color: Color(0xFFE0F4EF),
                child: Icon(Icons.image_not_supported_outlined),
              ),
            ),
    ),
  );

  IconData _iconFor(AdminRewardType type) => switch (type) {
    AdminRewardType.frame => Icons.crop_square,
    AdminRewardType.background => Icons.wallpaper,
    AdminRewardType.banner => Icons.view_carousel_outlined,
    AdminRewardType.badge => Icons.verified_outlined,
    AdminRewardType.theme => Icons.palette_outlined,
    AdminRewardType.color => Icons.color_lens_outlined,
    AdminRewardType.effect => Icons.auto_awesome_outlined,
  };
}

class _RewardFormDialog extends StatefulWidget {
  const _RewardFormDialog({this.reward, required this.repository});

  final AdminReward? reward;
  final AdminRepository repository;

  @override
  State<_RewardFormDialog> createState() => _RewardFormDialogState();
}

class _RewardFormDialogState extends State<_RewardFormDialog> {
  final _formKey = GlobalKey<FormState>();
  late final _nameController = TextEditingController(
    text: widget.reward?.name ?? '',
  );
  late final _descriptionController = TextEditingController(
    text: widget.reward?.description ?? '',
  );
  late final _imageUrlController = TextEditingController(
    text: widget.reward?.imageUrl ?? '',
  );
  late AdminRewardType _type = widget.reward?.type ?? AdminRewardType.frame;
  late int _level = widget.reward?.requiredLevel ?? 1;
  bool _isSaving = false;
  bool _isPreparingUpload = false;

  @override
  void dispose() {
    _nameController.dispose();
    _descriptionController.dispose();
    _imageUrlController.dispose();
    super.dispose();
  }

  Future<void> _prepareImageUpload() async {
    final fileNameController = TextEditingController();
    final fileName = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Imagen de recompensa'),
        content: TextField(
          controller: fileNameController,
          decoration: const InputDecoration(
            labelText: 'Nombre del archivo',
            hintText: 'recompensa.png',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () =>
                Navigator.pop(context, fileNameController.text.trim()),
            child: const Text('Solicitar URL'),
          ),
        ],
      ),
    );
    fileNameController.dispose();
    if (fileName == null || fileName.isEmpty || !mounted) return;
    setState(() => _isPreparingUpload = true);
    try {
      final url = await widget.repository.createRewardImageUploadUrl(fileName);
      if (!mounted) return;
      _imageUrlController.text = url;
      setState(() => _isPreparingUpload = false);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'URL de demostración lista. La subida real se conectará al flujo prefirmado del backend.',
          ),
        ),
      );
    } catch (error) {
      if (!mounted) return;
      setState(() => _isPreparingUpload = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('No se pudo preparar la imagen: $error')),
      );
    }
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _isSaving = true);
    final draft = AdminReward(
      id: widget.reward?.id ?? 'new',
      name: _nameController.text.trim(),
      description: _descriptionController.text.trim(),
      type: _type,
      requiredLevel: _level,
      enabled: widget.reward?.enabled ?? true,
      imageUrl: _imageUrlController.text.trim().isEmpty
          ? null
          : _imageUrlController.text.trim(),
    );
    try {
      final saved = await widget.repository.saveReward(draft);
      if (mounted) Navigator.pop(context, saved);
    } on AdminRepositoryException catch (error) {
      if (!mounted) return;
      setState(() => _isSaving = false);
      final message = switch (error.code) {
        'LEVEL_NOT_FOUND' =>
          'El nivel requerido no existe. Selecciona un nivel disponible.',
        'DUPLICATE_NAME' => 'Ya existe una recompensa con ese nombre.',
        _ => 'No se pudo guardar la recompensa (${error.code}).',
      };
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(message)));
    } catch (error) {
      if (!mounted) return;
      setState(() => _isSaving = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('No se pudo guardar la recompensa: $error')),
      );
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text(
      widget.reward == null ? 'Crear recompensa' : 'Editar recompensa',
    ),
    content: SizedBox(
      width: 480,
      child: Form(
        key: _formKey,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(labelText: 'Nombre'),
                validator: (value) => value == null || value.trim().isEmpty
                    ? 'Escribe un nombre.'
                    : null,
              ),
              TextFormField(
                controller: _descriptionController,
                minLines: 2,
                maxLines: 4,
                decoration: const InputDecoration(labelText: 'Descripción'),
                validator: (value) => value == null || value.trim().isEmpty
                    ? 'Escribe una descripción.'
                    : null,
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<AdminRewardType>(
                initialValue: _type,
                isExpanded: true,
                decoration: const InputDecoration(labelText: 'Tipo'),
                items: [
                  for (final type in AdminRewardType.values)
                    DropdownMenuItem(value: type, child: Text(type.label)),
                ],
                onChanged: (value) {
                  if (value != null) setState(() => _type = value);
                },
              ),
              DropdownButtonFormField<int>(
                initialValue: _level,
                isExpanded: true,
                decoration: const InputDecoration(labelText: 'Nivel requerido'),
                items: [
                  for (var level = 1; level <= 10; level++)
                    DropdownMenuItem(value: level, child: Text('Nivel $level')),
                ],
                onChanged: (value) {
                  if (value != null) setState(() => _level = value);
                },
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _imageUrlController,
                decoration: const InputDecoration(
                  labelText: 'URL de imagen',
                  hintText: 'Opcional',
                ),
                onChanged: (_) => setState(() {}),
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _isPreparingUpload
                          ? null
                          : _prepareImageUpload,
                      icon: _isPreparingUpload
                          ? const SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.upload_outlined),
                      label: const Text('Preparar carga'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  _preview(),
                ],
              ),
            ],
          ),
        ),
      ),
    ),
    actions: [
      TextButton(
        onPressed: _isSaving ? null : () => Navigator.pop(context),
        child: const Text('Cancelar'),
      ),
      FilledButton(
        onPressed: _isSaving ? null : _save,
        child: _isSaving
            ? const SizedBox(
                width: 18,
                height: 18,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : const Text('Guardar'),
      ),
    ],
  );

  Widget _preview() {
    final url = _imageUrlController.text.trim();
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: SizedBox(
        width: 56,
        height: 56,
        child: url.isEmpty
            ? const ColoredBox(
                color: Color(0xFFE0F4EF),
                child: Icon(Icons.image_outlined),
              )
            : Image.network(
                url,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) => const ColoredBox(
                  color: Color(0xFFE0F4EF),
                  child: Icon(Icons.image_not_supported_outlined),
                ),
              ),
      ),
    );
  }
}
