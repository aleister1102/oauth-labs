package middlewares

import (
	"log"
	"net/http"

	"github.com/cyllective/oauth-labs/oalib/scope"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab05/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab05/server/internal/services"
	"github.com/cyllective/oauth-labs/lab05/server/internal/session"
)

func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Pragma", "no-cache")
		c.Header("Cache-Control", "private, no-cache, no-store, max-age=0, no-transform")
		c.Header("Expires", "0")
		c.Next()
	}
}

func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		if !session.IsAuthenticated(s) {
			session.Delete(s)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func JWTRequired(tok *services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := tok.GetFromRequest(c)
		if err != nil {
			log.Printf("[middlewares.JWTRequired]: failed to extract and verify token from request: %s\n", err.Error())
			c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			c.Abort()
			return
		}

		c.Set("access_token", token)
		c.Next()
	}
}

func ScopeRequired(tok *services.TokenService, requiredScopes *scope.Scope) gin.HandlerFunc {
	abort := func(c *gin.Context) {
		c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		c.Abort()
	}

	return func(c *gin.Context) {
		// We don't verify the token here, ScopeRequired is only used after the JWTRequired middleware.
		maybeToken, ok := c.Get("access_token")
		if !ok {
			log.Printf("[middlewares.ScopeRequired]: token not found in request")
			abort(c)
			return
		}
		at, ok := maybeToken.(*dto.AccessToken)
		if !ok {
			log.Printf("[middlewares.ScopeRequired]: failed to cast token in request")
			abort(c)
			return
		}
		if !tok.HasRequiredScopes(at.Token, requiredScopes) {
			log.Println("[middlewares.ScopeRequired]: scope requirement not met")
			abort(c)
			return
		}

		c.Next()
	}
}
