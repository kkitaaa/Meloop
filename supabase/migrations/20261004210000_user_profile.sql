-- 20261004210000_user_profile.sql
-- Tabla dedicada para la personalización y diseño del perfil de usuario (RF-07)

CREATE TABLE IF NOT EXISTS PERFIL (
    id_perfil VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL UNIQUE,
    biografia TEXT,
    foto_perfil VARCHAR(255),
    banner VARCHAR(255),
    tema VARCHAR(50) DEFAULT 'default',
    colores JSONB DEFAULT '{}'::JSONB,
    fecha_actualizacion TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_perfil_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_perfil_usuario ON PERFIL(id_usuario);
