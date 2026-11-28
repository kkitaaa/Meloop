-- 1. Crear usuario lector ('1') y usuario amigo ('2') de prueba
INSERT INTO usuario (id_usuario, username, correo, contrasena_hash)
VALUES 
    ('1', 'k6_reader_1', 'k6_reader_1@meloop.test', 'hash_test_123'),
    ('2', 'k6_friend_2', 'k6_friend_2@meloop.test', 'hash_test_123')
ON CONFLICT (id_usuario) DO NOTHING;

-- 2. Crear amistad aceptada entre '1' y '2'
INSERT INTO amistad (id_amistad, id_usuario_1, id_usuario_2, estado)
VALUES ('amistad-k6-1-2', '1', '2', 'ACEPTADA')
ON CONFLICT (id_amistad) DO UPDATE SET estado = 'ACEPTADA';

-- 3. Asignar preferencia musical de género 'rock' al usuario '1'
INSERT INTO preferencia_musical (id_preferencia, id_usuario, tipo, spotify_id)
VALUES ('pref-k6-rock', '1', 'GENERO', 'rock')
ON CONFLICT (id_preferencia) DO UPDATE SET tipo = 'GENERO', spotify_id = 'rock';

-- 4. Insertar 2 publicaciones del amigo '2':
--    a) Una más antigua con afinidad 'rock' (relevance_score = 15)
--    b) Una más reciente sin afinidad (relevance_score = 0)
INSERT INTO publicacion (id_publicacion, id_usuario, texto, url_multimedia, fecha_creacion)
VALUES 
    ('post-k6-relevante', '2', 'Escuchando buen rock clasico hoy', '', NOW() - INTERVAL '2 hours'),
    ('post-k6-reciente', '2', 'Hola a todos desde Meloop', '', NOW())
ON CONFLICT (id_publicacion) DO UPDATE 
SET texto = EXCLUDED.texto,
    fecha_creacion = EXCLUDED.fecha_creacion;