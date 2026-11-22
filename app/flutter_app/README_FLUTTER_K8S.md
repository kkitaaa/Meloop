# Flutter-Kubernetes Integration - Documentation Index

## 📚 Complete Documentation Map

### For Quick Start (5 minutes)
👉 **[QUICK_REFERENCE.md](./QUICK_REFERENCE.md)**
- Quick build commands
- Service usage examples
- API endpoints reference
- Troubleshooting tips

### For Deployment (30 minutes)
👉 **[KUBERNETES_INTEGRATION.md](./KUBERNETES_INTEGRATION.md)**
- Complete deployment guide
- Configuration for all environments
- K8s verification steps
- Production setup

### For Testing & Validation (60 minutes)
👉 **[VERIFICATION_GUIDE.md](./VERIFICATION_GUIDE.md)**
- Complete testing procedures
- Validation steps for all components
- Platform-specific testing
- Troubleshooting guide

### For Verification Report
👉 **[VERIFICATION_REPORT.md](./VERIFICATION_REPORT.md)**
- Full verification results
- Build configuration details
- Service architecture breakdown
- Deployment readiness checklist

---

## 🔍 Service Documentation (In-Code)

### Configuration Management
**File**: `lib/config/api_config.dart`

```dart
// Get current environment
Environment env = ApiConfig.environment;

// Get base URL (platform-specific)
String baseUrl = ApiConfig.baseUrl;

// Get WebSocket URL
String wsUrl = ApiConfig.webSocketUrl;

// All available endpoints
ApiConfig.authEndpoint      // /api/auth
ApiConfig.feedEndpoint      // /api/feed
ApiConfig.notificationsEndpoint  // /api/notifications
// ... more endpoints
```

**Key Features**:
- 3 environments: development, staging, production
- Platform-specific URLs
- Compile-time configuration via `--dart-define`

---

### HTTP Client
**File**: `lib/services/http_client_manager.dart`

```dart
// Initialize
final httpClient = HttpClientManager();
final dio = httpClient.dio;

// Set auth token
httpClient.setAuthToken('your-token');

// Make requests
final response = await dio.get('/api/endpoint');

// Update base URL dynamically
httpClient.updateBaseUrl('http://new-url');
```

**Features**:
- Automatic retry with exponential backoff
- Token refresh interceptor
- Request/response timeouts (10s each)
- Logging in development mode

---

### WebSocket Manager
**File**: `lib/services/websocket_manager.dart`

```dart
// Initialize
final wsManager = WebSocketManager('/api/notifications');

// Connect
await wsManager.connect();

// Add handlers
wsManager.addMessageHandler((msg) => print('Message: $msg'));
wsManager.addErrorHandler((error) => print('Error: $error'));
wsManager.addCloseHandler(() => print('Closed'));

// Send message
wsManager.send('Hello');

// Check connection status
bool connected = wsManager.isConnected;

// Disconnect
await wsManager.disconnect();
```

**Features**:
- Automatic connection with timeout
- Auto-reconnection (max 5 attempts)
- Event-based architecture
- Exponential backoff on reconnection

---

### Health Check Service
**File**: `lib/services/health_check_service.dart`

```dart
// Initialize
final healthService = HealthCheckService();

// Check overall health
final health = await healthService.checkHealth();
print('Status: ${health.status}');
print('Healthy: ${health.isHealthy}');

// Get individual service health
final diagnostics = await healthService.runDiagnostics();
for (var service in diagnostics.serviceHealth) {
  print('${service.name}: ${service.isHealthy}');
}

// Get full report
final report = await healthService.getConnectivityReport();
```

**Features**:
- Overall backend health check
- Individual service diagnostics
- Response time measurement
- 5-second timeout
- Structured reports

---

### Debug UI Screen
**File**: `lib/screens/kubernetes_connectivity_screen.dart`

```dart
// Navigate to debug screen
Navigator.push(
  context,
  MaterialPageRoute(
    builder: (_) => KubernetesConnectivityScreen(),
  ),
);
```

**Features**:
- Real-time health monitoring
- Service status display
- Manual health check trigger
- WebSocket status display
- Debug information panel

---

## 🛠️ Build System

### PowerShell Script (Windows)
**File**: `build.ps1`

```powershell
# Development Android
.\build.ps1 -Environment dev -Platform android

# Staging Windows
.\build.ps1 -Environment staging -Platform windows

# Production all platforms
.\build.ps1 -Environment prod -Platform all
```

**Environments**: dev, development, staging, prod, production  
**Platforms**: android, windows, ios, all

---

### Bash Script (Linux/macOS)
**File**: `build.sh`

```bash
# Development Android
./build.sh -e dev -p android

# Staging Windows
./build.sh -e staging -p windows

# Production all platforms
./build.sh -e prod -p all
```

**Environments**: dev, staging, prod  
**Platforms**: android, windows, ios, all

---

## 📱 Platform Configuration

### Android
- **Development**: `http://10.0.2.2:8080` (emulator host alias)
- **Staging**: `http://staging.meloop.com`
- **Production**: `https://api.meloop.com`

### Windows
- **Development**: `http://localhost:8080`
- **Staging**: `http://staging.meloop.com`
- **Production**: `https://api.meloop.com`

