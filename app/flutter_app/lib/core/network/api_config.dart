/// Configuración del API Gateway.
///
/// Se puede cambiar al compilar sin modificar código:
/// `flutter run --dart-define=API_BASE_URL=http://192.168.1.10:8080`
abstract final class ApiConfig {
  static const String baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  static const Duration requestTimeout = Duration(seconds: 15);
}
