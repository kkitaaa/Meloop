#!/bin/sh
set -e

echo "==> Esperando disponibilidad del servicio MinIO..."
until /usr/bin/mc alias set local http://minio:9000 "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}"; do
  echo "MinIO aún no responde, reintentando en 2 segundos..."
  sleep 2
done

BUCKET_NAME="${MINIO_MEDIA_BUCKET:-meloop-media}"

echo "==> Creando bucket '${BUCKET_NAME}' si no existe..."
/usr/bin/mc mb "local/${BUCKET_NAME}" --ignore-existing

echo "==> Configurando bucket '${BUCKET_NAME}' como estrictamente privado (sin acceso anónimo)..."
/usr/bin/mc anonymous set none "local/${BUCKET_NAME}"

echo "==> Creando o actualizando usuario para media-service ('${MINIO_MEDIA_USER}')..."
/usr/bin/mc admin user add local "${MINIO_MEDIA_USER}" "${MINIO_MEDIA_PASSWORD}" || true

echo "==> Registrando política IAM para media-service..."
/usr/bin/mc admin policy create local media-service-policy /infrastructure/minio/media-service-policy.json || \
/usr/bin/mc admin policy update local media-service-policy /infrastructure/minio/media-service-policy.json || \
/usr/bin/mc admin policy add local media-service-policy /infrastructure/minio/media-service-policy.json || true

echo "==> Asociando política al usuario '${MINIO_MEDIA_USER}'..."
/usr/bin/mc admin policy attach local media-service-policy --user "${MINIO_MEDIA_USER}" || \
/usr/bin/mc admin policy set local media-service-policy user="${MINIO_MEDIA_USER}" || true

echo "==> Inicialización de MinIO completada exitosamente."
exit 0
