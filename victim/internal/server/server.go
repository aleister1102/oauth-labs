package server

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/cydave/staticfs"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/victim/internal/assets"
	"github.com/cyllective/oauth-labs/victim/internal/config"
	"github.com/cyllective/oauth-labs/victim/internal/dto"
	"github.com/cyllective/oauth-labs/victim/internal/victims"
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
	templ := template.New("")
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

func configureRoutes(s *Server) error {
	r := s.Engine

	knownVictims := victims.All()

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Victims": knownVictims,
		})
	})

	r.POST("/", func(c *gin.Context) {
		var req dto.VisitRequest
		if err := c.Bind(&req); err != nil {
			log.Printf("binding failed: %s\n", err)
			c.HTML(http.StatusInternalServerError, "index.tmpl", gin.H{
				"Victims": knownVictims,
				"Error":   "Hrm... something broke.",
			})
			return
		}
		if !victims.Exists(req.LabNumber) {
			c.HTML(http.StatusBadRequest, "index.tmpl", gin.H{
				"Victims": knownVictims,
				"Error":   "Invalid lab choice",
			})
			return
		}
		url, err := url.Parse(req.URL)
		if err != nil || !url.IsAbs() {
			c.HTML(http.StatusBadRequest, "index.tmpl", gin.H{
				"Victims": knownVictims,
				"Error":   "Invalid URL",
			})
			return
		}

		v, _ := victims.Get(req.LabNumber)
		if err := v.CheckURL(req.URL); err != nil {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"Victims": knownVictims,
				"Error":   err.Error(),
			})
			return
		}

		go func() { s.VisitChan <- &req }()
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Victims": knownVictims,
			"Message": "The victim will be with your shortly.",
		})
	})

	return nil
}

type Server struct {
	Engine    *gin.Engine
	VisitChan chan *dto.VisitRequest
}

func Init() (*Server, error) {
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

	server := &Server{
		Engine:    r,
		VisitChan: make(chan *dto.VisitRequest, 10),
	}
	if err := configureRoutes(server); err != nil {
		return nil, err
	}

	return server, nil
}
