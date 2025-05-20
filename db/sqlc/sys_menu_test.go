package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitMenu(t *testing.T) {
	_, err := testStore.ListInitSysMenus(context.Background(), 1)
	require.NoError(t, err)
}
