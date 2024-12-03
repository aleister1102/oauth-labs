package dto

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  string
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
}
