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

func createRandomArticle(t *testing.T) Article {
	user := createRandomUser(t)

	ID, err := uuid.NewRandom()
	require.NoError(t, err)

	arg := CreateArticleParams{
		ID:        ID,
		Title:     util.RandomString(6),
		Summary:   util.RandomString(10),
		Content:   util.RandomString(50),
		IsPublish: false,
		Owner:     user.ID,
	}

	post, err := testStore.CreateArticle(context.Background(), arg)

	require.NoError(t, err)
	require.Equal(t, arg.ID, post.ID)
	require.Equal(t, arg.Title, post.Title)
	require.Equal(t, arg.Summary, post.Summary)
	require.Equal(t, arg.Content, post.Content)
	require.Equal(t, int32(0), post.Views)
	require.Equal(t, int32(0), post.Likes)
	require.False(t, post.IsPublish)
	require.Equal(t, arg.Owner, post.Owner)
	require.NotZero(t, post.CreatedAt)
	require.True(t, post.UpdatedAt.IsZero())
	require.True(t, post.DeletedAt.IsZero())

	return post
}

func TestCreateArticle(t *testing.T) {
	createRandomArticle(t)
}

func TestGetArticle(t *testing.T) {

	post := createRandomArticle(t)

	getArticle, err := testStore.GetArticle(context.Background(), post.ID)

	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, post.ID, getArticle.ID)
	require.Equal(t, post.Title, getArticle.Title)
	require.Equal(t, post.Summary, getArticle.Summary)
	require.Equal(t, post.Content, getArticle.Content)
	require.Equal(t, post.Views, getArticle.Views)
	require.Equal(t, post.Likes, getArticle.Likes)
	require.Equal(t, post.IsPublish, getArticle.IsPublish)
	require.Equal(t, post.Owner, getArticle.Owner)
	require.WithinDuration(t, post.CreatedAt, getArticle.CreatedAt, time.Second)
	require.WithinDuration(t, post.UpdatedAt, getArticle.UpdatedAt, time.Second)
	require.WithinDuration(t, post.DeletedAt, getArticle.DeletedAt, time.Second)
}

func TestListArticles(t *testing.T) {
	var lastArticle Article
	for i := 0; i < 10; i++ {
		lastArticle = createRandomArticle(t)
	}

	arg := ListArticlesParams{
		Limit:  10,
		Offset: 0,
		IsPublish: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	}

	posts, err := testStore.ListArticles(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, posts)

	for _, post := range posts {
		require.NotEmpty(t, post)
	}

	require.Equal(t, posts[0].Owner, lastArticle.Owner)
}

func TestUpdateArticleOnlyTitle(t *testing.T) {
	oldArticle := createRandomArticle(t)

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
	oldArticle := createRandomArticle(t)

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
	oldArticle := createRandomArticle(t)

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
	oldArticle := createRandomArticle(t)

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
	oldArticle := createRandomArticle(t)

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
	require.WithinDuration(t, oldArticle.CreatedAt, updatedArticle.CreatedAt, time.Second)
	require.NotZero(t, updatedArticle.UpdatedAt)
}

func TestDeleteArticle(t *testing.T) {
	article := createRandomArticle(t)

	err := testStore.DeleteArticle(context.Background(), article.ID)
	require.NoError(t, err)

	getArticle, err := testStore.GetArticle(context.Background(), article.ID)
	require.Equal(t, err, ErrRecordNotFound)
	require.Empty(t, getArticle)
}
