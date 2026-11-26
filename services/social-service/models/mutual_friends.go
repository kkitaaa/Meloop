package models

type MutualFriend struct {
	ID        string `json:"id_usuario"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type MutualFriendsData struct {
	Items      []MutualFriend
	TotalCount int
	Hidden     bool
}
