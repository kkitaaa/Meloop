import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../theme/meloop_typography.dart';
import '../../../widgets/meloop_toast_pill.dart';
import '../../../widgets/paper/stacked_paper_card.dart';
import '../models/profile_customization.dart';
import 'profile_stickers.dart';

enum ProfileEditorTab { profile, style, decorations, stickers, layout }

/// Estudio en vivo para personalizar el perfil y el tema global de Meloop:
/// - Permite editar datos del perfil, estilo de la app, paleta de colores,
///   fondo, marcos, pines, stickers arrastrables y distribución acotada.
/// - Respeta desbloqueos por nivel y permite inspeccionar el JSON listo para
///   sincronizar con el servidor (`PUT /users/me/profile/customization`).
class ProfileEditorPanel extends StatefulWidget {
  const ProfileEditorPanel({
    super.key,
    required this.config,
    required this.userLevel,
    required this.rgbFrameUnlocked,
    required this.usernameController,
    required this.bioController,
    required this.selectedStickerId,
    required this.onConfigChanged,
    required this.onProfileTextChanged,
    required this.onAddSticker,
    required this.onUpdateSelectedSticker,
    required this.onRemoveSelectedSticker,
    required this.onDemoLevelUp,
    required this.onResetDefault,
    required this.onClose,
    this.initialTab = ProfileEditorTab.style,
  });

  final ProfileCustomizationConfig config;
  final int userLevel;
  final bool rgbFrameUnlocked;
  final TextEditingController usernameController;
  final TextEditingController bioController;
  final String? selectedStickerId;
  final ValueChanged<ProfileCustomizationConfig> onConfigChanged;
  final VoidCallback onProfileTextChanged;
  final ValueChanged<String> onAddSticker;
  final ValueChanged<ProfileStickerPlacement> onUpdateSelectedSticker;
  final VoidCallback onRemoveSelectedSticker;
  final VoidCallback onDemoLevelUp;
  final VoidCallback onResetDefault;
  final VoidCallback onClose;
  final ProfileEditorTab initialTab;

  @override
  State<ProfileEditorPanel> createState() => _ProfileEditorPanelState();
}

class _ProfileEditorPanelState extends State<ProfileEditorPanel> {
  static const List<String> _moodPresets = [
    'Feliz',
    'Relajado',
    'Inspirado',
    'Enérgico',
    'Nostálgico',
    'Loquilla',
    'Enojada',
  ];

  late ProfileEditorTab _activeTab;
  late final TextEditingController _songTitleController;
  late final TextEditingController _songAlbumController;
  late final TextEditingController _songArtistController;

  @override
  void initState() {
    super.initState();
    _activeTab = widget.initialTab;
    _songTitleController = TextEditingController(
      text: widget.config.favoriteSongTitle,
    );
    _songAlbumController = TextEditingController(
      text: widget.config.favoriteSongAlbum,
    );
    _songArtistController = TextEditingController(
      text: widget.config.favoriteSongArtist,
    );
  }

  @override
  void dispose() {
    _songTitleController.dispose();
    _songAlbumController.dispose();
    _songArtistController.dispose();
    super.dispose();
  }

  bool _isBorderUnlocked(ProfileBorderStyle border) {
    if (border == ProfileBorderStyle.rainbowRgb && widget.rgbFrameUnlocked) {
      return true;
    }
    return widget.userLevel >= border.requiredLevel;
  }

  void _showLockedNotice(int requiredLevel) {
    showMeloopToastPill(
      context,
      message: 'Se desbloquea al alcanzar el Nivel $requiredLevel',
      icon: Icons.lock_outline_rounded,
    );
  }

