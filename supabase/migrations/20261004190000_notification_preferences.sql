CREATE UNIQUE INDEX IF NOT EXISTS uq_configuracion_notificacion_usuario_tipo
    ON configuracion_notificacion(id_usuario, tipo_notificacion);