package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskNotifyAutomationDraft = "task:notify_automation_draft"

type PayloadNotifyAutomationDraft struct {
	Kind            string    `json:"kind"`
	ArticleID       uuid.UUID `json:"article_id"`
	Title           string    `json:"title"`
	CategoryName    string    `json:"category_name"`
	ReviewURL       string    `json:"review_url"`
	IdempotencyKey  string    `json:"idempotency_key"`
	GenerationModel string    `json:"generation_model"`
	ErrorMessage    string    `json:"error_message"`
	NotifyEmail     string    `json:"notify_email"`
}

func (distributor *RedisTaskDistributor) DistributeTaskNotifyAutomationDraft(
	ctx context.Context,
	payload *PayloadNotifyAutomationDraft,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskNotifyAutomationDraft, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskNotifyAutomationDraft(ctx context.Context, task *asynq.Task) error {
	var payload PayloadNotifyAutomationDraft
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	if payload.NotifyEmail == "" {
		return fmt.Errorf("missing notify email: %w", asynq.SkipRetry)
	}

	subject, content, err := buildAutomationDraftEmail(payload)
	if err != nil {
		return err
	}

	if err := processor.mailer.SendEmail(subject, content, []string{payload.NotifyEmail}, nil, nil, nil); err != nil {
		return fmt.Errorf("failed to send automation draft email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", payload.NotifyEmail).Msg("processed task")

	return nil
}

func buildAutomationDraftEmail(payload PayloadNotifyAutomationDraft) (string, string, error) {
	switch payload.Kind {
	case "success":
		subject := fmt.Sprintf("Nostalgia 自动化草稿待审核：%s", payload.Title)
		content := fmt.Sprintf(`自动化草稿已创建，请进入后台审核。<br/>
标题：%s<br/>
分类：%s<br/>
模型：%s<br/>
幂等键：%s<br/>
审核入口：<a target="blank" href="%s">%s</a><br/>`,
			payload.Title,
			payload.CategoryName,
			payload.GenerationModel,
			payload.IdempotencyKey,
			payload.ReviewURL,
			payload.ReviewURL,
		)
		return subject, content, nil
	case "failure":
		subject := fmt.Sprintf("Nostalgia 自动化草稿创建失败：%s", payload.Title)
		content := fmt.Sprintf(`自动化草稿创建失败，请检查自动化请求。<br/>
标题：%s<br/>
模型：%s<br/>
幂等键：%s<br/>
错误：%s<br/>`,
			payload.Title,
			payload.GenerationModel,
			payload.IdempotencyKey,
			payload.ErrorMessage,
		)
		return subject, content, nil
	default:
		return "", "", fmt.Errorf("invalid automation draft notification kind: %w", asynq.SkipRetry)
	}
}
