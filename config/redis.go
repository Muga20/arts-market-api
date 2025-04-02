package config

import (
	"github.com/hibiken/asynq"
	"log"
)

var RedisConfig = &asynq.RedisClientOpt{
	Addr:     "127.0.0.1:6379",
	Password: "",
	DB:       0,
}

func InitializeRedis() {
	// Create the Asynq client
	client := asynq.NewClient(*RedisConfig)
	defer client.Close()

	// Try pinging the Redis server
	if err := client.Ping(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully")
}
