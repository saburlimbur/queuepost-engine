package main

import (
	"os"

	"outpost-engine/internal/database"
	"outpost-engine/internal/worker"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	// Logger
	var logger *zap.Logger
	var err error
	if os.Getenv("LOG_LEVEL") == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("starting asynq worker")

	// DB (reuse same DB as server)
	db, err := database.ConnectDatabase(logger)
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}
	defer db.Close()

	// Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	processor := worker.NewProcessor(db, logger)

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypePublishPost, processor.HandlePublishPostTask)

	// Optional: health check / error handling
	// mux.HandleFunc(worker.TypeRefreshToken, processor.HandleRefreshTokenTask)

	if err := srv.Run(mux); err != nil {
		logger.Fatal("asynq server failed", zap.Error(err))
	}
}
