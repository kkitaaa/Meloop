# Media service: integración de cargas

El servicio crea URLs prefirmadas `PUT` para imágenes JPEG/PNG de hasta 5 MiB. Los
tipos admitidos son `profile_photo`, `banner` y `post_attachment`; las claves se
separan por propietario (`profiles/{userID}/`, `banners/{userID}/` y
`posts/{userID}/`). La verificación comprueba el objeto en MinIO y el user-service
rechaza asociar a un perfil una foto o banner de otro usuario.

## URL pública de MinIO

`MINIO_ENDPOINT` es el endpoint interno usado por el servicio para consultas a
MinIO. `MINIO_PUBLIC_ENDPOINT` es el host accesible desde el cliente y se usa
para firmar URLs temporales de carga y descarga. Debe ser una URL absoluta
HTTP(S) resoluble por el navegador/app; no configure el nombre interno
`minio:9000`.

En Docker Compose el valor predeterminado para desarrollo local es
`http://localhost:9000` y puede sobrescribirse:

```powershell
$env:MINIO_PUBLIC_ENDPOINT = "https://media.example.com"
docker compose up -d --build media-service
```

En Kubernetes, `k8s/ingress.yaml` publica el API de MinIO como
`http://minio.meloop.local`; `k8s/media.yaml` usa ese host para firmar URLs.
Configure DNS/hosts e Ingress para que `minio.meloop.local` sea accesible desde
el dispositivo cliente. Para producción, sustituya el host por el dominio TLS
real y habilite HTTPS en el Ingress y en `MINIO_PUBLIC_ENDPOINT`.

El bucket permanece privado. Después de guardar el `object_key` permanente en
un perfil o publicación, la app solicita `GET /media/download-url?key=<object_key>`
con su token de sesión. El servicio verifica que el objeto exista y responde
con `download_url`, una URL GET firmada que vence en 15 minutos. La app debe
solicitar una URL nueva cuando expire; no persista `download_url` en la base de
datos.

## Ejecutar la prueba de integración

Levante MinIO y los servicios necesarios. La prueba requiere dos sesiones JWT de
usuarios distintos, credenciales con permiso para actualizar perfil y crear una
publicación, y un archivo JPEG/PNG menor o igual a 5 MiB:

```powershell
$env:MELOOP_TEST_TOKEN = "<JWT del usuario propietario>"
$env:MELOOP_TEST_USER_ID = "<ID del usuario propietario>"
$env:MELOOP_OTHER_TEST_TOKEN = "<JWT de otro usuario>"
$env:MELOOP_OTHER_TEST_USER_ID = "<ID del otro usuario>"

.\services\media-service\tests\media_upload_integration.ps1 `
  -BearerToken $env:MELOOP_TEST_TOKEN `
  -UserID $env:MELOOP_TEST_USER_ID `
  -OtherUserBearerToken $env:MELOOP_OTHER_TEST_TOKEN `
  -OtherUserID $env:MELOOP_OTHER_TEST_USER_ID `
  -PublicMediaEndpoint "http://localhost:9000"
```

El script solicita URLs prefirmadas para una foto de perfil y un adjunto, sube
ambos archivos mediante `PUT`, comprueba existencia y tamaño, descarga ambos
con URLs GET firmadas, asocia la foto al perfil, crea una publicación con la
clave estable del adjunto y comprueba que las URLs firmadas usen el host
público. También comprueba los rechazos de MIME no permitido, tamaño
excedido y asociación de una foto perteneciente a otro usuario.

Si post-service no está publicado detrás del gateway, indique la ruta compatible
con el endpoint de creación mediante `-PostApiUrl`. El script modifica el perfil
de prueba y crea una publicación; use cuentas y archivos desechables.

Pruebas unitarias del servicio:

```powershell
Push-Location services\media-service
go test ./...
Pop-Location

Push-Location services\user-service
go test ./services
Pop-Location
```
