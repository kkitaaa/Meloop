import 'package:flutter/widgets.dart';

import '../state/notification_controller.dart';

class NotificationScope extends InheritedNotifier<NotificationController> {
  const NotificationScope({
    required NotificationController controller,
    required super.child,
    super.key,
  }) : super(notifier: controller);

  static NotificationController of(BuildContext context) {
    final scope = context
        .dependOnInheritedWidgetOfExactType<NotificationScope>();
    assert(scope != null, 'NotificationScope is missing above this context.');
    return scope!.notifier!;
  }
}
