ALTER TABLE AMISTAD
    ADD COLUMN IF NOT EXISTS fecha_amistad TIMESTAMPTZ;

UPDATE AMISTAD
SET fecha_amistad = NOW()
WHERE estado = 'ACEPTADA'
  AND fecha_amistad IS NULL;

CREATE INDEX IF NOT EXISTS idx_amistad_aceptada_usuario_1_fecha_id
    ON AMISTAD (id_usuario_1, fecha_amistad DESC, id_amistad DESC)
    WHERE estado = 'ACEPTADA';

CREATE INDEX IF NOT EXISTS idx_amistad_aceptada_usuario_2_fecha_id
    ON AMISTAD (id_usuario_2, fecha_amistad DESC, id_amistad DESC)
    WHERE estado = 'ACEPTADA';
