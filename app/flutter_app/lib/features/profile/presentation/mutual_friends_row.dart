import 'package:flutter/material.dart';

import 'mutual_friends_screen.dart';

class MutualFriendsRow extends StatelessWidget {
  final int targetUserId;
  final int mutualFriendsCount;
  final List<String> previewAvatars;

  const MutualFriendsRow({
    super.key,
    required this.targetUserId,
    required this.mutualFriendsCount,
    required this.previewAvatars,
  });

  @override
  Widget build(BuildContext context) {
    if (mutualFriendsCount == 0) return const SizedBox.shrink();

    final displayAvatars = previewAvatars.take(3).toList();

    return Semantics(
      button: true,
      label: "Ver la lista de $mutualFriendsCount amigos en común",
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => MutualFriendsScreen(
                targetUserId: targetUserId,
                initialCount: mutualFriendsCount,
              ),
            ),
          );
        },
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8.0, horizontal: 12.0),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: 24.0 + (displayAvatars.length - 1) * 16.0,
                height: 24,
                child: Stack(
                  children: List.generate(
                    displayAvatars.length,
                    (index) => Positioned(
                      left: index * 16.0,
                      child: Semantics(
                        label: "Avatar de amigo en común ${index + 1}",
                        image: true,
                        child: Container(
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            border: Border.all(color: Colors.white, width: 2),
                          ),
                          child: CircleAvatar(
                            radius: 10,
                            backgroundColor: Colors.grey[300],
                            child: const Icon(
                              Icons.person,
                              size: 12,
                              color: Colors.white,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                "$mutualFriendsCount amigos en común",
                style: const TextStyle(fontSize: 12, color: Colors.black87),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
