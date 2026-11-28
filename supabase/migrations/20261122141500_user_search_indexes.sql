CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_usuario_username_trgm
    ON USUARIO USING GIN (username gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_perfil_biografia_trgm
    ON PERFIL USING GIN (biografia gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_bloqueo_blocker_blocked
    ON BLOQUEO (id_usuario_bloqueador, id_usuario_bloqueado);

CREATE INDEX IF NOT EXISTS idx_bloqueo_blocked_blocker
    ON BLOQUEO (id_usuario_bloqueado, id_usuario_bloqueador);
