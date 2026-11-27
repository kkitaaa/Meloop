SELECT u.id_usuario, p.tipo, p.spotify_id
FROM USUARIO u
LEFT JOIN PREFERENCIA_MUSICAL p ON p.id_usuario = u.id_usuario
WHERE u.correo = 'seed.alex@example.test';