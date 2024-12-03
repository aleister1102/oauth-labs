package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab03/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab03/server/internal/services"
	"github.com/cyllective/oauth-labs/lab03/server/internal/utils"
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
		ID:        me.ID,
		AvatarURL: me.AvatarURL,
		Firstname: me.Firstname,
		Lastname:  me.Lastname,
		Email:     me.Email,
		Extra:     me.Extra,
	}
	c.JSON(http.StatusOK, profile)
}

// GET /api/users/:id
func (u *UserController) User(c *gin.Context) {
	me := u.getCurrentUser(c)
	id, err := utils.GetUUID(c.Param("id"))
	if err != nil {
		log.Printf("[UserController.User]: error: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}
	user, err := u.userService.Get(c.Request.Context(), id)
	if err != nil {
		log.Printf("[UserController.User]: error: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	// If this is the current user making the request, return all details.
	if me.ID == id {
		profile := &dto.Profile{
			ID:        me.ID,
			AvatarURL: me.AvatarURL,
			Firstname: me.Firstname,
			Lastname:  me.Lastname,
			Email:     me.Email,
			Extra:     me.Extra,
		}
		c.JSON(http.StatusOK, profile)
		return
	}

	// only return "public" profile information.
	profile := &dto.Profile{
		ID:        user.ID,
		AvatarURL: user.AvatarURL,
	}
	c.JSON(http.StatusOK, profile)
}
