package dto

type CreateConsent struct {
	ClientID string `form:"client_id" binding:"required"`
	ReturnTo string `form:"return_to" binding:"required"`
}

type RevokeConsent struct {
	ClientID string `form:"client_id" binding:"required"`
}

type Consent struct {
	UserID   string
	ClientID string
}

type UserConsents struct {
	UserID    string
	ClientIDs []string
}
