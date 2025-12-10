package db

import (
	"context"
	"fmt"
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

func TestSearchArticles(t *testing.T) {
	user := createRandomUser(t)
	category := createRandomCategory(t)

	// 1. 生成一个本次测试独有的随机特征码
	// 例如: "test_run_x8s7d6"
	uniqueTestTag := util.RandomString(10)

	// 2. 构造测试数据
	// 技巧：将 uniqueTestTag 拼接到 Title 或 Content 中

	// 文章 A: 命中 (标题包含特征码)
	articleA := CreateArticleParams{
		ID:         uuid.New(),
		Title:      fmt.Sprintf("Mastering Golang %s", uniqueTestTag), // 注入特征码
		Summary:    "Concurrency patterns",
		Content:    "Go is fast.",
		IsPublish:  true,
		Owner:      user.ID,
		CategoryID: category.ID,
	}
	_, err := testStore.CreateArticle(context.Background(), articleA)
	require.NoError(t, err)

	// 文章 B: 命中 (内容包含特征码)
	articleB := CreateArticleParams{
		ID:         uuid.New(),
		Title:      "Backend Dev",
		Summary:    "Summary info",
		Content:    fmt.Sprintf("We use %s for performance.", uniqueTestTag), // 注入特征码
		IsPublish:  true,
		Owner:      user.ID,
		CategoryID: category.ID,
	}
	_, err = testStore.CreateArticle(context.Background(), articleB)
	require.NoError(t, err)

	// 文章 C: 命中但被过滤 (未发布)
	articleC := CreateArticleParams{
		ID:         uuid.New(),
		Title:      fmt.Sprintf("Secret Draft %s", uniqueTestTag), // 注入特征码
		IsPublish:  false,                                         // 未发布
		Owner:      user.ID,
		CategoryID: category.ID,
		Content:    "Draft",
	}
	_, err = testStore.CreateArticle(context.Background(), articleC)
	require.NoError(t, err)

	// 文章 D: 噪音数据 (完全不包含特征码)
	articleD := CreateArticleParams{
		ID:         uuid.New(),
		Title:      "Rust Programming",
		Content:    "Safe memory without GC", // 这里的文本里没有 uniqueTestTag
		IsPublish:  true,
		Owner:      user.ID,
		CategoryID: category.ID,
	}
	_, err = testStore.CreateArticle(context.Background(), articleD)
	require.NoError(t, err)

	// -------------------------------------------------------
	// 测试场景 1: 搜索 特征码 (uniqueTestTag)
	// -------------------------------------------------------
	searchArg := SearchArticlesParams{
		Limit:   10,
		Offset:  0,
		Keyword: uniqueTestTag, // 只搜这个随机字符串
		IsPublish: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}

	results, err := testStore.SearchArticles(context.Background(), searchArg)
	require.NoError(t, err)

	// 断言：无论数据库里有多少脏数据，包含这个 uniqueTestTag 且已发布的，只有 A 和 B
	require.Len(t, results, 2)

	// ... 后续验证 ID 的逻辑不变 ...

	// -------------------------------------------------------
	// 测试场景 2: Count
	// -------------------------------------------------------
	countArg := CountSearchArticlesParams{
		Keyword: uniqueTestTag,
		IsPublish: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}
	count, err := testStore.CountSearchArticles(context.Background(), countArg)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}
