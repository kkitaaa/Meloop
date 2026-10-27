# Meloop en Kubernetes local

Este entorno de desarrollo ejecuta el API Gateway Go, el servicio de recomendaciones que lo conecta con ML, el servicio ML y Redis con almacenamiento persistente. El Gateway expone `/health` y `/recommendations`; las rutas de los otros microservicios de Meloop requieren que esos servicios también estén desplegados.

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
docker build -t meloop-recommendation-service:dev -f services\recommendation-service\Dockerfile .
docker build -t meloop-ml-service:dev .\ml-service

kind load docker-image meloop-api-gateway:dev meloop-recommendation-service:dev meloop-ml-service:dev --name meloop
```

## Desplegar y comprobar

El Namespace debe estar activo antes de crear los recursos que viven en él.
Créalo primero y espera a que Kubernetes lo registre:

```powershell
kubectl apply -f .\k8s\namespace.yaml
kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/meloop --timeout=60s
kubectl apply -f k8s\
kubectl get pods -n meloop -w
kubectl get services,ingress,pvc -n meloop
```

El manifiesto `secret.yaml` crea un Secret vacío; no contiene credenciales. Para habilitar la persistencia en una base Postgres/Supabase externa, crea el Secret con la URL obtenida desde tu gestor de secretos (no la guardes en el repositorio ni en un archivo versionado), luego reinicia el servicio:

```powershell
$env:DATABASE_URL = "<URL_DE_CONEXION>"
kubectl create secret generic meloop-secrets -n meloop --from-literal=DATABASE_URL="$env:DATABASE_URL" --dry-run=client -o yaml | kubectl apply -f -
kubectl rollout restart deployment/recommendation-service -n meloop
```

Configura `$env:DATABASE_URL` desde tu gestor de secretos y no guardes el valor en el repositorio. No vuelvas a aplicar `secret.yaml` después de crear la credencial, porque ese manifiesto representa deliberadamente un Secret vacío.

Prueba el Ingress agregando `127.0.0.1 meloop.local` al archivo `hosts` de Windows (`C:\Windows\System32\drivers\etc\hosts`) y ejecuta:

```powershell
curl.exe http://meloop.local/health
curl.exe http://meloop.local/recommendations/1/music
```

También puedes probar sin Ingress:

```powershell
kubectl port-forward svc/backend 8080:8080 -n meloop
curl.exe http://localhost:8080/health
```

Para verificar la comunicación Gateway → recomendaciones → ML, consulta un endpoint del servicio y revisa que los Pods estén listos:

```powershell
curl.exe http://localhost:8080/recommendations/1/music
kubectl get pods -n meloop
kubectl logs deployment/recommendation-service -n meloop
```

Las recomendaciones personalizadas requieren perfiles en la base de datos; sin `DATABASE_URL` una respuesta sin datos no indica un fallo de conectividad entre los servicios.

## Operación

```powershell
kubectl describe pod <POD> -n meloop
kubectl logs <POD> -n meloop
kubectl delete pod <POD> -n meloop
```

Los Deployments recrean los Pods eliminados. Redis usa un PVC de 1 GiB; comprueba el estado con `kubectl get pvc -n meloop`. Para eliminar el clúster local: `kind delete cluster --name meloop`.

## Imágenes de CI

El workflow `.github/workflows/k8s-images.yml` construye las tres imágenes en cada push a `main` y las publica en GHCR con etiquetas `latest` y el SHA del commit. El flujo local anterior usa imágenes `:dev` para evitar requerir un registry.
