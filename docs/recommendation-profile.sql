-- Tables consumed by services/recommendation-service.
-- Keep RLS enabled in Supabase and grant the service role read access only.

CREATE TABLE IF NOT EXISTS public.user_music_preferences (
  user_id bigint NOT NULL,
  preference_type text NOT NULL CHECK (preference_type IN ('genre', 'artist', 'song')),
  preference_value text NOT NULL,
  PRIMARY KEY (user_id, preference_type, preference_value)
);

CREATE TABLE IF NOT EXISTS public.user_interactions (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id bigint NOT NULL,
  interaction_type text NOT NULL CHECK (interaction_type IN ('like', 'friend_added', 'comment', 'post_interaction')),
  target_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS user_interactions_user_created_idx
  ON public.user_interactions (user_id, created_at DESC);

ALTER TABLE public.user_music_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.user_interactions ENABLE ROW LEVEL SECURITY;

-- The recommendation service should use a restricted server-side connection.
-- Do not expose these tables through the anon key.
