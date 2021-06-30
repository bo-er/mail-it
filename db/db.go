package db

import (
	"fmt"
	"os"

	"github.com/bo-er/mail-it/user"
	"github.com/go-redis/redis"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr, password string, db int) *RedisStore {
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	if password == "" {
		fmt.Fprintf(os.Stdout, "redis client is using empty password")
	}
	if db == 0 {
		fmt.Fprintf("redis client is using default database")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	return &RedisStore{
		client: redisClient,
	}
}

type EmailStore interface {
	Set(key string, bm user.MailBrief) error
	Get(key string, bm user.MailBrief) error
}

func (rs *RedisStore) Set(key string, bm *user.MailBrief) error {

	_, err := rs.client.HMSet(key, bm.MapFormat()).Result()
	return err
}
