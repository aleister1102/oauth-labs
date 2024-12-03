package dto

type LoginRequest struct {
	Username string `form:"username" binding:"required,max=60,alphanum"`
	Password string `form:"password" binding:"required,max=70"`
	ReturnTo string `form:"return_to" binding:"omitempty"`
}

type RegisterRequest struct {
	Username string `form:"username" binding:"required,max=60,alphanum"`
	Password string `form:"password" binding:"required,max=70"`
}
