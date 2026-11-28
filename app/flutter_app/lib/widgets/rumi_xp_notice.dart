import 'package:flutter/material.dart';

class RumiXpNotice extends StatelessWidget {
  final int? xpAwarded;

  const RumiXpNotice({super.key, this.xpAwarded});

  @override
  Widget build(BuildContext context) {
    final message = xpAwarded == null
        ? 'Tu publicación ya está en el feed.'
        : '¡Publicaste! Ganaste $xpAwarded XP.';

    return Row(
      children: [
        const CircleAvatar(
          radius: 16,
          backgroundColor: Color(0x331ABC9C),
          child: Icon(Icons.auto_awesome, color: Color(0xFF1ABC9C), size: 18),
        ),
        const SizedBox(width: 10),
        Expanded(child: Text('Rumi: $message')),
      ],
    );
  }
}
