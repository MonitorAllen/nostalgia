package db

import (
	"context"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateComment(t *testing.T) {
	article := createRandomArticle(t, false, 1)

	sendCommentUser := createRandomUser(t)

	arg := CreateCommentParams{
		Content:    util.RandomString(32),
		ArticleID:  article.ID,
		ParentID:   0,
		FromUserID: sendCommentUser.ID,
		ToUserID:   article.Owner,
	}

	comment, err := testStore.CreateComment(context.Background(), arg)

	require.NoError(t, err)
	require.NotZero(t, comment.ID)
	require.Equal(t, arg.Content, comment.Content)
	require.Equal(t, arg.ArticleID, comment.ArticleID)
	require.Zero(t, comment.ParentID)
	require.Equal(t, arg.FromUserID, comment.FromUserID)
	require.Equal(t, arg.ToUserID, comment.ToUserID)
	require.NotZero(t, comment.CreatedAt)
	require.True(t, comment.DeletedAt.IsZero())

}
