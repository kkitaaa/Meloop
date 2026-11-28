# Meloop en Kubernetes local

Este entorno de desarrollo ejecuta el API Gateway Go, el servicio de autenticación, el servicio de recomendaciones, ML y Redis. El Gateway es el punto de entrada HTTP de Flutter y enruta `/auth/*` a `auth-service`; las rutas de otros microservicios solo estarán disponibles cuando esos servicios también estén desplegados.

La aplicación actual usa Supabase/Postgres como base de datos. No se crea un PostgreSQL paralelo: el servicio de recomendaciones acepta `DATABASE_URL` opcional para conectarse a una base compatible externa. Sin esa variable, la API funciona sin persistir perfiles ni consumir eventos.

## Requisitos

- Docker Desktop iniciado.
- `kubectl` y `kind` instalados.
- Para Ingress, acceso a GitHub para instalar `ingress-nginx`.

Desde la raíz del repositorio, crea el clúster (el mapeo de puertos permite usar Ingress en `localhost`):

```powershell
kind create cluster --name meloop --config kind\kind-config.yaml
kubectl cluster-info --context kind-meloop
kubectl get nodes
```

Instala el controlador Ingress una vez por clúster y espera a que esté listo:

```powershell
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.1/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx --for=condition=Ready pod --selector=app.kubernetes.io/component=controller --timeout=180s
```

## Construir y cargar imágenes

Los contextos Go deben ser la raíz del repositorio porque sus módulos importan `services/common`.

```powershell
docker build -t meloop-api-gateway:dev -f services\api-gateway\Dockerfile .
docker build -t meloop-auth-service:dev -f services\auth-service\Dockerfile .
docker build -t meloop-recommendation-service:dev -f services\recommendation-service\Dockerfile .
docker build -t meloop-ml-service:dev .\ml-service

kind load docker-image meloop-api-gateway:dev meloop-auth-service:dev meloop-recommendation-service:dev meloop-ml-service:dev --name meloop
```

## Desplegar y comprobar

El Namespace debe estar activo antes de crear los recursos que viven en él.
Créalo primero y espera a que Kubernetes lo registre. Auth requiere una base PostgreSQL/Supabase existente con el esquema del servicio; no se versionan credenciales. Carga la URL de conexión desde tu gestor de secretos antes de desplegar:

```powershell
kubectl apply -f .\k8s\namespace.yaml
kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/meloop --timeout=60s
$env:DATABASE_URL = "<URL_DE_CONEXION>"
kubectl create secret generic meloop-secrets -n meloop --from-literal=DATABASE_URL="$env:DATABASE_URL" --dry-run=client -o yaml | kubectl apply -f -
Remove-Item Env:DATABASE_URL
$manifests = @(
  '.\k8s\configmap.yaml',
  '.\k8s\redis-service.yaml',
  '.\k8s\redis-statefulset.yaml',
  '.\k8s\ml-service.yaml',
  '.\k8s\ml-deployment.yaml',
  '.\k8s\recommendation-service.yaml',
  '.\k8s\recommendation-deployment.yaml',
  '.\k8s\auth-service.yaml',
  '.\k8s\backend-service.yaml',
  '.\k8s\backend-deployment.yaml',
  '.\k8s\ingress.yaml'
)
foreach ($manifest in $manifests) {
  kubectl apply -f $manifest
  if ($LASTEXITCODE -ne 0) { throw "No se pudo aplicar $manifest" }
}
kubectl get pods -n meloop -w
kubectl get services,ingress,pvc -n meloop
```

El Secret también se usa para la conexión de recomendaciones si ese servicio la requiere. No apliques `secret.yaml`: es una plantilla vacía y no debe reemplazar el Secret creado con credenciales.
No apliques todos los YAML indiscriminadamente: `gamification.yaml`, `media.yaml` y `minio.yaml` son opcionales y requieren que configures primero sus Secrets y dependencias.

Prueba el Ingress agregando `127.0.0.1 meloop.local` al archivo `hosts` de Windows (`C:\Windows\System32\drivers\etc\hosts`) y ejecuta:

```powershell
curl.exe http://meloop.local/health
curl.exe http://meloop.local/api/demo
curl.exe http://meloop.local/recommendations/1/music
```

Para usar el API Gateway desde Windows y el emulador Android sin configurar DNS local, deja este port-forward activo:

```powershell
kubectl port-forward svc/backend 8080:8080 -n meloop
curl.exe http://localhost:8080/health
curl.exe http://localhost:8080/api/demo
```

En otra terminal, desde `app\flutter_app`, inicia Flutter. Windows usa `localhost`; Android Emulator usa `10.0.2.2` automáticamente:

```powershell
flutter run -d windows --dart-define=FLUTTER_ENV=development
flutter run -d <id-del-emulador-android> --dart-define=FLUTTER_ENV=development
```

Flutter utiliza `/auth/login`, `/auth/register`, `/auth/refresh` y `/auth/logout` a través del Gateway. El refresh es una renovación deslizante de la sesión activa: debe ocurrir antes de que expire el token y no sustituye un refresh-token OAuth. Las credenciales de sesión se guardan en el almacenamiento seguro del dispositivo.

Para verificar la comunicación Gateway → recomendaciones → ML, consulta un endpoint del servicio y revisa que los Pods estén listos:

```powershell
curl.exe http://localhost:8080/recommendations/1/music
kubectl get pods -n meloop
kubectl logs deployment/recommendation-service -n meloop
kubectl logs deployment/auth-service -n meloop
```

Los manifests actuales no despliegan chat ni notificaciones WebSocket; la app solo intentará esa comprobación si se configura `WEBSOCKET_PATH` con una ruta real habilitada en el Gateway. No se debe interpretar una ruta WebSocket no configurada como una prueba exitosa.

Para compilar staging/production, define una URL real del Gateway mediante `K8S_API_GATEWAY_URL_STAGING` / `K8S_API_GATEWAY_URL_PROD` o el host mediante `K8S_INGRESS_HOST_STAGING` / `K8S_INGRESS_HOST_PROD`. Production exige HTTPS; no hay un dominio público de producción configurado por estos manifests.

## Operación

```powershell
kubectl describe pod <POD> -n meloop
kubectl logs <POD> -n meloop
kubectl delete pod <POD> -n meloop
```

Los Deployments recrean los Pods eliminados. Redis usa un PVC de 1 GiB; comprueba el estado con `kubectl get pvc -n meloop`. Para eliminar el clúster local: `kind delete cluster --name meloop`.

## Imágenes de CI

El workflow `.github/workflows/k8s-images.yml` construye las tres imágenes en cada push a `main` y las publica en GHCR con etiquetas `latest` y el SHA del commit. El flujo local anterior usa imágenes `:dev` para evitar requerir un registry.
