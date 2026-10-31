-- 1. TABLAS INDEPENDIENTES (No dependen de nadie)
CREATE TABLE IF NOT EXISTS NIVEL (
    id_nivel INT PRIMARY KEY,
    numero INT NOT NULL,
    experiencia_requerida INT NOT NULL
);

CREATE TABLE IF NOT EXISTS CANCION (
    id_cancion VARCHAR(64) PRIMARY KEY,
    spotify_id VARCHAR(50),
    titulo VARCHAR(150),
    artista_nombre VARCHAR(150)
);

-- 2. TABLAS PRINCIPALES (Dependen de las independientes)
CREATE TABLE IF NOT EXISTS USUARIO (
    id_usuario VARCHAR(64) PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    correo VARCHAR(100) NOT NULL UNIQUE,
    contrasena_hash VARCHAR(255) NOT NULL,
    id_nivel INT,
    experiencia INT DEFAULT 0,
    CONSTRAINT fk_usuario_nivel FOREIGN KEY (id_nivel) REFERENCES NIVEL(id_nivel) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS RECOMPENSA (
    id_recompensa VARCHAR(64) PRIMARY KEY,
    tipo VARCHAR(50),
    id_nivel INT,
    CONSTRAINT fk_recompensa_nivel FOREIGN KEY (id_nivel) REFERENCES NIVEL(id_nivel) ON DELETE CASCADE
);

-- 3. TABLAS DE CONTENIDO E INTERACCIONES (Dependen de Usuario y Canción)
CREATE TABLE IF NOT EXISTS PUBLICACION (
    id_publicacion VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL,
    id_cancion VARCHAR(64),
    texto TEXT,
    url_multimedia VARCHAR(255),
    fecha_creacion TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_publicacion_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_publicacion_cancion FOREIGN KEY (id_cancion) REFERENCES CANCION(id_cancion) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS INTERACCION (
    id_interaccion VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL,
    id_publicacion VARCHAR(64) NOT NULL,
    tipo VARCHAR(50), -- Ej: 'LIKE', 'COMENTARIO'
    texto TEXT,
    fecha_creacion TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_interaccion_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_interaccion_publicacion FOREIGN KEY (id_publicacion) REFERENCES PUBLICACION(id_publicacion) ON DELETE CASCADE
);

-- 4. TABLAS SOCIALES Y DE COMUNICACIÓN (Dependen de Usuario)
CREATE TABLE IF NOT EXISTS AMISTAD (
    id_amistad VARCHAR(64) PRIMARY KEY,
    id_usuario_1 VARCHAR(64) NOT NULL,
    id_usuario_2 VARCHAR(64) NOT NULL,
    estado VARCHAR(50), -- Ej: 'PENDIENTE', 'ACEPTADA'
    CONSTRAINT fk_amistad_u1 FOREIGN KEY (id_usuario_1) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_amistad_u2 FOREIGN KEY (id_usuario_2) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS MENSAJE_CHAT (
    id_mensaje VARCHAR(64) PRIMARY KEY,
    id_emisor VARCHAR(64) NOT NULL,
    id_receptor VARCHAR(64) NOT NULL,
    contenido TEXT,
    fecha_envio TIMESTAMPTZ DEFAULT NOW(),
    estado VARCHAR(50), -- Ej: 'ENVIADO', 'LEIDO'
    CONSTRAINT fk_mensaje_emisor FOREIGN KEY (id_emisor) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_mensaje_receptor FOREIGN KEY (id_receptor) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS BLOQUEO (
    id_bloqueo VARCHAR(64) PRIMARY KEY,
    id_usuario_bloqueador VARCHAR(64) NOT NULL,
    id_usuario_bloqueado VARCHAR(64) NOT NULL,
    fecha_creacion TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_bloqueo_u1 FOREIGN KEY (id_usuario_bloqueador) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_bloqueo_u2 FOREIGN KEY (id_usuario_bloqueado) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

-- 5. TABLAS DE SISTEMA, GAMIFICACIÓN Y PRIVACIDAD
CREATE TABLE IF NOT EXISTS INVENTARIO (
    id_inventario VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL,
    id_recompensa VARCHAR(64) NOT NULL,
    equipada BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_inventario_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_inventario_recompensa FOREIGN KEY (id_recompensa) REFERENCES RECOMPENSA(id_recompensa) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS PREFERENCIA_MUSICAL (
    id_preferencia VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL,
    tipo VARCHAR(50), -- Ej: 'ARTISTA', 'GENERO'
    spotify_id VARCHAR(50),
    CONSTRAINT fk_preferencia_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS CONFIGURACION_PRIVACIDAD (
    id_privacidad VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL UNIQUE,
    visibilidad_perfil VARCHAR(50) DEFAULT 'PUBLICO',
    visibilidad_publicaciones VARCHAR(50) DEFAULT 'PUBLICO',
    visibilidad_interacciones VARCHAR(50) DEFAULT 'PUBLICO',
    permitir_mensajes VARCHAR(50) DEFAULT 'AMIGOS',
    permitir_solicitudes_amistad VARCHAR(50) DEFAULT 'TODOS',
    CONSTRAINT fk_privacidad_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS CONFIGURACION_NOTIFICACION (
    id_configuracion VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL,
    tipo_notificacion VARCHAR(50),
    habilitada BOOLEAN DEFAULT TRUE,
    CONSTRAINT fk_conf_notif_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS NOTIFICACION (
    id_notificacion VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL,
    tipo VARCHAR(50),
    id_objetivo INT,
    leida BOOLEAN DEFAULT FALSE,
    contador INT DEFAULT 1,
    fecha_creacion TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_notificacion_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

-- 6. TABLAS DE MODERACIÓN
CREATE TABLE IF NOT EXISTS REPORTE (
    id_reporte VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL, -- El denunciante
    id_objetivo INT NOT NULL, -- ID polimórfico (puede ser post o comentario)
    motivo VARCHAR(255),
    estado VARCHAR(50) DEFAULT 'PENDIENTE', 
    CONSTRAINT fk_reporte_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ACCION_MODERACION (
    id_accion VARCHAR(64) PRIMARY KEY,
    id_reporte VARCHAR(64) NOT NULL,
    id_moderador VARCHAR(64) NOT NULL,
    id_usuario_afectado VARCHAR(64) NOT NULL,
    tipo_accion VARCHAR(50),
    motivo VARCHAR(255),
    fecha_creacion TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_accion_reporte FOREIGN KEY (id_reporte) REFERENCES REPORTE(id_reporte) ON DELETE CASCADE,
    CONSTRAINT fk_accion_moderador FOREIGN KEY (id_moderador) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_accion_afectado FOREIGN KEY (id_usuario_afectado) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

-- 7. TABLA DE RECUPERACIÓN
CREATE TABLE IF NOT EXISTS TOKEN_RECUPERACION (
    id_recuperacion VARCHAR(64) PRIMARY KEY,
    id_usuario VARCHAR(64) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expira_en TIMESTAMPTZ NOT NULL,
    usado BOOLEAN NOT NULL DEFAULT FALSE,
    usado_en TIMESTAMPTZ,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_token_recuperacion_usuario FOREIGN KEY (id_usuario) REFERENCES USUARIO(id_usuario) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_token_recuperacion_hash ON TOKEN_RECUPERACION(token_hash);
CREATE INDEX IF NOT EXISTS idx_token_recuperacion_usuario ON TOKEN_RECUPERACION(id_usuario);