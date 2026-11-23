# Informe de verificación Flutter-Kubernetes

## Entorno comprobado

La integración se probó en Windows contra un clúster Kubernetes local `kind`
(`meloop`) mediante el API Gateway en `http://localhost:8080`. El port-forward
del Gateway debe permanecer activo para repetir estas pruebas. Este resultado
no representa una validación contra un clúster o dominio de producción.

## Resultados

- `flutter analyze`: aprobado.
- `flutter test`: aprobado.
- `go test -p 1 ./...` en `services/recommendation-service`: aprobado.
- `flutter test integration_test\api_gateway_test.dart -d windows` con
  `MELOOP_RUN_AUTH_TESTS=1`: las cuatro pruebas pasaron.
- La integración Windows verificó GET/POST demo, feed, recomendaciones y el
  ciclo real de registro, inicio de sesión, validación, renovación y cierre de
  sesión a través del Gateway de Kubernetes.
- `/recommendations/1/music` respondió HTTP 200 con origen `ml` y el modelo
  `profile_compatibility`. La prueba encontró que los arreglos vacíos se
  serializaban como `null`, lo que causaba HTTP 422 en ML; se corrigió la
  serialización y se agregaron pruebas del contrato.
- El Ingress local `meloop.local` respondió HTTP 200 para `/health` y para
  recomendaciones.
- Los builds locales de Windows y Android debug se habían completado; en esta
  ejecución se validó en vivo solo Windows.

## Pendiente / límites

- No hay un servicio de WebSocket desplegado en estos manifests, por lo que esa
  conexión no se pudo validar.
- No están desplegados aquí los servicios de usuarios/social ni todos los
  servicios opcionales. La prueba cubre únicamente los servicios actualmente
  listos en el clúster.
- No se validó Android en vivo ni un endpoint público de staging/producción.
  Production exige un Gateway HTTPS real configurado en el entorno.

## Evaluación

La ruta Windows → API Gateway → auth/recommendations/ML está verificada en el
clúster local. La aceptación de WebSocket, Android en vivo y servicios no
desplegados sigue pendiente; no se debe presentar esta prueba local como
validación completa de producción.
