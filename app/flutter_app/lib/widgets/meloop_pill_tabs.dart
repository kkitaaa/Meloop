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
class MeloopPillTabs<T> extends StatelessWidget {
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
  Widget build(BuildContext context) {
    final buttonHeight = height - padding * 2;
    final double? sharedWidth = fixedButtonWidth ??
        (equalWidths && items.isNotEmpty
            ? measureUniformButtonWidth(
                context: context,
                labels: items.map((e) => e.label).toList(),
                fontSize: fontSize,
                horizontalItemPadding: horizontalItemPadding,
              )
            : null);

    return Container(
      height: height,
      padding: EdgeInsets.all(padding),
      decoration: BoxDecoration(
        color: pillColor,
        borderRadius: BorderRadius.circular(height / 2 + 8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: items.map((item) {
          final isSelected = item.value == selectedValue;
          final inner = AnimatedContainer(
            duration: const Duration(milliseconds: 180),
            curve: Curves.easeOutCubic,
            width: sharedWidth,
            height: buttonHeight,
            padding: sharedWidth == null
                ? EdgeInsets.symmetric(horizontal: horizontalItemPadding)
                : EdgeInsets.zero,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: isSelected ? Colors.white : Colors.transparent,
              gradient: isSelected
                  ? const LinearGradient(
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                      colors: [Color(0xFFFFFFFF), Color(0xFFE6E6E6)],
                    )
                  : null,
              borderRadius: BorderRadius.circular(buttonHeight / 2 + 4),
            ),
            child: Text(
              item.label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: MeloopTypography.lexend(
                color: isSelected ? const Color(0xFF141B1B) : Colors.white,
                fontSize: fontSize,
                fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
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
                onTap: disabled ? null : () => onSelected(item.value),
                borderRadius: BorderRadius.circular(buttonHeight / 2 + 4),
                child: inner,
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}
