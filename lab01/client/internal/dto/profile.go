package dto

type Profile struct {
	ID        string
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Extra     string `json:"extra"`
}
