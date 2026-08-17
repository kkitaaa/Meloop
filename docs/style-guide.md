# Guía de estilo del proyecto

## Objetivo

Mantener un código uniforme entre los integrantes del equipo y facilitar el trabajo colaborativo en frontend, backend y servicio de ML.

## JavaScript / TypeScript

- Usar ESLint y Prettier con la configuración compartida del repositorio.
- Indentación de 2 espacios.
- Longitud máxima de línea: 100 caracteres.
- Usar `singleQuote: true` y `semi: true`.
- Preferir `const` y `let` sobre `var`.
- Evitar variables sin uso y `console.log` en código de producción.
- Nombrar funciones y variables con `camelCase`.

## Dart / Flutter

- Formatear con `dart format`.
- Preferir `const` en constructores y objetos inmutables.
- Usar nombres descriptivos con `camelCase`.
- Mantener la lógica separada por responsabilidades y evitar código duplicado.
- Ejecutar `dart analyze` antes de abrir un PR.

## Python

- Usar Ruff para linting y Black para formato.
- Indentación de 4 espacios.
- Longitud máxima de línea: 100 caracteres.
- Usar `snake_case` para variables y funciones.
- Preferir tipos explícitos cuando aporten claridad.
- Ejecutar `ruff check` y `black --check` antes de mergear.

## Revisión antes de cada entrega

1. Ejecutar `npm run format`.
2. Ejecutar `npm run lint`.
3. Revisar el diff para asegurar que no haya cambios accidentales.
4. Confirmar que el código cumple con las convenciones del repositorio.
