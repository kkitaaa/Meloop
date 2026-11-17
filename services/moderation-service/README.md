# Moderation Service

The service accepts user reports for posts, comments, and user profiles at
`POST /v1/moderation/reports`. It requires a valid bearer session in
`Authorization`; moderation validates it with auth-service and derives the
reporter ID from that session. A client-supplied `X-User-ID` is not trusted.

```http
POST /v1/moderation/reports
Content-Type: application/json
Authorization: Bearer <session-token>

{
  "target_type": "post",
  "target_id": "post-123",
  "reason": "spam",
  "details": ""
}
```

`target_type` is `post`, `comment`, or `user`. The allowed reasons are `spam`,
`acoso`, `odio`, `contenido_sexual`, `violencia`, `desinformacion`,
`suplantacion`, and `otro`; `details` is required for `otro` and may contain at
most 500 characters. A successful response is `201` with the report ID,
`pendiente` status, and server-generated `due_at` timestamp. Self-reports return
`422 CANNOT_REPORT_SELF`, invisible or missing targets return `404`,
duplicate open reports return `409 ALREADY_REPORTED`, and rate-limited users
return `429`.

## Configuration

- `DATABASE_URL`: PostgreSQL/Supabase connection string. Apply the migration in
  `supabase/migrations/20261116230000_moderation_reports.sql` before startup.
- `REDIS_URL`: Redis connection URL used for the rolling limit of 10 reports per
  reporter per hour.
- `RABBITMQ_URL`: RabbitMQ connection URL for the `meloop.events` topic exchange.
- `POST_SERVICE_URL`: post-service base URL. Defaults to `http://localhost:8084`.
- `USER_SERVICE_URL`: user-service base URL. Defaults to `http://localhost:8082`.
- `AUTH_SERVICE_URL`: auth-service base URL. Defaults to `http://localhost:8083`.
- `INTERNAL_SERVICE_TOKEN`: required shared token used to authenticate moderation
  lookups to post-service. Configure the same value in post-service and moderation-service.
- `PORT`: HTTP port, defaults to `8089`.

Reports are stored with server-generated timestamps and a 48-hour deadline.
Report creation and the `report.created` event outbox record are committed in
one database transaction; an asynchronous dispatcher publishes pending events
to RabbitMQ and retries failures.
