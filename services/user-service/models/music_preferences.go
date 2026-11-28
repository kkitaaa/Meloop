package models

const MusicPreferenceLimit = 30

type MusicPreference struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	ArtworkURL *string `json:"artwork_url,omitempty"`
}

type MusicPreferences struct {
	Songs   []MusicPreference `json:"songs"`
	Artists []MusicPreference `json:"artists"`
	Genres  []MusicPreference `json:"genres"`
}

type MusicPreferencesResponse struct {
	Preferences MusicPreferences `json:"preferences"`
	Limits      map[string]int   `json:"limits"`
}

type AddMusicPreferenceRequest struct {
	Type       string  `json:"type"`
	ItemID     string  `json:"item_id"`
	Name       string  `json:"name"`
	ArtworkURL *string `json:"artwork_url"`
}
