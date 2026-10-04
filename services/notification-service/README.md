# Notification Service

Consumes events from the durable `meloop.events` topic exchange and stores notifications in Supabase PostgreSQL.

## Configuration

Set `DATABASE_URL` to the Supabase PostgreSQL connection string, `RABBITMQ_URL` to the AMQP URL, and `AUTH_SERVICE_URL` to the auth service base URL. For local development the HTTP API listens on port `8089`; requests require an `Authorization` session token validated by `auth-service`.

Apply `supabase/migrations/20261003120000_notification_event_delivery.sql` and `supabase/migrations/20261004190000_notification_preferences.sql` before starting the service.

## HTTP API

All routes are available through the API Gateway under `/notifications` and require `Authorization: Bearer <session-token>`.

- `GET /notifications?limit=50&offset=0` returns the user's notifications newest first, including `read` state.
- `PATCH /notifications/:id/read` marks one of the user's notifications as read.
- `PATCH /notifications/read` marks all of the user's notifications as read.
- `GET /notifications/preferences` lists supported notification types; types without a saved preference are enabled by default.
- `PUT /notifications/preferences/:type` accepts `{"enabled":false}` or `{"enabled":true}`.

Example: `PUT /notifications/preferences/post.liked` with body `{"enabled":false}` silences post-like alerts for the authenticated user.

For local development, call the Gateway at `http://localhost:8080/notifications`; the notification service listens directly on `http://localhost:8089/notifications`.

## Supported events

The consumer binds `post.liked`, `comment.created`, `comment.commented`, `comment.replied`, `friend.requested`, `friend.accepted`, `message.sent`, `report.created`, `moderation.action`, `user.level_up`, and `reward.unlocked`.

Messages are JSON. IDs accept camelCase or snake_case. Use `eventId` (or AMQP `MessageId`) as a unique event identifier when identical event bodies can represent different actions. If neither is present, the service derives a stable key from the event and payload.

- `post.liked`: `userId`, `postId`; the service looks up the publication owner.
- `comment.created` / `comment.commented`: `userId`, `postId`; the service looks up the publication owner.
- `comment.replied`: `userId`, `replyToId` (or `parentCommentId`); the service looks up the parent interaction owner.
- `friend.requested`: `sender_id`, `receiver_id`; the receiver is notified.
- `friend.accepted`: `sender_id`, `receiver_id`; the original sender is notified.
- `user.level_up` / `reward.unlocked`: `user_id`; that user is notified.
- Other events: provide `recipientId` or the relevant `receiverId` / `affectedUserId`. Actor fields are `userId`, `actorId`, `senderId`, or `moderatorId`.

Processed messages are acknowledged after the database write. Invalid events are dead-lettered to `notification.events.dead`; transient database failures are requeued.

## Run and test

From this directory:

```powershell
go test ./...
$env:DATABASE_URL = "postgresql://postgres:postgres@localhost:15422/postgres?sslmode=disable"
$env:RABBITMQ_URL = "amqp://guest:guest@localhost:5672/"
$env:AUTH_SERVICE_URL = "http://localhost:8083"
go run .
```

Publish a test event to the `meloop.events` exchange with routing key `user.level_up` and payload `{"eventId":"test-level-001","userId":"<existing-user-id>","level":2}`. Verify a row in Supabase with `tipo = 'user.level_up'`, the target `id_usuario`, `leida = false`, and `id_evento = 'test-level-001'`. Publish it again with the same ID to verify that it is not duplicated.