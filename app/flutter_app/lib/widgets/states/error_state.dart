import 'package:flutter/material.dart';

class ErrorState extends StatelessWidget {
  final int errorCode;
  final VoidCallback onRetry;
  final Color tealAccent;

  const ErrorState({
    super.key,
    required this.errorCode,
    required this.onRetry,
    required this.tealAccent,
  });

  String _translateErrorCode(int code) {
    switch (code) {
      case 401:
        return "Tu sesión ha expirado. Por favor, vuelve a iniciar sesión.";
      case 403:
        return "No tienes los permisos necesarios para ver este contenido.";
      case 429:
        return "Has realizado demasiadas solicitudes. Espera unos minutos.";
      default:
        if (code >= 500 && code < 600) {
          return "El servidor está experimentando problemas ($code). Estamos trabajando en ello.";
        }
        return "No se pudo cargar la información ($code). Revisa tu conexión a internet.";
    }
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.wifi_tethering_error_rounded,
              size: 56,
              color: Colors.red[300],
            ),
            const SizedBox(height: 16),
            const Text(
              "¡Ups! Algo salió mal",
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: Colors.black87,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              _translateErrorCode(errorCode),
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 13, color: Colors.grey),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: tealAccent,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(20),
                ),
              ),
              onPressed: onRetry,
              icon: const Icon(Icons.refresh, color: Colors.white, size: 16),
              label: const Text(
                "Reintentar",
                style: TextStyle(color: Colors.white),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
