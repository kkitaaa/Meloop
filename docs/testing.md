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

Las pruebas del post-service se ejecutan desde la raíz con `make test-post` (o `mingw32-make test-post` en Windows). Para correr también las pruebas de persistencia contra PostgreSQL, configura `TEST_DATABASE_URL` con una base de pruebas que ya tenga aplicadas las migraciones de Supabase; sin esa variable, Go ejecuta el resto de la suite y omite únicamente la integración PostgreSQL.

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

GitHub Actions ejecuta las pruebas de cada plataforma en workflows separados. Los módulos Go se validan en un job general. Además, un job dedicado levanta PostgreSQL, aplica las migraciones de Supabase y ejecuta las pruebas de integración del post-service con `REQUIRE_TEST_DATABASE_URL=1`, de modo que CI falla si no llega a probar la persistencia real.

## Reglas para nuevas pruebas

- Probar comportamiento observable y códigos HTTP.
- Cubrir al menos un caso exitoso y uno de error por endpoint.
- Evitar mocks incompletos; aislar únicamente dependencias externas.
- No incluir contraseñas, tokens ni datos personales reales en fixtures.
- Mantener las pruebas deterministas y rápidas.
