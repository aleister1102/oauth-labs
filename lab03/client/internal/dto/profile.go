package dto

type Profile struct {
	ID        string `json:"user_id"`
	AvatarURL string `json:"avatar_url"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Extra     string `json:"extra"`
}
