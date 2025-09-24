package db

import (
	"context"
	"testing"
	"time"

	"github.com/MonitorAllen/nostalgia/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomArticle(t *testing.T, isPublish bool, categoryID int64) Article {
	user := createRandomUser(t)

	// 如果没有指定分类，则随机生成一个
	if categoryID == 0 {
		category := createRandomCategory(t)
		categoryID = category.ID
	}

	ID, err := uuid.NewRandom()
	require.NoError(t, err)

	arg := CreateArticleParams{
		ID:         ID,
		Title:      util.RandomString(6),
		Summary:    util.RandomString(10),
		Content:    util.RandomString(50),
		IsPublish:  isPublish,
		Owner:      user.ID,
		CategoryID: categoryID,
	}

	article, err := testStore.CreateArticle(context.Background(), arg)

	require.NoError(t, err)
	require.Equal(t, arg.ID, article.ID)
	require.Equal(t, arg.Title, article.Title)
	require.Equal(t, arg.Summary, article.Summary)
	require.Equal(t, arg.Content, article.Content)
	require.Equal(t, int32(0), article.Views)
	require.Equal(t, int32(0), article.Likes)
	require.Equal(t, isPublish, article.IsPublish)
	require.Equal(t, arg.Owner, article.Owner)
	require.Equal(t, arg.CategoryID, article.CategoryID)
	require.NotZero(t, article.CreatedAt)
	require.True(t, article.UpdatedAt.IsZero())
	require.True(t, article.DeletedAt.IsZero())

	return article
}

func TestCreateArticle(t *testing.T) {
	createRandomArticle(t, false, 0)
}

func TestGetArticle(t *testing.T) {

	article := createRandomArticle(t, false, 0)

	getArticle, err := testStore.GetArticle(context.Background(), article.ID)

	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, article.ID, getArticle.ID)
	require.Equal(t, article.Title, getArticle.Title)
	require.Equal(t, article.Summary, getArticle.Summary)
	require.Equal(t, article.Content, getArticle.Content)
	require.Equal(t, article.Views, getArticle.Views)
	require.Equal(t, article.Likes, getArticle.Likes)
	require.Equal(t, article.IsPublish, getArticle.IsPublish)
	require.Equal(t, article.Owner, getArticle.Owner)
	require.WithinDuration(t, article.CreatedAt, getArticle.CreatedAt, time.Second)
	require.WithinDuration(t, article.UpdatedAt, getArticle.UpdatedAt, time.Second)
	require.WithinDuration(t, article.DeletedAt, getArticle.DeletedAt, time.Second)
}

func TestListArticles(t *testing.T) {
	var lastArticle Article
	for i := 0; i < 10; i++ {
		lastArticle = createRandomArticle(t, false, 1)
	}

	arg := ListArticlesParams{
		Limit:  10,
		Offset: 0,
		IsPublish: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	}

	articles, err := testStore.ListArticles(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, articles)

	for _, article := range articles {
		require.NotEmpty(t, article)
	}

	require.Equal(t, articles[0].Owner, lastArticle.Owner)
}

func TestUpdateArticleOnlyTitle(t *testing.T) {
	oldArticle := createRandomArticle(t, false, 1)

	arg := UpdateArticleParams{
		Title: pgtype.Text{
			String: util.RandomString(6),
			Valid:  true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ID: oldArticle.ID,
	}

	updatedArticle, err := testStore.UpdateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateArticle)
	require.NotEqual(t, oldArticle.Title, updatedArticle.Title)
	require.Equal(t, oldArticle.Summary, updatedArticle.Summary)
	require.Equal(t, oldArticle.Content, updatedArticle.Content)
	require.Equal(t, oldArticle.IsPublish, updatedArticle.IsPublish)
	require.WithinDuration(t, oldArticle.CreatedAt, updatedArticle.CreatedAt, time.Second)
	require.NotZero(t, updatedArticle.UpdatedAt)
}

