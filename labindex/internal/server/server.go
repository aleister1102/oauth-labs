package server

import (
	"html/template"
	"net/http"

	"github.com/cydave/staticfs"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/labindex/internal/assets"
	"github.com/cyllective/oauth-labs/labindex/internal/config"
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
		"Labname": func() string { return "server-00" },
	}
	templ := template.New("").Funcs(funcMaps)
	templ, err := templ.ParseFS(assets.Templates, "templates/*.tmpl")
	if err != nil {
		return err
	}
	r.SetHTMLTemplate(templ)
	return nil
}

func configureMiddlewares(r *gin.Engine) error {
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	return nil
}

func configureRoutes(r *gin.Engine) error {
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Labs": config.GetLabs(),
		})
	})

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
	if err := configureRoutes(r); err != nil {
		return nil, err
	}

	return r, nil
}
