package main

import (
	"context"
	"log"

	"task_api/internal/config"
	"task_api/internal/database"
	"task_api/internal/queue"
	workerPkg "task_api/internal/worker"
)

func main() {
	log.Println("starting worker process...")

	ctx := context.Background()

	redisConfig := config.NewRedisConfig()

	redisClient := database.ConnectRedis(redisConfig)

	redisQueue := queue.NewRedisQueue(redisClient)

	worker := workerPkg.NewNotificationWorker(redisQueue)

	worker.Start(ctx)
}