package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab02/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab02/server/internal/services"
)

type ProfileController struct {
	authenticationService *services.AuthenticationService
	profileService        *services.ProfileService
}

func NewProfileController(authenticationService *services.AuthenticationService, profileService *services.ProfileService) *ProfileController {
	return &ProfileController{authenticationService, profileService}
}

// GET /profile
func (pc *ProfileController) GetProfile(c *gin.Context) {
	user, err := pc.authenticationService.GetUserFromSession(c)
	if err != nil {
		log.Printf("[ProfileController.GetProfile]: error: %s", err.Error())
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": 500,
			"Error":  "Hrm... something broke.",
		})
		return
	}

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"ActiveNavigation": "profile",
		"User":             user,
	})
}

// POST /profile
func (pc *ProfileController) UpdateProfile(c *gin.Context) {
	user, err := pc.authenticationService.GetUserFromSession(c)
	if err != nil {
		log.Printf("[ProfileController.UpdateProfile]: error: %s", err.Error())
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": 500,
			"Error":  "Hrm... something broke.",
		})
	}

	var request dto.UpdateProfile
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("[ProfileController.UpdateProfile]: error: %s", err.Error())
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": 500,
			"Error":  "Hrm... something broke.",
		})
		return
	}
	request.UserID = user.ID
	if err := pc.profileService.Update(c.Request.Context(), &request); err != nil {
		log.Printf("[ProfileController.UpdateProfile]: error: %s", err.Error())
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": 500,
			"Error":  "Hrm... something broke.",
		})
		return
	}
	c.Redirect(http.StatusFound, "/profile")
}
