ALTER TABLE usuario
    ADD COLUMN IF NOT EXISTS rol VARCHAR(20) NOT NULL DEFAULT 'USER',
    ADD COLUMN IF NOT EXISTS suspendido BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS puede_moderar BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'ck_usuario_rol' AND conrelid = 'usuario'::regclass
    ) THEN
        ALTER TABLE usuario
            ADD CONSTRAINT ck_usuario_rol CHECK (rol IN ('USER', 'ADMIN'));
    END IF;
END $$;

ALTER TABLE recompensa
    ADD COLUMN IF NOT EXISTS nombre VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS descripcion TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS disponible BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE recompensa
SET nombre = tipo
WHERE nombre = '' AND tipo IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_usuario_rol ON usuario(rol);
CREATE INDEX IF NOT EXISTS idx_recompensa_disponible_nivel ON recompensa(disponible, id_nivel);