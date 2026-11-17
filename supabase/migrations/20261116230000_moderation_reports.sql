CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id VARCHAR(64),
    target_type TEXT NOT NULL CHECK (target_type IN ('post', 'comment', 'user')),
    target_id TEXT NOT NULL,
    reason TEXT NOT NULL CHECK (
        reason IN (
            'spam',
            'acoso',
            'odio',
            'contenido_sexual',
            'violencia',
            'desinformacion',
            'suplantacion',
            'otro'
        )
    ),
    details TEXT NOT NULL DEFAULT '' CHECK (char_length(details) <= 500),
    status TEXT NOT NULL DEFAULT 'pendiente' CHECK (
        status IN ('pendiente', 'en_revision', 'resuelto', 'escalado')
    ),
    assigned_moderator_id VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp(),
    due_at TIMESTAMPTZ NOT NULL DEFAULT (transaction_timestamp() + INTERVAL '48 hours'),
    resolved_at TIMESTAMPTZ,
    resolved_by VARCHAR(64),
    resolution TEXT,
    content_snapshot JSONB NOT NULL,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    CONSTRAINT reports_other_requires_details
        CHECK (reason <> 'otro' OR length(btrim(details)) > 0),
    CONSTRAINT reports_resolution_fields_consistent
        CHECK (
            (status <> 'resuelto' AND resolved_at IS NULL AND resolved_by IS NULL)
            OR (status = 'resuelto' AND resolved_at IS NOT NULL AND resolved_by IS NOT NULL)
        )
);

CREATE UNIQUE INDEX IF NOT EXISTS reports_open_reporter_target_unique
    ON reports (reporter_id, target_type, target_id)
    WHERE reporter_id IS NOT NULL
      AND status IN ('pendiente', 'en_revision', 'escalado');

CREATE INDEX IF NOT EXISTS reports_moderation_queue_idx
    ON reports (status, due_at, created_at);

CREATE TABLE IF NOT EXISTS report_event_outbox (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    routing_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp(),
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT
);
