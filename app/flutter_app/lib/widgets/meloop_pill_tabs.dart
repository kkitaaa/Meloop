import 'package:flutter/material.dart';

import '../theme/meloop_typography.dart';

class MeloopPillItem<T> {
  const MeloopPillItem({required this.value, required this.label});

  final T value;
  final String label;
}

/// Barra segmentada estilo píldora reutilizada tanto en la navbar global
/// (`Inicio`, `Amigos`, `Artistas y canciones`) como en los filtros de la
/// sección de blogs (`Recientes`, `Tendencia`, `Amigos`).
///
/// Por defecto todos los botones tienen el mismo ancho, calculado a partir de
/// la etiqueta más larga más margen suficiente.
class MeloopPillTabs<T> extends StatefulWidget {
  const MeloopPillTabs({
    super.key,
    required this.items,
    required this.selectedValue,
    required this.onSelected,
    this.height = 48,
    this.padding = 4.0,
    this.fontSize = 14,
    this.horizontalItemPadding = 22,
    this.fixedButtonWidth,
    this.pillColor = const Color(0xFF16877F),
    this.equalWidths = true,
    this.disabled = false,
  });

  final List<MeloopPillItem<T>> items;
  final T? selectedValue;
  final ValueChanged<T> onSelected;
  final double height;
  final double padding;
  final double fontSize;
  final double horizontalItemPadding;
  final double? fixedButtonWidth;
  final Color pillColor;
  final bool equalWidths;
  final bool disabled;

  /// Calcula el ancho total exacto de la píldora cuando `equalWidths == true`,
  /// útil para centrar y posicionar elementos adyacentes (como la lupa).
  static double measureTotalWidth({
    required BuildContext context,
    required List<String> labels,
    required double fontSize,
    required double horizontalItemPadding,
    required double padding,
  }) {
    final buttonWidth = measureUniformButtonWidth(
      context: context,
      labels: labels,
      fontSize: fontSize,
      horizontalItemPadding: horizontalItemPadding,
    );
    return buttonWidth * labels.length + padding * 2;
  }

  static double measureUniformButtonWidth({
    required BuildContext context,
    required List<String> labels,
    required double fontSize,
    required double horizontalItemPadding,
  }) {
    final style = MeloopTypography.lexend(
      fontSize: fontSize,
      fontWeight: FontWeight.w600,
    );
    final maxTextWidth = labels
        .map(
          (label) => (TextPainter(
            text: TextSpan(text: label, style: style),
            textDirection: Directionality.of(context),
            textScaler: MediaQuery.textScalerOf(context),
          )..layout()).width,
        )
        .reduce((a, b) => a > b ? a : b);
    return (maxTextWidth + horizontalItemPadding * 2).ceilToDouble();
  }

  @override
  State<MeloopPillTabs<T>> createState() => _MeloopPillTabsState<T>();
}

class _MeloopPillTabsState<T> extends State<MeloopPillTabs<T>> {
  int _lastValidIndex = 0;

  @override
  Widget build(BuildContext context) {
    final buttonHeight = widget.height - widget.padding * 2;
    final double? sharedWidth =
        widget.fixedButtonWidth ??
        (widget.equalWidths && widget.items.isNotEmpty
            ? MeloopPillTabs.measureUniformButtonWidth(
                context: context,
                labels: widget.items.map((e) => e.label).toList(),
                fontSize: widget.fontSize,
                horizontalItemPadding: widget.horizontalItemPadding,
              )
            : null);

    final selectedIndex = widget.selectedValue == null
        ? -1
        : widget.items.indexWhere((item) => item.value == widget.selectedValue);
    if (selectedIndex >= 0) {
      _lastValidIndex = selectedIndex;
    }
    final clampedIndex = _lastValidIndex.clamp(
      0,
      widget.items.isEmpty ? 0 : widget.items.length - 1,
    );
    final useSlidingIndicator = sharedWidth != null;

    return Container(
      height: widget.height,
      padding: EdgeInsets.all(widget.padding),
      decoration: BoxDecoration(
        color: widget.pillColor,
        borderRadius: BorderRadius.circular(widget.height / 2 + 8),
      ),
      child: Stack(
        children: [
          if (useSlidingIndicator)
            AnimatedPositioned(
              duration: const Duration(milliseconds: 240),
              curve: Curves.easeOutCubic,
              left: clampedIndex * sharedWidth,
              top: 0,
              bottom: 0,
              width: sharedWidth,
              child: IgnorePointer(
                child: AnimatedOpacity(
                  duration: const Duration(milliseconds: 160),
                  opacity: selectedIndex >= 0 ? 1.0 : 0.0,
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                        colors: [Color(0xFFFFFFFF), Color(0xFFE6E6E6)],
                      ),
                      borderRadius: BorderRadius.circular(buttonHeight / 2 + 4),
                    ),
                  ),
                ),
              ),
            ),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: widget.items.map((item) {
              final isSelected = item.value == widget.selectedValue;
              final inner = AnimatedContainer(
                duration: const Duration(milliseconds: 180),
                curve: Curves.easeOutCubic,
                width: sharedWidth,
                height: buttonHeight,
                padding: sharedWidth == null
                    ? EdgeInsets.symmetric(
                        horizontal: widget.horizontalItemPadding,
                      )
                    : EdgeInsets.zero,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: !useSlidingIndicator && isSelected
                      ? Colors.white
                      : Colors.transparent,
                  gradient: !useSlidingIndicator && isSelected
                      ? const LinearGradient(
                          begin: Alignment.topCenter,
                          end: Alignment.bottomCenter,
                          colors: [Color(0xFFFFFFFF), Color(0xFFE6E6E6)],
                        )
                      : null,
                  borderRadius: BorderRadius.circular(buttonHeight / 2 + 4),
                ),
                child: AnimatedDefaultTextStyle(
                  duration: const Duration(milliseconds: 180),
                  curve: Curves.easeOutCubic,
                  style: MeloopTypography.lexend(
                    color: isSelected ? const Color(0xFF141B1B) : Colors.white,
                    fontSize: widget.fontSize,
                    fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
                  ),
                  child: Text(
                    item.label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              );

              return Semantics(
                button: true,
                selected: isSelected,
                label: item.label,
                child: Material(
                  color: Colors.transparent,
                  child: InkWell(
                    onTap: widget.disabled
                        ? null
                        : () => widget.onSelected(item.value),
                    borderRadius: BorderRadius.circular(buttonHeight / 2 + 4),
                    child: inner,
                  ),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }
}
