import 'dart:async';
import 'dart:math' as math;
import 'dart:ui' as ui;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../core/navigation/responsive_navigation_shell.dart';
import '../features/moderation/data/moderation_models.dart';
import '../features/notifications/presentation/notification_scope.dart';
import '../services/api_client.dart';
import '../services/auth_session.dart';
import '../theme/meloop_typography.dart';
import '../utils/web_splash.dart';
import '../widgets/paper/stacked_paper_card.dart';
import '../widgets/parallax_album_covers.dart';
import 'feed_screen.dart';
import 'kubernetes_connectivity_screen.dart';

class AuthScreen extends StatefulWidget {
  final ApiClient? apiClient;
  final VoidCallback? onReadyForSplashDismiss;
  final Future<void> Function(BuildContext context, bool isDesktop)?
  preloadImages;

  const AuthScreen({
    super.key,
    this.apiClient,
    this.onReadyForSplashDismiss,
    this.preloadImages,
  });

  @override
  State<AuthScreen> createState() => _AuthScreenState();
}

class _AuthScreenState extends State<AuthScreen>
    with SingleTickerProviderStateMixin {
  final TextEditingController _emailController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _usernameController = TextEditingController();
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  late final AnimationController _bgController;

  bool _isRegisterMode = false;
  bool _isSubmitting = false;
  bool _preloadStarted = false;
  bool _assetsReady = false;
  bool _splashDismissed = false;
  String? _authError;

  static bool get _isTestBinding {
    try {
      return WidgetsBinding.instance.runtimeType.toString().contains('Test');
    } catch (_) {
      return false;
    }
  }

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _bgController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 20),
    );
    if (AuthSession.shared.isAuthenticated) {
      _restoreSession();
    }
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_preloadStarted) return;
    _preloadStarted = true;

    if (_isTestBinding &&
        widget.onReadyForSplashDismiss == null &&
        widget.preloadImages == null) {
      _assetsReady = true;
      return;
    }

    final isDesktop = MediaQuery.sizeOf(context).width > 800;
    unawaited(_preloadLoginAssets(isDesktop));
  }

  Future<void> _preloadLoginAssets(bool isDesktop) async {
    try {
      final preloadTask = widget.preloadImages != null
          ? widget.preloadImages!(context, isDesktop)
          : Future.wait<void>([
              MeloopLogoSticker.preload(
                context,
                height: isDesktop ? 76 : 56,
              ),
              if (isDesktop) ParallaxAlbumCovers.preload(context),
              if (!_isTestBinding)
                GoogleFonts.pendingFonts()
                    .timeout(const Duration(seconds: 3))
                    .then<void>((_) {}, onError: (_, _) {}),
            ]);
      await preloadTask.timeout(const Duration(seconds: 12));
    } catch (_) {
      // Si ocurre algún error o timeout, revelamos igualmente el login.
    }

    if (!mounted) return;
    setState(() => _assetsReady = true);
    if (!_isTestBinding && !_bgController.isAnimating) {
      _bgController.repeat();
    }
    // Esperamos un cuadro completo para que Flutter pinte las imágenes ya
    // decodificadas antes de desvanecer la pantalla de carga inicial.
    await WidgetsBinding.instance.endOfFrame;
    if (!mounted) return;
    _dismissSplashOnce();
  }

  void _dismissSplashOnce() {
    if (_splashDismissed) return;
    _splashDismissed = true;
    (widget.onReadyForSplashDismiss ?? dismissWebSplash)();
  }

  @override
  void dispose() {
    _bgController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _usernameController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  void _switchMode({required bool isRegister}) {
    if (_isRegisterMode == isRegister) return;
    setState(() {
      _isRegisterMode = isRegister;
      _authError = null;
    });
  }

  Future<void> _restoreSession() async {
    try {
      await _apiClient.validateSession();
      if (mounted) _openHome();
    } on ApiException catch (error) {
      if (mounted && error.statusCode != 401) {
        setState(() {
          _authError = 'No se pudo validar la sesión: ${error.message}';
        });
      }
    } on Exception catch (error) {
      if (mounted) {
        setState(() => _authError = 'No se pudo validar la sesión: $error');
      }
    }
  }

  Future<void> _login() async {
    final email = _emailController.text.trim();
    final password = _passwordController.text;
    if (email.isEmpty || password.isEmpty) {
      setState(() => _authError = 'Ingresa tu correo y contraseña.');
      return;
    }

    setState(() {
      _isSubmitting = true;
      _authError = null;
    });
    try {
      await _apiClient.login(email: email, password: password);
      if (mounted) _openHome();
    } on ApiException catch (error) {
      if (mounted) setState(() => _authError = error.message);
    } on Exception catch (error) {
      if (mounted) {
        setState(() => _authError = 'No se pudo iniciar sesión: $error');
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  Future<void> _register() async {
    final username = _usernameController.text.trim();
    final email = _emailController.text.trim();
    final password = _passwordController.text;
    if (username.isEmpty || email.isEmpty || password.isEmpty) {
      setState(() => _authError = 'Completa usuario, correo y contraseña.');
      return;
    }

    setState(() {
      _isSubmitting = true;
      _authError = null;
    });
    try {
      await _apiClient.register(
        username: username,
        email: email,
        password: password,
      );
      await _apiClient.login(email: email, password: password);
      if (mounted) _openHome();
    } on ApiException catch (error) {
      if (mounted) setState(() => _authError = error.message);
    } on Exception catch (error) {
      if (mounted) {
        setState(() => _authError = 'No se pudo crear la cuenta: $error');
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  void _openHome() {
    final context = this.context;
    NotificationScope.maybeOf(context)?.start();
    const userRole = ModerationRole.user;
    final Widget destination = defaultTargetPlatform == TargetPlatform.android
        ? const ResponsiveNavigationShell(userRole: userRole)
        : FeedScreen(userRole: userRole);
    Navigator.pushReplacement(
      context,
      MaterialPageRoute(builder: (_) => destination),
    );
    _dismissSplashOnce();
  }

  void _openDiagnostics() {
    Navigator.push<void>(
      context,
      MaterialPageRoute<void>(
        builder: (_) => const KubernetesConnectivityScreen(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.sizeOf(context).width;
    final isDesktop = screenWidth > 800;

    return Scaffold(
      resizeToAvoidBottomInset: false,
      backgroundColor: const Color(0xFF004562),
      body: Stack(
        fit: StackFit.expand,
        children: [
          // Fondo radial animado (un solo shader draw en su propio RepaintBoundary
          // para consumo mínimo de GPU).
          RepaintBoundary(
            child: _AnimatedTealRadialBackground(animation: _bgController),
          ),

          // Carátulas con paralaje concentradas en el área visible izquierda,
          // con un leve solapamiento (70px) detrás del borde izquierdo del panel
          // derecho opaco para un corte natural sin desperdiciar portadas ocultas.
          if (isDesktop)
            Positioned(
              left: 0,
              top: 0,
              bottom: 0,
              width: math.max(
                0.0,
                screenWidth - (screenWidth * 0.30).clamp(400.0, 560.0) + 70.0,
              ),
              child: RepaintBoundary(
                child: ParallaxAlbumCovers(
                  animate: !_isTestBinding && _assetsReady,
                ),
              ),
            ),

          if (isDesktop)
            _buildDesktopLayout(screenWidth)
          else
            _buildMobileLayout(screenWidth),

          // Botón de diagnóstico oculto en la esquina inferior derecha:
          // solo aparece y se vuelve clickeable tras 3 segundos de hover.
          Positioned(
            right: 16,
            bottom: 14,
            child: _HiddenDiagnosticCorner(onPressed: _openDiagnostics),
          ),
        ],
      ),
    );
  }

  Widget _buildDesktopLayout(double screenWidth) {
    final double panelWidth = (screenWidth * 0.30).clamp(400.0, 560.0);

    return Row(
      children: [
        const Expanded(child: SizedBox.shrink()),
        Container(
          width: panelWidth,
          height: double.infinity,
          decoration: BoxDecoration(
            color: const Color(0xFF3E524F),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.28),
                blurRadius: 32,
                offset: const Offset(-8, 0),
              ),
            ],
          ),
          child: SafeArea(
            child: Center(
              child: SingleChildScrollView(
                padding: const EdgeInsets.symmetric(
                  horizontal: 40,
                  vertical: 44,
                ),
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 380),
                  child: _buildAuthCardContent(isDesktop: true),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildMobileLayout(double screenWidth) {
    return SafeArea(
      child: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 28),
          child: Container(
            width: math.min(screenWidth * 0.92, 420),
            padding: const EdgeInsets.symmetric(horizontal: 26, vertical: 34),
            decoration: BoxDecoration(
              color: const Color(0xB32E4240),
              borderRadius: BorderRadius.circular(24),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.32),
                  blurRadius: 24,
                  offset: const Offset(0, 10),
                ),
              ],
            ),
            child: _buildAuthCardContent(isDesktop: false),
          ),
        ),
      ),
    );
  }

  Widget _buildAuthCardContent({required bool isDesktop}) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        MeloopLogoSticker(
          height: isDesktop ? 76 : 56,
          strokeWidth: isDesktop ? 3.2 : 2.6,
        ),
        SizedBox(height: isDesktop ? 44 : 28),

        // Pestañas con indicador deslizante suave
        _AuthModeTabs(
          isRegister: _isRegisterMode,
          onSelect: (isRegister) => _switchMode(isRegister: isRegister),
        ),
        SizedBox(height: isDesktop ? 36 : 24),

        // Altura mínima fija para que el logo y las pestañas nunca se muevan
        // al cambiar entre Iniciar Sesión y Registrarse.
        ConstrainedBox(
          constraints: BoxConstraints(minHeight: isDesktop ? 440 : 410),
          child: AnimatedSwitcher(
            duration: const Duration(milliseconds: 220),
            reverseDuration: const Duration(milliseconds: 160),
            switchInCurve: Curves.easeOutCubic,
            switchOutCurve: Curves.easeOutCubic,
            layoutBuilder: (currentChild, previousChildren) => Stack(
              alignment: Alignment.topCenter,
              children: [
                ...previousChildren,
                ?currentChild,
              ],
            ),
            transitionBuilder: (child, animation) {
              final isIncomingRegister =
                  child.key == const ValueKey('register-form');
              final direction = isIncomingRegister ? 1.0 : -1.0;
              final slideTween = Tween<Offset>(
                begin: Offset(0.06 * direction, 0),
                end: Offset.zero,
              );
              return FadeTransition(
                opacity: animation,
                child: SlideTransition(
                  position: slideTween.animate(animation),
                  child: child,
                ),
              );
            },
            child: _isRegisterMode
                ? _buildRegisterForm(key: const ValueKey('register-form'))
                : _buildLoginForm(key: const ValueKey('login-form')),
          ),
        ),
      ],
    );
  }

  Widget _buildLoginForm({required Key key}) {
    return Column(
      key: key,
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        _fieldLabel('Correo Electrónico:'),
        const SizedBox(height: 8),
        _buildTextField(
          controller: _emailController,
          keyboardType: TextInputType.emailAddress,
          textInputAction: TextInputAction.next,
        ),
        const SizedBox(height: 22),
        _fieldLabel('Contraseña:'),
        const SizedBox(height: 8),
        _buildTextField(
          controller: _passwordController,
          obscureText: true,
          textInputAction: TextInputAction.done,
          onSubmitted: (_) {
            if (!_isSubmitting) _login();
          },
        ),
        if (_authError != null) ...[
          const SizedBox(height: 16),
          _buildErrorBanner(_authError!),
        ],
        const SizedBox(height: 36),
        Center(
          child: _buildSubmitButton(
            label: 'Iniciar sesión',
            onPressed: _isSubmitting ? null : _login,
          ),
        ),
        const SizedBox(height: 28),
        Center(
          child: _buildSwitchLink(
            prompt: '¿No tienes cuenta? ',
            actionLabel: 'Crea una aquí',
            onTap: () => _switchMode(isRegister: true),
          ),
        ),
      ],
    );
  }

  Widget _buildRegisterForm({required Key key}) {
    return Column(
      key: key,
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        _fieldLabel('Usuario:'),
        const SizedBox(height: 8),
        _buildTextField(
          controller: _usernameController,
          textInputAction: TextInputAction.next,
        ),
        const SizedBox(height: 18),
        _fieldLabel('Correo Electrónico:'),
        const SizedBox(height: 8),
        _buildTextField(
          controller: _emailController,
          keyboardType: TextInputType.emailAddress,
          textInputAction: TextInputAction.next,
        ),
        const SizedBox(height: 18),
        _fieldLabel('Contraseña:'),
        const SizedBox(height: 8),
        _buildTextField(
          controller: _passwordController,
          obscureText: true,
          textInputAction: TextInputAction.done,
          onSubmitted: (_) {
            if (!_isSubmitting) _register();
          },
        ),
        if (_authError != null) ...[
          const SizedBox(height: 16),
          _buildErrorBanner(_authError!),
        ],
        const SizedBox(height: 32),
        Center(
          child: _buildSubmitButton(
            label: 'Crear cuenta',
            onPressed: _isSubmitting ? null : _register,
          ),
        ),
        const SizedBox(height: 24),
        Center(
          child: _buildSwitchLink(
            prompt: '¿Ya tienes una cuenta? ',
            actionLabel: 'Inicia sesión aquí',
            onTap: () => _switchMode(isRegister: false),
          ),
        ),
      ],
    );
  }

  Widget _fieldLabel(String text) {
    return Text(
      text,
      style: MeloopTypography.lexend(
        color: Colors.white,
        fontSize: 15,
        fontWeight: FontWeight.w400,
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    bool obscureText = false,
    TextInputType? keyboardType,
    TextInputAction? textInputAction,
    ValueChanged<String>? onSubmitted,
  }) {
    return TextField(
      controller: controller,
      obscureText: obscureText,
      keyboardType: keyboardType,
      textInputAction: textInputAction,
      onSubmitted: onSubmitted,
      style: MeloopTypography.lexend(
        color: const Color(0xFF172222),
        fontSize: 15,
      ),
      decoration: InputDecoration(
        filled: true,
        fillColor: Colors.white,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 18,
          vertical: 15,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: Color(0xFF68B6AA), width: 1.5),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: Color(0xFF68B6AA), width: 1.5),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: Color(0xFF56E4C8), width: 2.2),
        ),
      ),
    );
  }

  Widget _buildSubmitButton({
    required String label,
    required VoidCallback? onPressed,
  }) {
    return DecoratedBox(
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [Color(0xFF5CE6CC), Color(0xFF0FB49A)],
        ),
        borderRadius: BorderRadius.circular(14),
      ),
      child: ElevatedButton(
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.transparent,
          disabledBackgroundColor: Colors.transparent,
          shadowColor: Colors.transparent,
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 44, vertical: 16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(14),
          ),
        ),
        onPressed: onPressed,
        child: _isSubmitting
            ? const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2.2,
                  color: Colors.white,
                ),
              )
            : Text(
                label,
                style: MeloopTypography.lexend(
                  color: Colors.white,
                  fontSize: 16,
                  fontWeight: FontWeight.w500,
                ),
              ),
      ),
    );
  }

  Widget _buildErrorBanner(String message) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0x33FF6B6B),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0x88FF8A8A)),
      ),
      child: Text(
        message,
        style: MeloopTypography.lexend(color: Colors.white, fontSize: 13),
        textAlign: TextAlign.center,
      ),
    );
  }

  Widget _buildSwitchLink({
    required String prompt,
    required String actionLabel,
    required VoidCallback onTap,
  }) {
    return Wrap(
      alignment: WrapAlignment.center,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        Text(
          prompt,
          style: MeloopTypography.lexend(color: Colors.white, fontSize: 14),
        ),
        MouseRegion(
          cursor: SystemMouseCursors.click,
          child: GestureDetector(
            onTap: onTap,
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 2),
              child: Text(
                actionLabel,
                style: MeloopTypography.lexend(
                  color: const Color(0xFF39E0C1),
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _AuthModeTabs extends StatelessWidget {
  const _AuthModeTabs({required this.isRegister, required this.onSelect});

  final bool isRegister;
  final ValueChanged<bool> onSelect;

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 320),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              Expanded(
                child: _AuthTabItem(
                  label: 'Iniciar Sesión',
                  selected: !isRegister,
                  onTap: () => onSelect(false),
                ),
              ),
              Expanded(
                child: _AuthTabItem(
                  label: 'Registrarse',
                  selected: isRegister,
                  onTap: () => onSelect(true),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          SizedBox(
            height: 7,
            child: Stack(
              alignment: Alignment.center,
              children: [
                Container(
                  height: 3,
                  width: double.infinity,
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                AnimatedAlign(
                  duration: const Duration(milliseconds: 280),
                  curve: Curves.easeOutCubic,
                  alignment: isRegister
                      ? Alignment.centerRight
                      : Alignment.centerLeft,
                  child: FractionallySizedBox(
                    widthFactor: 0.5,
                    child: Container(
                      height: 7,
                      decoration: BoxDecoration(
                        color: const Color(0xFF56E4C8),
                        borderRadius: BorderRadius.circular(4),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _AuthTabItem extends StatelessWidget {
  const _AuthTabItem({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 4),
          child: AnimatedDefaultTextStyle(
            duration: const Duration(milliseconds: 200),
            style: MeloopTypography.lexend(
              color: selected
                  ? Colors.white
                  : Colors.white.withValues(alpha: 0.72),
              fontSize: 19,
              fontWeight: selected ? FontWeight.w500 : FontWeight.w400,
            ),
            child: Text(label, textAlign: TextAlign.center),
          ),
        ),
      ),
    );
  }
}

/// Fondo radial con el mismo gradiente de la pantalla de carga (`#248686` →
/// `#004562`) cuyo punto luminoso se desplaza suavemente con bajo costo de GPU.
class _AnimatedTealRadialBackground extends StatelessWidget {
  const _AnimatedTealRadialBackground({required this.animation});

  final Animation<double> animation;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: animation,
      builder: (context, _) => CustomPaint(
        painter: _MovingRadialGradientPainter(progress: animation.value),
        child: const SizedBox.expand(),
      ),
    );
  }
}

class _MovingRadialGradientPainter extends CustomPainter {
  const _MovingRadialGradientPainter({required this.progress});

  final double progress;

  static const _brightColor = Color(0xFF248686);
  static const _deepColor = Color(0xFF004562);

  @override
  void paint(Canvas canvas, Size size) {
    if (size.isEmpty) return;

    final t = progress * 2 * math.pi;
    // Órbita suave tipo Lissajous sobre el lienzo izquierdo/central
    final centerX = size.width * (0.26 + 0.16 * math.sin(t));
    final centerY = size.height * (0.42 + 0.24 * math.sin(2 * t + 0.5));
    final radiusX = size.width * 0.72;
    final radiusY = size.height * 0.78;

    canvas.save();
    canvas.translate(centerX, centerY);
    canvas.scale(radiusX, radiusY);
    final shader = ui.Gradient.radial(
      Offset.zero,
      1.0,
      const [_brightColor, _deepColor],
      const [0.0, 1.0],
    );
    canvas.drawRect(
      Rect.fromLTRB(
        -centerX / radiusX,
        -centerY / radiusY,
        (size.width - centerX) / radiusX,
        (size.height - centerY) / radiusY,
      ),
      Paint()..shader = shader,
    );
    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant _MovingRadialGradientPainter oldDelegate) =>
      oldDelegate.progress != progress;
}

/// Esquina oculta de diagnóstico: invisible y no clickeable hasta mantener el
/// cursor encima durante 3 segundos continuos.
class _HiddenDiagnosticCorner extends StatefulWidget {
  const _HiddenDiagnosticCorner({required this.onPressed});

  final VoidCallback onPressed;

  @override
  State<_HiddenDiagnosticCorner> createState() =>
      _HiddenDiagnosticCornerState();
}

class _HiddenDiagnosticCornerState extends State<_HiddenDiagnosticCorner> {
  Timer? _hoverTimer;
  bool _revealed = false;

  void _onEnter(PointerEvent _) {
    if (_revealed) return;
    _hoverTimer?.cancel();
    _hoverTimer = Timer(const Duration(seconds: 3), () {
      if (mounted) setState(() => _revealed = true);
    });
  }

  void _onExit(PointerEvent _) {
    _hoverTimer?.cancel();
    _hoverTimer = null;
    if (_revealed) {
      setState(() => _revealed = false);
    }
  }

  @override
  void dispose() {
    _hoverTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: _onEnter,
      onExit: _onExit,
      child: ConstrainedBox(
        constraints: const BoxConstraints(minWidth: 210, minHeight: 44),
        child: Align(
          alignment: Alignment.bottomRight,
          child: IgnorePointer(
            ignoring: !_revealed,
            child: AnimatedOpacity(
              opacity: _revealed ? 1.0 : 0.0,
              duration: const Duration(milliseconds: 240),
              curve: Curves.easeOut,
              child: TextButton.icon(
                style: TextButton.styleFrom(
                  foregroundColor: Colors.white.withValues(alpha: 0.9),
                  backgroundColor: Colors.black.withValues(alpha: 0.28),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 10,
                  ),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10),
                  ),
                ),
                onPressed: widget.onPressed,
                icon: const Icon(Icons.monitor_heart_outlined, size: 16),
                label: Text(
                  'Diagnóstico de conexión',
                  style: MeloopTypography.lexend(
                    fontSize: 12.5,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
