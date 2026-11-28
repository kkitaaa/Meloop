# Estrategia de pruebas

La estrategia comienza con pruebas rápidas y aisladas, y deja preparada la ampliación hacia integración y end-to-end.

## Capas

- **Unitarias:** validan lógica y controladores sin depender de infraestructura externa.
- **API básicas:** usan `httptest` en Go y `TestClient` en FastAPI para comprobar estados HTTP y respuestas.
- **Integración:** conectan servicios reales con Redis, RabbitMQ, Docker Compose y bases de datos cuando corresponda.
- **End-to-end futura:** validará flujos completos desde el gateway o Flutter.

## Ejecutar localmente

### Go

Desde cada módulo:

```powershell
go test ./...
```

Las pruebas del post-service se ejecutan desde la raíz con `make test-post` (o `mingw32-make test-post` en Windows). Para correr también su prueba de persistencia contra PostgreSQL, configura `TEST_DATABASE_URL` con una base de pruebas que ya tenga aplicadas las migraciones de Supabase; sin esa variable, Go ejecuta el resto de la suite y omite únicamente la integración PostgreSQL.

Para revisar todos los módulos desde la raíz:

```powershell
Get-ChildItem services -Directory | ForEach-Object { if (Test-Path (Join-Path $_.FullName 'go.mod')) { Push-Location $_.FullName; go test ./...; Pop-Location } }
```

### ML service

```powershell
.\.venv\Scripts\Activate.ps1
pip install -r ml-service/requirements-dev.txt
pytest ml-service/tests
```

### Flutter

```powershell
cd app/flutter_app
flutter pub get
flutter test
flutter analyze
```

## CI

GitHub Actions ejecuta las pruebas de cada plataforma en workflows separados. Los módulos Go se validan en un único job: para cada módulo bajo `services/`, CI revisa el formato con `gofmt` y ejecuta `go test ./...`; los fallos de un módulo no impiden revisar los demás. Las pruebas unitarias no requieren Docker. Las pruebas de integración que necesiten Redis, RabbitMQ u otros servicios deberán levantar explícitamente los servicios requeridos con Docker Compose.

## Reglas para nuevas pruebas

- Probar comportamiento observable y códigos HTTP.
- Cubrir al menos un caso exitoso y uno de error por endpoint.
- Evitar mocks incompletos; aislar únicamente dependencias externas.
- No incluir contraseñas, tokens ni datos personales reales en fixtures.
- Mantener las pruebas deterministas y rápidas.
