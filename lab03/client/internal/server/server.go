package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/cydave/staticfs"
	"github.com/gin-contrib/sessions"
	sredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab03/client/internal/assets"
	"github.com/cyllective/oauth-labs/lab03/client/internal/client"
	"github.com/cyllective/oauth-labs/lab03/client/internal/config"
	"github.com/cyllective/oauth-labs/lab03/client/internal/constants"
	"github.com/cyllective/oauth-labs/lab03/client/internal/controllers"
	"github.com/cyllective/oauth-labs/lab03/client/internal/middlewares"
	"github.com/cyllective/oauth-labs/lab03/client/internal/services"
	"github.com/cyllective/oauth-labs/lab03/client/internal/session"
	"github.com/cyllective/oauth-labs/lab03/client/internal/utils"
)

func configureStaticFS(r *gin.Engine) error {
	// Set caching headers for resources that are found.
	okCallback := func(c *gin.Context, _ string) {
		c.Header("Cache-Control", "private,max-age=3600")
	}
	// Set no-cache headers for resources that were not found.
	errCallback := func(c *gin.Context, _ error) {
		c.Header("Pragma", "no-cache")
		c.Header("Cache-Control", "private, no-cache, no-store, max-age=0, no-transform")
		c.Header("Expires", "0")
	}
	// Create staticfs with the above defined callbacks.
	sfs := staticfs.New(assets.Static).
		WithRootAliases().
		WithOKCallback(okCallback).
		WithErrCallback(errCallback)
	sfs.Configure(r)
	return nil
}

func configureTemplating(r *gin.Engine) error {
	funcMaps := template.FuncMap{
		"Labname": func() string { return fmt.Sprintf("client-%s", constants.LabNumber) },
	}
	templ := template.New("").Funcs(funcMaps)
	templ, err := templ.ParseFS(assets.Templates, "templates/*.tmpl")
	if err != nil {
		return err
	}
	r.SetHTMLTemplate(templ)
	return nil
}

func configureSessions(r *gin.Engine) error {
	opts := config.GetRedisConfig()
	store, err := sredis.NewStoreWithDB(10, opts.Network, opts.Addr, opts.Password, fmt.Sprintf("%d", opts.DB), config.GetSessionSecret(), config.GetSessionSecret())
	if err != nil {
		return fmt.Errorf("failed to configure redis session store: %w", err)
	}
	if err := sredis.SetKeyPrefix(store, fmt.Sprintf("client%s:session:", constants.LabNumber)); err != nil {
		return fmt.Errorf("failed to set redis key prefix: %w", err)
	}

	store.Options(config.GetSessionOptions())
	r.Use(sessions.Sessions(fmt.Sprintf("client-%s", constants.LabNumber), store))
	return nil
}

func configureRoutes(r *gin.Engine) error {
	oauthConfig := config.GetOAuthConfig()
	tokenService := services.NewTokenService()

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{
			"Status": "404",
			"Error":  "Not Found",
		})
	})

	{
		c := controllers.NewHealthController()
		r.GET("/health", c.Health)
	}

	r.GET("/", func(c *gin.Context) {
		s := sessions.Default(c)
		if !session.IsAuthenticated(s) {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		tokens, err := tokenService.Get(s)
		if err != nil {
			log.Printf("failed to get tokens from session: %s", err.Error())
			session.Delete(s)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
				"Status": "500",
				"Error":  "Hrm... something broke.",
			})
			return
		}
		accessToken, err := tokenService.Parse(tokens.AccessToken)
		if err != nil {
			log.Printf("failed to parse access token: %s", err.Error())
			session.Delete(s)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
				"Status": "500",
				"Error":  "Hrm... something broke.",
			})
			return
		}

		c.Redirect(http.StatusFound, "/profile/"+accessToken.Subject())
	})

	r.GET("/profile/:id", middlewares.NoCache(), func(c *gin.Context) {
		s := sessions.Default(c)
		id, err := utils.GetUUID(c.Param("id"))
		if err != nil {
			log.Printf("[Profile]: failed to convert id to uuid: %s", err.Error())
			c.HTML(http.StatusNotFound, "error.tmpl", gin.H{
				"Status": "404",
				"Error":  "Profile not found",
			})
			return
		}

		// Check if user is authenticated and viewing own profile
		tokens, authErr := tokenService.Get(s)
		if authErr == nil {
			accessToken, parseErr := tokenService.Parse(tokens.AccessToken)
			if parseErr == nil && id == accessToken.Subject() {
				// Authenticated user viewing own profile - show full profile
				clnt := client.NewAPIClient(c.Request.Context(), tokens)
				profile, err := clnt.GetProfile(id)
				if err != nil {
					log.Printf("[Profile]: failed to get own profile: %s", err.Error())
					session.Delete(s)
					c.Redirect(http.StatusFound, "/")
					return
				}

				cfg := config.Get()
				c.HTML(http.StatusOK, "profile.tmpl", gin.H{
					"Profile":                profile,
					"IsAuthenticated":        true,
					"AuthorizationServerURL": cfg.GetString("authorization_server.issuer"),
				})
				return
			}
		}

		// Not authenticated OR viewing other's profile → Show public private profile
		clnt := client.NewAPIClient(c.Request.Context(), nil)
		profile, err := clnt.GetPublicProfile(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.tmpl", gin.H{
				"Status": "404",
				"Error":  "Profile not found",
			})
			return
		}

		c.HTML(http.StatusOK, "private_profile.tmpl", gin.H{
			"Profile":         profile,
			"IsAuthenticated": authErr == nil,
		})
	})

	{
		c := controllers.NewOAuthController(oauthConfig, tokenService)
		r.GET("/login", middlewares.NoCache(), c.Login)
		r.POST("/logout", middlewares.NoCache(), c.Logout)
		r.GET("/callback", middlewares.NoCache(), c.Callback)
	}

	return nil
}

func Init() (*gin.Engine, error) {
	cfg := config.Get()
	if env := cfg.GetString("environment"); env == "" || env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	if err := configureStaticFS(r); err != nil {
		return nil, err
	}
	if err := configureTemplating(r); err != nil {
		return nil, err
	}
	if err := configureSessions(r); err != nil {
		return nil, err
	}
	if err := configureRoutes(r); err != nil {
		return nil, err
	}

	return r, nil
}
