package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyllective/oauth-labs/victim/internal/config"
	"github.com/cyllective/oauth-labs/victim/internal/server"
	"github.com/cyllective/oauth-labs/victim/internal/victims"
)

func main() {
	// Prepare configuration.
	cfg, err := config.InitFrom(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Prepare victim configuration
	victims.Init()

	// Prepare gin engine
	server, err := server.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare server
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.GetString("server.host"), cfg.GetInt("server.port")),
		Handler:           server.Engine,
		ReadHeaderTimeout: time.Duration(5) * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Spawn victim worker.
	go func() { victims.Handle(server.VisitChan) }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error shutting down server: %s", err.Error())
	}

	log.Println("Server exiting")
}
