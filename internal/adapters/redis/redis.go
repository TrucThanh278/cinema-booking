package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewClient(address string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: address,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("error redis: %v", err)
	}
	log.Printf("connected to redis at address: %v", address)
	return rdb
}
