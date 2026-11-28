CREATE TABLE IF NOT EXISTS saved_posts (
    user_id VARCHAR(64) NOT NULL,
    post_id VARCHAR(64) NOT NULL,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_saved_posts PRIMARY KEY (user_id, post_id),
    CONSTRAINT fk_saved_posts_user FOREIGN KEY (user_id)
        REFERENCES usuario (id_usuario) ON DELETE CASCADE,
    CONSTRAINT fk_saved_posts_post FOREIGN KEY (post_id)
        REFERENCES publicacion (id_publicacion) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_saved_posts_user_saved_at
    ON saved_posts (user_id, saved_at DESC, post_id DESC);
