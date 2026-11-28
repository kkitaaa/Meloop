# Notification Service

Consumes events from the durable `meloop.events` topic exchange and stores notifications in Supabase PostgreSQL. It also exposes an authenticated REST API for each user's notification inbox.

## Configuration

Set `DATABASE_URL` to the Supabase PostgreSQL connection string, `RABBITMQ_URL` to the AMQP URL, and `AUTH_SERVICE_URL` to the Auth service URL. The HTTP server listens on `PORT` (default `8088`). For local development, the defaults target PostgreSQL, RabbitMQ, and Auth on `localhost`.

Apply `supabase/migrations/20261003120000_notification_event_delivery.sql` before starting the service.

## HTTP API

All `/notifications` routes require a Bearer session token validated by Auth:

- `GET /notifications?limit=20&offset=0` lists the authenticated user's notifications, newest first. `limit` is capped at 100.
- `PATCH /notifications/{id}/read` marks one of that user's notifications as read. IDs belonging to another user are not exposed.
- `PATCH /notifications/read` marks all unread notifications for the authenticated user as read.

Notification preferences remain managed through `GET` and `PATCH /users/me/notifications/settings` in user-service. Event generation checks those persisted settings before inserting an alert. If a preference row does not yet exist, notifications remain enabled by default, matching user-service defaults.

## Supported events

The consumer binds `post.liked`, `comment.created`, `comment.commented`, `comment.replied`, `friend.requested`, `friend.accepted`, `message.sent`, `report.created`, `moderation.action`, `user.level_up`, and `reward.unlocked`.

Messages are JSON. IDs accept camelCase or snake_case. Use `eventId` (or AMQP `MessageId`) as a unique event identifier when identical event bodies can represent different actions. If neither is present, the service derives a stable key from the event and payload.

- `post.liked`: `userId`, `postId`; the service looks up the publication owner. The post service publishes a unique `event_id`/AMQP `MessageId` and persistent message so redelivery is idempotent.
- `comment.created`: the post service payload uses `comment_author_id`, `target_author_id`, `target_type`, and `target_id`. For `target_type: "post"`, the service looks up the publication owner; for `target_type: "comment"`, it creates a reply notification for the parent interaction owner. Placeholder target author IDs are ignored and resolved from PostgreSQL.
- `comment.commented`: `userId`, `postId`; the service looks up the publication owner.
- `comment.replied`: `userId`, `replyToId` (or `parentCommentId`); the service looks up the parent interaction owner.
- `friend.requested`: `sender_id`, `receiver_id`; the receiver is notified.
- `friend.accepted`: `sender_id`, `receiver_id`; the original sender is notified.
- `user.level_up` / `reward.unlocked`: `user_id`; that user is notified.
- Other events: provide `recipientId` or the relevant `receiverId` / `affectedUserId`. Actor fields are `userId`, `actorId`, `senderId`, or `moderatorId`.

Processed messages are acknowledged after the database write or when the matching preference is disabled. Invalid events and permanent database constraint failures are dead-lettered to `notification.events.dead`; transient database failures are requeued.

## Run and test

From this directory:

```powershell
go test ./...
$env:DATABASE_URL = "<Supabase PostgreSQL connection string>"
$env:RABBITMQ_URL = "amqp://guest:guest@localhost:5672/"
$env:AUTH_SERVICE_URL = "http://localhost:8083"
go run .
```

### End-to-end k6 test

The repository-root `notification_k6_test.js` registers two temporary users, publishes all supported notification event routes through RabbitMQ's management HTTP API, checks delivery in the authenticated inbox, and republishes one event to verify deduplication. It requires a running API Gateway, Auth and Notification services, RabbitMQ with the management plugin, and a database with the notification migration applied.

Run it from the repository root in PowerShell:

```powershell
$env:RABBITMQ_USER = "<RabbitMQ user>"
$env:RABBITMQ_PASSWORD = "<RabbitMQ password>"
k6 run .\notification_k6_test.js
```

Optional settings: `BASE_URL` (default `http://localhost:8080`), `AUTH_URL` (default `http://localhost:8083`), `RABBITMQ_MANAGEMENT_URL` (default `http://localhost:15672`), `RABBITMQ_VHOST` (default `/`), `VUS` and `ITERATIONS` (default one end-to-end flow), and `POLL_TIMEOUT_SECONDS` (default 15). The test creates users and notification records in the configured test database.

Publish a test event to the `meloop.events` exchange with routing key `user.level_up` and payload `{"eventId":"test-level-001","userId":"<existing-user-id>","level":2}`. Verify a row in Supabase with `tipo = 'user.level_up'`, the target `id_usuario`, `leida = false`, and `id_evento = 'test-level-001'`. Publish it again with the same ID to verify that it is not duplicated.
