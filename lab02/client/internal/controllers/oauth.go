package controllers

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab02/client/internal/services"
	"github.com/cyllective/oauth-labs/lab02/client/internal/session"
)

type OAuthController struct {
	oauthConfig  *oauth2.Config
	tokenService *services.TokenService
}

func NewOAuthController(oauthConfig *oauth2.Config, tokenService *services.TokenService) *OAuthController {
	return &OAuthController{oauthConfig, tokenService}
}

// GET /callback
func (o *OAuthController) Callback(c *gin.Context) {
	ctx := c.Request.Context()
	s := sessions.Default(c)

	if err := c.Query("error"); err != "" {
		log.Printf("[OAuthController.Callback]: authorization server responded with an error: %s\n", err)
		session.Delete(s)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"Status": "400",
			"Error":  "Authorization request failed: " + err,
		})
		return
	}

	code := c.Query("code")
	var tokens *oauth2.Token
	tokens, err := o.oauthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("[OAuthController.Callback]: failed to exchange authorization code for tokens: %s", err.Error())
		session.Delete(s)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"Status": "400",
			"Error":  "Token request failed: " + err.Error(),
		})
		return
	}

	if _, err := o.tokenService.Parse(tokens.AccessToken); err != nil {
		log.Printf("[OAuthController.Callback]: failed to decode access_token: %s", err.Error())
		session.Delete(s)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": "500",
			"Error":  "Hrm... something broke.",
		})
		return
	}

	if err = session.SetTokens(s, tokens); err != nil {
		log.Printf("[OAuthController.Callback]: failed to save tokens to session: %s", err.Error())
		session.Delete(s)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": "500",
			"Error":  "Hrm... something broke.",
		})
		return
	}

	c.Redirect(http.StatusFound, "/")
}

// GET /login
func (o *OAuthController) Login(c *gin.Context) {
	s := sessions.Default(c)
	if session.IsAuthenticated(s) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	s.Clear()
	loginURL := o.oauthConfig.AuthCodeURL("")
	_ = s.Save()
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"LoginURL": loginURL,
	})
}

// POST /logout
func (o *OAuthController) Logout(c *gin.Context) {
	s := sessions.Default(c)
	if session.IsAuthenticated(s) {
		tokens, err := o.tokenService.Get(s)
		if err != nil {
			log.Printf("[OAuthController.Logout]: error: %s", err.Error())
			return
		}
		if errs := o.tokenService.Revoke(c.Request.Context(), tokens); errs != nil {
			for _, err := range errs {
				if err != nil {
					log.Printf("[OAuthController.Logout]: warning while revoking: %s", err.Error())
				}
			}
		}
	}

	session.Delete(s)
	c.Redirect(http.StatusFound, "/login")
}
