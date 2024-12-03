package dto

type Profile struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Extra     string `json:"extra,omitempty"`
}

type UpdateProfile struct {
	Firstname string `form:"firstname" binding:"omitempty,max=60"`
	Lastname  string `form:"lastname" binding:"omitempty,max=60"`
	Email     string `form:"email" binding:"omitempty,max=100,email"`

	UserID string
}
