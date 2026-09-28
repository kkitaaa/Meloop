# 6. Flutter App

En este documento se explica cómo configurar y ejecutar la aplicación Flutter móvil.

**Tiempo estimado:** 20-30 minutos

---

## Flutter App Overview

**Tecnología:** Flutter 3.13.0+ con Dart  
**Plataformas soportadas:** Android, iOS, Web, Windows  
**Puerto:** Depende de la plataforma (web: 5000, simulador: emulador)

---

## Verificar Instalación de Flutter

```bash
flutter --version
# Output: Flutter 3.13.x • channel stable • ...

dart --version
# Output: Dart 3.1.x (stable) (...)

flutter doctor
# Deberías ver información sobre tu entorno
```

---

## 📁 Estructura de la App

```
app/
└── flutter_app/
    ├── lib/                    # Código fuente
    │   └── main.dart          # Punto de entrada
    ├── test/                  # Tests unitarios
    ├── web/                   # Versión web
    ├── windows/               # Versión Windows
    ├── android/               # Versión Android
    ├── linux/                 # Versión Linux
    ├── pubspec.yaml          # Dependencias
    ├── analysis_options.yaml  # Análisis de código
    └── README.md
```

---

## 📦 Instalar Dependencias

### Paso 1: Obtener dependencias

```bash
cd app/flutter_app

# Obtener todas las dependencias
flutter pub get

# Salida esperada:
# Running \"flutter pub get\" in flutter_app...
# Got dependencies!
```

### Paso 2: Actualizar dependencias (opcional)

```bash
flutter pub upgrade

# Para actualizar a la última versión disponible
flutter pub upgrade --major-versions
```

---

## 🏃 Ejecutar la App

### Opción A: En navegador (Web)

```bash
cd app/flutter_app

flutter run -d chrome

# Salida esperada:
# Launching lib/main.dart on Chrome in debug mode...
# Building for chrome...
# Web app built successfully!
# Running on Chrome...
```

**Acceso:** http://localhost:5000

### Opción B: En Windows (Desktop)

```bash
flutter run -d windows

# Salida esperada:
# Launching lib/main.dart on Windows in debug mode...
# Building for Windows...
# App launched successfully!
```

### Opción C: En Linux (Desktop)

```bash
flutter run -d linux
```

### Opción D: En Android Emulator

```bash
# Primero, abre el Android Emulator (requiere Android SDK)
emulator -avd <nombre_emulador> &

# Luego ejecuta Flutter
flutter run
```

---

## Configuración de Conectividad al API

Necesitas configurar la app para conectarse a tu backend:

### Paso 1: Crear archivo de configuración

```dart
// lib/config/api_config.dart
class ApiConfig {
  // En desarrollo, usa localhost
  static const String devApiUrl = 'http://localhost:8080';
  
  // En producción, usa tu servidor
  static const String prodApiUrl = 'https://api.meloop.app';
  
  // Cambiar basado en el entorno
  static String get apiUrl => _isProduction ? prodApiUrl : devApiUrl;
  
  static const bool _isProduction = false; // Cambiar a true en producción
  
  // Endpoints
  static const String healthCheck = '/health';
  static const String loginEndpoint = '/auth/login';
  static const String registerEndpoint = '/auth/register';
  static const String usersEndpoint = '/users';
  static const String postsEndpoint = '/posts';
}
```

### Paso 2: Implementar cliente HTTP

```dart
// lib/services/api_service.dart
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:flutter/foundation.dart';
import '../config/api_config.dart';

class ApiService {
  static final ApiService _instance = ApiService._internal();
  
  factory ApiService() {
    return _instance;
  }
  
  ApiService._internal();
  
  // Headers por defecto
  Map<String, String> get defaultHeaders => {
    'Content-Type': 'application/json',
  };
  
  // GET request
  Future<T> get<T>(
    String endpoint, {
    Map<String, String>? headers,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.apiUrl}$endpoint');
      final response = await http.get(
        url,
        headers: {...defaultHeaders, ...?headers},
      ).timeout(Duration(seconds: 30));
      
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return fromJson(data);
      } else {
        throw Exception('Error: ${response.statusCode} - ${response.body}');
      }
    } catch (e) {
      debugPrint('GET Error: \$e');
      rethrow;
    }
  }
  
  // POST request
  Future<T> post<T>(
    String endpoint, {
    Map<String, String>? headers,
    required Map<String, dynamic> body,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.apiUrl}$endpoint');
      final response = await http.post(
        url,
        headers: {...defaultHeaders, ...?headers},
        body: jsonEncode(body),
      ).timeout(Duration(seconds: 30));
      
      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = jsonDecode(response.body);
        return fromJson(data);
      } else {
        throw Exception('Error: ${response.statusCode} - ${response.body}');
      }
    } catch (e) {
      debugPrint('POST Error: \$e');
      rethrow;
    }
  }
  
  // Validar conexión
  Future<bool> healthCheck() async {
    try {
      final url = Uri.parse('${ApiConfig.apiUrl}${ApiConfig.healthCheck}');
      final response = await http.get(url).timeout(Duration(seconds: 5));
      return response.statusCode == 200;
    } catch (e) {
      debugPrint('Health check failed: \$e');
      return false;
    }
  }
}
```

