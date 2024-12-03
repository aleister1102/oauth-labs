package controllers

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab01/client/internal/client"
	"github.com/cyllective/oauth-labs/lab01/client/internal/services"
	"github.com/cyllective/oauth-labs/lab01/client/internal/session"
	"github.com/cyllective/oauth-labs/lab01/client/internal/utils"
)

type OAuthController struct {
	oauthConfig  *oauth2.Config
	tokenService *services.TokenService
	userService  *services.UserService
}

func NewOAuthController(oauthConfig *oauth2.Config, tokenService *services.TokenService, userService *services.UserService) *OAuthController {
	return &OAuthController{oauthConfig, tokenService, userService}
}

// GET /callback
func (o *OAuthController) Callback(c *gin.Context) {
	ctx := c.Request.Context()
	s := sessions.Default(c)
	state, _ := session.GetString(s, "state")
	if c.Query("state") != state {
		session.Delete(s)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"Status": "400",
			"Error":  "Invalid state",
		})
		return
	}

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
	verifier, _ := session.GetString(s, "verifier")
	tokens, err := o.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
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

	// Derive user identity and update database if needed.
	// We use the email claim to distinguish between users.
	cl := client.NewAPIClient(ctx, tokens)
	profile, err := cl.GetProfile()
	if err != nil {
		log.Printf("[OAuthController.Callback]: error: failed to fetch user profile: %s", err.Error())
		session.Delete(s)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Status": "500",
			"Error":  "Hrm... something broke.",
		})
		return
	}
	if err := o.userService.Create(ctx, profile); err != nil {
		log.Printf("[OAuthController.Callback]: error: failed to establish user identity: %s", err.Error())
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
	state := utils.RandomBase64URL(32)
	s.Set("state", state)
	verifier := oauth2.GenerateVerifier()
	s.Set("verifier", verifier)
	loginURL := o.oauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	_ = s.Save()
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"LoginURL": loginURL,
	})
}

// POST /logout
func (o *OAuthController) Logout(c *gin.Context) {
	s := sessions.Default(c)
	if session.IsAuthenticated(s) {
		o.revokeTokens(c, s)
	}

	session.Delete(s)
	c.Redirect(http.StatusFound, "/login")
}

func (o *OAuthController) revokeTokens(c *gin.Context, session sessions.Session) {
	tokens, err := o.tokenService.Get(session)
	if err != nil {
		log.Printf("[OAuthController.revokeTokens]: error: %s", err.Error())
		return
	}
	if errs := o.tokenService.Revoke(c.Request.Context(), tokens); errs != nil {
		for _, err := range errs {
			if err != nil {
				log.Printf("[OAuthController.revokeTokens]: warning while revoking: %s", err.Error())
			}
		}
	}
}
