package db

import (
	"fmt"
	"os"

	"github.com/bo-er/mail-it/models"
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
		fmt.Fprintf(os.Stdout, "redis client is using empty password\n")
	}
	if db == 0 {
		fmt.Fprintf(os.Stdout, "redis client is using default database\n")
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
	Set(key string, mb models.MailBrief) error
	Get(key string, fileds ...string) ([]interface{}, error)
	LPush(key string, values ...interface{}) (int64, error)
	LGetAll(key string) ([]string, error)
	FlushDB() error
	SimpleSort(key, order string) ([]string, error)
}

func (rs *RedisStore) Set(key string, mb models.MailBrief) error {
	_, err := rs.client.HMSet(key, mb.MapFormat()).Result()
	return err
}

func (rs *RedisStore) Get(key string, fileds ...string) ([]interface{}, error) {
	results, err := rs.client.HMGet(key, fileds...).Result()
	return results, err

}

func (rs *RedisStore) LPush(key string, values ...interface{}) (int64, error) {
	return rs.client.LPush(key, values...).Result()
}

func (rs *RedisStore) LGetAll(key string) ([]string, error) {
	return rs.client.LRange(key, 0, -1).Result()
}
func (rs *RedisStore) FlushDB() error {
	return rs.client.FlushDB().Err()
}

func (rs *RedisStore) SimpleSort(key, order string) ([]string, error) {
	return rs.client.Sort(key, &redis.Sort{
		Order: order,
	}).Result()
}
