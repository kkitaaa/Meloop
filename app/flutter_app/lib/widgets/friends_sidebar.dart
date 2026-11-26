import 'package:flutter/material.dart';

import '../theme/meloop_typography.dart';
import 'paper/stacked_paper_card.dart';

/// Modelo de vista (solo frontend) para un amigo conectado en la barra lateral.
///
/// TODO(backend): Cuando el backend implemente presencia en tiempo real y el
/// listado de amigos conectados (p. ej. vía social-service + WebSocketManager),
/// mapear la respuesta JSON a este modelo con los campos sugeridos:
/// ```json
/// {
///   "user_id": "uuid",
///   "username": "NombreUsuario",
///   "avatar_url": "https://...",
///   "joined_label": "Se unió hace 20 minutos.",
///   "friends_count": 12,
///   "blogs_count": 4,
///   "is_online": true
/// }
/// ```
class OnlineFriendPreview {
  const OnlineFriendPreview({
    required this.username,
    required this.joinedLabel,
    required this.friendsCount,
    required this.blogsCount,
    this.userId,
    this.isOnline = true,
  });

  final String? userId;
  final String username;
  final String joinedLabel;
  final int friendsCount;
  final int blogsCount;
  final bool isOnline;
}

/// Modelo de vista (solo frontend) para perfiles sugeridos según afinidad con
/// las recomendaciones de Spotify del usuario autenticado.
///
/// TODO(backend): Reemplazar los datos estáticos con un endpoint de sugerencias
/// basadas en Spotify (p. ej. `GET /v1/social/recommendations?source=spotify`
/// o equivalente en el API Gateway) que devuelva perfiles ordenados por
/// compatibilidad musical:
/// ```json
/// {
///   "user_id": "uuid",
///   "username": "NombreUsuario",
///   "avatar_url": "https://...",
///   "joined_label": "Se unió hace 20 minutos.",
///   "friends_count": 0,
///   "blogs_count": 0,
///   "affinity_score": 0.92
/// }
/// ```
class RecommendedProfilePreview {
  const RecommendedProfilePreview({
    required this.username,
    required this.joinedLabel,
    required this.friendsCount,
    required this.blogsCount,
    this.userId,
  });

  final String? userId;
  final String username;
  final String joinedLabel;
  final int friendsCount;
  final int blogsCount;
}

/// Barra lateral derecha reutilizable para escritorio con dos tarjetas de papel
/// apilado (`assets/textures/paper.png`):
/// 1. **Amigos conectados** (parte superior, papel con tinte menta suave).
/// 2. **Perfiles que te podrían interesar** (parte inferior, basado en
///    recomendaciones de Spotify, papel blanco).
///
/// TODO(backend): Para conectar esta barra al backend sin romper el diseño:
/// - Inyectar `onlineFriends` y `recommendedProfiles` desde un controlador o
///   `FutureBuilder`/`StreamBuilder` en la pantalla contenedora (o pasar un
///   `ApiClient` opcional a este widget).
/// - Mantener un estado de carga discreto dentro de cada tarjeta y un estado
///   vacío amigable si el usuario aún no tiene amigos conectados o Spotify
///   vinculado.
class DesktopFriendsSidebar extends StatelessWidget {
  const DesktopFriendsSidebar({
    super.key,
    this.onlineFriends = defaultOnlineFriends,
    this.recommendedProfiles = defaultRecommendedProfiles,
    this.onSeeMoreOnlineTap,
    this.onSeeMoreRecommendedTap,
    this.onProfileTap,
  });

  final List<OnlineFriendPreview> onlineFriends;
  final List<RecommendedProfilePreview> recommendedProfiles;
  final VoidCallback? onSeeMoreOnlineTap;
  final VoidCallback? onSeeMoreRecommendedTap;
  final void Function(String username, String? userId)? onProfileTap;

