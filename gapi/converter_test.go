package gapi

import (
	"testing"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestConvertArticleIncludesAutomationFields(t *testing.T) {
	row := db.ListAllArticlesRow{
		ID:                  uuid.New(),
		Title:               "Automation draft",
		CreatedByAutomation: true,
		AutomationStatus:    "pending_review",
	}

	article := convertArticle(row)

	require.True(t, article.GetCreatedByAutomation())
	require.Equal(t, "pending_review", article.GetAutomationStatus())
}
