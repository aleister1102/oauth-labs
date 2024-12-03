package middlewares

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab00/client/internal/services"
	"github.com/cyllective/oauth-labs/lab00/client/internal/session"
)

func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Pragma", "no-cache")
		c.Header("Cache-Control", "private, no-cache, no-store, max-age=0, no-transform")
		c.Header("Expires", "0")
		c.Next()
	}
}

func TokensRequired(tok *services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		_, err := tok.Get(s)
		if err != nil && !errors.Is(err, services.ErrAccessTokenExpired) {
			log.Printf("failed to retrieve tokens: %s", err.Error())
			session.Delete(s)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
