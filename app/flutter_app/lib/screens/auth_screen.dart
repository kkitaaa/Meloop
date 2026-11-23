import 'package:flutter/material.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/foundation.dart';

import '../core/navigation/responsive_navigation_shell.dart';
import '../features/moderation/data/moderation_models.dart';
import '../features/notifications/presentation/notification_scope.dart';
import '../services/api_client.dart';
import '../services/auth_session.dart';
import 'feed_screen.dart';
import 'kubernetes_connectivity_screen.dart';

class AuthScreen extends StatefulWidget {
  final ApiClient? apiClient;

  const AuthScreen({super.key, this.apiClient});

  @override
  State<AuthScreen> createState() => _AuthScreenState();
}

class _AuthScreenState extends State<AuthScreen> {
  final TextEditingController _emailController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _usernameController = TextEditingController();
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;

  bool _isSubmitting = false;
  String? _authError;

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    if (AuthSession.shared.isAuthenticated) {
      _restoreSession();
    }
  }

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    _usernameController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
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
  }

  @override
  Widget build(BuildContext context) {
    double screenWidth = MediaQuery.of(context).size.width;
    bool isDesktop = screenWidth > 800;

    // Anchos y paddings dinámicos para respirar en celular
    double containerWidth = isDesktop ? 420 : screenWidth * 0.92;
    double horizontalPadding = isDesktop ? 40.0 : 24.0;
    double verticalPadding = isDesktop ? 40.0 : 32.0;

    return Scaffold(
      // Evita que la imagen de fondo se comprima al abrir el teclado
      resizeToAvoidBottomInset: false,
      body: Stack(
        children: [
          Positioned.fill(
            child: Image.asset(
              'assets/imagenes/fondo.png',
              fit: BoxFit.cover,
              alignment: isDesktop ? Alignment.centerLeft : Alignment.center,
            ),
          ),

          // Capa de oscurecimiento suave solo para móvil
          if (!isDesktop)
            Positioned.fill(
              child: Container(color: Colors.black.withValues(alpha: 0.15)),
            ),

          SafeArea(
            child: Center(
              child: SingleChildScrollView(
                child: Align(
                  alignment: isDesktop
                      ? Alignment.centerRight
                      : Alignment.center,
                  child: Padding(
                    padding: EdgeInsets.only(
                      right: isDesktop ? screenWidth * 0.1 : 0,
                    ),
                    child: Container(
                      width: containerWidth,
                      padding: EdgeInsets.symmetric(
                        horizontal: horizontalPadding,
                        vertical: verticalPadding,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFF455A55),
                        borderRadius: BorderRadius.circular(
                          isDesktop ? 12 : 24,
                        ),
                        // Sombra elegante para dar volumen en el celular
                        boxShadow: [
                          if (!isDesktop)
                            BoxShadow(
                              color: Colors.black.withValues(alpha: 0.4),
                              blurRadius: 20,
                              offset: const Offset(0, 10),
                            ),
                        ],
                      ),
                      child: DefaultTabController(
                        length: 2,
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Image.asset(
                              'assets/imagenes/meloop.png',
                              height: isDesktop ? 55 : 45,
                            ),
                            const SizedBox(height: 24),

                            const TabBar(
                              indicatorColor: Color(0xFF1ABC9C),
                              indicatorWeight: 3,
                              labelColor: Colors.white,
                              unselectedLabelColor: Colors.white60,
                              dividerColor: Colors.transparent, // Elimina la línea base gris nativa
                              tabs: [
                                Tab(text: 'Iniciar Sesión'),
                                Tab(text: 'Registrarse'),
                              ],
                            ),
                            const SizedBox(height: 24),

                            SizedBox(
                              height: 380,
                              child: TabBarView(
                                children: [
                                  _buildLoginForm(context),
                                  _buildRegisterForm(),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLoginForm(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 8),
        const Text(
          'Correo Electronico:',
          style: TextStyle(color: Colors.white, fontSize: 13),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _emailController,
          keyboardType: TextInputType
              .emailAddress, // Muestra el teclado con el "@" en celular
          decoration: InputDecoration(
            filled: true,
            fillColor: Colors.white,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 14,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide.none,
            ),
          ),
        ),
        const SizedBox(height: 20),
        const Text(
          'Contraseña:',
          style: TextStyle(color: Colors.white, fontSize: 13),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _passwordController,
          obscureText: true,
          decoration: InputDecoration(
            filled: true,
            fillColor: Colors.white,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 14,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide.none,
            ),
          ),
        ),
        const SizedBox(height: 32),

        Center(
          child: ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF1ABC9C),
              padding: const EdgeInsets.symmetric(horizontal: 48, vertical: 14),
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(24),
              ),
            ),
            onPressed: _isSubmitting ? null : _login,
            child: _isSubmitting
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text(
                    'Iniciar sesión',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
          ),
        ),
        if (_authError != null) ...[
          const SizedBox(height: 12),
          Text(
            _authError!,
            style: const TextStyle(color: Colors.white, fontSize: 13),
            textAlign: TextAlign.center,
          ),
        ],
        const Spacer(),

        Center(
          child: RichText(
            text: TextSpan(
              text: 'No tienes cuenta? ',
              style: const TextStyle(color: Colors.white70, fontSize: 13),
              children: [
                TextSpan(
                  text: 'Crea una aqui',
                  style: const TextStyle(
                    color: Color(0xFF1ABC9C),
                    fontSize: 13,
                    fontWeight: FontWeight.bold,
                  ),
                  recognizer: TapGestureRecognizer()..onTap = () {},
                ),
              ],
            ),
          ),
        ),
        if (kDebugMode)
          Align(
            alignment: Alignment.centerRight,
            child: TextButton.icon(
              onPressed: () {
                Navigator.push<void>(
                  context,
                  MaterialPageRoute<void>(
                    builder: (_) => const KubernetesConnectivityScreen(),
                  ),
                );
              },
              icon: const Icon(Icons.monitor_heart, size: 16),
              label: const Text('Diagnóstico de conexión'),
            ),
          ),
        const SizedBox(height: 8),
      ],
    );
  }

  Widget _buildRegisterForm() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 8),
        const Text('Usuario', style: TextStyle(color: Colors.white)),
        const SizedBox(height: 6),
        TextField(
          controller: _usernameController,
          decoration: const InputDecoration(
            filled: true,
            fillColor: Colors.white,
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
        ),
        const SizedBox(height: 12),
        const Text('Correo electrónico', style: TextStyle(color: Colors.white)),
        const SizedBox(height: 6),
        TextField(
          controller: _emailController,
          keyboardType: TextInputType.emailAddress,
          decoration: const InputDecoration(
            filled: true,
            fillColor: Colors.white,
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
        ),
        const SizedBox(height: 12),
        const Text('Contraseña', style: TextStyle(color: Colors.white)),
        const SizedBox(height: 6),
        TextField(
          controller: _passwordController,
          obscureText: true,
          decoration: const InputDecoration(
            filled: true,
            fillColor: Colors.white,
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
        ),
        if (_authError != null) ...[
          const SizedBox(height: 8),
          Text(
            _authError!,
            style: const TextStyle(color: Colors.white, fontSize: 12),
          ),
        ],
        const Spacer(),
        Center(
          child: ElevatedButton(
            onPressed: _isSubmitting ? null : _register,
            child: _isSubmitting
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Crear cuenta'),
          ),
        ),
      ],
    );
  }
}
