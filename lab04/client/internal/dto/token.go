package dto

type SetTokenRequest struct {
	AccessToken  string `json:"access_token" binding:"required,jwt"`
	RefreshToken string `json:"refresh_token" binding:"required,base64rawurl"`
	Scope        string `json:"scope" binding:"required"`
	ExpiresIn    int    `json:"expires_in" binding:"required,min=0,max=10000"`
}
