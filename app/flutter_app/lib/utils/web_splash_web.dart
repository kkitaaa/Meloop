import 'dart:js_interop';

@JS('dismissMeloopSplash')
external void _dismissMeloopSplash();

void dismissWebSplash() {
  try {
    _dismissMeloopSplash();
  } catch (_) {
    // Ignora si el splash ya fue removido o la función global no existe.
  }
}
