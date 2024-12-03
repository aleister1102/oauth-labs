package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab04/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab04/server/internal/services"
)

type OAuthController struct {
	oauthService *services.OAuthService
}

func NewOAuthController(oas *services.OAuthService) *OAuthController {
	return &OAuthController{oas}
}

// GET /.well-known/oauth-authorization-server
func (oa *OAuthController) Metadata(c *gin.Context) {
	meta := oa.oauthService.Metadata()
	c.JSON(http.StatusOK, meta)
}

// GET /.well-known/jwks.json
func (oa *OAuthController) JWKs(c *gin.Context) {
	jwks := oa.oauthService.JWKs()
	c.JSON(http.StatusOK, jwks)
}

// GET /oauth/authorize
func (oa *OAuthController) Authorize(c *gin.Context) {
	var request dto.AuthorizeRequest
	if err := c.ShouldBind(&request); err != nil {
		panic(err)
	}
	res, err := oa.oauthService.Authorize(c, &request)
	if err != nil {
		// Prompt for authentication.
		if errors.Is(err, services.ErrAuthenticationRequired) {
			log.Println("[OAuthController.Authorize]: user did not authenticate yet, redirecting to oauth login page.")
			returnTo := url.QueryEscape(c.Request.URL.String())
			c.Redirect(http.StatusFound, fmt.Sprintf("/login?return_to=%s", returnTo))
			return
		}

		// Prompt for consent.
		if errors.Is(err, services.ErrConsentRequired) {
			log.Println("[OAuthController.Authorize]: user did not consent to client yet, redirecting to consent page.")
			returnTo := url.QueryEscape(c.Request.URL.String())
			c.Redirect(http.StatusFound, fmt.Sprintf("/login?return_to=%s", returnTo))
			return
		}

		// Abort the authorize request, validation failed and we can't proceed.
		if errors.Is(err, services.ErrAbortAuthorize) {
			log.Println("[OAuthController.Authorize]: authorization request failed validation, aborting.")
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"Status": "400",
				"Error":  "Authorize request failed:" + err.Error(),
			})
			return
		}

		// Handle client errors.
		if aerr, ok := err.(services.AuthorizeError); ok {
			log.Printf("[OAuthController.Authorize]: authorization request had client specific errors, informing client: %#v", aerr)
			c.Redirect(http.StatusFound, aerr.RedirectURI)
			return
		}
		if aerr, ok := err.(*services.AuthorizeError); ok {
			log.Printf("[OAuthController.Authorize]: *authorization request had client specific errors, informing client: %#v", aerr)
			c.Redirect(http.StatusFound, aerr.RedirectURI)
			return
		}

		log.Printf("[OAuthController.Authorize]: authorization request threw an uncaught error: %#v", err)
		panic(err)
	}

	c.Redirect(http.StatusFound, res.RedirectURI)
}

// POST /oauth/token
func (oa *OAuthController) Token(c *gin.Context) {
	var request dto.TokenRequest
	if err := c.ShouldBind(&request); err != nil {
		panic(err)
	}
	res, err := oa.oauthService.Token(c, &request)
	if err != nil {
		log.Printf("[OAuthController.Token]: request failed: %#v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// POST /oauth/revoke
func (oa *OAuthController) Revoke(c *gin.Context) {
	var request dto.OAuthRevoke
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("[OAuthController.Revoke]: warning: revocation request was invalid, no tokens revoked: %s", err.Error())
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	if err := oa.oauthService.Revoke(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// POST /oauth/register
func (oa *OAuthController) Register(c *gin.Context) {
	var request dto.OAuthRegisterClient
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("[OAuthController.Register]: dynamic client registration request failed to bind: %#v\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, secret, ok := c.Request.BasicAuth()
	if !ok {
		log.Println("[OAuthController.Register]: register request failed: client_id and client_secret not found in basic auth")
		c.JSON(http.StatusBadRequest, gin.H{"error": "basic auth missing"})
		return
	}
	request.ClientID = id
	request.ClientSecret = secret
	request.RegisterSecret = c.GetHeader("x-register-key")

	res, err := oa.oauthService.Register(c.Request.Context(), request)
	if err != nil {
		log.Printf("[OAuthController.Register]: dynamic client registration request failed: %#v\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
