BEGIN;

-- La contrasena de las cuentas seed usa bcrypt DefaultCost, igual que auth-service.
-- IDs alfanumericos de 32 caracteres para coincidir con los que genera auth-service.
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

-- Reemplaza las cuentas seed anteriores (IDs numericos) y sus datos
-- dependientes para evitar colisiones por correo/username al actualizar el seed.
DELETE FROM USUARIO
WHERE correo IN (
    'seed.alex@example.test',
    'seed.bianca@example.test',
    'seed.casey@example.test',
    'seed.diego@example.test',
    'seed.emery@example.test'
)
OR username IN (
    'alex_seed',
    'bianca_seed',
    'casey_seed',
    'diego_seed',
    'emery_seed'
);

INSERT INTO USUARIO (
    id_usuario, username, correo, contrasena_hash, id_nivel, experiencia,
    rol, suspendido, es_moderador, fecha_creacion, fecha_ultimo_login
)
VALUES
    ('a0000000000000000000000000000001', 'alex_seed', 'seed.alex@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 1, 65, 'usuario', FALSE, FALSE, '2026-11-01 09:00:00+00', '2026-11-23 09:00:00+00'),
    ('b0000000000000000000000000000002', 'bianca_seed', 'seed.bianca@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 2, 180, 'usuario', FALSE, FALSE, '2026-11-02 09:00:00+00', '2026-11-22 12:00:00+00'),
    ('c0000000000000000000000000000003', 'casey_seed', 'seed.casey@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 3, 420, 'usuario', FALSE, FALSE, '2026-11-03 09:00:00+00', '2026-11-21 15:00:00+00'),
    ('d0000000000000000000000000000004', 'diego_seed', 'seed.diego@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 4, 760, 'usuario', FALSE, FALSE, '2026-11-04 09:00:00+00', '2026-11-20 18:00:00+00'),
    ('e0000000000000000000000000000005', 'emery_seed', 'seed.emery@example.test', '$2a$10$ChkSRVWFRcMkjX4zwmgbDu6j.7VsLia8VXwHLbIzjwjwKa9Rgu2UO', 5, 1300, 'moderador', FALSE, TRUE, '2026-11-05 09:00:00+00', '2026-11-23 10:00:00+00')
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
    ('401', 'a0000000000000000000000000000001', 'Descubriendo musica nueva.', 'https://example.test/seed/alex.png', 'https://example.test/seed/alex-banner.png', 'midnight', '{"accent":"#5B8DEF"}'::JSONB, '2026-11-20 10:00:00+00'),
    ('402', 'b0000000000000000000000000000002', 'Listas para cada estado de animo.', 'https://example.test/seed/bianca.png', 'https://example.test/seed/bianca-banner.png', 'aurora', '{"accent":"#C06C84"}'::JSONB, '2026-11-20 11:00:00+00'),
    ('403', 'c0000000000000000000000000000003', 'Amante del indie y los conciertos.', 'https://example.test/seed/casey.png', 'https://example.test/seed/casey-banner.png', 'default', '{"accent":"#4A9B7F"}'::JSONB, '2026-11-20 12:00:00+00'),
    ('404', 'd0000000000000000000000000000004', 'Siempre buscando nuevos sonidos.', 'https://example.test/seed/diego.png', 'https://example.test/seed/diego-banner.png', 'sunset', '{"accent":"#E08E45"}'::JSONB, '2026-11-20 13:00:00+00'),
    ('405', 'e0000000000000000000000000000005', 'Moderacion y musica para todos.', 'https://example.test/seed/emery.png', 'https://example.test/seed/emery-banner.png', 'default', '{"accent":"#7559A6"}'::JSONB, '2026-11-20 14:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO PUBLICACION (
    id_publicacion, id_usuario, id_cancion, texto, url_multimedia, fecha_creacion
)
VALUES
    ('101', 'a0000000000000000000000000000001', '101', 'Esta cancion acompana mis caminatas.', NULL, '2026-11-18 10:00:00+00'),
    ('102', 'b0000000000000000000000000000002', '102', 'Recomendacion para escuchar con calma.', NULL, '2026-11-19 10:00:00+00'),
    ('103', 'c0000000000000000000000000000003', '103', 'Un descubrimiento de esta semana.', NULL, '2026-11-20 10:00:00+00'),
    ('104', 'd0000000000000000000000000000004', '104', 'Musica nueva para compartir.', NULL, '2026-11-21 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO INTERACCION (
    id_interaccion, id_usuario, id_publicacion, tipo, texto, fecha_creacion
)
VALUES
    ('601', 'b0000000000000000000000000000002', '101', 'like', NULL, '2026-11-18 11:00:00+00'),
    ('602', 'c0000000000000000000000000000003', '101', 'comentario', 'La agregue a mi lista.', '2026-11-18 12:00:00+00'),
    ('603', 'a0000000000000000000000000000001', '102', 'like', NULL, '2026-11-19 11:00:00+00'),
    ('604', 'd0000000000000000000000000000004', '102', 'comentario', 'Buena recomendacion.', '2026-11-19 12:00:00+00'),
    ('605', 'b0000000000000000000000000000002', '103', 'like', NULL, '2026-11-20 11:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO AMISTAD (
    id_amistad, id_usuario_1, id_usuario_2, estado, fecha_amistad
)
VALUES
    ('701', 'a0000000000000000000000000000001', 'b0000000000000000000000000000002', 'ACEPTADA', '2026-11-10 10:00:00+00'),
    ('702', 'a0000000000000000000000000000001', 'c0000000000000000000000000000003', 'ACEPTADA', '2026-11-11 10:00:00+00'),
    ('703', 'b0000000000000000000000000000002', 'd0000000000000000000000000000004', 'ACEPTADA', '2026-11-12 10:00:00+00'),
    ('704', 'd0000000000000000000000000000004', 'e0000000000000000000000000000005', 'PENDIENTE', NULL)
ON CONFLICT DO NOTHING;

INSERT INTO BLOQUEO (
    id_bloqueo, id_usuario_bloqueador, id_usuario_bloqueado, fecha_creacion
)
VALUES
    ('801', 'c0000000000000000000000000000003', 'e0000000000000000000000000000005', '2026-11-15 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO PREFERENCIA_MUSICAL (id_preferencia, id_usuario, tipo, spotify_id, fecha_creacion)
VALUES
    ('901', 'a0000000000000000000000000000001', 'GENERO', 'indie-pop', '2026-11-18 10:00:00+00'),
    ('902', 'a0000000000000000000000000000001', 'ARTISTA', 'Artista Semilla', '2026-11-18 10:00:00+00'),
    ('903', 'a0000000000000000000000000000001', 'CANCION', '101', '2026-11-18 10:00:00+00'),
    ('904', 'b0000000000000000000000000000002', 'GENERO', 'alternative-rock', '2026-11-19 10:00:00+00'),
    ('905', 'b0000000000000000000000000000002', 'ARTISTA', 'Los Ejemplos', '2026-11-19 10:00:00+00'),
    ('906', 'c0000000000000000000000000000003', 'GENERO', 'indie-pop', '2026-11-20 10:00:00+00'),
    ('907', 'd0000000000000000000000000000004', 'CANCION', '104', '2026-11-21 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO INVENTARIO (
    id_inventario, id_usuario, id_recompensa, equipada, fecha_desbloqueo
)
VALUES
    ('1001', 'b0000000000000000000000000000002', '301', TRUE, '2026-11-12 10:00:00+00'),
    ('1002', 'c0000000000000000000000000000003', '302', TRUE, '2026-11-13 10:00:00+00'),
    ('1003', 'd0000000000000000000000000000004', '303', FALSE, '2026-11-14 10:00:00+00'),
    ('1004', 'e0000000000000000000000000000005', '304', TRUE, '2026-11-15 10:00:00+00')
ON CONFLICT DO NOTHING;

INSERT INTO CONFIGURACION_PRIVACIDAD (
    id_privacidad, id_usuario, visibilidad_perfil, visibilidad_publicaciones,
    visibilidad_interacciones, permitir_mensajes, permitir_solicitudes_amistad,
    visibilidad_amigos
)
VALUES
    ('1101', 'a0000000000000000000000000000001', 'PUBLICO', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'TODOS', 'AMIGOS'),
    ('1102', 'b0000000000000000000000000000002', 'PUBLICO', 'PUBLICO', 'PUBLICO', 'TODOS', 'TODOS', 'PUBLICO'),
    ('1103', 'c0000000000000000000000000000003', 'PRIVADO', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'PRIVADO'),
    ('1104', 'd0000000000000000000000000000004', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'AMIGOS', 'TODOS', 'AMIGOS'),
    ('1105', 'e0000000000000000000000000000005', 'PUBLICO', 'PUBLICO', 'PUBLICO', 'TODOS', 'TODOS', 'PUBLICO')
ON CONFLICT DO NOTHING;

INSERT INTO CONFIGURACION_NOTIFICACION (id_configuracion, id_usuario, tipo_notificacion, habilitada)
VALUES
    ('1201', 'a0000000000000000000000000000001', 'LIKE', TRUE),
    ('1202', 'a0000000000000000000000000000001', 'COMENTARIO', TRUE),
    ('1203', 'a0000000000000000000000000000001', 'MENSAJE', FALSE),
    ('1204', 'b0000000000000000000000000000002', 'SOLICITUD_AMISTAD', TRUE),
    ('1205', 'c0000000000000000000000000000003', 'LIKE', TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO NOTIFICACION (
    id_notificacion, id_usuario, tipo, id_objetivo, leida, contador,
    fecha_creacion, id_emisor, id_evento, datos
)
VALUES
    ('1301', 'a0000000000000000000000000000001', 'LIKE', '101', FALSE, 1, '2026-11-18 11:00:00+00', 'b0000000000000000000000000000002', 'seed:event:like:001', '{"source":"seed"}'::JSONB),
    ('1302', 'b0000000000000000000000000000002', 'COMENTARIO', '102', TRUE, 1, '2026-11-19 12:00:00+00', 'd0000000000000000000000000000004', 'seed:event:comment:001', '{"source":"seed"}'::JSONB),
    ('1303', 'd0000000000000000000000000000004', 'SOLICITUD_AMISTAD', '704', FALSE, 1, '2026-11-21 11:00:00+00', 'e0000000000000000000000000000005', 'seed:event:friend-request:001', '{"source":"seed"}'::JSONB)
ON CONFLICT DO NOTHING;

INSERT INTO MENSAJE_CHAT (
    id_mensaje, id_emisor, id_receptor, contenido, fecha_envio, estado
)
VALUES
    ('1401', 'a0000000000000000000000000000001', 'b0000000000000000000000000000002', 'Te comparto una recomendacion musical.', '2026-11-20 16:00:00+00', 'LEIDO'),
    ('1402', 'b0000000000000000000000000000002', 'a0000000000000000000000000000001', 'Gracias, la escuchare hoy.', '2026-11-20 16:05:00+00', 'ENVIADO')
ON CONFLICT DO NOTHING;

INSERT INTO REPORTE (id_reporte, id_usuario, id_objetivo, motivo, estado)
VALUES
    ('1501', 'a0000000000000000000000000000001', 1, 'Reporte ficticio para probar moderacion.', 'PENDIENTE')
ON CONFLICT DO NOTHING;

INSERT INTO ACCION_MODERACION (
    id_accion, id_reporte, id_moderador, id_usuario_afectado,
    tipo_accion, motivo, fecha_creacion
)
VALUES
    ('1502', '1501', 'e0000000000000000000000000000005', 'a0000000000000000000000000000001', 'REVISION', 'Accion de prueba sin sancion.', '2026-11-22 10:00:00+00')
ON CONFLICT DO NOTHING;

COMMIT;
