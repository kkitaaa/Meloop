# Recommendation service

The service consumes platform activity from the durable RabbitMQ queue
`recommendation.events`, bound to `meloop.events` for `like.created`,
`comment.created`, and `post.created`.

Configure `DATABASE_URL`, `REDIS_URL`, and `RABBITMQ_URL` when running the
service. The consumer persists eligible activity in `INTERACCION` and removes
the user's calculated recommendation cache entries from Redis. If either
operation fails, RabbitMQ retries the message; malformed events are sent to
`recommendation.events.dead`.

Events should provide a user identifier (`userId`, `user_id`, `actorId`,
`actor_id`, `authorId`, or `creatorId`) and a post identifier (`postId`,
`post_id`, `publicationId`, or `id_publicacion`). `post.created` may use `id`
as its post identifier. IDs may be strings or JSON numbers. RabbitMQ's
`message_id` or an event-specific ID is used to make persistence idempotent.

Activity is not persisted when `CONFIGURACION_PRIVACIDAD.visibilidad_interacciones`
is `PRIVADO`. Missing privacy settings use the schema's public default. The
profile loader also excludes activity while interactions are private and only
uses recent activity from the last 30 days.

Friend recommendations are checked against `USUARIO`, bidirectional
`BLOQUEO` records, accepted `AMISTAD` records, and inactive `ACCION_MODERACION`
actions in one batched database query. Recommendations from ML, cache, and
backup are all revalidated before being returned. The service keeps the ML
ranking order and takes later eligible candidates from the fetched candidate
batch to satisfy the requested limit.
