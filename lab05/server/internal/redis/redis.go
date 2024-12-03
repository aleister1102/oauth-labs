package redis

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/cyllective/oauth-labs/lab05/server/internal/config"
)

var rdb *redis.Client

func Init() (*redis.Client, error) {
	if err := WaitForBackoff(10); err != nil {
		return nil, err
	}
	opts := config.GetRedisConfig()
	rdb = redis.NewClient(opts)
	return rdb, nil
}

func Get() *redis.Client {
	return rdb
}

func WaitForBackoff(tries int) error {
	opts := config.GetRedisConfig()
	ctx := context.Background()
	errs := 0
	for i := 0; i < tries; i++ {
		r := redis.NewClient(opts)
		if err := r.Ping(ctx).Err(); err == nil {
			log.Println("[redis]: connected!")
			return nil
		}

		errs++
		time.Sleep(time.Duration(errs) * time.Second)
		log.Printf("[redis]: failed to connect, retrying...")
	}

	log.Printf("[redis]: failed to connect, aborting.")
	return errors.New("redis failed to connect")
}
