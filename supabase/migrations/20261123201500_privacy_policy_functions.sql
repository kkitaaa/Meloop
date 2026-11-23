CREATE OR REPLACE FUNCTION meloop_privacy_allows_visibility(
    p_owner_id TEXT,
    p_viewer_id TEXT,
    p_visibility TEXT
)
RETURNS BOOLEAN
LANGUAGE SQL
STABLE
AS $$
    SELECT CASE
        WHEN p_owner_id = p_viewer_id THEN TRUE
        WHEN COALESCE(NULLIF(UPPER(BTRIM(p_visibility)), ''), 'PUBLICO') = 'PUBLICO' THEN TRUE
        WHEN COALESCE(NULLIF(UPPER(BTRIM(p_visibility)), ''), 'PUBLICO') = 'AMIGOS' THEN EXISTS (
            SELECT 1
            FROM AMISTAD a
            WHERE (
                (a.id_usuario_1::TEXT = p_owner_id AND a.id_usuario_2::TEXT = p_viewer_id)
                OR (a.id_usuario_1::TEXT = p_viewer_id AND a.id_usuario_2::TEXT = p_owner_id)
            )
              AND UPPER(BTRIM(a.estado)) = 'ACEPTADA'
        )
        ELSE FALSE
    END
$$;

CREATE OR REPLACE FUNCTION meloop_privacy_allows_friend_request(
    p_sender_id TEXT,
    p_receiver_id TEXT,
    p_reception TEXT
)
RETURNS BOOLEAN
LANGUAGE SQL
STABLE
AS $$
    SELECT CASE
        WHEN p_sender_id = p_receiver_id THEN FALSE
        WHEN COALESCE(NULLIF(UPPER(BTRIM(p_reception)), ''), 'TODOS') = 'TODOS' THEN TRUE
        WHEN COALESCE(NULLIF(UPPER(BTRIM(p_reception)), ''), 'TODOS') = 'AMIGOS' THEN EXISTS (
            SELECT 1
            FROM (
                SELECT a.id_usuario_2::TEXT AS friend_id
                FROM AMISTAD a
                WHERE a.id_usuario_1::TEXT = p_sender_id
                  AND UPPER(BTRIM(a.estado)) = 'ACEPTADA'
                UNION
                SELECT a.id_usuario_1::TEXT AS friend_id
                FROM AMISTAD a
                WHERE a.id_usuario_2::TEXT = p_sender_id
                  AND UPPER(BTRIM(a.estado)) = 'ACEPTADA'
            ) sender_friends
            JOIN (
                SELECT a.id_usuario_2::TEXT AS friend_id
                FROM AMISTAD a
                WHERE a.id_usuario_1::TEXT = p_receiver_id
                  AND UPPER(BTRIM(a.estado)) = 'ACEPTADA'
                UNION
                SELECT a.id_usuario_1::TEXT AS friend_id
                FROM AMISTAD a
                WHERE a.id_usuario_2::TEXT = p_receiver_id
                  AND UPPER(BTRIM(a.estado)) = 'ACEPTADA'
            ) receiver_friends USING (friend_id)
            WHERE sender_friends.friend_id NOT IN (p_sender_id, p_receiver_id)
              AND receiver_friends.friend_id NOT IN (p_sender_id, p_receiver_id)
        )
        ELSE FALSE
    END
$$;

CREATE OR REPLACE FUNCTION meloop_privacy_allows_message(
    p_sender_id TEXT,
    p_receiver_id TEXT,
    p_reception TEXT
)
RETURNS BOOLEAN
LANGUAGE SQL
STABLE
AS $$
    SELECT CASE
        WHEN p_sender_id = p_receiver_id THEN FALSE
        WHEN COALESCE(NULLIF(UPPER(BTRIM(p_reception)), ''), 'AMIGOS') = 'TODOS' THEN TRUE
        WHEN COALESCE(NULLIF(UPPER(BTRIM(p_reception)), ''), 'AMIGOS') = 'AMIGOS' THEN EXISTS (
            SELECT 1
            FROM AMISTAD a
            WHERE (
                (a.id_usuario_1::TEXT = p_sender_id AND a.id_usuario_2::TEXT = p_receiver_id)
                OR (a.id_usuario_1::TEXT = p_receiver_id AND a.id_usuario_2::TEXT = p_sender_id)
            )
              AND UPPER(BTRIM(a.estado)) = 'ACEPTADA'
        )
        ELSE FALSE
    END
$$;
