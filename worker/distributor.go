package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskDelayDeleteCache(ctx context.Context, payload *PayloadDelayDeleteCache, opts ...asynq.Option) error
	// DistributeTaskDelayDeleteCacheDefault 使用默认配置分发缓存删除任务
	DistributeTaskDelayDeleteCacheDefault(ctx context.Context, keys ...string) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{client: client}
}
