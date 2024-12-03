package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs"

	"github.com/cyllective/oauth-labs/lab05/server/internal/config"
	"github.com/cyllective/oauth-labs/lab05/server/internal/database"
	"github.com/cyllective/oauth-labs/lab05/server/internal/redis"
	"github.com/cyllective/oauth-labs/lab05/server/internal/server"
)

func main() {
	// Prepare configuration.
	cfg, err := config.InitFrom(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Prepare database.
	db, err := database.Init()
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %s", err.Error())
		return
	}

	// Prepare redis.
	rdb, err := redis.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare gin engine
	router, err := server.Init()
	if err != nil {
		log.Fatal(err)
	}

	defer closeDB(db)
	defer closeRedis(rdb)

	// Prepare server
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.GetString("server.host"), cfg.GetInt("server.port")),
		Handler:           router,
		ReadHeaderTimeout: time.Duration(5) * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

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

func closeDB(db io.Closer) {
	if err := db.Close(); err != nil {
		log.Printf("failed to close database connection: %s", err.Error())
	}
}

func closeRedis(rdb io.Closer) {
	if err := rdb.Close(); err != nil {
		log.Printf("failed to close redis connection: %s", err.Error())
	}
}
