ALTER TABLE notificacion
    ADD COLUMN IF NOT EXISTS id_emisor VARCHAR(64),
    ADD COLUMN IF NOT EXISTS id_evento TEXT,
    ADD COLUMN IF NOT EXISTS datos JSONB NOT NULL DEFAULT '{}'::JSONB;

ALTER TABLE notificacion
    ALTER COLUMN id_objetivo TYPE VARCHAR(64)
    USING id_objetivo::TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_notificacion_emisor'
    ) THEN
        ALTER TABLE notificacion
            ADD CONSTRAINT fk_notificacion_emisor
            FOREIGN KEY (id_emisor) REFERENCES usuario(id_usuario) ON DELETE SET NULL;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_notificacion_id_evento
    ON notificacion(id_evento);

CREATE INDEX IF NOT EXISTS idx_notificacion_usuario_lectura_fecha
    ON notificacion(id_usuario, leida, fecha_creacion DESC);