  // TODO(backend): Datos placeholder temporales de solo frontend. Eliminar o
  // usar solo como fallback cuando existan los endpoints reales de presencia y
  // recomendaciones de Spotify.
  static const List<OnlineFriendPreview> defaultOnlineFriends = [
    OnlineFriendPreview(
      username: 'Usuario N°1',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
    OnlineFriendPreview(
      username: 'Usuario N°2',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
    OnlineFriendPreview(
      username: 'Usuario N°3',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
    OnlineFriendPreview(
      username: 'Usuario N°4',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
  ];

  static const List<RecommendedProfilePreview> defaultRecommendedProfiles = [
    RecommendedProfilePreview(
      username: 'Usuario N°1',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
    RecommendedProfilePreview(
      username: 'Usuario N°2',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
    RecommendedProfilePreview(
      username: 'Usuario N°3',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
    RecommendedProfilePreview(
      username: 'Usuario N°4',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
    RecommendedProfilePreview(
      username: 'Usuario N°5',
      joinedLabel: 'Se unió hace 20 minutos.',
      friendsCount: 0,
      blogsCount: 0,
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        _buildOnlineFriendsCard(),
        const SizedBox(height: 22),
        _buildRecommendedProfilesCard(),
      ],
    );
  }

  Widget _buildOnlineFriendsCard() {
    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      tintColor: const Color(0xFFD9F7F0),
      padding: const EdgeInsets.fromLTRB(16, 18, 16, 16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Amigos conectados',
            textAlign: TextAlign.center,
            style: MeloopTypography.lexend(
              fontSize: 16.5,
              fontWeight: FontWeight.w600,
              color: const Color(0xFF141A1A),
            ),
          ),
          const SizedBox(height: 10),
          Divider(
            height: 1,
            thickness: 1,
            color: Colors.black.withValues(alpha: 0.1),
          ),
          const SizedBox(height: 12),
          for (final friend in onlineFriends)
            _ProfilePreviewTile(
              username: friend.username,
              joinedLabel: friend.joinedLabel,
              statsLabel:
                  '${friend.friendsCount} amigos · ${friend.blogsCount} blogs',
              showOnlineIndicator: friend.isOnline,
              onTap: onProfileTap == null
                  ? null
                  : () => onProfileTap!(friend.username, friend.userId),
            ),
          const SizedBox(height: 4),
          _SidebarActionButton(label: 'Ver más', onPressed: onSeeMoreOnlineTap),
        ],
      ),
    );
  }

  Widget _buildRecommendedProfilesCard() {
    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      padding: const EdgeInsets.fromLTRB(16, 18, 16, 16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Perfiles que te podrían interesar',
            textAlign: TextAlign.center,
            style: MeloopTypography.lexend(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: const Color(0xFF141A1A),
            ),
          ),
          const SizedBox(height: 10),
          Divider(
            height: 1,
            thickness: 1,
            color: Colors.black.withValues(alpha: 0.1),
          ),
          const SizedBox(height: 12),
          for (final profile in recommendedProfiles)
            _ProfilePreviewTile(
              username: profile.username,
              joinedLabel: profile.joinedLabel,
              statsLabel:
                  '${profile.friendsCount} amigos · ${profile.blogsCount} blogs',
              showOnlineIndicator: false,
              onTap: onProfileTap == null
                  ? null
                  : () => onProfileTap!(profile.username, profile.userId),
            ),
          const SizedBox(height: 2),
          Text(
            'En base a tus recomendaciones de Spotify',
            textAlign: TextAlign.center,
            style: MeloopTypography.lexend(
              fontSize: 11.5,
              color: const Color(0xFF6B7777),
            ),
          ),
          const SizedBox(height: 10),
          _SidebarActionButton(
            label: 'Ver más',
            onPressed: onSeeMoreRecommendedTap,
          ),
        ],
      ),
    );
  }
}

class _ProfilePreviewTile extends StatelessWidget {
  const _ProfilePreviewTile({
    required this.username,
    required this.joinedLabel,
    required this.statsLabel,
    required this.showOnlineIndicator,
    this.onTap,
  });

  final String username;
  final String joinedLabel;
  final String statsLabel;
  final bool showOnlineIndicator;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.zero,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.06),
            blurRadius: 6,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
            child: Row(
              children: [
                // Avatar placeholder con marco blanco cuadrado estilo foto
                Container(
                  width: 46,
                  height: 46,
                  padding: const EdgeInsets.all(3),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.14),
                        blurRadius: 4,
                        offset: const Offset(0, 1),
                      ),
                    ],
                  ),
                  child: const ColoredBox(
                    color: Color(0xFFDFE4E4),
                    child: Icon(
                      Icons.person,
                      size: 28,
                      color: Color(0xFF9EA7A7),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        username,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: 13.5,
                          fontWeight: FontWeight.w600,
                          color: const Color(0xFF1A2020),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        joinedLabel,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: 11,
                          color: const Color(0xFFA0AAAA),
                        ),
                      ),
                      Text(
                        statsLabel,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: 11,
                          color: const Color(0xFFA0AAAA),
                        ),
                      ),
                    ],
                  ),
                ),
                if (showOnlineIndicator) ...[
                  const SizedBox(width: 8),
                  Container(
                    width: 11,
                    height: 11,
                    decoration: const BoxDecoration(
                      color: Color(0xFF16A34A),
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 4),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _SidebarActionButton extends StatelessWidget {
  const _SidebarActionButton({required this.label, this.onPressed});

  final String label;
  final VoidCallback? onPressed;

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      style: ElevatedButton.styleFrom(
        backgroundColor: const Color(0xFF00AE92),
        foregroundColor: Colors.white,
        elevation: 0,
        shadowColor: Colors.transparent,
        padding: const EdgeInsets.symmetric(horizontal: 34, vertical: 10),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      ),
      onPressed: onPressed,
      child: Text(
        label,
        style: MeloopTypography.lexend(
          fontSize: 13.5,
          fontWeight: FontWeight.w600,
          color: Colors.white,
        ),
      ),
    );
  }
}
