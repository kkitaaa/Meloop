import 'web_splash_stub.dart'
    if (dart.library.js_interop) 'web_splash_web.dart'
    as impl;

/// Oculta la pantalla de carga inicial de Web (`#meloop-splash`) una vez que
/// los recursos visuales de la primera pantalla terminaron de cargarse.
/// En plataformas no Web es una operación sin efecto.
void dismissWebSplash() => impl.dismissWebSplash();
