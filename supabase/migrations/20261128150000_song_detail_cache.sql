ALTER TABLE CANCION
  ADD COLUMN IF NOT EXISTS spotify_detail JSONB,
  ADD COLUMN IF NOT EXISTS spotify_detail_market VARCHAR(2),
  ADD COLUMN IF NOT EXISTS spotify_detail_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_cancion_spotify_id
  ON CANCION (spotify_id);

CREATE INDEX IF NOT EXISTS idx_cancion_spotify_detail_expiry
  ON CANCION (spotify_detail_expires_at)
  WHERE spotify_detail IS NOT NULL;
