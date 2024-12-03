package dto

type CreateRefreshToken struct {
	UserID   string
	ClientID string
	Scope    string
}

type RevokeRefreshTokens struct {
	UserID   string
	ClientID string
}

type RefreshToken struct {
	SignedToken string
	ID          string
	UserID      string
	ClientID    string
	Scope       string
	CreatedAt   int
}

type RawRefreshToken struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	ClientID  string `json:"client_id"`
	Scope     string `json:"scope"`
	CreatedAt int    `json:"created_at"`
}
