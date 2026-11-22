# Quick Reference - Flutter-Kubernetes Integration

## 🚀 Quick Start

### Build Development APK
```bash
flutter build apk --dart-define=FLUTTER_ENV=development --debug
```

### Build Staging
```powershell
.\build.ps1 -Environment staging -Platform android
```

### Build Production
```powershell
.\build.ps1 -Environment production -Platform windows
```

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| [KUBERNETES_INTEGRATION.md](./KUBERNETES_INTEGRATION.md) | Complete deployment guide |
| [VERIFICATION_GUIDE.md](./VERIFICATION_GUIDE.md) | Testing and validation procedures |
| [VERIFICATION_REPORT.md](./VERIFICATION_REPORT.md) | Full verification results |

## 🔧 Core Services

### ApiConfig
- **File**: `lib/config/api_config.dart`
- **Purpose**: Centralized configuration management
- **Features**: 
  - 3 environments (dev/staging/prod)
  - Platform-specific URLs
  - Compile-time configuration via dart-define

```dart
// Usage
String baseUrl = ApiConfig.baseUrl; // Returns platform-specific URL
Environment env = ApiConfig.environment; // Returns current environment
```

### HttpClientManager
- **File**: `lib/services/http_client_manager.dart`
- **Purpose**: HTTP client with resilience patterns
- **Features**:
  - Automatic retry with exponential backoff
  - Token refresh interceptor
  - Request/response timeouts
  - Development logging

```dart
// Usage
final httpClient = HttpClientManager();
final dio = httpClient.dio;
httpClient.setAuthToken('token');
```

### WebSocketManager
- **File**: `lib/services/websocket_manager.dart`
- **Purpose**: Real-time WebSocket communication
- **Features**:
  - Automatic reconnection
  - Event handlers (message, error, close)
  - Max 5 reconnection attempts
  - Exponential backoff

```dart
// Usage
final wsManager = WebSocketManager('/api/notifications');
await wsManager.connect();
wsManager.addMessageHandler((msg) => print('Message: $msg'));
```

### HealthCheckService
- **File**: `lib/services/health_check_service.dart`
- **Purpose**: Backend connectivity diagnostics
- **Features**:
  - Overall health check
  - Individual service diagnostics
  - Response time measurement
  - Structured reports

```dart
// Usage
final healthService = HealthCheckService();
final health = await healthService.checkHealth();
print('Backend healthy: ${health.isHealthy}');
```

## 📱 Platform-Specific URLs

| Platform | Development | Staging | Production |
|----------|-------------|---------|------------|
| Android | 10.0.2.2:8080 | staging.meloop.com | api.meloop.com |
| Windows | localhost:8080 | staging.meloop.com | api.meloop.com |
| K8s | localhost:8080 | backend.meloop.svc.cluster.local | backend.meloop.svc.cluster.local |

## 🛠️ Build Scripts

### PowerShell (Windows)
```powershell
# Syntax
.\build.ps1 -Environment <env> -Platform <platform>

# Examples
.\build.ps1 -Environment dev -Platform android
.\build.ps1 -Environment staging -Platform windows
.\build.ps1 -Environment prod -Platform all
```

### Bash (Linux/macOS)
```bash
# Syntax
./build.sh -e <env> -p <platform>

# Examples
./build.sh -e dev -p android
./build.sh -e staging -p windows
./build.sh -e prod -p all
```

## 🧪 Quick Health Check

```dart
import 'package:flutter_app/services/health_check_service.dart';

void main() async {
  final health = HealthCheckService();
  final report = await health.checkHealth();
  
  print('Status: ${report.status}');
  print('Healthy: ${report.isHealthy}');
  print('Services: ${report.services}');
}
```

## 🔗 API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/health` | GET | Health check |
| `/api/auth` | POST | Authentication |
| `/api/demo` | GET | Demo endpoint |
| `/api/feed` | GET | Feed endpoint |
| `/api/posts` | GET/POST | Posts |
| `/api/users` | GET/POST | Users |
| `/api/profile` | GET/PUT | User profile |
| `/api/notifications` | WS | WebSocket notifications |
| `/api/music` | GET/POST | Music service |

## 🐛 Debugging

### Enable Logging
Logging is automatically enabled in development mode:
```dart
// No action needed - ApiConfig.enableLogging handles it
```

### Debug Screen
Navigate to `KubernetesConnectivityScreen` for real-time debugging:
```dart
import 'package:flutter_app/screens/kubernetes_connectivity_screen.dart';

// Push to screen
Navigator.push(
  context,
  MaterialPageRoute(builder: (_) => KubernetesConnectivityScreen()),
);
```

## 📋 Verification Checklist

- [ ] Environment variables are set correctly
- [ ] Backend is accessible from Flutter
- [ ] WebSocket connections work
- [ ] Health check responds
- [ ] API endpoints are reachable
- [ ] Authentication works
- [ ] Error handling is functional
- [ ] Builds succeed for all platforms

## 🚨 Troubleshooting

### "Cannot reach backend"
1. Check environment configuration: `ApiConfig.baseUrl`
2. Verify network connectivity
3. Run health check: `HealthCheckService.checkHealth()`
4. Check K8s Ingress configuration

### "WebSocket timeout"
1. Verify WebSocket path configuration
2. Check server WebSocket support
3. Review firewall rules
4. Check network connectivity

### "Build fails"
1. Run `flutter clean`
2. Run `flutter pub get`
3. Verify Gradle version (9.5.0+)
4. Check Android Gradle Plugin version (9.1.1+)
5. Verify dart-define syntax

## 📞 Support Resources

- [Flutter Documentation](https://flutter.dev/docs)
- [Kubernetes Documentation](https://kubernetes.io/docs)
- [Dio Library](https://github.com/flutterchina/dio)
- [Project Wiki](../../../wiki)

---

**Last Updated**: 2024-11-27  
**Version**: 1.0.0  
**Status**: Production Ready
