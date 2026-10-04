# Gamification administration

The HTTP API listens on port `8088`, consumes RabbitMQ events in parallel, and reads/writes Supabase PostgreSQL. Set `DATABASE_URL`, `RABBITMQ_URL`, and `AUTH_SERVICE_URL` for the environment.

All administration routes require `Authorization: Bearer <session-token>` and an `ADMIN` role in PostgreSQL. Apply `supabase/migrations/20261004210000_admin_controls.sql` first. The initial administrator is promoted manually through a trusted database session; see `services/user-service/README.md`.

Routes are exposed through the API Gateway:

- `GET /admin/rewards?include_unavailable=true` lists rewards; unavailable rewards are hidden by default.
- `GET /admin/rewards/:id` retrieves one reward.
- `POST /admin/rewards` creates a reward with `nombre`, `tipo`, and `nivel_requerido` (an existing `NIVEL.id_nivel`); `descripcion` is optional and `disponible` defaults to true.
- `PUT /admin/rewards/:id` replaces reward fields, including `disponible`.
- `PATCH /admin/rewards/:id/availability` accepts `{"disponible":false}` or `true`.
- `DELETE /admin/rewards/:id` marks a reward unavailable without deleting inventory references.