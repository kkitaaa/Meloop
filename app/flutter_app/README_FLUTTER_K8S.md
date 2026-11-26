# Flutter ↔ Kubernetes

Guía de la integración actual entre Flutter y el API Gateway. La configuración
no incluye dominios de ejemplo: staging y producción requieren una URL real del
Gateway.

## Documentación

- [KUBERNETES_INTEGRATION.md](./KUBERNETES_INTEGRATION.md): configuración y
  rutas actualmente soportadas.
- [VERIFICATION_GUIDE.md](./VERIFICATION_GUIDE.md): ejecución de pruebas con
  un clúster accesible.
- [TEST_PLAN.md](./TEST_PLAN.md): comandos y criterios de aceptación.
- [VERIFICATION_REPORT.md](./VERIFICATION_REPORT.md): resultados verificados y
  limitaciones pendientes.
- [Despliegue Kubernetes](../../k8s/README.md): creación del clúster, secretos,
  manifests y comprobaciones del Gateway.

## Resumen

- En desarrollo, Android Emulator usa `http://10.0.2.2:8080` y Windows usa
  `http://localhost:8080`; ambos esperan un port-forward al servicio `backend`.
- `API_BASE_URL` o `K8S_API_GATEWAY_URL` permiten configurar una URL explícita.
  Production exige HTTPS; no se presupone que exista un Ingress público.
- La autenticación se envía al Gateway en `/auth/*`. La sesión se guarda en
  almacenamiento seguro y se renueva de forma deslizante antes de expirar.
- El Gateway publica rutas como `/health`, `/api/demo` y `/api/feed`. Otras
  funcionalidades solo operan si el servicio correspondiente está desplegado.
- Los manifests actuales no despliegan un servidor WebSocket. La app informa
  que WebSocket no está configurado, salvo que se indique una ruta real.

La compilación y los tests locales no demuestran que un clúster real, sus
secretos, su base de datos o todas las funcionalidades estén operativos.
