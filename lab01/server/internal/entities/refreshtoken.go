package entities

type RefreshToken struct {
	ID       string
	ClientID string
	UserID   string
	Data     string
	Revoked  bool
}
