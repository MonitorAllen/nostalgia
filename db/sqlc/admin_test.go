package db

import (
	"context"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomAdmin(t *testing.T) Admin {
	hashPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateAdminParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashPassword,
		IsActive:       true,
		RoleID:         util.RandomInt(1, 2),
	}

	admin, err := testStore.CreateAdmin(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, admin)

	require.NotZero(t, admin.ID)
	require.Equal(t, arg.Username, admin.Username)
	require.Equal(t, arg.HashedPassword, admin.HashedPassword)
	require.True(t, admin.IsActive)
	require.Equal(t, arg.RoleID, admin.RoleID)

	require.True(t, admin.UpdatedAt.IsZero())
	require.NotZero(t, admin.CreatedAt)

	return admin
}

func TestCreateAdmin(t *testing.T) {
	_ = createRandomAdmin(t)
}

func TestGetAdmin(t *testing.T) {
	admin := createRandomAdmin(t)

	gotAdmin, err := testStore.GetAdmin(context.Background(), admin.Username)
	require.NoError(t, err)
	require.Equal(t, admin.ID, gotAdmin.ID)
	require.Equal(t, admin.Username, gotAdmin.Username)
	require.Equal(t, admin.HashedPassword, gotAdmin.HashedPassword)
	require.Equal(t, admin.IsActive, gotAdmin.IsActive)
	require.Equal(t, admin.RoleID, gotAdmin.RoleID)
	require.Equal(t, admin.UpdatedAt, gotAdmin.UpdatedAt)
	require.WithinDuration(t, admin.CreatedAt, gotAdmin.CreatedAt, time.Second)
}
