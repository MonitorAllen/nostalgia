package gapi

import (
	"testing"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestConvertArticleIncludesListMetadata(t *testing.T) {
	row := db.ListAllArticlesRow{
		ID:                  uuid.New(),
		Title:               "Automation draft",
		Cover:               "/resources/articles/cover.webp",
		CreatedByAutomation: true,
		AutomationStatus:    "pending_review",
	}

	article := convertArticle(row)

	require.Equal(t, "/resources/articles/cover.webp", article.GetCover())
	require.True(t, article.GetCreatedByAutomation())
	require.Equal(t, "pending_review", article.GetAutomationStatus())
}

func TestConvertOnlyArticleIncludesCover(t *testing.T) {
	row := db.Article{
		ID:    uuid.New(),
		Title: "Article draft",
		Cover: "/resources/articles/draft-cover.webp",
	}

	article := convertOnlyArticle(row, false)

	require.Equal(t, "/resources/articles/draft-cover.webp", article.GetCover())
}

func TestConvertArticleWithCategoryIncludesCover(t *testing.T) {
	row := db.GetArticleRow{
		ID:    uuid.New(),
		Title: "Preview article",
		Cover: "/resources/articles/preview-cover.webp",
	}

	article := convertArticleWithCategory(row, false)

	require.Equal(t, "/resources/articles/preview-cover.webp", article.GetCover())
}
