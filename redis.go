package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

type redisStruct struct {
	addr    string
	pass    string
	channel string
	db      int
	client  *redis.Client
}

// create a redis connection
func connectRedis(ctx context.Context, rdis *redisStruct) {
	client := redis.NewClient(&redis.Options{
		Addr:     rdis.addr,
		Password: rdis.pass,
		DB:       rdis.db,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalln("Error ", err)
	}
	rdis.client = client
}
