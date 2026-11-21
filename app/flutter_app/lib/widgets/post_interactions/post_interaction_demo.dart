import 'package:flutter/material.dart';

import '../../features/gamification/presentation/gamification_demo_controller.dart';

/// Demo boundary for post likes.
///
/// Vicente: keep this demo boundary and its interaction visuals. When the post
/// API is ready, replace `setLike` with the request and retain the optimistic
/// state, button animation, XP effect, and level-up presentation.
class PostInteractionDemoService {
  const PostInteractionDemoService();

  Future<void> setLike({required String postId, required bool liked}) async {
    await Future<void>.delayed(const Duration(milliseconds: 180));
  }
}

class DemoLikeButton extends StatefulWidget {
  const DemoLikeButton({
    super.key,
    required this.postId,
    required this.likes,
    required this.initiallyLiked,
    this.onLikesChanged,
    this.onLikedChanged,
    this.service = const PostInteractionDemoService(),
  });

  final String postId;
  final int likes;
  final bool initiallyLiked;
  final ValueChanged<int>? onLikesChanged;
  final ValueChanged<bool>? onLikedChanged;
  final PostInteractionDemoService service;

  @override
  State<DemoLikeButton> createState() => _DemoLikeButtonState();
}

class _DemoLikeButtonState extends State<DemoLikeButton>
    with SingleTickerProviderStateMixin {
  static const _teal = Color(0xFF00AF92);
  static const _like = Color(0xFFFF6262);

  late final AnimationController _animationController;
  late bool _liked;
  late int _likes;
  bool _isUpdating = false;
  bool _showXp = false;

  @override
  void initState() {
    super.initState();
    _liked = widget.initiallyLiked;
    _likes = widget.likes;
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 520),
    );
  }

  @override
  void didUpdateWidget(covariant DemoLikeButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.postId != widget.postId) {
      _liked = widget.initiallyLiked;
      _likes = widget.likes;
    } else if (!_isUpdating && oldWidget.likes != widget.likes) {
      _liked = widget.initiallyLiked;
      _likes = widget.likes;
    }
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  Future<void> _toggleLike() async {
    if (_isUpdating) return;
    final previousLiked = _liked;
    final previousLikes = _likes;
    final liked = !previousLiked;
    setState(() {
      _liked = liked;
      _likes = (previousLikes + (liked ? 1 : -1)).clamp(0, 0x7fffffff);
      _isUpdating = true;
      _showXp = liked;
    });
    widget.onLikesChanged?.call(_likes);
    widget.onLikedChanged?.call(liked);
    if (liked) _animationController.forward(from: 0);

    try {
      await widget.service.setLike(postId: widget.postId, liked: liked);
      if (!mounted) return;
      setState(() => _isUpdating = false);
      if (liked) {
        // The global overlay consumes the pending reward and presents the
        // original full-screen confetti banner without blocking the feed.
        GamificationDemoController.instance.grantLikeXp();
      }
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _liked = previousLiked;
        _likes = previousLikes;
        _isUpdating = false;
        _showXp = false;
      });
      widget.onLikesChanged?.call(previousLikes);
      widget.onLikedChanged?.call(previousLiked);
      ScaffoldMessenger.maybeOf(context)?.showSnackBar(
        SnackBar(content: Text('No se pudo actualizar el like: $error')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Tooltip(
          message: _liked ? 'Quitar like' : 'Dar like',
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: _isUpdating ? null : _toggleLike,
              borderRadius: BorderRadius.circular(6),
              child: SizedBox(
                width: 34,
                height: 36,
                child: Stack(
                  clipBehavior: Clip.none,
                  alignment: Alignment.center,
                  children: [
                    AnimatedScale(
                      scale: _liked ? 1.08 : 1,
                      duration: const Duration(milliseconds: 180),
                      curve: Curves.easeOutBack,
                      child: Icon(
                        _liked ? Icons.favorite : Icons.favorite_border,
                        color: _liked ? _like : _teal,
                        size: 23,
                      ),
                    ),
                    if (_liked)
                      IgnorePointer(
                        child: AnimatedBuilder(
                          animation: _animationController,
                          builder: (context, child) => Opacity(
                            opacity: (1 - _animationController.value)
                                .clamp(0, 1)
                                .toDouble(),
                            child: Transform.scale(
                              scale: 1 + _animationController.value * .7,
                              child: const Icon(
                                Icons.favorite,
                                color: _like,
                                size: 23,
                              ),
                            ),
                          ),
                        ),
                      ),
                    if (_showXp)
                      IgnorePointer(
                        child: AnimatedBuilder(
                          animation: _animationController,
                          builder: (context, child) => Opacity(
                            opacity: (1 - _animationController.value)
                                .clamp(0, 1)
                                .toDouble(),
                            child: Transform.translate(
                              offset: Offset(
                                0,
                                -24 * _animationController.value,
                              ),
                              child: const SizedBox(
                                width: 70,
                                child: Text(
                                  '+10 XP',
                                  textAlign: TextAlign.center,
                                  style: TextStyle(
                                    color: _teal,
                                    fontWeight: FontWeight.bold,
                                    fontSize: 11,
                                  ),
                                ),
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
        ),
        const SizedBox(width: 4),
        Text(
          '$_likes Likes',
          style: const TextStyle(color: _teal, fontSize: 11),
        ),
      ],
    );
  }
}
