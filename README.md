# Meloop

Proyecto Meloop.

## Estándar de calidad y estilo

Este repositorio define una convención compartida para frontend, backend y servicio de ML con el objetivo de mantener un código consistente entre todos los integrantes del equipo.

### Herramientas base

- ESLint + Prettier para JavaScript/TypeScript.
- Dart formatter + análisis de Flutter para frontend.
- Ruff + Black para Python del servicio ML.

### Comandos principales

```bash
npm run format
npm run lint
```

### Dependencias recomendadas

```bash
python -m pip install ruff black
```

Si se trabaja con Flutter, instalar el SDK de Dart/Flutter en la máquina del desarrollador para habilitar `dart analyze` y `dart format`.

### Reglas básicas

- Usar 2 espacios para JavaScript/TypeScript y 4 espacios para Python.
- Mantener líneas de hasta 100 caracteres.
- Usar comillas simples en TypeScript y JavaScript cuando la configuración lo permita.
- Preferir `const`/`final` para valores inmutables.
- Evitar variables sin uso y logs de depuración en código final.
- Usar nombres de funciones descriptivos y camelCase/ snake_case según el lenguaje.
- Ejecutar el formateo antes de cada pull request.

### Documentación adicional

- [docs/style-guide.md](docs/style-guide.md)
