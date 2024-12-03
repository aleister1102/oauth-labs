package controllers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab04/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab04/server/internal/services"
)

type ConsentController struct {
	authenticationService *services.AuthenticationService
	clientService         *services.ClientService
	consentService        *services.ConsentService
	tokenService          *services.TokenService
}

func NewConsentController(authenticationService *services.AuthenticationService, clientService *services.ClientService, consentService *services.ConsentService, tokenService *services.TokenService) *ConsentController {
	return &ConsentController{authenticationService, clientService, consentService, tokenService}
}

// GET /consents
func (cc *ConsentController) Get(c *gin.Context) {
	user, err := cc.authenticationService.GetUserFromSession(c)
	if err != nil {
		log.Printf("[ConsentController.Get]: error: %s", err.Error())
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": "500",
			"Error":  "Hrm... something broke.",
		})
		return
	}

	ctx := c.Request.Context()
	consents, err := cc.consentService.GetAll(ctx, user.ID)
	if err != nil {
		log.Printf("[ConsentController.GetAll]: error: %s", err.Error())
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": "500",
			"Error":  "Hrm... something broke.",
		})
		return
	}
	clients, err := cc.clientService.GetMany(ctx, consents.ClientIDs...)
	if err != nil {
		log.Printf("[ConsentController.GetAll]: error: %s", err.Error())
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": "500",
			"Error":  "Hrm... something broke.",
		})
		return
	}

	c.HTML(http.StatusOK, "consents.tmpl", gin.H{
		"Clients":          clients,
		"ActiveNavigation": "consents",
	})
}

// POST /consents
func (cc *ConsentController) Create(c *gin.Context) {
	user, err := cc.authenticationService.GetUserFromSession(c)
	if err != nil {
		log.Printf("[ConsentController.Create]: error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get user from session: " + err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	var request dto.CreateConsent
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("[ConsentController.Create]: error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "request failed to bind: " + err.Error(),
		})
		return
	}
	client, err := cc.clientService.Get(ctx, request.ClientID)
	if err != nil {
		log.Printf("[ConsentController.Create]: error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get client: " + err.Error(),
		})
		return
	}

	consent := &dto.Consent{
		UserID:   user.ID,
		ClientID: client.ID,
	}
	err = cc.consentService.Create(ctx, consent)
	if err != nil {
		log.Printf("[ConsentController.Create]: error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to create consent: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, request.ReturnTo)
}

// POST /consents/revoke
func (cc *ConsentController) Revoke(c *gin.Context) {
	user, err := cc.authenticationService.GetUserFromSession(c)
	if err != nil {
		log.Printf("[ConsentController.Revoke]: error: %s", err.Error())
		panic(err)
	}

	ctx := c.Request.Context()
	var request dto.RevokeConsent
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("[ConsentController.Revoke]: error: %s", err.Error())
		c.Redirect(http.StatusBadRequest, "/consents?error="+url.QueryEscape(err.Error()))
		return
	}

	client, err := cc.clientService.Get(ctx, request.ClientID)
	if err != nil {
		log.Printf("[ConsentController.Revoke]: error: %s", err.Error())
		c.Redirect(http.StatusBadRequest, "/consents?error="+url.QueryEscape(err.Error()))
		return
	}
	consent := &dto.Consent{UserID: user.ID, ClientID: client.ID}
	if !cc.consentService.HasConsent(ctx, consent) {
		c.Redirect(http.StatusFound, "/consents")
		return
	}

	// Revoke the user's consent for this client
	if err := cc.consentService.Revoke(ctx, consent); err != nil {
		log.Printf("[ConsentController.Revoke]: error: %s", err.Error())
		c.Redirect(http.StatusBadRequest, "/consents?error="+url.QueryEscape(err.Error()))
		return
	}

	// Revoke all issued access_tokens and refresh_tokens for this client + user combination
	if err := cc.tokenService.RevokeAll(ctx, consent.ClientID, consent.UserID); err != nil {
		log.Printf("[ConsentController.Revoke]: warning: failed to revoke tokens: %s", err.Error())
		c.Redirect(http.StatusBadRequest, "/consents?error="+url.QueryEscape(err.Error()))
		return
	}

	c.Redirect(http.StatusFound, "/consents")
}
