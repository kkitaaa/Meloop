import 'package:flutter/material.dart';

import 'post_interaction_api.dart';

class PostInteractionBar extends StatefulWidget {
  const PostInteractionBar({
    super.key,
    required this.postId,
    required this.likes,
    required this.comments,
    this.initiallyLiked = false,
    this.initiallySaved = false,
    this.onCommentsTap,
    this.api,
  });

  final Object postId;
  final int likes;
  final int comments;
  final bool initiallyLiked;
  final bool initiallySaved;
  final VoidCallback? onCommentsTap;
  final PostInteractionApi? api;

  @override
  State<PostInteractionBar> createState() => _PostInteractionBarState();
}

class _PostInteractionBarState extends State<PostInteractionBar> {
  static const _teal = Color(0xFF00AF92);
  static const _like = Color(0xFFFF6262);

  late bool _liked;
  late bool _saved;
  late int _likes;
  late final PostInteractionApi _api;

  @override
  void initState() {
    super.initState();
    _liked = widget.initiallyLiked;
    _saved = widget.initiallySaved;
    _likes = widget.likes;
    _api = widget.api ?? PostInteractionApi();
  }

  @override
  void didUpdateWidget(covariant PostInteractionBar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.postId != widget.postId) {
      _liked = widget.initiallyLiked;
      _saved = widget.initiallySaved;
      _likes = widget.likes;
    }
  }

  Future<void> _toggleLike() async {
    final previous = _liked;
    setState(() {
      _liked = !previous;
      _likes += _liked ? 1 : -1;
    });
    try {
      await _api.setLike(postId: widget.postId.toString(), liked: _liked);
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _liked = previous;
        _likes += previous ? 1 : -1;
      });
      _showError('No se pudo actualizar el like. Inténtalo otra vez.');
    }
  }

  Future<void> _toggleBookmark() async {
    final previous = _saved;
    setState(() => _saved = !previous);
    try {
      await _api.setBookmark(postId: widget.postId.toString(), saved: _saved);
    } catch (_) {
      if (!mounted) return;
      setState(() => _saved = previous);
      _showError('No se pudo guardar la publicación. Inténtalo otra vez.');
    }
  }

  void _showError(String message) {
    ScaffoldMessenger.maybeOf(context)?.showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        _AnimatedInteractionButton(
          active: _liked,
          activeColor: _like,
          inactiveColor: _teal,
          inactiveIcon: Icons.favorite_border,
          activeIcon: Icons.favorite,
          tooltip: _liked ? 'Quitar like' : 'Dar like',
          onTap: _toggleLike,
        ),
        const SizedBox(width: 7),
        Text(
          '$_likes Likes',
          style: const TextStyle(color: _teal, fontSize: 11),
        ),
        const SizedBox(width: 22),
        InkWell(
          onTap: widget.onCommentsTap,
          borderRadius: BorderRadius.circular(4),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Row(
              children: [
                const Icon(
                  Icons.mode_comment_outlined,
                  color: _teal,
                  size: 20,
                ),
                const SizedBox(width: 4),
                Text(
                  '${widget.comments} Comentarios',
                  style: const TextStyle(color: _teal, fontSize: 11),
                ),
              ],
            ),
          ),
        ),
        const Spacer(),
        _AnimatedInteractionButton(
          active: _saved,
          activeColor: _teal,
          inactiveColor: _teal,
          inactiveIcon: Icons.bookmark_border,
          activeIcon: Icons.bookmark,
          tooltip: _saved ? 'Quitar de guardados' : 'Guardar publicación',
          onTap: _toggleBookmark,
        ),
      ],
    );
  }
}

class _AnimatedInteractionButton extends StatefulWidget {
  const _AnimatedInteractionButton({
    required this.active,
    required this.activeColor,
    required this.inactiveColor,
    required this.inactiveIcon,
    required this.activeIcon,
    required this.tooltip,
    required this.onTap,
  });

  final bool active;
  final Color activeColor;
  final Color inactiveColor;
  final IconData inactiveIcon;
  final IconData activeIcon;
  final String tooltip;
  final VoidCallback onTap;

  @override
  State<_AnimatedInteractionButton> createState() => _AnimatedInteractionButtonState();
}

class _AnimatedInteractionButtonState extends State<_AnimatedInteractionButton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: const Duration(milliseconds: 420));
  }

  @override
  void didUpdateWidget(covariant _AnimatedInteractionButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!oldWidget.active && widget.active) _controller.forward(from: 0);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final color = widget.active ? widget.activeColor : widget.inactiveColor;
    final icon = widget.active ? widget.activeIcon : widget.inactiveIcon;
    return Tooltip(
      message: widget.tooltip,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: widget.onTap,
          borderRadius: BorderRadius.circular(4),
          child: SizedBox(
            width: 28,
            height: 36,
            child: Stack(
              alignment: Alignment.center,
              children: [
                Icon(
                  icon,
                  color: color,
                  size: 25,
                ),
                if (widget.active)
                  IgnorePointer(
                    child: AnimatedBuilder(
                      animation: _controller,
                      builder: (context, child) => Opacity(
                        opacity: (1 - _controller.value).clamp(0, 1).toDouble(),
                        child: Transform.scale(
                          scale: 1 + (_controller.value * .65),
                          child: Icon(
                            widget.activeIcon,
                            color: widget.activeColor,
                            size: 25,
                          ),
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
