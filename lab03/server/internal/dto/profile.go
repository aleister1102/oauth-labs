package dto

type Profile struct {
	ID        string `json:"user_id"`
	AvatarURL string `json:"avatar_url"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Extra     string `json:"extra"`
}

type UpdateProfile struct {
	AvatarURL string `form:"avatar_url" binding:"omitempty,url,max=240"`
	Firstname string `form:"firstname" binding:"omitempty,max=60"`
	Lastname  string `form:"lastname" binding:"omitempty,max=60"`
	Email     string `form:"email" binding:"omitempty,max=100,email"`

	UserID string
}