### Paso 3: Usar el servicio en tu aplicación

```dart
// lib/screens/home_screen.dart
import 'package:flutter/material.dart';
import '../services/api_service.dart';

class HomeScreen extends StatefulWidget {
  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  late ApiService _apiService;
  bool _isConnected = false;
  String _statusMessage = '';
  
  @override
  void initState() {
    super.initState();
    _apiService = ApiService();
    _checkConnection();
  }
  
  Future<void> _checkConnection() async {
    final isConnected = await _apiService.healthCheck();
    setState(() {
      _isConnected = isConnected;
      _statusMessage = isConnected 
          ? '✅ Conectado al API Gateway'
          : '❌ No se puede conectar al API Gateway';
    });
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Meloop')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(_statusMessage),
            SizedBox(height: 20),
            ElevatedButton(
              onPressed: _checkConnection,
              child: Text('Validar Conexión'),
            ),
          ],
        ),
      ),
    );
  }
}
```

---

## 🧪 Testing

### Ejecutar tests unitarios

```bash
cd app/flutter_app

# Tests unitarios
flutter test

# Tests con cobertura
flutter test --coverage

# Ver reporte de cobertura
start coverage/index.html # Windows
xdg-open coverage/index.html # Linux
```

### Ejemplo de test

```dart
// test/services/api_service_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:meloop_app/services/api_service.dart';

void main() {
  group('ApiService', () {
    test('Health check retorna true/false', () async {
      final apiService = ApiService();
      final result = await apiService.healthCheck();
      expect(result, isA<bool>());
    });
  });
}
```

---

## Análisis de Código

### Ejecutar análisis

```bash
cd app/flutter_app

# Análisis de código
flutter analyze

# Debería mostrar:
# No issues found!
```

### Corregir automáticamente problemas

```bash
dart fix --apply
```

---

## 🚀 Modos de Ejecución

### Debug (Desarrollo)

```bash
flutter run
# Ejecuta con hot reload y debugging
```

### Release (Optimizado)

```bash
flutter run --release
# Compila en modo release (más rápido, no debugging)
```

### Profile (Rendimiento)

```bash
flutter run --profile
# Compila con optimizaciones pero permite profiling
```

---

## 📱 Compilar para Diferentes Plataformas

### Web

```bash
flutter build web
# Output en: build/web/
```

### Windows Desktop

```bash
flutter build windows
# Output en: build/windows/runner/Release/
```

### Linux Desktop

```bash
flutter build linux
# Output en: build/linux/x64/release/bundle/
```

### Android

```bash
# Requiere Android SDK
flutter build apk
flutter build appbundle
# Output en: build/app/outputs/
```

---

## Troubleshooting

### "Unable to locate a browser"

```bash
# Instalar Chrome primero
# Luego especificar el navegador
flutter run -d web --web-renderer=html
```

### "Multiple devices detected"

```bash
# Listar dispositivos disponibles
flutter devices

# Ejecutar en dispositivo específico
flutter run -d <device-id>
```

### "Device is offline"

```bash
# Reconectar dispositivo
flutter clean
flutter pub get
flutter run
```

### App se ve distorsionada o no renderiza

```bash
flutter clean
flutter pub get
flutter run --release
```

---

## 🌐 Conectar a API Gateway en Diferentes Escenarios

### Desarrollo Local

```dart
static const String devApiUrl = 'http://localhost:8080';
```

### Desarrollo en Docker (desde host)

```dart
// Windows con Docker Desktop
static const String devApiUrl = 'http://host.docker.internal:8080';
```

### Desarrollo en Emulador

```dart
// Android Emulator
static const String devApiUrl = 'http://10.0.2.2:8080';
```

### Producción

```dart
static const String prodApiUrl = 'https://api.meloop.app';
```

---

## Hot Reload / Hot Restart

Mientras la app está ejecutándose:

```bash
# Hot reload (recarga código manteniendo estado)
r

# Hot restart (recarga completa)
R

# Salir
q
```

---

## 🚀 Próximo Paso

Ahora que tienes la app Flutter compilada y conectada:

👉 **Ve a [7. Inicio Rápido →](07-quick-start.md)**

---

## Checklist de Flutter

```
□ Flutter 3.13.0+ instalado
□ Dart SDK compatible
□ flutter doctor pasa sin errores críticos
□ Dependencias obtenidas: flutter pub get
□ App compila sin errores
□ App ejecutándose en web/desktop
□ ApiConfig configurado correctamente
□ ApiService implementado
□ Health check conecta correctamente
□ Tests pasan: flutter test
□ Análisis pasa: flutter analyze
□ Hot reload/restart funciona
□ UI se ve correctamente
```

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0
