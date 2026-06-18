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

func createRandomUser(t *testing.T) User {
	t.Helper()

	hashPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		ID:             util.RandUserID(),
		Username:       util.RandomOwner(),
		HashedPassword: hashPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.UpdatedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func createRandomSession(t *testing.T, userID uuid.UUID) Session {
	t.Helper()

	arg := CreateSessionParams{
		ID:           uuid.New(),
		UserID:       userID,
		RefreshToken: uuid.NewString(),
		UserAgent:    "db-test",
		ClientIp:     "127.0.0.1",
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	session, err := testStore.CreateSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, session)
	require.Equal(t, arg.ID, session.ID)
	require.Equal(t, arg.UserID, session.UserID)
	require.False(t, session.IsBlocked)

	return session
}

func collectUserIDs(users []ListAdminUsersRow) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testStore.GetUser(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.UpdatedAt, user2.UpdatedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUserOnlyFullName(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := util.RandomOwner()
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		ID: oldUser.ID,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	newEmail := util.RandomEmail()
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		ID: oldUser.ID,
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newHashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		ID: oldUser.ID,
		HashedPassword: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := util.RandomOwner()
	newEmail := util.RandomEmail()
	newHashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		ID: oldUser.ID,
		HashedPassword: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)

	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)

	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
}

func TestListAdminUsersFiltersVisitorsOnly(t *testing.T) {
	ctx := context.Background()
	visitor := createRandomUser(t)
	admin := getOrCreateAdminUser(t)

	users, err := testStore.ListAdminUsers(ctx, ListAdminUsersParams{
		Limit:  10,
		Offset: 0,
		Status: "all",
	})

	require.NoError(t, err)
	require.NotEmpty(t, users)
	require.Contains(t, collectUserIDs(users), visitor.ID)
	require.NotContains(t, collectUserIDs(users), admin.ID)
}

func TestDisableVisitorUserTxBlocksSessions(t *testing.T) {
	ctx := context.Background()
	user := createRandomUser(t)
	session := createRandomSession(t, user.ID)

	result, err := testStore.DisableVisitorUserTx(ctx, DisableVisitorUserTxParams{
		ID:             user.ID,
		DisabledReason: "spam",
	})

	require.NoError(t, err)
	require.Equal(t, user.ID, result.User.ID)
	require.Equal(t, "spam", result.User.DisabledReason)
	require.True(t, result.User.DisabledAt.Valid)

	blockedSession, err := testStore.GetSession(ctx, session.ID)
	require.NoError(t, err)
	require.True(t, blockedSession.IsBlocked)
}

func TestEnableVisitorUserClearsDisabledState(t *testing.T) {
	ctx := context.Background()
	user := createRandomUser(t)
	disabled, err := testStore.DisableVisitorUser(ctx, DisableVisitorUserParams{
		ID:             user.ID,
		DisabledReason: "temporary",
	})
	require.NoError(t, err)
	require.True(t, disabled.DisabledAt.Valid)

	enabled, err := testStore.EnableVisitorUser(ctx, user.ID)

	require.NoError(t, err)
	require.False(t, enabled.DisabledAt.Valid)
	require.Empty(t, enabled.DisabledReason)
}
