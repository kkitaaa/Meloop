-- B5: dominio canónico de USUARIO.rol y coherencia con es_moderador.
-- Los alias conocidos se normalizan; cualquier otro valor detiene la migración
-- para que sea revisado manualmente.

DO $$
DECLARE
    unknown_roles bigint;
BEGIN
    SELECT count(*)
    INTO unknown_roles
    FROM usuario
    WHERE rol IS NULL
       OR lower(btrim(rol)) NOT IN (
            'usuario', 'moderador', 'admin',
            'administrador', 'administrator', 'moderator'
       );

    IF unknown_roles > 0 THEN
        RAISE EXCEPTION
            'USUARIO contiene % roles desconocidos; se requiere revisión manual antes de aplicar B5',
            unknown_roles;
    END IF;
END $$;

UPDATE usuario
SET rol = CASE lower(btrim(rol))
    WHEN 'administrador' THEN 'admin'
    WHEN 'administrator' THEN 'admin'
    WHEN 'moderator' THEN 'moderador'
    WHEN 'usuario' THEN 'usuario'
    WHEN 'moderador' THEN 'moderador'
    WHEN 'admin' THEN 'admin'
END;

UPDATE usuario
SET es_moderador = (rol = 'moderador');

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'usuario_rol_canonico'
          AND conrelid = 'usuario'::regclass
    ) THEN
        ALTER TABLE usuario
            ADD CONSTRAINT usuario_rol_canonico
            CHECK (rol IN ('usuario', 'moderador', 'admin'));
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'usuario_es_moderador_coherente'
          AND conrelid = 'usuario'::regclass
    ) THEN
        ALTER TABLE usuario
            ADD CONSTRAINT usuario_es_moderador_coherente
            CHECK (es_moderador = (rol = 'moderador'));
    END IF;
END $$;
