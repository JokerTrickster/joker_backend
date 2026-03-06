package _redis

import (
	"context"
	"fmt"
	"os"

	_aws "github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/JokerTrickster/joker_backend/shared/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var Client *redis.Client

const RankingKey = "food:rankings"

func InitRedis() error {
	ctx := context.Background()
	isLocal := os.Getenv("IS_LOCAL")
	var connectionString string
	if isLocal == "true" {
		connectionString = fmt.Sprintf("redis://%s:%s@localhost:6379/1", os.Getenv("REDIS_USER"), os.Getenv("REDIS_PASSWORD"))
	} else {
		dbInfos, err := _aws.AwsSsmGetParams(ctx, []string{"dev_backend_redis_user", "dev_backend_redis_password", "dev_backend_redis_host", "dev_backend_redis_port", "dev_backend_redis_db"})
		if err != nil {
			return err
		}
		connectionString = fmt.Sprintf("redis://:%s@%s:%s/%s",
			dbInfos[3], //password
			dbInfos[0], //host
			dbInfos[1], //port
			dbInfos[2], //db
		)
	}

	opt, err := redis.ParseURL(connectionString)
	if err != nil {
		logger.Error("Failed to parse Redis URL", zap.Error(err))
		return err
	}

	Client = redis.NewClient(opt)

	_, err = Client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	logger.Info("Connected to Redis")

	return nil
}
