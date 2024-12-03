package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"github.com/cyllective/oauth-labs/lab01/server/internal/config"
)

var (
	db *sql.DB
	//go:embed schema.sql
	schema string
)

func Init() (*sql.DB, error) {
	if err := WaitForBackoff(10); err != nil {
		return nil, err
	}
	d, err := getClient()
	if err != nil {
		return nil, err
	}
	db = d
	db.SetConnMaxIdleTime(time.Duration(10) * time.Second)
	db.SetConnMaxLifetime(time.Duration(10) * time.Second)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	return db, nil
}

func getClient() (*sql.DB, error) {
	d, err := sql.Open("mysql", config.GetDatabaseURI())
	if err != nil {
		return nil, err
	}
	if err := d.Ping(); err != nil {
		return nil, err
	}
	return d, nil
}

func WaitForBackoff(tries int) error {
	errs := 0
	for i := 0; i < tries; i++ {
		if _, err := getClient(); err == nil {
			log.Printf("[database]: connected!")
			return nil
		}

		errs++
		time.Sleep(time.Duration(errs) * time.Second)
		log.Printf("[database]: failed to connect, retrying...")
	}

	log.Printf("[database]: failed to connect, aborting.")
	return errors.New("failed to connect to database")
}

// Here be sql.DB dragons.
func Migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, command := range strings.Split(schema, ";") {
		cmd := strings.TrimSpace(command)
		if cmd == "" {
			continue
		}
		_, err := tx.Exec(cmd + ";")
		if err != nil {
			log.Println(err.Error())
		}
	}

	// Seed the admin user.
	cfg := config.Get()
	bcryptCost := cfg.GetInt("server.bcrypt_cost")
	adminPW := cfg.GetString("server.admin_password")
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(adminPW), bcryptCost)
	if err != nil {
		panic(fmt.Errorf("failed to generate password hash: %w", err))
	}
	hash := string(hashBytes)
	_, err = tx.Exec(`UPDATE users SET password = ? WHERE username = 'admin'`, hash)
	if err != nil {
		log.Println(err.Error())
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return nil
}

func Get() *sql.DB {
	return db
}

func MakePlaceholders(count int) string {
	if count <= 0 {
		panic("unable to make placeholders for zero or negative count")
	}
	if count == 1 {
		return "?"
	}
	return "?" + strings.Repeat(",?", count-1)
}

func StringSliceToArgs(s []string) []any {
	args := make([]interface{}, len(s))
	for i, id := range s {
		args[i] = id
	}
	return args
}