  void _showServerPayloadDialog() {
    final jsonText = widget.config.toPrettyJson();
    showDialog<void>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
        title: Row(
          children: [
            Icon(
              Icons.data_object_rounded,
              color: widget.config.colorTheme.primary,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'Payload JSON para el servidor',
                style: MeloopTypography.lexend(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          ],
        ),
        content: SizedBox(
          width: 520,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Contrato sugerido: PUT /users/me/profile/customization',
                style: MeloopTypography.lexend(
                  fontSize: 12,
                  color: const Color(0xFF5D726F),
                ),
              ),
              const SizedBox(height: 10),
              Container(
                constraints: const BoxConstraints(maxHeight: 360),
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFF151C24),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: SingleChildScrollView(
                  child: SelectableText(
                    jsonText,
                    style: const TextStyle(
                      fontFamily: 'monospace',
                      fontSize: 12,
                      color: Color(0xFF9DF2DF),
                      height: 1.4,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
        actions: [
          TextButton.icon(
            onPressed: () {
              Clipboard.setData(ClipboardData(text: jsonText));
              Navigator.of(dialogContext).pop();
              showMeloopToastPill(
                context,
                message: 'JSON copiado al portapapeles',
                icon: Icons.copy_rounded,
              );
            },
            icon: const Icon(Icons.copy_rounded, size: 17),
            label: const Text('Copiar JSON'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: widget.config.colorTheme.primary,
            ),
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: const Text('Cerrar'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final accent = widget.config.colorTheme.primary;

    return PhotoPaperSheetStack(
      sheetCount: 1,
      shadowBlur: 18,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Cabecera del estudio
          Container(
            padding: const EdgeInsets.fromLTRB(16, 14, 12, 12),
            decoration: BoxDecoration(
              color: widget.config.colorTheme.softCardTint,
              border: const Border(
                bottom: BorderSide(color: Color(0xFFDCE7E5)),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(Icons.auto_awesome_rounded, color: accent, size: 20),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        'Estudio de perfil',
                        style: MeloopTypography.lexend(
                          fontSize: 16,
                          fontWeight: FontWeight.w700,
                          color: const Color(0xFF182626),
                        ),
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 9,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: accent.withValues(alpha: 0.14),
                        borderRadius: BorderRadius.circular(14),
                      ),
                      child: Text(
                        'Nivel ${widget.userLevel}',
                        style: MeloopTypography.lexend(
                          fontSize: 11.5,
                          fontWeight: FontWeight.w700,
                          color: accent,
                        ),
                      ),
                    ),
                    const SizedBox(width: 4),
                    IconButton(
                      tooltip: 'Cerrar editor',
                      onPressed: widget.onClose,
                      icon: const Icon(Icons.close_rounded, size: 20),
                      visualDensity: VisualDensity.compact,
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 6,
                  runSpacing: 6,
                  children: [
                    _MiniActionChip(
                      icon: Icons.bolt_rounded,
                      label: '+1 Nivel (Demo)',
                      accent: accent,
                      onTap: widget.onDemoLevelUp,
                    ),
                    _MiniActionChip(
                      icon: Icons.data_object_rounded,
                      label: 'Ver JSON',
                      accent: const Color(0xFF3E5E5C),
                      onTap: _showServerPayloadDialog,
                    ),
                    _MiniActionChip(
                      icon: Icons.restart_alt_rounded,
                      label: 'Base',
                      accent: const Color(0xFF6B7E7C),
                      onTap: widget.onResetDefault,
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Barra de pestañas del editor (todas visibles sin scroll horizontal)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Wrap(
              spacing: 6,
              runSpacing: 6,
              children: [
                _EditorTabChip(
                  label: 'Estilo',
                  icon: Icons.palette_outlined,
                  selected: _activeTab == ProfileEditorTab.style,
                  accent: accent,
                  onTap: () =>
                      setState(() => _activeTab = ProfileEditorTab.style),
                ),
                _EditorTabChip(
                  label: 'Decorar',
                  icon: Icons.auto_fix_high_rounded,
                  selected: _activeTab == ProfileEditorTab.decorations,
                  accent: accent,
                  onTap: () =>
                      setState(() => _activeTab = ProfileEditorTab.decorations),
                ),
                _EditorTabChip(
                  label: 'Stickers',
                  icon: Icons.emoji_emotions_outlined,
                  selected: _activeTab == ProfileEditorTab.stickers,
                  accent: accent,
                  onTap: () =>
                      setState(() => _activeTab = ProfileEditorTab.stickers),
                ),
                _EditorTabChip(
                  label: 'Diseño',
                  icon: Icons.dashboard_customize_outlined,
                  selected: _activeTab == ProfileEditorTab.layout,
                  accent: accent,
                  onTap: () =>
                      setState(() => _activeTab = ProfileEditorTab.layout),
                ),
                _EditorTabChip(
                  label: 'Perfil',
                  icon: Icons.person_outline_rounded,
                  selected: _activeTab == ProfileEditorTab.profile,
                  accent: accent,
                  onTap: () =>
                      setState(() => _activeTab = ProfileEditorTab.profile),
                ),
              ],
            ),
          ),
          const Divider(height: 1, color: Color(0xFFE3ECEA)),

          // Contenido scrollable de la pestaña activa
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.fromLTRB(16, 14, 16, 20),
              child: switch (_activeTab) {
                ProfileEditorTab.style => _buildStyleTab(accent),
                ProfileEditorTab.decorations => _buildDecorationsTab(accent),
                ProfileEditorTab.stickers => _buildStickersTab(accent),
                ProfileEditorTab.layout => _buildLayoutTab(accent),
                ProfileEditorTab.profile => _buildProfileTextTab(accent),
              },
            ),
          ),

          // Botón inferior Guardar / Listo
          Container(
            padding: const EdgeInsets.fromLTRB(16, 10, 16, 12),
            decoration: const BoxDecoration(
              color: Colors.white,
              border: Border(top: BorderSide(color: Color(0xFFE3ECEA))),
            ),
            child: FilledButton.icon(
              style: FilledButton.styleFrom(
                backgroundColor: accent,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(14),
                ),
              ),
              onPressed: () {
                widget.onProfileTextChanged();
                widget.onClose();
                showMeloopToastPill(
                  context,
                  message: 'Personalización guardada en tu perfil',
                  icon: Icons.check_circle_rounded,
                );
              },
              icon: const Icon(Icons.check_rounded, size: 18),
              label: Text(
                'Guardar y cerrar',
                style: MeloopTypography.lexend(
                  fontSize: 13.5,
                  fontWeight: FontWeight.w700,
                  color: Colors.white,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStyleTab(Color accent) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _sectionTitle('Estilo global de la app'),
        const SizedBox(height: 6),
        Text(
          'Cambia la apariencia de las tarjetas en toda la aplicación.',
          style: MeloopTypography.lexend(
            fontSize: 11.5,
            color: const Color(0xFF6B807D),
          ),
        ),
        const SizedBox(height: 10),
        for (final style in MeloopAppStyle.values) ...[
          _buildUnlockableStyleTile(style, accent),
          const SizedBox(height: 8),
        ],
        const SizedBox(height: 14),
        _sectionTitle('Color de toda la app'),
        const SizedBox(height: 10),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            for (final palette in MeloopColorTheme.values)
              _buildColorSwatch(palette),
          ],
        ),
        const SizedBox(height: 18),
        _sectionTitle('Fondo de pantalla'),
        const SizedBox(height: 10),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            for (final pattern in MeloopBackgroundPattern.values)
              _buildBackgroundChip(pattern, accent),
          ],
        ),
      ],
    );
  }

  Widget _buildUnlockableStyleTile(MeloopAppStyle style, Color accent) {
    final unlocked = widget.userLevel >= style.requiredLevel;
    final selected = widget.config.appStyle == style;
    return Material(
      color: selected ? accent.withValues(alpha: 0.10) : Colors.white,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () {
          if (!unlocked) {
            _showLockedNotice(style.requiredLevel);
            return;
          }
          widget.onConfigChanged(widget.config.copyWith(appStyle: style));
        },
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: selected ? accent : const Color(0xFFDCE7E5),
              width: selected ? 1.8 : 1,
            ),
          ),
          child: Row(
            children: [
              Icon(
                switch (style) {
                  MeloopAppStyle.paper => Icons.description_outlined,
                  MeloopAppStyle.flatSquare => Icons.crop_square_rounded,
                  MeloopAppStyle.flatSquircle => Icons.rounded_corner_rounded,
                  MeloopAppStyle.frutigerAero => Icons.bubble_chart_outlined,
                  MeloopAppStyle.neonNight => Icons.nightlight_round,
                },
                size: 20,
                color: unlocked ? accent : const Color(0xFF9FB0AD),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      style.label,
                      style: MeloopTypography.lexend(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: unlocked
                            ? const Color(0xFF182626)
                            : const Color(0xFF8C9E9B),
                      ),
                    ),
                    Text(
                      style.description,
                      style: MeloopTypography.lexend(
                        fontSize: 11,
                        color: const Color(0xFF6B807D),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 6),
              if (!unlocked)
                _lockBadge(style.requiredLevel)
              else if (selected)
                Icon(Icons.check_circle_rounded, size: 18, color: accent),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildColorSwatch(MeloopColorTheme palette) {
    final unlocked = widget.userLevel >= palette.requiredLevel;
    final selected = widget.config.colorTheme == palette;

    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () {
        if (!unlocked) {
          _showLockedNotice(palette.requiredLevel);
          return;
        }
        widget.onConfigChanged(widget.config.copyWith(colorTheme: palette));
      },
      child: Container(
        width: 106,
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: selected ? palette.primary : const Color(0xFFDCE7E5),
            width: selected ? 2 : 1,
          ),
        ),
        child: Column(
          children: [
            Container(
              height: 24,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [palette.headerTop, palette.primary],
                ),
                borderRadius: BorderRadius.circular(8),
              ),
              alignment: Alignment.center,
              child: !unlocked
                  ? const Icon(
                      Icons.lock_rounded,
                      size: 13,
                      color: Colors.white,
                    )
                  : selected
                  ? const Icon(
                      Icons.check_rounded,
                      size: 15,
                      color: Colors.white,
                    )
                  : null,
            ),
            const SizedBox(height: 5),
            Text(
              palette.label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: MeloopTypography.lexend(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: unlocked
                    ? const Color(0xFF182626)
                    : const Color(0xFF8D9F9C),
              ),
            ),
            Text(
              unlocked ? 'Desbloqueado' : 'Nv. ${palette.requiredLevel}',
              style: MeloopTypography.lexend(
                fontSize: 9.5,
                color: unlocked ? palette.primary : const Color(0xFF8D9F9C),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBackgroundChip(MeloopBackgroundPattern pattern, Color accent) {
    final unlocked = widget.userLevel >= pattern.requiredLevel;
    final selected = widget.config.backgroundPattern == pattern;
    return ChoiceChip(
      label: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (!unlocked) ...[
            const Icon(Icons.lock_rounded, size: 12, color: Color(0xFF7B8E8B)),
            const SizedBox(width: 4),
          ],
          Text(
            unlocked
                ? pattern.label
                : '${pattern.label} (Nv.${pattern.requiredLevel})',
            style: MeloopTypography.lexend(
              fontSize: 11.5,
              fontWeight: FontWeight.w600,
              color: selected
                  ? Colors.white
                  : unlocked
                  ? const Color(0xFF182626)
                  : const Color(0xFF8C9E9B),
            ),
          ),
        ],
      ),
      selected: selected,
      selectedColor: accent,
      backgroundColor: Colors.white,
      showCheckmark: false,
      side: BorderSide(color: selected ? accent : const Color(0xFFD5E3E0)),
      onSelected: (_) {
        if (!unlocked) {
          _showLockedNotice(pattern.requiredLevel);
          return;
        }
        widget.onConfigChanged(
          widget.config.copyWith(backgroundPattern: pattern),
        );
      },
    );
  }

  Widget _buildDecorationsTab(Color accent) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _sectionTitle('Marco de foto de perfil'),
        const SizedBox(height: 10),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            for (final border in ProfileBorderStyle.values)
              _buildBorderCard(border, accent),
          ],
        ),
        const SizedBox(height: 18),
        _sectionTitle('Pin personalizado'),
        const SizedBox(height: 6),
        Text(
          'Tu pin se muestra en tu Polaroid y en paneles fijados de la app.',
          style: MeloopTypography.lexend(
            fontSize: 11.5,
            color: const Color(0xFF6B807D),
          ),
        ),
        const SizedBox(height: 10),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            for (final pin in kProfilePinCatalog) _buildPinChip(pin, accent),
          ],
        ),
      ],
    );
  }

  Widget _buildBorderCard(ProfileBorderStyle border, Color accent) {
    final unlocked = _isBorderUnlocked(border);
    final selected = widget.config.borderStyle == border;

    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () {
        if (!unlocked) {
          _showLockedNotice(border.requiredLevel);
          return;
        }
        widget.onConfigChanged(widget.config.copyWith(borderStyle: border));
      },
      child: Container(
        width: 106,
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: selected ? accent.withValues(alpha: 0.08) : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: selected ? accent : const Color(0xFFDCE7E5),
            width: selected ? 2 : 1,
          ),
        ),
        child: Column(
          children: [
            Container(
              width: 34,
              height: 34,
              decoration: BoxDecoration(
                color: const Color(0xFFF3F7F6),
                borderRadius: BorderRadius.circular(6),
                border: Border.all(color: border.previewColor, width: 3),
              ),
              child: unlocked
                  ? Icon(
                      selected
                          ? Icons.check_rounded
                          : Icons.crop_portrait_rounded,
                      size: 16,
                      color: selected ? accent : const Color(0xFF5D726F),
                    )
                  : const Icon(
                      Icons.lock_rounded,
                      size: 15,
                      color: Color(0xFF8C9E9B),
                    ),
            ),
            const SizedBox(height: 6),
            Text(
              border.label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: MeloopTypography.lexend(
                fontSize: 10.5,
                fontWeight: FontWeight.w600,
                color: unlocked
                    ? const Color(0xFF182626)
                    : const Color(0xFF8D9F9C),
              ),
            ),
            if (!unlocked)
              Text(
                'Nv. ${border.requiredLevel}',
                style: MeloopTypography.lexend(
                  fontSize: 9.5,
                  color: const Color(0xFF8D9F9C),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildPinChip(ProfilePinOption pin, Color accent) {
    final unlocked = widget.userLevel >= pin.requiredLevel;
    final selected = widget.config.pinId == pin.id;
    return ChoiceChip(
      avatar: unlocked
          ? CustomizablePin(color: pin.color, size: 22)
          : const Icon(Icons.lock_rounded, size: 13, color: Color(0xFF8C9E9B)),
      label: Text(
        unlocked ? pin.label : '${pin.label} (Nv.${pin.requiredLevel})',
        style: MeloopTypography.lexend(
          fontSize: 11.5,
          fontWeight: FontWeight.w600,
          color: selected ? Colors.white : const Color(0xFF182626),
        ),
      ),
      selected: selected,
      selectedColor: accent,
      backgroundColor: Colors.white,
      showCheckmark: false,
      side: BorderSide(color: selected ? accent : const Color(0xFFDCE7E5)),
      onSelected: (_) {
        if (!unlocked) {
          _showLockedNotice(pin.requiredLevel);
          return;
        }
        widget.onConfigChanged(widget.config.copyWith(pinId: pin.id));
      },
    );
  }

  Widget _buildStickersTab(Color accent) {
    final maxStickers = ProfileCustomizationConfig.maxStickersForLevel(
      widget.userLevel,
    );
    final placedCount = widget.config.stickers.length;
    ProfileStickerPlacement? selectedPlacement;
    for (final s in widget.config.stickers) {
      if (s.id == widget.selectedStickerId) {
        selectedPlacement = s;
        break;
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: _sectionTitle('Stickers ($placedCount/$maxStickers)'),
            ),
            Text(
              'Arrástralos sobre el perfil',
              style: MeloopTypography.lexend(
                fontSize: 11,
                color: accent,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Text(
          'Cupos por nivel: Nv.1 → 1 · Nv.5 → 2 · Nv.10 → 3 · Nv.15 → 5.',
          style: MeloopTypography.lexend(
            fontSize: 11,
            color: const Color(0xFF6B807D),
          ),
        ),
        if (selectedPlacement != null) ...[
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: widget.config.colorTheme.softCardTint,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: accent.withValues(alpha: 0.4)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Sticker seleccionado: ${stickerDefinitionById(selectedPlacement.stickerId).label}',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                          color: const Color(0xFF182626),
                        ),
                      ),
                    ),
                    IconButton(
                      tooltip: 'Quitar sticker',
                      onPressed: widget.onRemoveSelectedSticker,
                      icon: const Icon(
                        Icons.delete_outline_rounded,
                        size: 18,
                        color: Color(0xFFD64545),
                      ),
                      visualDensity: VisualDensity.compact,
                    ),
                  ],
                ),
                Row(
                  children: [
                    Text(
                      'Tamaño',
                      style: MeloopTypography.lexend(fontSize: 11),
                    ),
                    Expanded(
                      child: Slider(
                        value: selectedPlacement.scale,
                        min: 0.65,
                        max: 1.55,
                        activeColor: accent,
                        onChanged: (v) => widget.onUpdateSelectedSticker(
                          selectedPlacement!.copyWith(scale: v),
                        ),
                      ),
                    ),
                  ],
                ),
                Row(
                  children: [
                    Text(
                      'Ángulo',
                      style: MeloopTypography.lexend(fontSize: 11),
                    ),
                    Expanded(
                      child: Slider(
                        value: selectedPlacement.rotation,
                        min: -math.pi / 4,
                        max: math.pi / 4,
                        activeColor: accent,
                        onChanged: (v) => widget.onUpdateSelectedSticker(
                          selectedPlacement!.copyWith(rotation: v),
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
        const SizedBox(height: 12),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            for (final def in kProfileStickerCatalog)
              _buildStickerCatalogTile(
                def,
                accent: accent,
                canPlaceMore: placedCount < maxStickers,
              ),
          ],
        ),
      ],
    );
  }

  Widget _buildStickerCatalogTile(
    ProfileStickerDefinition def, {
    required Color accent,
    required bool canPlaceMore,
  }) {
    final unlocked = widget.userLevel >= def.requiredLevel;
    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () {
        if (!unlocked) {
          _showLockedNotice(def.requiredLevel);
          return;
        }
        if (!canPlaceMore) {
          showMeloopToastPill(
            context,
            message: 'Límite de stickers alcanzado para tu nivel',
            icon: Icons.info_outline_rounded,
          );
          return;
        }
        widget.onAddSticker(def.id);
      },
      child: Container(
        width: 106,
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 8),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xFFDCE7E5)),
        ),
        child: Column(
          children: [
            Stack(
              alignment: Alignment.center,
              children: [
                Opacity(
                  opacity: unlocked ? 1.0 : 0.42,
                  child: ProfileStickerArtwork(stickerId: def.id, size: 52),
                ),
                if (!unlocked) _lockBadge(def.requiredLevel),
              ],
            ),
            const SizedBox(height: 4),
            Text(
              def.label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: MeloopTypography.lexend(
                fontSize: 10.5,
                fontWeight: FontWeight.w600,
                color: unlocked
                    ? const Color(0xFF182626)
                    : const Color(0xFF8C9E9B),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLayoutTab(Color accent) {
    final layout = widget.config.layout;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _sectionTitle('Distribución segura del perfil'),
        const SizedBox(height: 6),
        Text(
          'Elige un formato equilibrado y decide qué módulos mostrar u ocultar.',
          style: MeloopTypography.lexend(
            fontSize: 11.5,
            color: const Color(0xFF6B807D),
          ),
        ),
        const SizedBox(height: 10),
        for (final preset in ProfileLayoutPreset.values) ...[
          Material(
            color: layout.preset == preset
                ? accent.withValues(alpha: 0.10)
                : Colors.white,
            borderRadius: BorderRadius.circular(12),
            child: InkWell(
              borderRadius: BorderRadius.circular(12),
              onTap: () => widget.onConfigChanged(
                widget.config.copyWith(layout: layout.copyWith(preset: preset)),
              ),
              child: Container(
                padding: const EdgeInsets.all(11),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: layout.preset == preset
                        ? accent
                        : const Color(0xFFDCE7E5),
                    width: layout.preset == preset ? 1.8 : 1,
                  ),
                ),
                child: Row(
                  children: [
                    Icon(
                      layout.preset == preset
                          ? Icons.radio_button_checked_rounded
                          : Icons.radio_button_unchecked_rounded,
                      color: layout.preset == preset
                          ? accent
                          : const Color(0xFF8D9F9C),
                      size: 19,
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            preset.label,
                            style: MeloopTypography.lexend(
                              fontSize: 13,
                              fontWeight: FontWeight.w700,
                              color: const Color(0xFF182626),
                            ),
                          ),
                          Text(
                            preset.description,
                            style: MeloopTypography.lexend(
                              fontSize: 11,
                              color: const Color(0xFF6B807D),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 8),
        ],
        const SizedBox(height: 12),
        _sectionTitle('Secciones visibles'),
        const SizedBox(height: 6),
        _layoutSwitch(
          title: 'Mostrar lista de amigos',
          subtitle: 'Si la ocultas, liberas espacio para stickers.',
          value: layout.showFriends,
          accent: accent,
          onChanged: (v) => widget.onConfigChanged(
            widget.config.copyWith(layout: layout.copyWith(showFriends: v)),
          ),
        ),
        _layoutSwitch(
          title: 'Mostrar canción favorita',
          subtitle: 'Tarjeta destacada con tu tema favorito.',
          value: layout.showFavoriteSong,
          accent: accent,
          onChanged: (v) => widget.onConfigChanged(
            widget.config.copyWith(
              layout: layout.copyWith(showFavoriteSong: v),
            ),
          ),
        ),
        _layoutSwitch(
          title: 'Mostrar artistas que me gustan',
          subtitle: 'Grilla de artistas favoritos de tu perfil.',
          value: layout.showFavoriteArtists,
          accent: accent,
          onChanged: (v) => widget.onConfigChanged(
            widget.config.copyWith(
              layout: layout.copyWith(showFavoriteArtists: v),
            ),
          ),
        ),
        _layoutSwitch(
          title: 'Mostrar estado «Me siento»',
          subtitle: 'Debajo de tu nombre en la polaroid.',
          value: layout.showMood,
          accent: accent,
          onChanged: (v) => widget.onConfigChanged(
            widget.config.copyWith(layout: layout.copyWith(showMood: v)),
          ),
        ),
        _layoutSwitch(
          title: 'Artistas arriba de «Sobre mí»',
          subtitle: 'Intercambia el orden del bloque central.',
          value: layout.artistsBeforeBio,
          accent: accent,
          onChanged: (v) => widget.onConfigChanged(
            widget.config.copyWith(
              layout: layout.copyWith(artistsBeforeBio: v),
            ),
          ),
        ),
      ],
    );
  }

  Widget _layoutSwitch({
    required String title,
    required String subtitle,
    required bool value,
    required Color accent,
    required ValueChanged<bool> onChanged,
  }) {
    return Material(
      type: MaterialType.transparency,
      child: SwitchListTile.adaptive(
        dense: true,
        contentPadding: EdgeInsets.zero,
        activeTrackColor: accent,
        title: Text(
          title,
          style: MeloopTypography.lexend(
            fontSize: 12.5,
            fontWeight: FontWeight.w600,
            color: const Color(0xFF182626),
          ),
        ),
        subtitle: Text(
          subtitle,
          style: MeloopTypography.lexend(
            fontSize: 10.5,
            color: const Color(0xFF6B807D),
          ),
        ),
        value: value,
        onChanged: onChanged,
      ),
    );
  }

  Widget _buildProfileTextTab(Color accent) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _sectionTitle('Información pública'),
        const SizedBox(height: 10),
        TextField(
          controller: widget.usernameController,
          onChanged: (_) => widget.onProfileTextChanged(),
          decoration: InputDecoration(
            labelText: 'Nombre de usuario',
            isDense: true,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
        ),
        const SizedBox(height: 12),
        TextField(
          controller: widget.bioController,
          maxLines: 4,
          onChanged: (_) => widget.onProfileTextChanged(),
          decoration: InputDecoration(
            labelText: 'Sobre mí (biografía)',
            isDense: true,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
        ),
        const SizedBox(height: 16),
        _sectionTitle('Estado «Me siento»'),
        const SizedBox(height: 8),
        Wrap(
          spacing: 6,
          runSpacing: 6,
          children: [
            for (final mood in _moodPresets)
              ChoiceChip(
                label: Text(
                  mood,
                  style: MeloopTypography.lexend(
                    fontSize: 11.5,
                    fontWeight: FontWeight.w600,
                    color: widget.config.mood == mood
                        ? Colors.white
                        : const Color(0xFF182626),
                  ),
                ),
                selected: widget.config.mood == mood,
                selectedColor: accent,
                backgroundColor: Colors.white,
                showCheckmark: false,
                onSelected: (_) =>
                    widget.onConfigChanged(widget.config.copyWith(mood: mood)),
              ),
          ],
        ),
        const SizedBox(height: 16),
        _sectionTitle('Canción favorita'),
        const SizedBox(height: 10),
        TextField(
          controller: _songTitleController,
          onChanged: (v) => widget.onConfigChanged(
            widget.config.copyWith(favoriteSongTitle: v),
          ),
          decoration: InputDecoration(
            labelText: 'Título de la canción',
            isDense: true,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: _songAlbumController,
                onChanged: (v) => widget.onConfigChanged(
                  widget.config.copyWith(favoriteSongAlbum: v),
                ),
                decoration: InputDecoration(
                  labelText: 'Álbum',
                  isDense: true,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: TextField(
                controller: _songArtistController,
                onChanged: (v) => widget.onConfigChanged(
                  widget.config.copyWith(favoriteSongArtist: v),
                ),
                decoration: InputDecoration(
                  labelText: 'Artista',
                  isDense: true,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _sectionTitle(String text) {
    return Text(
      text,
      style: MeloopTypography.lexend(
        fontSize: 13.5,
        fontWeight: FontWeight.w700,
        color: const Color(0xFF182626),
      ),
    );
  }

  Widget _lockBadge(int requiredLevel) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
      decoration: BoxDecoration(
        color: const Color(0xFF202B2B).withValues(alpha: 0.82),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.lock_rounded, size: 11, color: Colors.white),
          const SizedBox(width: 3),
          Text(
            'Nv.$requiredLevel',
            style: MeloopTypography.lexend(
              fontSize: 10,
              fontWeight: FontWeight.w700,
              color: Colors.white,
            ),
          ),
        ],
      ),
    );
  }
}

class _EditorTabChip extends StatelessWidget {
  const _EditorTabChip({
    required this.label,
    required this.icon,
    required this.selected,
    required this.accent,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool selected;
  final Color accent;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 6),
      child: InkWell(
        borderRadius: BorderRadius.circular(18),
        onTap: onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 7),
          decoration: BoxDecoration(
            color: selected ? accent : const Color(0xFFF1F6F5),
            borderRadius: BorderRadius.circular(18),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                icon,
                size: 15,
                color: selected ? Colors.white : const Color(0xFF3E5553),
              ),
              const SizedBox(width: 5),
              Text(
                label,
                style: MeloopTypography.lexend(
                  fontSize: 11.5,
                  fontWeight: FontWeight.w600,
                  color: selected ? Colors.white : const Color(0xFF263838),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _MiniActionChip extends StatelessWidget {
  const _MiniActionChip({
    required this.icon,
    required this.label,
    required this.accent,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final Color accent;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(14),
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: accent.withValues(alpha: 0.35)),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 13.5, color: accent),
            const SizedBox(width: 4),
            Text(
              label,
              style: MeloopTypography.lexend(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: accent,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
