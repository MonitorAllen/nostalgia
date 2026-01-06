package db

import (
	"context"
	"testing"

	"github.com/MonitorAllen/nostalgia/util"
	"github.com/stretchr/testify/require"
)

func createRandomCategory(t *testing.T) Category {
	t.Helper()
	name := util.RandomString(6)

	category, err := testStore.CreateCategory(context.Background(), name)
	require.NoError(t, err)
	require.NotEmpty(t, category.ID)
	require.Equal(t, name, category.Name)
	require.NotZero(t, category.CreatedAt)
	return category
}

func TestCreateCategory(t *testing.T) {
	createRandomCategory(t)
}

func TestDeleteCategory(t *testing.T) {
	category := createRandomCategory(t)

	createRandomArticle(t, true, category.ID)
	createRandomArticle(t, true, category.ID)

	arg := DeleteCategoryTxParams{
		ID: category.ID,
		AfterDelete: func() error {
			return nil
		},
	}

	err := testStore.DeleteCategoryTx(context.Background(), arg)
	require.NoError(t, err)

	listArg := ListArticlesByCategoryIDParams{
		CategoryID: category.ID,
		Offset:     0,
		Limit:      2,
	}
	articleList, err := testStore.ListArticlesByCategoryID(context.Background(), listArg)
	require.NoError(t, err)
	require.Empty(t, articleList)
}

func TestUpdateCategory(t *testing.T) {
	category := createRandomCategory(t)

	arg := UpdateCategoryParams{
		Name: util.RandomString(6),
		ID:   category.ID,
	}

	updateCategory, err := testStore.UpdateCategory(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, updateCategory.ID, category.ID)
	require.NotEqual(t, updateCategory.Name, category.Name)
	require.Equal(t, updateCategory.CreatedAt, category.CreatedAt)
	require.NotZero(t, updateCategory.UpdatedAt)
}

func TestListAllCategories(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomCategory(t)
	}

	categories, err := testStore.ListAllCategories(context.Background())
	require.NoError(t, err)

	count, err := testStore.CountCategories(context.Background())
	require.NoError(t, err)

	require.Len(t, categories, int(count))
}

func TestListCategoriesCountArticles(t *testing.T) {
	cate1 := createRandomCategory(t)
	cate2 := createRandomCategory(t)

	for i := 0; i < 10; i++ {
		createRandomArticle(t, true, cate1.ID)
	}

	for i := 0; i < 5; i++ {
		createRandomArticle(t, true, cate2.ID)
	}

	categories, err := testStore.ListCategoriesCountArticles(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, categories)
}

func TestListArticlesByCategoryID(t *testing.T) {
	category := createRandomCategory(t)
	for i := 0; i < 10; i++ {
		createRandomArticle(t, true, category.ID)
	}

	arg := ListArticlesByCategoryIDParams{
		CategoryID: category.ID,
		Limit:      10,
		Offset:     0,
	}

	articles, err := testStore.ListArticlesByCategoryID(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, articles, 10)
	for _, article := range articles {
		require.NotEmpty(t, article)
		require.Equal(t, article.CategoryID, category.ID)
	}
}

func TestGetCategoryByName(t *testing.T) {
	category := createRandomCategory(t)

	gotCategory, err := testStore.GetCategoryByName(context.Background(), category.Name)
	require.NoError(t, err)
	require.Equal(t, category.Name, gotCategory.Name)
}
