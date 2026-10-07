-- Keep current musical preferences timestamped so cold-start popularity can
-- use only preferences updated during the recent activity window.
ALTER TABLE PREFERENCIA_MUSICAL
  ADD COLUMN IF NOT EXISTS fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_preferencia_musical_recency
  ON PREFERENCIA_MUSICAL (fecha_creacion DESC, tipo);

CREATE INDEX IF NOT EXISTS idx_publicacion_recency_song_user
  ON PUBLICACION (fecha_creacion DESC, id_cancion, id_usuario);
