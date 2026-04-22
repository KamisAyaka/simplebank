package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributorTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOps asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOps)
	return &RedisTaskDistributor{
		client: client,
	}
}