func TestUpdateArticleOnlySummary(t *testing.T) {
	oldArticle := createRandomArticle(t, false, 1)

	arg := UpdateArticleParams{
		Summary: pgtype.Text{
			String: util.RandomString(6),
			Valid:  true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ID: oldArticle.ID,
	}

	updatedArticle, err := testStore.UpdateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateArticle)
	require.Equal(t, oldArticle.Title, updatedArticle.Title)
	require.NotEqual(t, oldArticle.Summary, updatedArticle.Summary)
	require.Equal(t, oldArticle.Content, updatedArticle.Content)
	require.Equal(t, oldArticle.IsPublish, updatedArticle.IsPublish)
	require.WithinDuration(t, oldArticle.CreatedAt, updatedArticle.CreatedAt, time.Second)
	require.NotZero(t, updatedArticle.UpdatedAt)
}

func TestUpdateArticleOnlyContent(t *testing.T) {
	oldArticle := createRandomArticle(t, false, 1)

	arg := UpdateArticleParams{
		Content: pgtype.Text{
			String: util.RandomString(32),
			Valid:  true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ID: oldArticle.ID,
	}

	updatedArticle, err := testStore.UpdateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateArticle)
	require.Equal(t, oldArticle.Title, updatedArticle.Title)
	require.Equal(t, oldArticle.Summary, updatedArticle.Summary)
	require.NotEqual(t, oldArticle.Content, updatedArticle.Content)
	require.Equal(t, oldArticle.IsPublish, updatedArticle.IsPublish)
	require.WithinDuration(t, oldArticle.CreatedAt, updatedArticle.CreatedAt, time.Second)
	require.NotZero(t, updatedArticle.UpdatedAt)
}

func TestUpdateArticleOnlyIsPublish(t *testing.T) {
	oldArticle := createRandomArticle(t, false, 1)

	arg := UpdateArticleParams{
		IsPublish: pgtype.Bool{
			Bool:  !oldArticle.IsPublish,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ID: oldArticle.ID,
	}

	updatedArticle, err := testStore.UpdateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateArticle)
	require.Equal(t, oldArticle.Title, updatedArticle.Title)
	require.Equal(t, oldArticle.Summary, updatedArticle.Summary)
	require.Equal(t, oldArticle.Content, updatedArticle.Content)
	require.NotEqual(t, oldArticle.IsPublish, updatedArticle.IsPublish)
	require.WithinDuration(t, oldArticle.CreatedAt, updatedArticle.CreatedAt, time.Second)
	require.NotZero(t, updatedArticle.UpdatedAt)
}

func TestUpdateArticleAllFields(t *testing.T) {
	oldArticle := createRandomArticle(t, false, 1)

	newCategory := createRandomCategory(t)

	arg := UpdateArticleParams{
		Title: pgtype.Text{
			String: util.RandomString(6),
			Valid:  true,
		},
		Summary: pgtype.Text{
			String: util.RandomString(6),
			Valid:  true,
		},
		Content: pgtype.Text{
			String: util.RandomString(32),
			Valid:  true,
		},
		IsPublish: pgtype.Bool{
			Bool:  !oldArticle.IsPublish,
			Valid: true,
		},
		CategoryID: pgtype.Int8{
			Int64: newCategory.ID,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ID: oldArticle.ID,
	}

	updatedArticle, err := testStore.UpdateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateArticle)
	require.NotEqual(t, oldArticle.Title, updatedArticle.Title)
	require.NotEqual(t, oldArticle.Summary, updatedArticle.Summary)
	require.NotEqual(t, oldArticle.Content, updatedArticle.Content)
	require.NotEqual(t, oldArticle.IsPublish, updatedArticle.IsPublish)
	require.NotEqual(t, oldArticle.CategoryID, updatedArticle.CategoryID)
	require.WithinDuration(t, oldArticle.CreatedAt, updatedArticle.CreatedAt, time.Second)
	require.NotZero(t, updatedArticle.UpdatedAt)
}

func TestDeleteArticle(t *testing.T) {
	article := createRandomArticle(t, false, 1)

	err := testStore.DeleteArticle(context.Background(), article.ID)
	require.NoError(t, err)

	getArticle, err := testStore.GetArticle(context.Background(), article.ID)
	require.Equal(t, err, ErrRecordNotFound)
	require.Empty(t, getArticle)
}

func TestIncrementArticleLikes(t *testing.T) {
	article := createRandomArticle(t, false, 1)

	err := testStore.IncrementArticleLikes(context.Background(), article.ID)
	require.NoError(t, err)
}

func TestIncrementArticleViews(t *testing.T) {
	article := createRandomArticle(t, false, 1)

	err := testStore.IncrementArticleViews(context.Background(), article.ID)
	require.NoError(t, err)
}
