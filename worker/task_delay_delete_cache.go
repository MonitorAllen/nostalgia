package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskDelayDeleteCache = "task:delay_delete_cache"

type PayloadDelayDeleteCache struct {
	Keys []string `json:"keys"`
}

func (distributor *RedisTaskDistributor) DistributeTaskDelayDeleteCache(ctx context.Context, payload *PayloadDelayDeleteCache, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskDelayDeleteCache, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskDelayDeleteCache(ctx context.Context, task *asynq.Task) error {
	var payload PayloadDelayDeleteCache
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	for _, key := range payload.Keys {
		_ = processor.cache.Del(ctx, key)
	}

	return nil
}

// DistributeTaskDelayDeleteCacheDefault 使用默认配置分发缓存删除任务
// 默认配置：MaxRetry=3, Timeout=3s, Queue=critical
func (distributor *RedisTaskDistributor) DistributeTaskDelayDeleteCacheDefault(ctx context.Context, keys ...string) error {
	payload := &PayloadDelayDeleteCache{Keys: keys}
	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(3 * time.Second),
		asynq.Queue(QueueCritical),
	}
	return distributor.DistributeTaskDelayDeleteCache(ctx, payload, opts...)
}
