ALTER TABLE CONFIGURACION_PRIVACIDAD
    ALTER COLUMN permitir_mensajes SET DEFAULT 'AMIGOS',
    ALTER COLUMN permitir_solicitudes_amistad SET DEFAULT 'TODOS';

DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    FOR constraint_name IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'configuracion_privacidad'::regclass
          AND conname IN (
              'ck_privacidad_visibilidad_perfil',
              'ck_privacidad_visibilidad_publicaciones',
              'ck_privacidad_visibilidad_interacciones',
              'ck_privacidad_permitir_mensajes',
              'ck_privacidad_permitir_solicitudes_amistad'
          )
          AND NOT convalidated
    LOOP
        EXECUTE format(
            'ALTER TABLE CONFIGURACION_PRIVACIDAD DROP CONSTRAINT %I',
            constraint_name
        );
    END LOOP;
END $$;

UPDATE CONFIGURACION_PRIVACIDAD
SET visibilidad_perfil = CASE
        WHEN UPPER(BTRIM(COALESCE(visibilidad_perfil, ''))) IN ('PUBLICO', 'AMIGOS', 'PRIVADO')
            THEN UPPER(BTRIM(visibilidad_perfil))
        ELSE 'PUBLICO'
    END,
    visibilidad_publicaciones = CASE
        WHEN UPPER(BTRIM(COALESCE(visibilidad_publicaciones, ''))) IN ('PUBLICO', 'AMIGOS', 'PRIVADO')
            THEN UPPER(BTRIM(visibilidad_publicaciones))
        ELSE 'PUBLICO'
    END,
    visibilidad_interacciones = CASE
        WHEN UPPER(BTRIM(COALESCE(visibilidad_interacciones, ''))) IN ('PUBLICO', 'AMIGOS', 'PRIVADO')
            THEN UPPER(BTRIM(visibilidad_interacciones))
        ELSE 'PUBLICO'
    END,
    permitir_mensajes = CASE
        WHEN UPPER(BTRIM(COALESCE(permitir_mensajes, ''))) IN ('TODOS', 'AMIGOS', 'NADIE')
            THEN UPPER(BTRIM(permitir_mensajes))
        ELSE 'AMIGOS'
    END,
    permitir_solicitudes_amistad = CASE
        WHEN UPPER(BTRIM(COALESCE(permitir_solicitudes_amistad, ''))) IN ('TODOS', 'AMIGOS', 'NADIE')
            THEN UPPER(BTRIM(permitir_solicitudes_amistad))
        ELSE 'TODOS'
    END;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'configuracion_privacidad'::regclass
          AND conname = 'ck_privacidad_visibilidad_perfil'
    ) THEN
        ALTER TABLE CONFIGURACION_PRIVACIDAD
            ADD CONSTRAINT ck_privacidad_visibilidad_perfil
            CHECK (visibilidad_perfil IS NOT NULL AND UPPER(TRIM(visibilidad_perfil)) IN ('PUBLICO', 'AMIGOS', 'PRIVADO'));
    END IF;
    ALTER TABLE CONFIGURACION_PRIVACIDAD
        VALIDATE CONSTRAINT ck_privacidad_visibilidad_perfil;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'configuracion_privacidad'::regclass
          AND conname = 'ck_privacidad_visibilidad_publicaciones'
    ) THEN
        ALTER TABLE CONFIGURACION_PRIVACIDAD
            ADD CONSTRAINT ck_privacidad_visibilidad_publicaciones
            CHECK (visibilidad_publicaciones IS NOT NULL AND UPPER(TRIM(visibilidad_publicaciones)) IN ('PUBLICO', 'AMIGOS', 'PRIVADO'));
    END IF;
    ALTER TABLE CONFIGURACION_PRIVACIDAD
        VALIDATE CONSTRAINT ck_privacidad_visibilidad_publicaciones;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'configuracion_privacidad'::regclass
          AND conname = 'ck_privacidad_visibilidad_interacciones'
    ) THEN
        ALTER TABLE CONFIGURACION_PRIVACIDAD
            ADD CONSTRAINT ck_privacidad_visibilidad_interacciones
            CHECK (visibilidad_interacciones IS NOT NULL AND UPPER(TRIM(visibilidad_interacciones)) IN ('PUBLICO', 'AMIGOS', 'PRIVADO'));
    END IF;
    ALTER TABLE CONFIGURACION_PRIVACIDAD
        VALIDATE CONSTRAINT ck_privacidad_visibilidad_interacciones;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'configuracion_privacidad'::regclass
          AND conname = 'ck_privacidad_permitir_mensajes'
    ) THEN
        ALTER TABLE CONFIGURACION_PRIVACIDAD
            ADD CONSTRAINT ck_privacidad_permitir_mensajes
            CHECK (permitir_mensajes IS NOT NULL AND UPPER(TRIM(permitir_mensajes)) IN ('TODOS', 'AMIGOS', 'NADIE'));
    END IF;
    ALTER TABLE CONFIGURACION_PRIVACIDAD
        VALIDATE CONSTRAINT ck_privacidad_permitir_mensajes;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'configuracion_privacidad'::regclass
          AND conname = 'ck_privacidad_permitir_solicitudes_amistad'
    ) THEN
        ALTER TABLE CONFIGURACION_PRIVACIDAD
            ADD CONSTRAINT ck_privacidad_permitir_solicitudes_amistad
            CHECK (permitir_solicitudes_amistad IS NOT NULL AND UPPER(TRIM(permitir_solicitudes_amistad)) IN ('TODOS', 'AMIGOS', 'NADIE'));
    END IF;
    ALTER TABLE CONFIGURACION_PRIVACIDAD
        VALIDATE CONSTRAINT ck_privacidad_permitir_solicitudes_amistad;
END $$;
