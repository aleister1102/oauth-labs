package controllers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab05/client/internal/client"
	"github.com/cyllective/oauth-labs/lab05/client/internal/config"
	"github.com/cyllective/oauth-labs/lab05/client/internal/dto"
	"github.com/cyllective/oauth-labs/lab05/client/internal/services"
	"github.com/cyllective/oauth-labs/lab05/client/internal/session"
	"github.com/cyllective/oauth-labs/lab05/client/internal/utils"
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
	s := sessions.Default(c)
	state, ok := session.GetString(s, "state")
	if !ok {
		session.Delete(s)
		c.Redirect(http.StatusFound, "/")
		return
	}
	verifier, ok := session.GetString(s, "verifier")
	if !ok {
		session.Delete(s)
		c.Redirect(http.StatusFound, "/")
		return
	}

	cfg := config.Get()
	creds := fmt.Sprintf("%s:%s", cfg.GetString("client.id"), cfg.GetString("client.secret"))
	b64creds := base64.StdEncoding.EncodeToString([]byte(creds))
	c.HTML(http.StatusOK, "callback.tmpl", gin.H{
		"ClientCredentials": b64creds,
		"RedirectURI":       cfg.GetString("client.redirect_uri"),
		"TokenURI":          cfg.GetString("authorization_server.token_uri"),
		"Verifier":          verifier,
		"State":             state,
	})
}

// POST /set-tokens
func (o *OAuthController) SetTokens(c *gin.Context) {
	s := sessions.Default(c)
	ctx := c.Request.Context()

	var req dto.SetTokenRequest
	if err := c.BindJSON(&req); err != nil {
		log.Printf("[OAuthController.SetToken]: error: %s", err.Error())
		session.Delete(s)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "access_token and refresh_token required",
		})
		return
	}
	if _, err := o.tokenService.Parse(req.AccessToken); err != nil {
		log.Printf("[OAuthController.SetToken]: error: %s", err.Error())
		session.Delete(s)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tokens",
		})
		return
	}

	expiry := time.Now().UTC().Add(time.Duration(req.ExpiresIn) * time.Second)
	tokens := &oauth2.Token{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       expiry,
	}
	api := client.NewAPIClient(ctx, tokens)
	if _, err := api.GetProfile(); err != nil {
		log.Printf("[OAuthController.SetToken]: error: %s", err.Error())
		session.Delete(s)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tokens",
		})
		return
	}
	if err := session.SetTokens(s, tokens); err != nil {
		log.Printf("[OAuthController.SetToken]: error: %s", err.Error())
		session.Delete(s)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tokens",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
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
