package main

import (
	"context"
	"log"

	"task_api/internal/config"
	"task_api/internal/database"
	"task_api/internal/queue"
	"task_api/internal/repositories"
	workerPkg "task_api/internal/worker"
)

func main() {
	log.Println("starting worker process...")

	ctx := context.Background()
	db := database.ConnectPostgres()

	redisConfig := config.NewRedisConfig()

	redisClient := database.ConnectRedis(redisConfig)

	redisQueue := queue.NewRedisQueue(redisClient)
	notificationRepo := repositories.NewNotificationRepository(db)
	worker := workerPkg.NewNotificationWorker(redisQueue, notificationRepo)

	worker.Start(ctx)
}
