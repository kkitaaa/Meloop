package models

type UserSearchResult struct {
	IDUsuario  string `json:"id_usuario"`
	Username   string `json:"username"`
	FotoPerfil string `json:"foto_perfil,omitempty"`
	Biografia  string `json:"biografia,omitempty"`
}

type UserSearchPage struct {
	Query      string             `json:"query"`
	Users      []UserSearchResult `json:"users"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	HasMore    bool               `json:"has_more"`
	NextOffset *int               `json:"next_offset,omitempty"`
}
