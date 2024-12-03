package controllers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab00/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab00/server/internal/services"
	"github.com/cyllective/oauth-labs/lab00/server/internal/session"
)

type AuthenticationController struct {
	authenticationService *services.AuthenticationService
	clientService         *services.ClientService
	consentService        *services.ConsentService
}

func NewAuthenticationController(authenticationService *services.AuthenticationService, clientService *services.ClientService, consentService *services.ConsentService) *AuthenticationController {
	return &AuthenticationController{authenticationService, clientService, consentService}
}

// GET /login
func (ac *AuthenticationController) GetLogin(c *gin.Context) {
	ctx := c.Request.Context()
	s := sessions.Default(c)
	isAuthenticated := session.IsAuthenticated(s)
	returnToURL, ok := getReturnToURL(c)
	if !ok || returnToURL == "" {
		if isAuthenticated {
			c.Redirect(http.StatusFound, "/")
			return
		}

		c.HTML(http.StatusOK, "login.tmpl", gin.H{})
		return
	}

	client, err := ac.clientService.GetFromURL(ctx, returnToURL)
	if err != nil {
		// If we are authenticated and the client doesn't exist, we render a 404.
		if isAuthenticated {
			c.HTML(http.StatusNotFound, "error.tmpl", gin.H{
				"Status": "404",
				"Error":  "Client not found.",
			})
			return
		}

		// If we are not authenticated, we present the normal login screen,
		// without a return_to param. The client didn't exist, so we don't
		// act on the return_to.
		c.HTML(http.StatusOK, "login.tmpl", gin.H{})
		return
	}

	if !isAuthenticated {
		c.HTML(http.StatusOK, "oauth_login.tmpl", gin.H{
			"Client":   client,
			"ReturnTo": returnToURL,
		})
		return
	}
	user, err := ac.authenticationService.GetUserFromSession(c)
	if err != nil {
		panic(err)
	}
	// Prompt for consent if required.
	if !ac.consentService.HasConsent(ctx, &dto.Consent{UserID: user.ID, ClientID: client.ID}) {
		c.HTML(http.StatusOK, "oauth_consent.tmpl", gin.H{
			"Client":    client,
			"ReturnTo":  returnToURL,
			"CancelURL": "",
		})
		return
	}

	c.Redirect(http.StatusFound, returnToURL)
}

// POST /login
func (ac *AuthenticationController) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBind(&request); err != nil {
		c.Redirect(http.StatusFound, "/login?error=missing+or+invalid+params")
		return
	}

	if err := ac.authenticationService.Login(c, &request); err != nil {
		log.Printf("[AuthenticationController.Login]: error: %s", err.Error())
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	redirectURI := "/"
	if request.ReturnTo != "" {
		redirectURI = request.ReturnTo
	}
	if !returnToURLAllowed(redirectURI) {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.Redirect(http.StatusFound, redirectURI)
}

// GET /register
func (ac *AuthenticationController) GetRegister(c *gin.Context) {
	s := sessions.Default(c)
	if session.IsAuthenticated(s) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "register.tmpl", gin.H{})
}

// POST /register
func (ac *AuthenticationController) Register(c *gin.Context) {
	var request dto.RegisterRequest
	if err := c.ShouldBind(&request); err != nil {
		c.Redirect(http.StatusFound, "/register?error=missing+or+invalid+params")
		return
	}
	if err := ac.authenticationService.Register(c.Request.Context(), &request); err != nil {
		c.HTML(http.StatusOK, "register.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

// POST /logout
func (ac *AuthenticationController) Logout(c *gin.Context) {
	s := sessions.Default(c)
	session.Delete(s)
	c.Redirect(http.StatusFound, "/login")
}

func returnToURLAllowed(u string) bool {
	p, e := url.Parse(u)
	if e != nil || p.IsAbs() {
		return false
	}
	return p.Path == "/oauth/authorize"
}

func getReturnToURL(c *gin.Context) (string, bool) {
	returnTo := c.Query("return_to")
	if returnTo == "" {
		return "", false
	}
	if !returnToURLAllowed(returnTo) {
		return "", false
	}
	return returnTo, true
}
