package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab00/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab00/server/internal/services"
)

type UserController struct {
	userService  *services.UserService
	tokenService *services.TokenService
}

func NewUserController(usr *services.UserService, tok *services.TokenService) *UserController {
	return &UserController{usr, tok}
}

func (u *UserController) getCurrentUser(c *gin.Context) *dto.User {
	token := c.MustGet("access_token").(*dto.AccessToken)
	user, err := u.userService.Get(c.Request.Context(), token.UserID)
	if err != nil {
		panic(err)
	}
	return user
}

// GET /api/users/me
func (u *UserController) Me(c *gin.Context) {
	me := u.getCurrentUser(c)
	profile := &dto.Profile{
		Firstname: me.Firstname,
		Lastname:  me.Lastname,
		Email:     me.Email,
	}
	c.JSON(http.StatusOK, profile)
}