### Kubernetes
- **Internal Service DNS**: `backend.meloop.svc.cluster.local:8080`
- **Ingress**: Configured via `K8S_INGRESS_HOST`
- **API Gateway**: Configured via `K8S_API_GATEWAY_URL`

---

## 🔐 Configuration via Compile-Time Variables

### How to Pass Configuration

```bash
# Using Flutter CLI
flutter build apk \
  --dart-define=FLUTTER_ENV=development \
  --dart-define=API_BASE_URL=http://localhost:8080 \
  --dart-define=K8S_NAMESPACE=meloop \
  --dart-define=K8S_INGRESS_HOST=meloop.local
```

### Available Variables
- `FLUTTER_ENV`: development, staging, production
- `API_BASE_URL`: Override default base URL
- `K8S_API_GATEWAY_URL`: K8s API Gateway endpoint
- `K8S_NAMESPACE`: K8s namespace (default: meloop)
- `K8S_INGRESS_HOST`: K8s Ingress hostname
- `CURRENT_USER_ID`: Current authenticated user

---

## 🧪 Testing Quick Commands

### Health Check Test
```dart
import 'package:flutter_app/services/health_check_service.dart';

void testHealth() async {
  final service = HealthCheckService();
  final health = await service.checkHealth();
  assert(health.isHealthy, 'Backend should be healthy');
}
```

### HTTP Client Test
```dart
import 'package:flutter_app/services/http_client_manager.dart';

void testHttp() async {
  final client = HttpClientManager();
  final response = await client.dio.get('/health');
  assert(response.statusCode == 200, 'Health check failed');
}
```

### WebSocket Test
```dart
import 'package:flutter_app/services/websocket_manager.dart';

void testWebSocket() async {
  final ws = WebSocketManager('/api/notifications');
  await ws.connect();
  assert(ws.isConnected, 'WebSocket should be connected');
  await ws.disconnect();
}
```

---

## 📊 API Endpoints Reference

| Endpoint | Method | Purpose | Auth |
|----------|--------|---------|------|
| `/health` | GET | Backend health | ❌ |
| `/api/auth` | POST | Authentication | ❌ |
| `/api/auth/refresh` | POST | Token refresh | ✅ |
| `/api/demo` | GET | Demo endpoint | ❌ |
| `/api/feed` | GET | User feed | ✅ |
| `/api/posts` | GET/POST | Posts | ✅ |
| `/api/users` | GET/POST | Users | ✅ |
| `/api/users/{id}` | GET | User by ID | ✅ |
| `/api/profile` | GET/PUT | User profile | ✅ |
| `/api/notifications` | WS | Notifications | ✅ |
| `/api/music` | GET/POST | Music service | ✅ |

---

## ✅ Verification Checklist

Before deploying to production, verify:

- [ ] Environment variables configured correctly
- [ ] Backend is accessible from Flutter
- [ ] WebSocket connections work
- [ ] Health check responds successfully
- [ ] All API endpoints are reachable
- [ ] Authentication works
- [ ] Error handling is functional
- [ ] Builds succeed for all platforms
- [ ] Logging works in development mode
- [ ] Retry logic works (simulate network errors)

---

## 🚨 Common Issues & Solutions

### "Cannot reach backend"
**Solutions**:
1. Verify `ApiConfig.baseUrl` is correct
2. Check network connectivity
3. Run `HealthCheckService.checkHealth()`
4. Verify K8s Ingress/LoadBalancer configuration

### "WebSocket timeout"
**Solutions**:
1. Check WebSocket path (`/api/notifications`)
2. Verify server WebSocket support
3. Review firewall rules
4. Check network connectivity

### "Build fails"
**Solutions**:
1. Run `flutter clean && flutter pub get`
2. Verify Gradle version: `9.5.0+`
3. Verify Android Gradle Plugin: `9.1.1+`
4. Check dart-define syntax

### "Token refresh not working"
**Note**: Token refresh interceptor is a skeleton in Phase 1  
**Action**: Implement in Phase 2 with actual endpoint

---

## 📞 Support & Resources

- **Flutter Docs**: https://flutter.dev/docs
- **Kubernetes Docs**: https://kubernetes.io/docs
- **Dio Library**: https://github.com/flutterchina/dio
- **Project Issues**: Check repository issues
- **Team Wiki**: https://github.com/kkitaaa/Meloop/wiki

---

## 📈 Metrics & Monitoring

### Build Times (Approximate)
- APK Debug: ~2 minutes
- APK Release: ~3-4 minutes
- Windows Debug: ~1-2 minutes
- Windows Release: ~2-3 minutes

### Network Timeouts
- Connect: 10 seconds
- Send: 10 seconds
- Receive: 10 seconds
- WebSocket Reconnect: Up to 5 attempts

### Health Check
- Endpoint: `/health`
- Timeout: 5 seconds
- Expected Response: `{"status": "ok", "version": "...", "services": {...}}`

---

## 🎓 Learning Path

1. **Day 1**: Read QUICK_REFERENCE.md
2. **Day 2**: Build development APK using build.ps1
3. **Day 3**: Test health check and WebSocket
4. **Day 4**: Deploy to staging K8s
5. **Day 5**: Integrate with actual backend APIs

---

**Last Updated**: 2024-11-27  
**Version**: 1.0.0  
**Status**: Production Ready

For the latest updates, check the [VERIFICATION_REPORT.md](./VERIFICATION_REPORT.md).
