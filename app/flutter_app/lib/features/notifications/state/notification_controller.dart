import 'dart:async';

import 'package:flutter/widgets.dart';
import 'package:just_audio/just_audio.dart';

import 'app_notification.dart';
import 'notification_repository.dart';

class NotificationController extends ChangeNotifier
    with WidgetsBindingObserver {
  NotificationController({
    required this.repository,
    this.pollInterval = const Duration(minutes: 1),
    this.bannerDuration = const Duration(seconds: 5),
    AudioPlayer? audioPlayer,
    this.playSound = true,
  }) : _ownsAudioPlayer = audioPlayer == null && playSound,
       _audioPlayer = audioPlayer ?? (playSound ? AudioPlayer() : null);

  final NotificationRepository repository;
  final AudioPlayer? _audioPlayer;
  final bool _ownsAudioPlayer;
  final bool playSound;
  final Map<String, bool> _preferences = {
    'like': true,
    'comment': true,
    'friend': true,
  };
  final List<AppNotification> _notifications = [];

  Duration pollInterval;
  Duration bannerDuration;
  Timer? _pollTimer;
  Timer? _bannerTimer;
  bool _started = false;
  bool _isPolling = false;
  bool _appIsResumed = false;
  bool _disposed = false;
  bool _audioAssetLoaded = false;
  String? _lastError;
  AppNotification? _activeBanner;
  int _notificationPulse = 0;

  List<AppNotification> get notifications => List.unmodifiable(_notifications);
  AppNotification? get activeBanner => _activeBanner;
  int get notificationPulse => _notificationPulse;
  String? get lastError => _lastError;

  int get unreadCount => _notifications
      .where(
        (notification) =>
            !notification.isRead &&
            _preferences[AppNotification.preferenceKey(notification.kind)]!,
      )
      .length;

  bool preference(String key) => _preferences[key] ?? true;

  void start() {
    if (_started) return;
    _started = true;
    WidgetsBinding.instance.addObserver(this);
    _onLifecycleChanged(WidgetsBinding.instance.lifecycleState);
  }

  Future<void> setPreference(String key, bool enabled) async {
    if (_preferences[key] == enabled) return;
    try {
      await repository.updatePreference(type: key, enabled: enabled);
    } catch (error, stackTrace) {
      _lastError = error.toString();
      debugPrint(
        'Error al guardar preferencia de notificaciones: $error\n$stackTrace',
      );
      notifyListeners();
      return;
    }
    _preferences[key] = enabled;
    _lastError = null;
    notifyListeners();
  }

  void markRead(String id) {
    final index = _notifications.indexWhere((item) => item.id == id);
    if (index == -1 || _notifications[index].isRead) return;
    _notifications[index] = _notifications[index].copyWith(isRead: true);
    notifyListeners();
  }

  void dismissBanner() {
    _bannerTimer?.cancel();
    _bannerTimer = null;
    if (_activeBanner == null) return;
    _activeBanner = null;
    notifyListeners();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    _onLifecycleChanged(state);
  }

  void _onLifecycleChanged(AppLifecycleState? state) {
    final resumed = state == AppLifecycleState.resumed;
    if (resumed == _appIsResumed) return;
    _appIsResumed = resumed;

    _pollTimer?.cancel();
    _pollTimer = null;
    if (!resumed) return;

    unawaited(_poll());
    _pollTimer = Timer.periodic(pollInterval, (_) => unawaited(_poll()));
  }

  Future<void> _poll() async {
    if (!_appIsResumed || _isPolling) return;
    _isPolling = true;
    try {
      final notifications = await repository.fetchNotifications(
        since: _notifications.isEmpty ? null : _notifications.first.createdAt,
      );
      if (_disposed || !_appIsResumed) return;
      _lastError = null;
      final existingIds = _notifications.map((item) => item.id).toSet();
      final incoming = notifications
          .where((item) => !existingIds.contains(item.id))
          .toList();
      if (incoming.isEmpty) return;

      _notifications.insertAll(0, incoming);
      _notifications.sort((a, b) => b.createdAt.compareTo(a.createdAt));
      _showBanner(incoming.first);
      _notificationPulse++;
      notifyListeners();
      if (playSound) unawaited(_playNotificationSound());
    } catch (error, stackTrace) {
      if (_disposed) return;
      _lastError = error.toString();
      debugPrint('Error al consultar notificaciones: $error\n$stackTrace');
      notifyListeners();
    } finally {
      _isPolling = false;
    }
  }

  void _showBanner(AppNotification notification) {
    _bannerTimer?.cancel();
    _activeBanner = notification;
    _bannerTimer = Timer(bannerDuration, dismissBanner);
  }

  Future<void> _playNotificationSound() async {
    try {
      final audioPlayer = _audioPlayer;
      if (audioPlayer == null) return;
      if (!_audioAssetLoaded) {
        await audioPlayer.setAsset('assets/audio/chan.wav');
        _audioAssetLoaded = true;
      }
      await audioPlayer.seek(Duration.zero);
      await audioPlayer.play();
    } catch (error, stackTrace) {
      if (_disposed) return;
      _lastError = 'No se pudo reproducir el sonido de notificación: $error';
      debugPrint('Error de audio de notificación: $error\n$stackTrace');
      notifyListeners();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    WidgetsBinding.instance.removeObserver(this);
    _pollTimer?.cancel();
    _bannerTimer?.cancel();
    if (_ownsAudioPlayer) unawaited(_audioPlayer?.dispose());
    repository.dispose();
    super.dispose();
  }
}
