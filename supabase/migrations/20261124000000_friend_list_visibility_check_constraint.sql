-- Normalizar valores nulos o inválidos a 'PRIVADO' para evitar exposición de listas con configuración desconocida
UPDATE CONFIGURACION_PRIVACIDAD
SET visibilidad_amigos = 'PRIVADO'
WHERE visibilidad_amigos IS NULL
   OR visibilidad_amigos NOT IN ('PUBLICO', 'AMIGOS', 'PRIVADO');

-- Incorporar restricción CHECK específica para la tabla CONFIGURACION_PRIVACIDAD
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_configuracion_privacidad_visibilidad_amigos'
          AND conrelid = 'configuracion_privacidad'::regclass
    ) THEN
        ALTER TABLE CONFIGURACION_PRIVACIDAD
            ADD CONSTRAINT chk_configuracion_privacidad_visibilidad_amigos
            CHECK (visibilidad_amigos IN ('PUBLICO', 'AMIGOS', 'PRIVADO'));
    END IF;
END $$;
