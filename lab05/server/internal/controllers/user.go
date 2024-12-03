package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab05/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab05/server/internal/services"
)

type ApiController struct {
	userService  *services.UserService
	tokenService *services.TokenService
}

func NewUserController(usr *services.UserService, tok *services.TokenService) *ApiController {
	return &ApiController{usr, tok}
}

func (a *ApiController) getCurrentUser(c *gin.Context) *dto.User {
	ctx := c.Request.Context()
	token := c.MustGet("access_token").(*dto.AccessToken)
	user, err := a.userService.GetByUsername(ctx, token.Token.Subject())
	if err != nil {
		panic(err)
	}
	return user
}

// GET /api/users/me
func (a *ApiController) Me(c *gin.Context) {
	me := a.getCurrentUser(c)
	profile := &dto.Profile{
		Firstname: me.Firstname,
		Lastname:  me.Lastname,
		Email:     me.Email,
		Extra:     me.Extra,
	}
	c.JSON(http.StatusOK, profile)
}
