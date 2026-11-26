import 'package:flutter/material.dart';

import '../../../theme/meloop_typography.dart';
import '../../../widgets/paper/stacked_paper_card.dart';
import '../state/app_notification.dart';
import '../state/notification_controller.dart';

class NotificationFeedbackOverlay extends StatefulWidget {
  const NotificationFeedbackOverlay({
    required this.controller,
    required this.child,
    super.key,
  });

  final NotificationController controller;
  final Widget child;

  @override
  State<NotificationFeedbackOverlay> createState() =>
      _NotificationFeedbackOverlayState();
}

class _NotificationFeedbackOverlayState
    extends State<NotificationFeedbackOverlay>
    with SingleTickerProviderStateMixin {
  late final AnimationController _animController;
  late final Animation<Offset> _slide;
  late final Animation<double> _opacity;

  AppNotification? _displayedNotification;
  int _displayedUnreadCount = 1;
  bool _isExiting = false;
  bool _swipedDismissed = false;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 340),
      reverseDuration: const Duration(milliseconds: 260),
    );
    final curved = CurvedAnimation(
      parent: _animController,
      curve: Curves.easeOutCubic,
      reverseCurve: Curves.easeInCubic,
    );
    _slide = Tween<Offset>(
      begin: const Offset(0, -0.55),
      end: Offset.zero,
    ).animate(curved);
    _opacity = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(
        parent: _animController,
        curve: const Interval(0.0, 0.75, curve: Curves.easeOut),
        reverseCurve: Curves.easeIn,
      ),
    );
    _animController.addStatusListener(_handleAnimationStatus);

    widget.controller.addListener(_onControllerChanged);
    _syncFromController(initial: true);
  }

  @override
  void didUpdateWidget(covariant NotificationFeedbackOverlay oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.controller != widget.controller) {
      oldWidget.controller.removeListener(_onControllerChanged);
      widget.controller.addListener(_onControllerChanged);
      _syncFromController();
    }
  }

  @override
  void dispose() {
    widget.controller.removeListener(_onControllerChanged);
    _animController.removeStatusListener(_handleAnimationStatus);
    _animController.dispose();
    super.dispose();
  }

  void _handleAnimationStatus(AnimationStatus status) {
    if (status == AnimationStatus.dismissed && _isExiting && mounted) {
      setState(() {
        _displayedNotification = null;
        _isExiting = false;
      });
    }
  }

  void _onControllerChanged() {
    if (!mounted) return;
    _syncFromController();
  }

  void _syncFromController({bool initial = false}) {
    final incoming = widget.controller.activeBanner;
    final unread = widget.controller.unreadCount;

    if (incoming != null) {
      final isNew = _displayedNotification?.id != incoming.id || _isExiting;
      _displayedNotification = incoming;
      _displayedUnreadCount = unread;
      _isExiting = false;
      _swipedDismissed = false;
      if (isNew) {
        _animController.forward(from: 0.0);
      }
      if (!initial) setState(() {});
      return;
    }

    // Si el usuario deslizó manualmente hacia arriba con Dismissible, ya salió.
    if (_swipedDismissed) {
      _swipedDismissed = false;
      _displayedNotification = null;
      _isExiting = false;
      _animController.reset();
      if (!initial) setState(() {});
      return;
    }

    // Al expirar el temporizador o presionar la X: revertimos suavemente la
    // animación de entrada (deslizar hacia arriba + fade out).
    if (_displayedNotification != null && !_isExiting) {
      _isExiting = true;
      if (_animController.value <= 0.01) {
        _displayedNotification = null;
        _isExiting = false;
      } else {
        _animController.reverse();
      }
      if (!initial) setState(() {});
    }
  }

  void _handleSwipeDismiss() {
    _swipedDismissed = true;
    widget.controller.dismissBanner();
  }

  @override
  Widget build(BuildContext context) {
    final notification = _displayedNotification;
    final isActive = widget.controller.activeBanner != null && !_isExiting;
    final mediaQuery = MediaQuery.of(context);
    final isDesktop = mediaQuery.size.width > 900;
    final top = mediaQuery.viewPadding.top + (isDesktop ? 98.0 : 56.0);
    final sheetCount = _displayedUnreadCount.clamp(1, 3);

    Widget? bannerContent;
    if (notification != null) {
      final visualBanner = Stack(
        clipBehavior: Clip.none,
        children: [
          SlideTransition(
            position: _slide,
            child: FadeTransition(
              opacity: _opacity,
              child: Dismissible(
                key: ValueKey('dismiss-${notification.id}'),
                direction: DismissDirection.up,
                onDismissed: (_) => _handleSwipeDismiss(),
                child: PhotoPaperSheetStack(
                  key: ValueKey<String>(
                    'notification-paper-sheets-$sheetCount',
                  ),
                  sheetCount: sheetCount,
                  shadowBlur: 16,
                  padding: const EdgeInsets.fromLTRB(16, 13, 44, 13),
                  child: Row(
                    children: [
                      CircleAvatar(
                        backgroundColor: const Color(
                          0xFF00AF92,
                        ).withValues(alpha: .14),
                        child: Icon(
                          switch (notification.kind) {
                            NotificationKind.like => Icons.favorite,
                            NotificationKind.comment => Icons.mode_comment,
                            NotificationKind.friendRequest =>
                              Icons.person_add_alt_1,
                            NotificationKind.levelUp => Icons.trending_up,
                            NotificationKind.reward => Icons.card_giftcard,
                          },
                          color: switch (notification.kind) {
                            NotificationKind.like => const Color(0xFFFF6262),
                            NotificationKind.comment => const Color(0xFF00AF92),
                            NotificationKind.friendRequest => const Color(
                              0xFF288B83,
                            ),
                            NotificationKind.levelUp => const Color(0xFFFFB020),
                            NotificationKind.reward => const Color(0xFF00AF92),
                          },
                          size: 20,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              notification.title,
                              style: MeloopTypography.lexend(
                                fontWeight: FontWeight.w700,
                                fontSize: 13.5,
                                color: const Color(0xFF182222),
                              ),
                            ),
                            if (notification.detail.isNotEmpty) ...[
                              const SizedBox(height: 3),
                              Text(
                                notification.detail,
                                style: MeloopTypography.lexend(
                                  color: const Color(0xFF5D6E6D),
                                  fontSize: 12,
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),
                      if (_displayedUnreadCount > 1)
                        Padding(
                          padding: const EdgeInsets.only(left: 8),
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 7,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: const Color(0xFFE0F6F2),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Text(
                              '$_displayedUnreadCount sin leer',
                              style: MeloopTypography.lexend(
                                fontSize: 10.5,
                                fontWeight: FontWeight.w700,
                                color: const Color(0xFF11857A),
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
          Positioned(
            top: 2,
            right: 2,
            child: FadeTransition(
              opacity: _opacity,
              child: Semantics(
                label: 'Cerrar notificación',
                button: true,
                child: IconButton(
                  key: const ValueKey('notification-banner-close'),
                  visualDensity: VisualDensity.compact,
                  constraints: const BoxConstraints(
                    minWidth: 32,
                    minHeight: 32,
                  ),
                  padding: EdgeInsets.zero,
                  onPressed: isActive ? widget.controller.dismissBanner : null,
                  icon: const Icon(
                    Icons.close,
                    color: Color(0xFF7B8B8B),
                    size: 18,
                  ),
                ),
              ),
            ),
          ),
        ],
      );

      bannerContent = Material(
        color: Colors.transparent,
        child: isActive
            ? Card(
                elevation: 0,
                margin: EdgeInsets.zero,
                color: Colors.transparent,
                surfaceTintColor: Colors.transparent,
                shadowColor: Colors.transparent,
                clipBehavior: Clip.none,
                child: visualBanner,
              )
            : IgnorePointer(child: visualBanner),
      );
    }

    return Stack(
      children: [
        widget.child,
        if (bannerContent != null)
          Positioned(
            top: top,
            left: 16,
            right: 16,
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 520),
                child: bannerContent,
              ),
            ),
          ),
      ],
    );
  }
}
