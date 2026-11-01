-- RF-12: permite ocultar la lista de amigos sin ocultar el conteo de amigos en común.
ALTER TABLE CONFIGURACION_PRIVACIDAD
    ADD COLUMN IF NOT EXISTS visibilidad_amigos VARCHAR(50) NOT NULL DEFAULT 'PUBLICO';
