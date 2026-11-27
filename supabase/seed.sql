BEGIN;

-- La contrasena de las cuentas seed usa bcrypt DefaultCost, igual que auth-service.
-- Credenciales de desarrollo: seed.alex@example.test, seed.bianca@example.test,
-- seed.casey@example.test, seed.diego@example.test y seed.emery@example.test.
INSERT INTO NIVEL (id_nivel, numero, experiencia_requerida)
VALUES
    (1, 1, 0),
    (2, 2, 100),
    (3, 3, 250),
    (4, 4, 500),
    (5, 5, 1000)
ON CONFLICT (id_nivel) DO NOTHING;

INSERT INTO CANCION (id_cancion, spotify_id, titulo, artista_nombre)
VALUES
    ('101', '101', 'Luz de invierno', 'Artista Semilla'),
    ('102', 'seed-track-002', 'Rutas paralelas', 'Los Ejemplos'),
    ('103', 'seed-track-003', 'Cielo abierto', 'Artista Semilla'),
    ('104', '104', 'Pequenos comienzos', 'Banda de Prueba')
ON CONFLICT DO NOTHING;

INSERT INTO USUARIO (
    id_usuario, username, correo, contrasena_hash, id_nivel, experiencia,
    rol, suspendido, es_moderador, fecha_creacion, fecha_ultimo_login
)
VALUES
    ('1', 'alex_seed', 'seed.alex@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 1, 65, 'usuario', FALSE, FALSE, '2026-11-01 09:00:00+00', '2026-11-23 09:00:00+00'),
    ('2', 'bianca_seed', 'seed.bianca@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 2, 180, 'usuario', FALSE, FALSE, '2026-11-02 09:00:00+00', '2026-11-22 12:00:00+00'),
    ('3', 'casey_seed', 'seed.casey@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 3, 420, 'usuario', FALSE, FALSE, '2026-11-03 09:00:00+00', '2026-11-21 15:00:00+00'),
    ('4', 'diego_seed', 'seed.diego@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 4, 760, 'usuario', FALSE, FALSE, '2026-11-04 09:00:00+00', '2026-11-20 18:00:00+00'),
    ('5', 'emery_seed', 'seed.emery@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 5, 1300, 'moderador', FALSE, TRUE, '2026-11-05 09:00:00+00', '2026-11-23 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO RECOMPENSA (
    id_recompensa, tipo, id_nivel, nombre, descripcion, habilitada, disponible
)
VALUES
    ('301', 'INSIGNIA', 2, 'Primeros pasos', 'Insignia de nivel 2', TRUE, TRUE),
    ('302', 'TEMA', 3, 'Tema aurora', 'Tema de perfil de prueba', TRUE, TRUE),
    ('303', 'INSIGNIA', 4, 'Explorador musical', 'Insignia de nivel 4', TRUE, TRUE),
    ('304', 'MARCO', 5, 'Marco celebracion', 'Marco de perfil de prueba', TRUE, TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO PERFIL (
    id_perfil, id_usuario, biografia, foto_perfil, banner, tema, colores, fecha_actualizacion
)
VALUES
    ('401', '1', 'Descubriendo musica nueva.', 'https://example.test/seed/alex.png', 'https://example.test/seed/alex-banner.png', 'midnight', '{"accent":"#5B8DEF"}'::JSONB, '2026-11-20 10:00:00+00'),
    ('402', '2', 'Listas para cada estado de animo.', 'https://example.test/seed/bianca.png', 'https://example.test/seed/bianca-banner.png', 'aurora', '{"accent":"#C06C84"}'::JSONB, '2026-11-20 11:00:00+00'),
    ('403', '3', 'Amante del indie y los conciertos.', 'https://example.test/seed/casey.png', 'https://example.test/seed/casey-banner.png', 'default', '{"accent":"#4A9B7F"}'::JSONB, '2026-11-20 12:00:00+00'),
    ('404', '4', 'Siempre buscando nuevos sonidos.', 'https://example.test/seed/diego.png', 'https://example.test/seed/diego-banner.png', 'sunset', '{"accent":"#E08E45"}'::JSONB, '2026-11-20 13:00:00+00'),
    ('405', '5', 'Moderacion y musica para todos.', 'https://example.test/seed/emery.png', 'https://example.test/seed/emery-banner.png', 'default', '{"accent":"#7559A6"}'::JSONB, '2026-11-20 14:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO PUBLICACION (
    id_publicacion, id_usuario, id_cancion, texto, url_multimedia, fecha_creacion
)
VALUES
    ('101', '1', '101', 'Esta cancion acompana mis caminatas.', NULL, '2026-11-18 10:00:00+00'),
    ('102', '2', '102', 'Recomendacion para escuchar con calma.', NULL, '2026-11-19 10:00:00+00'),
    ('103', '3', '103', 'Un descubrimiento de esta semana.', NULL, '2026-11-20 10:00:00+00'),
    ('104', '4', '104', 'Musica nueva para compartir.', NULL, '2026-11-21 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO INTERACCION (
    id_interaccion, id_usuario, id_publicacion, tipo, texto, fecha_creacion
)
VALUES
    ('601', '2', '101', 'like', NULL, '2026-11-18 11:00:00+00'),
    ('602', '3', '101', 'comentario', 'La agregue a mi lista.', '2026-11-18 12:00:00+00'),
    ('603', '1', '102', 'like', NULL, '2026-11-19 11:00:00+00'),
    ('604', '4', '102', 'comentario', 'Buena recomendacion.', '2026-11-19 12:00:00+00'),
    ('605', '2', '103', 'like', NULL, '2026-11-20 11:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO AMISTAD (
    id_amistad, id_usuario_1, id_usuario_2, estado, fecha_amistad
)
VALUES
    ('701', '1', '2', 'ACEPTADA', '2026-11-10 10:00:00+00'),
    ('702', '1', '3', 'ACEPTADA', '2026-11-11 10:00:00+00'),
    ('703', '2', '4', 'ACEPTADA', '2026-11-12 10:00:00+00'),
    ('704', '4', '5', 'PENDIENTE', NULL)
ON CONFLICT DO NOTHING;

INSERT INTO BLOQUEO (
    id_bloqueo, id_usuario_bloqueador, id_usuario_bloqueado, fecha_creacion
)
VALUES
    ('801', '3', '5', '2026-11-15 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO PREFERENCIA_MUSICAL (id_preferencia, id_usuario, tipo, spotify_id, fecha_creacion)
VALUES
    ('901', '1', 'GENERO', 'indie-pop', '2026-11-18 10:00:00+00'),
    ('902', '1', 'ARTISTA', 'Artista Semilla', '2026-11-18 10:00:00+00'),
    ('903', '1', 'CANCION', '101', '2026-11-18 10:00:00+00'),
    ('904', '2', 'GENERO', 'alternative-rock', '2026-11-19 10:00:00+00'),
    ('905', '2', 'ARTISTA', 'Los Ejemplos', '2026-11-19 10:00:00+00'),
    ('906', '3', 'GENERO', 'indie-pop', '2026-11-20 10:00:00+00'),
    ('907', '4', 'CANCION', '104', '2026-11-21 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO INVENTARIO (
    id_inventario, id_usuario, id_recompensa, equipada, fecha_desbloqueo
)
VALUES
    ('1001', '2', '301', TRUE, '2026-11-12 10:00:00+00'),
    ('1002', '3', '302', TRUE, '2026-11-13 10:00:00+00'),
    ('1003', '4', '303', FALSE, '2026-11-14 10:00:00+00'),
    ('1004', '5', '304', TRUE, '2026-11-15 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO CONFIGURACION_PRIVACIDAD (
    id_privacidad, id_usuario, visibilidad_perfil, visibilidad_publicaciones,
    visibilidad_interacciones, permitir_mensajes, permitir_solicitudes_amistad,
    visibilidad_amigos
)
VALUES
    ('1101', '1', 'PUBLICO', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'TODOS', 'AMIGOS'),
    ('1102', '2', 'PUBLICO', 'PUBLICO', 'PUBLICO', 'TODOS', 'TODOS', 'PUBLICO'),
    ('1103', '3', 'PRIVADO', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'PRIVADO'),
    ('1104', '4', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'TODOS', 'AMIGOS'),
    ('1105', '5', 'PUBLICO', 'PUBLICO', 'PUBLICO', 'TODOS', 'TODOS', 'PUBLICO')
ON CONFLICT DO NOTHING;

INSERT INTO CONFIGURACION_NOTIFICACION (id_configuracion, id_usuario, tipo_notificacion, habilitada)
VALUES
    ('1201', '1', 'LIKE', TRUE),
    ('1202', '1', 'COMENTARIO', TRUE),
    ('1203', '1', 'MENSAJE', FALSE),
    ('1204', '2', 'SOLICITUD_AMISTAD', TRUE),
    ('1205', '3', 'LIKE', TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO NOTIFICACION (
    id_notificacion, id_usuario, tipo, id_objetivo, leida, contador,
    fecha_creacion, id_emisor, id_evento, datos
)
VALUES
    ('1301', '1', 'LIKE', '101', FALSE, 1, '2026-11-18 11:00:00+00', '2', 'seed:event:like:001', '{"source":"seed"}'::JSONB),
    ('1302', '2', 'COMENTARIO', '102', TRUE, 1, '2026-11-19 12:00:00+00', '4', 'seed:event:comment:001', '{"source":"seed"}'::JSONB),
    ('1303', '4', 'SOLICITUD_AMISTAD', '704', FALSE, 1, '2026-11-21 11:00:00+00', '5', 'seed:event:friend-request:001', '{"source":"seed"}'::JSONB)
ON CONFLICT DO NOTHING;

INSERT INTO MENSAJE_CHAT (
    id_mensaje, id_emisor, id_receptor, contenido, fecha_envio, estado
)
VALUES
    ('1401', '1', '2', 'Te comparto una recomendacion musical.', '2026-11-20 16:00:00+00', 'LEIDO'),
    ('1402', '2', '1', 'Gracias, la escuchare hoy.', '2026-11-20 16:05:00+00', 'ENVIADO')
ON CONFLICT DO NOTHING;

INSERT INTO REPORTE (id_reporte, id_usuario, id_objetivo, motivo, estado)
VALUES
    ('1501', '1', 1, 'Reporte ficticio para probar moderacion.', 'PENDIENTE')
ON CONFLICT DO NOTHING;

INSERT INTO ACCION_MODERACION (
    id_accion, id_reporte, id_moderador, id_usuario_afectado,
    tipo_accion, motivo, fecha_creacion
)
VALUES
    ('1502', '1501', '5', '1', 'REVISION', 'Accion de prueba sin sancion.', '2026-11-22 10:00:00+00')
ON CONFLICT DO NOTHING;

COMMIT;
