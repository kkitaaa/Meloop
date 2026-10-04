# Notification Service

Consumes events from the durable `meloop.events` topic exchange and stores notifications in Supabase PostgreSQL.

## Configuration

Set `DATABASE_URL` to the Supabase PostgreSQL connection string and `RABBITMQ_URL` to the AMQP URL. For local development, the defaults target PostgreSQL and RabbitMQ on `localhost`.

Apply `supabase/migrations/20261003120000_notification_event_delivery.sql` before starting the service.

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
$env:DATABASE_URL = "<Supabase PostgreSQL connection string>"
$env:RABBITMQ_URL = "amqp://guest:guest@localhost:5672/"
go run .
```

Publish a test event to the `meloop.events` exchange with routing key `user.level_up` and payload `{"eventId":"test-level-001","userId":"<existing-user-id>","level":2}`. Verify a row in Supabase with `tipo = 'user.level_up'`, the target `id_usuario`, `leida = false`, and `id_evento = 'test-level-001'`. Publish it again with the same ID to verify that it is not duplicated.