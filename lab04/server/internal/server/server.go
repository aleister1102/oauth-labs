package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/cydave/staticfs"
	"github.com/cyllective/oauth-labs/oalib/metadata"
	"github.com/cyllective/oauth-labs/oalib/scope"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	sredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab04/server/internal/assets"
	"github.com/cyllective/oauth-labs/lab04/server/internal/config"
	"github.com/cyllective/oauth-labs/lab04/server/internal/constants"
	"github.com/cyllective/oauth-labs/lab04/server/internal/controllers"
	"github.com/cyllective/oauth-labs/lab04/server/internal/database"
	"github.com/cyllective/oauth-labs/lab04/server/internal/middlewares"
	"github.com/cyllective/oauth-labs/lab04/server/internal/redis"
	"github.com/cyllective/oauth-labs/lab04/server/internal/repositories"
	"github.com/cyllective/oauth-labs/lab04/server/internal/services"
)

func configureStaticFS(r *gin.Engine) error {
	// Set caching headers for resources that are found.
	okCallback := func(c *gin.Context, _ string) {
		c.Header("Cache-Control", "private, max-age=3600")
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
		"Labname": func() string { return fmt.Sprintf("server-%s", constants.LabNumber) },
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
	if err := sredis.SetKeyPrefix(store, fmt.Sprintf("server%s:session:", constants.LabNumber)); err != nil {
		return fmt.Errorf("failed to set redis key prefix: %w", err)
	}
	store.Options(config.GetSessionOptions())
	r.Use(sessions.Sessions(fmt.Sprintf("server-%s", constants.LabNumber), store))
	return nil
}

func configureMiddlewares(r *gin.Engine) error {
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "null" || origin == "https://client-04.oauth.labs" || origin == "http://127.0.0.1:3001"
		},
		MaxAge: 12 * time.Hour,
	}))
	return nil
}

func configureRoutes(r *gin.Engine) error {
	cfg := config.Get()
	db := database.Get()
	rdb := redis.Get()
	meta := metadata.New(cfg.GetString("oauth.issuer")).
		WithGrantTypes("refresh_token", "authorization_code").
		WithResponseTypes("code").
		WithTokenEndpointAuthMethodsSupported("client_secret_basic", "client_secret_post").
		WithRevocationEndpointAuthMethodsSupported("client_secret_basic", "client_secret_post").
		WithCodeChallengeMethods("plain", "S256").
		WithScopes("read:profile").
		WithEndpoints(&metadata.Endpoints{
			JwksURI:               "/.well-known/jwks.json",
			RegistrationEndpoint:  "/oauth/register",
			AuthorizationEndpoint: "/oauth/authorize",
			TokenEndpoint:         "/oauth/token",
			RevocationEndpoint:    "/oauth/revoke",
		})

	// Repositories
	userRepository := repositories.NewUserRepository(db)
	clientRepository := repositories.NewClientRepository(db)
	consentRepository := repositories.NewConsentRepository(db)
	accessTokenRepository := repositories.NewAcccessTokenRepository(db)
	refreshTokenRepository := repositories.NewRefreshTokenRepository(db)

	// Services
	cryptoService := services.NewCryptoService()
	jwkService := services.NewJWKService()
	userService := services.NewUserService(userRepository)
	profileService := services.NewProfileService(userRepository)
	clientService := services.NewClientService(clientRepository)
	authService := services.NewAuthenticationService(userService)
	consentService := services.NewConsentService(consentRepository)
	azcodeService := services.NewAuthorizationCodeService(rdb, jwkService)
	accessTokenService := services.NewAccessTokenService(accessTokenRepository, clientRepository, userRepository, meta, cryptoService, jwkService)
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepository, cryptoService, jwkService)
	tokenService := services.NewTokenService(db, meta, jwkService, cryptoService, consentService, accessTokenService, refreshTokenService)
	oauthService := services.NewOAuthService(meta, authService, clientService, consentService, tokenService, jwkService, azcodeService)

	loginRequired := middlewares.LoginRequired()

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{
			"Status": "404",
			"Error":  "Not Found",
		})
	})

	r.GET("/", loginRequired, func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"ActiveNavigation": "index",
		})
	})

	{
		c := controllers.NewHealthController()
		r.GET("/health", middlewares.NoCache(), c.Health)
	}

	{
		c := controllers.NewAuthenticationController(authService, clientService, consentService)
		r.GET("/login", c.GetLogin)
		r.POST("/login", c.Login)
		r.GET("/register", c.GetRegister)
		r.POST("/register", c.Register)
		r.POST("/logout", c.Logout)
	}

	{
		c := controllers.NewProfileController(authService, profileService)
		r.GET("/profile", middlewares.NoCache(), loginRequired, c.GetProfile)
		r.POST("/profile", middlewares.NoCache(), loginRequired, c.UpdateProfile)
	}

	{
		c := controllers.NewConsentController(authService, clientService, consentService, tokenService)
		r.GET("/consents", middlewares.NoCache(), loginRequired, c.Get)
		r.POST("/consents", middlewares.NoCache(), loginRequired, c.Create)
		r.POST("/consents/revoke", middlewares.NoCache(), loginRequired, c.Revoke)
	}

	{
		c := controllers.NewOAuthController(oauthService)
		r.GET("/.well-known/oauth-authorization-server", c.Metadata)
		r.GET("/.well-known/jwks.json", c.JWKs)
		r.POST("/oauth/register", middlewares.NoCache(), c.Register)
		r.GET("/oauth/authorize", middlewares.NoCache(), c.Authorize)
		r.POST("/oauth/token", middlewares.NoCache(), c.Token)
		r.POST("/oauth/revoke", middlewares.NoCache(), c.Revoke)
	}

	{
		scopesRequired := func(scopes ...string) gin.HandlerFunc {
			requiredScopes := scope.NewWith(scopes...)
			return middlewares.ScopeRequired(tokenService, requiredScopes)
		}

		c := controllers.NewUserController(userService, tokenService)
		api := r.Group("/api", middlewares.NoCache(), middlewares.JWTRequired(tokenService))
		api.GET("/users/me", scopesRequired("read:profile"), c.Me)
	}

	return nil
}

func Init() (*gin.Engine, error) {
	cfg := config.Get()
	if env := cfg.GetString("environment"); env == "" || env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	if err := configureMiddlewares(r); err != nil {
		return nil, err
	}
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
