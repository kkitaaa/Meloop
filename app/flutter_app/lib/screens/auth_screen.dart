import 'package:flutter/material.dart';
import 'package:flutter/gestures.dart';
import 'feed_screen.dart';

class AuthScreen extends StatelessWidget {
  const AuthScreen({super.key});

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
              child: Container(color: Colors.black.withOpacity(0.15)),
            ),
            
          SafeArea(
            child: Center(
              child: SingleChildScrollView(
                child: Align(
                  alignment: isDesktop ? Alignment.centerRight : Alignment.center,
                  child: Padding(
                    padding: EdgeInsets.only(right: isDesktop ? screenWidth * 0.1 : 0),
                    child: Container(
                      width: containerWidth,
                      padding: EdgeInsets.symmetric(
                        horizontal: horizontalPadding, 
                        vertical: verticalPadding
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFF455A55),
                        borderRadius: BorderRadius.circular(isDesktop ? 12 : 24),
                        // Sombra elegante para dar volumen en el celular
                        boxShadow: [
                          if (!isDesktop)
                            BoxShadow(
                              color: Colors.black.withOpacity(0.4),
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
          keyboardType: TextInputType.emailAddress, // Muestra el teclado con el "@" en celular
          decoration: InputDecoration(
            filled: true,
            fillColor: Colors.white,
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
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
          obscureText: true,
          decoration: InputDecoration(
            filled: true,
            fillColor: Colors.white,
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
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
            onPressed: () {
              Navigator.pushReplacement(
                context,
                MaterialPageRoute(builder: (context) => const FeedScreen()),
              );
            },
            child: const Text(
              'Iniciar Sesion',
              style: TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.w600),
            ),
          ),
        ),
        const Spacer(),
        
        Center(
          child: RichText(
            text: TextSpan(
              text: 'No tienes cuenta? ',
              style: const TextStyle(color: Colors.white70, fontSize: 13),
              children: [
                TextSpan(
                  text: 'Crea una aqui',
                  style: const TextStyle(color: Color(0xFF1ABC9C), fontSize: 13, fontWeight: FontWeight.bold),
                  recognizer: TapGestureRecognizer()..onTap = () {},
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 8),
      ],
    );
  }

  Widget _buildRegisterForm() {
    return const Center(
      child: Text(
        "Aquí irá la columna con\nlos campos de registro",
        textAlign: TextAlign.center,
        style: TextStyle(color: Colors.white),
      ),
    );
  }
}