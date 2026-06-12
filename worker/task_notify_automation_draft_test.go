package worker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type sentEmail struct {
	subject string
	content string
	to      []string
}

type fakeAutomationDraftMailer struct {
	messages []sentEmail
}

func (m *fakeAutomationDraftMailer) SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error {
	m.messages = append(m.messages, sentEmail{
		subject: subject,
		content: content,
		to:      to,
	})
	return nil
}

func TestProcessTaskNotifyAutomationDraftSuccess(t *testing.T) {
	mailer := &fakeAutomationDraftMailer{}
	processor := &RedisTaskProcessor{mailer: mailer}
	payload := &PayloadNotifyAutomationDraft{
		Kind:            "success",
		ArticleID:       uuid.New(),
		Title:           "Redis cache invalidation",
		CategoryName:    "Go",
		ReviewURL:       "https://example.com/backend/articles/123",
		IdempotencyKey:  "daily-redis-cache",
		GenerationModel: "codex-automation",
		NotifyEmail:     "owner@example.com",
	}
	task := newNotifyAutomationDraftTask(t, payload)

	err := processor.ProcessTaskNotifyAutomationDraft(context.Background(), task)
	require.NoError(t, err)
	require.Len(t, mailer.messages, 1)
	require.Equal(t, []string{"owner@example.com"}, mailer.messages[0].to)
	require.Contains(t, mailer.messages[0].subject, "自动化草稿")
	require.Contains(t, mailer.messages[0].content, payload.Title)
	require.Contains(t, mailer.messages[0].content, payload.CategoryName)
	require.Contains(t, mailer.messages[0].content, payload.ReviewURL)
	require.Contains(t, mailer.messages[0].content, payload.IdempotencyKey)
	require.Contains(t, mailer.messages[0].content, payload.GenerationModel)
}

func TestProcessTaskNotifyAutomationDraftFailure(t *testing.T) {
	mailer := &fakeAutomationDraftMailer{}
	processor := &RedisTaskProcessor{mailer: mailer}
	payload := &PayloadNotifyAutomationDraft{
		Kind:            "failure",
		Title:           "Redis cache invalidation",
		IdempotencyKey:  "daily-redis-cache",
		GenerationModel: "codex-automation",
		ErrorMessage:    "category does not exist",
		NotifyEmail:     "owner@example.com",
	}
	task := newNotifyAutomationDraftTask(t, payload)

	err := processor.ProcessTaskNotifyAutomationDraft(context.Background(), task)
	require.NoError(t, err)
	require.Len(t, mailer.messages, 1)
	require.Equal(t, []string{"owner@example.com"}, mailer.messages[0].to)
	require.Contains(t, mailer.messages[0].subject, "自动化草稿创建失败")
	require.Contains(t, mailer.messages[0].content, payload.Title)
	require.Contains(t, mailer.messages[0].content, payload.ErrorMessage)
	require.Contains(t, mailer.messages[0].content, payload.IdempotencyKey)
}

func TestProcessTaskNotifyAutomationDraftInvalidPayloadSkipsRetry(t *testing.T) {
	processor := &RedisTaskProcessor{mailer: &fakeAutomationDraftMailer{}}
	task := asynq.NewTask(TaskNotifyAutomationDraft, []byte(`{`))

	err := processor.ProcessTaskNotifyAutomationDraft(context.Background(), task)
	require.ErrorIs(t, err, asynq.SkipRetry)
}

func newNotifyAutomationDraftTask(t *testing.T, payload *PayloadNotifyAutomationDraft) *asynq.Task {
	t.Helper()

	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return asynq.NewTask(TaskNotifyAutomationDraft, data)
}
