# Admin User Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a `/backend/users` admin module that lets the owner list, search, edit, disable, and restore public `visitor` accounts without changing the single-admin model.

**Architecture:** Add explicit disable state to `users`, expose admin-only gRPC-Gateway endpoints under `/v1/users`, block sessions when disabling, and reject disabled accounts during public login. The Vue admin page follows the existing dense backend table/modal style and calls the new `/v1` API through `adminHttp`.

**Tech Stack:** Go, Gin, gRPC-Gateway, protobuf, sqlc, PostgreSQL, gomock, Vue 3, TypeScript, Vite, Bun, Tailwind CSS, existing `AppButton`/`AppInput`/`AppBadge`/`ConfirmDialog` UI components.

---

## File Structure

- Create: `db/migration/000014_add_user_disabled_state.up.sql` - adds nullable disable state and listing index.
- Create: `db/migration/000014_add_user_disabled_state.down.sql` - removes disable state and index.
- Modify: `sqlc.yaml` - generate nullable timestamps as `pgtype.Timestamptz`.
- Modify: `db/query/user.sql` - visitor list/count/update/disable/enable queries.
- Modify: `db/query/session.sql` - query for blocking all sessions owned by a user.
- Generated: `db/sqlc/*.go` - sqlc query/model/store updates.
- Modify: `db/sqlc/store.go` - add transaction method to `Store`.
- Create: `db/sqlc/tx_admin_user.go` - disable-user transaction that updates user and blocks sessions atomically.
- Generated: `db/mock/store.go` - gomock Store interface refresh.
- Create: `proto/user.proto` - admin user messages and request/response types.
- Modify: `proto/service_nostalgia.proto` - import `user.proto` and register `/v1/users` RPCs.
- Generated: `pb/*.go` - protobuf, gRPC, and gateway outputs.
- Modify: `gapi/converter.go` - convert sqlc user rows to `pb.User`.
- Create: `gapi/rpc_list_users.go` - admin list handler.
- Create: `gapi/rpc_update_user.go` - admin profile update handler.
- Create: `gapi/rpc_disable_user.go` - admin disable handler.
- Create: `gapi/rpc_enable_user.go` - admin enable handler.
- Create: `gapi/rpc_list_users_test.go` - list auth/filter/pagination tests.
- Create: `gapi/rpc_update_user_test.go` - update visitor-only and duplicate-email tests.
- Create: `gapi/rpc_disable_enable_user_test.go` - disable/enable visitor-only/idempotent tests.
- Modify: `api/user.go` - reject disabled users during login.
- Modify: `api/user_test.go` - add disabled-login regression test.
- Modify: `web/frontend/src/admin/types.ts` - admin user types and request/response shapes.
- Create: `web/frontend/src/admin/api/adminUserApi.ts` - list/update/disable/enable API helpers.
- Create: `web/frontend/src/admin/adminUserManagement.test.ts` - source contract tests for route/sidebar/API/UI affordances.
- Modify: `web/frontend/src/router/index.ts` - add `adminUsers` route.
- Modify: `web/frontend/src/views/admin/AdminLayout.vue` - add sidebar "用户" nav after "分类".
- Create: `web/frontend/src/views/admin/AdminUserManagementView.vue` - backend user-management page.

---

## Task 1: Database Disable State and sqlc Queries

**Files:**
- Create: `db/migration/000014_add_user_disabled_state.up.sql`
- Create: `db/migration/000014_add_user_disabled_state.down.sql`
- Modify: `sqlc.yaml`
- Modify: `db/query/user.sql`
- Modify: `db/query/session.sql`
- Generated: `db/sqlc/*.go`

- [ ] **Step 1: Add the migration**

Create `db/migration/000014_add_user_disabled_state.up.sql`:

```sql
ALTER TABLE users
ADD COLUMN disabled_at timestamptz,
ADD COLUMN disabled_reason text NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS users_role_disabled_created_idx
ON users(role, disabled_at, created_at DESC);
```

Create `db/migration/000014_add_user_disabled_state.down.sql`:

```sql
DROP INDEX IF EXISTS users_role_disabled_created_idx;

ALTER TABLE users
DROP COLUMN IF EXISTS disabled_reason,
DROP COLUMN IF EXISTS disabled_at;
```

- [ ] **Step 2: Add visitor user-management SQL**

Append these queries to `db/query/user.sql`:

```sql
-- name: ListAdminUsers :many
SELECT
    id,
    username,
    full_name,
    email,
    is_email_verified,
    role,
    created_at,
    updated_at,
    disabled_at,
    disabled_reason
FROM users
WHERE role = 'visitor'
  AND (
    sqlc.arg(status)::text = 'all'
    OR (sqlc.arg(status)::text = 'enabled' AND disabled_at IS NULL)
    OR (sqlc.arg(status)::text = 'disabled' AND disabled_at IS NOT NULL)
  )
  AND (
    sqlc.narg(q)::text IS NULL
    OR username ILIKE '%' || sqlc.narg(q)::text || '%'
    OR full_name ILIKE '%' || sqlc.narg(q)::text || '%'
    OR email ILIKE '%' || sqlc.narg(q)::text || '%'
  )
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAdminUsersByFilter :one
SELECT count(*)
FROM users
WHERE role = 'visitor'
  AND (
    sqlc.arg(status)::text = 'all'
    OR (sqlc.arg(status)::text = 'enabled' AND disabled_at IS NULL)
    OR (sqlc.arg(status)::text = 'disabled' AND disabled_at IS NOT NULL)
  )
  AND (
    sqlc.narg(q)::text IS NULL
    OR username ILIKE '%' || sqlc.narg(q)::text || '%'
    OR full_name ILIKE '%' || sqlc.narg(q)::text || '%'
    OR email ILIKE '%' || sqlc.narg(q)::text || '%'
  );

-- name: UpdateVisitorUser :one
UPDATE users
SET
    full_name = sqlc.arg(full_name),
    email = sqlc.arg(email),
    is_email_verified = sqlc.arg(is_email_verified),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND role = 'visitor'
RETURNING *;

-- name: DisableVisitorUser :one
UPDATE users
SET
    disabled_at = COALESCE(disabled_at, now()),
    disabled_reason = sqlc.arg(disabled_reason),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND role = 'visitor'
RETURNING *;

-- name: EnableVisitorUser :one
UPDATE users
SET
    disabled_at = NULL,
    disabled_reason = '',
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND role = 'visitor'
RETURNING *;
```

- [ ] **Step 3: Configure nullable timestamp generation**

Modify the `timestamptz` overrides in `sqlc.yaml` so nullable timestamp columns generate as `pgtype.Timestamptz`, while existing non-null timestamp columns keep `time.Time`:

```yaml
        overrides:
          - db_type: "timestamptz"
            nullable: false
            go_type: "time.Time"
          - db_type: "timestamptz"
            nullable: true
            go_type:
              import: "github.com/jackc/pgx/v5/pgtype"
              type: "Timestamptz"
          - db_type: "uuid"
            go_type:
              import: "github.com/google/uuid"
              type: "UUID"
          - db_type: "bigint"
            go_type: "int64"
```

This is required because `users.disabled_at` is nullable and must preserve a real `NULL` state. Do not represent disabled state by comparing a nullable database value to Go's zero time.

- [ ] **Step 4: Add session blocking SQL**

Append this query to `db/query/session.sql`:

```sql
-- name: BlockUserSessions :exec
UPDATE sessions
SET is_blocked = true
WHERE user_id = $1;
```

- [ ] **Step 5: Generate sqlc code**

Run:

```bash
make sqlc
```

Expected:

- `db/sqlc/models.go` has `User.DisabledAt` and `User.DisabledReason`.
- `User.DisabledAt` is `pgtype.Timestamptz` and callers can distinguish `NULL` with `.Valid`.
- `db/sqlc/user.sql.go` contains `ListAdminUsers`, `CountAdminUsersByFilter`, `UpdateVisitorUser`, `DisableVisitorUser`, and `EnableVisitorUser`.
- `db/sqlc/session.sql.go` contains `BlockUserSessions`.
- `db/sqlc/querier.go` includes the new methods.

---

## Task 2: Disable Transaction and Database Tests

**Files:**
- Modify: `db/sqlc/store.go`
- Create: `db/sqlc/tx_admin_user.go`
- Create or modify: `db/sqlc/user_test.go`
- Generated: `db/mock/store.go`

- [ ] **Step 1: Add transaction method to Store**

Add this method to the `Store` interface in `db/sqlc/store.go`:

```go
DisableVisitorUserTx(ctx context.Context, arg DisableVisitorUserTxParams) (DisableVisitorUserTxResult, error)
```

- [ ] **Step 2: Implement the transaction**

Create `db/sqlc/tx_admin_user.go`:

```go
package db

import (
	"context"

	"github.com/google/uuid"
)

type DisableVisitorUserTxParams struct {
	ID             uuid.UUID `json:"id"`
	DisabledReason string    `json:"disabled_reason"`
}

type DisableVisitorUserTxResult struct {
	User User `json:"user"`
}

func (store *SQLStore) DisableVisitorUserTx(ctx context.Context, arg DisableVisitorUserTxParams) (DisableVisitorUserTxResult, error) {
	var result DisableVisitorUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		user, err := q.DisableVisitorUser(ctx, DisableVisitorUserParams{
			ID:             arg.ID,
			DisabledReason: arg.DisabledReason,
		})
		if err != nil {
			return err
		}

		if err := q.BlockUserSessions(ctx, arg.ID); err != nil {
			return err
		}

		result.User = user
		return nil
	})

	return result, err
}
```

If `store.execTx` has a different callback signature, follow the existing transaction files in `db/sqlc/tx_*.go` exactly.

- [ ] **Step 3: Add database tests**

Add focused tests to `db/sqlc/user_test.go`:

```go
func TestListAdminUsersFiltersVisitorsOnly(t *testing.T) {
	ctx := context.Background()
	visitor := createRandomUser(t)
	admin := createRandomAdminUser(t)

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
```

If the repository helper names differ, use existing `db/sqlc/*_test.go` helper style and keep the assertions above.

- [ ] **Step 4: Run database tests**

Run:

```bash
go test ./db/sqlc -count=1
```

Expected:

- PASS when PostgreSQL test database is available.
- If local database is unavailable, note the connection error and continue only after generating code compiles in later tasks.

- [ ] **Step 5: Regenerate mocks**

Run:

```bash
make mock
```

Expected:

- `db/mock/store.go` includes `DisableVisitorUserTx`.
- Existing gapi/api tests compile against the refreshed Store interface.

---

## Task 3: Proto Surface for Admin Users

**Files:**
- Create: `proto/user.proto`
- Modify: `proto/service_nostalgia.proto`
- Generated: `pb/*.go`

- [ ] **Step 1: Add user proto messages**

Create `proto/user.proto`:

```proto
syntax = "proto3";

package pb;

import "google/protobuf/timestamp.proto";

option go_package = "github.com/MonitorAllen/nostalgia/pb";

message User {
  string id = 1;
  string username = 2;
  string full_name = 3;
  string email = 4;
  bool is_email_verified = 5;
  string role = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
  google.protobuf.Timestamp disabled_at = 9;
  string disabled_reason = 10;
}

message ListUsersRequest {
  string q = 1;
  string status = 2;
  int32 page = 3;
  int32 limit = 4;
}

message ListUsersResponse {
  repeated User users = 1;
  int64 count = 2;
}

message UpdateUserRequest {
  string id = 1;
  string full_name = 2;
  string email = 3;
  bool is_email_verified = 4;
}

message UpdateUserResponse {
  User user = 1;
}

message DisableUserRequest {
  string id = 1;
  string reason = 2;
}

message DisableUserResponse {
  User user = 1;
}

message EnableUserRequest {
  string id = 1;
}

message EnableUserResponse {
  User user = 1;
}
```

- [ ] **Step 2: Register RPCs**

Modify `proto/service_nostalgia.proto`:

```proto
import "user.proto";
```

Add these RPCs inside `service Nostalgia`:

```proto
rpc ListUsers (ListUsersRequest) returns (ListUsersResponse) {
  option (google.api.http) = {
    get: "/v1/users"
  };
  option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
    description: "Use this API to list visitor users for admin management";
    summary: "list users";
    tags: "User";
  };
}

rpc UpdateUser (UpdateUserRequest) returns (UpdateUserResponse) {
  option (google.api.http) = {
    patch: "/v1/users/{id}"
    body: "*"
  };
  option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
    description: "Use this API to update visitor user profile fields";
    summary: "update user";
    tags: "User";
  };
}

rpc DisableUser (DisableUserRequest) returns (DisableUserResponse) {
  option (google.api.http) = {
    post: "/v1/users/{id}/disable"
    body: "*"
  };
  option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
    description: "Use this API to disable a visitor user and block sessions";
    summary: "disable user";
    tags: "User";
  };
}

rpc EnableUser (EnableUserRequest) returns (EnableUserResponse) {
  option (google.api.http) = {
    post: "/v1/users/{id}/enable"
    body: "*"
  };
  option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
    description: "Use this API to restore a disabled visitor user";
    summary: "enable user";
    tags: "User";
  };
}
```

- [ ] **Step 3: Generate protobuf code**

Run:

```bash
make proto
```

Expected:

- `pb/user.pb.go` exists.
- `pb/service_nostalgia.pb.go`, `pb/service_nostalgia_grpc.pb.go`, and `pb/service_nostalgia.pb.gw.go` include the four new RPCs.

---

## Task 4: gapi Admin User Handlers

**Files:**
- Modify: `gapi/converter.go`
- Create: `gapi/rpc_list_users.go`
- Create: `gapi/rpc_update_user.go`
- Create: `gapi/rpc_disable_user.go`
- Create: `gapi/rpc_enable_user.go`

- [ ] **Step 1: Add user converter**

Add to `gapi/converter.go`:

```go
func optionalTimestamp(value pgtype.Timestamptz) *timestamppb.Timestamp {
	if !value.Valid {
		return nil
	}
	return timestamppb.New(value.Time)
}

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Id:              user.ID.String(),
		Username:        user.Username,
		FullName:        user.FullName,
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		Role:            user.Role,
		CreatedAt:       timestamppb.New(user.CreatedAt),
		UpdatedAt:       timestamppb.New(user.UpdatedAt),
		DisabledAt:      optionalTimestamp(user.DisabledAt),
		DisabledReason:  user.DisabledReason,
	}
}

func convertAdminUserRow(user db.ListAdminUsersRow) *pb.User {
	return &pb.User{
		Id:              user.ID.String(),
		Username:        user.Username,
		FullName:        user.FullName,
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		Role:            user.Role,
		CreatedAt:       timestamppb.New(user.CreatedAt),
		UpdatedAt:       timestamppb.New(user.UpdatedAt),
		DisabledAt:      optionalTimestamp(user.DisabledAt),
		DisabledReason:  user.DisabledReason,
	}
}
```

Also add `github.com/jackc/pgx/v5/pgtype` to the imports in `gapi/converter.go`.

- [ ] **Step 2: Implement list handler**

Create `gapi/rpc_list_users.go`:

```go
package gapi

import (
	"context"
	"strings"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func normalizeUserListPage(page int32) int32 {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeUserListLimit(limit int32) int32 {
	switch limit {
	case 10, 20, 50:
		return limit
	default:
		return 20
	}
}

func normalizeUserStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "enabled", "disabled":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "all"
	}
}

func optionalSearchText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

func (server *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	page := normalizeUserListPage(req.GetPage())
	limit := normalizeUserListLimit(req.GetLimit())
	statusValue := normalizeUserStatus(req.GetStatus())
	q := optionalSearchText(req.GetQ())

	arg := db.ListAdminUsersParams{
		Limit:  limit,
		Offset: (page - 1) * limit,
		Status: statusValue,
		Q:      q,
	}

	users, err := server.store.ListAdminUsers(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	count, err := server.store.CountAdminUsersByFilter(ctx, db.CountAdminUsersByFilterParams{
		Status: statusValue,
		Q:      q,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count users: %v", err)
	}

	resp := &pb.ListUsersResponse{
		Users: make([]*pb.User, 0, len(users)),
		Count: count,
	}
	for _, user := range users {
		resp.Users = append(resp.Users, convertAdminUserRow(user))
	}

	return resp, nil
}
```

- [ ] **Step 3: Implement update handler**

Create `gapi/rpc_update_user.go`:

```go
package gapi

import (
	"context"
	"strings"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	fullName := strings.TrimSpace(req.GetFullName())
	email := strings.TrimSpace(req.GetEmail())
	if fullName == "" {
		return nil, status.Error(codes.InvalidArgument, "full name is required")
	}
	if email == "" || !strings.Contains(email, "@") {
		return nil, status.Error(codes.InvalidArgument, "valid email is required")
	}

	user, err := server.store.UpdateVisitorUser(ctx, db.UpdateVisitorUserParams{
		ID:              id,
		FullName:        fullName,
		Email:           email,
		IsEmailVerified: req.GetIsEmailVerified(),
	})
	if err != nil {
		if err == db.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.UpdateUserResponse{User: convertUser(user)}, nil
}
```

- [ ] **Step 4: Implement disable handler**

Create `gapi/rpc_disable_user.go`:

```go
package gapi

import (
	"context"
	"strings"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) DisableUser(ctx context.Context, req *pb.DisableUserRequest) (*pb.DisableUserResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	result, err := server.store.DisableVisitorUserTx(ctx, db.DisableVisitorUserTxParams{
		ID:             id,
		DisabledReason: strings.TrimSpace(req.GetReason()),
	})
	if err != nil {
		if err == db.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to disable user: %v", err)
	}

	return &pb.DisableUserResponse{User: convertUser(result.User)}, nil
}
```

- [ ] **Step 5: Implement enable handler**

Create `gapi/rpc_enable_user.go`:

```go
package gapi

import (
	"context"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) EnableUser(ctx context.Context, req *pb.EnableUserRequest) (*pb.EnableUserResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := server.store.EnableVisitorUser(ctx, id)
	if err != nil {
		if err == db.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to enable user: %v", err)
	}

	return &pb.EnableUserResponse{User: convertUser(user)}, nil
}
```

- [ ] **Step 6: Format**

Run:

```bash
gofmt -w gapi/converter.go gapi/rpc_list_users.go gapi/rpc_update_user.go gapi/rpc_disable_user.go gapi/rpc_enable_user.go db/sqlc/tx_admin_user.go db/sqlc/store.go
```

Expected: no output.

---

## Task 5: Backend Handler Tests and Login Guard

**Files:**
- Create: `gapi/rpc_list_users_test.go`
- Create: `gapi/rpc_update_user_test.go`
- Create: `gapi/rpc_disable_enable_user_test.go`
- Modify: `api/user.go`
- Modify: `api/user_test.go`

- [ ] **Step 1: Add gapi list tests**

Create `gapi/rpc_list_users_test.go` with these assertions:

```go
func TestListUsersNormalizesFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		ListAdminUsers(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.ListAdminUsersParams) ([]db.ListAdminUsersRow, error) {
			require.Equal(t, int32(20), arg.Limit)
			require.Equal(t, int32(0), arg.Offset)
			require.Equal(t, "all", arg.Status)
			require.True(t, arg.Q.Valid)
			require.Equal(t, "allen", arg.Q.String)
			return []db.ListAdminUsersRow{{ID: uuid.New(), Username: "allen", Role: util.Visitor}}, nil
		})
	store.EXPECT().
		CountAdminUsersByFilter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.CountAdminUsersByFilterParams) (int64, error) {
			require.Equal(t, "all", arg.Status)
			require.True(t, arg.Q.Valid)
			return 1, nil
		})

	resp, err := server.ListUsers(ctx, &pb.ListUsersRequest{
		Q:      " allen ",
		Status: "unknown",
		Page:   -1,
		Limit:  99,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), resp.GetCount())
	require.Len(t, resp.GetUsers(), 1)
}
```

- [ ] **Step 2: Add gapi update tests**

Create `gapi/rpc_update_user_test.go` with cases:

```go
func TestUpdateUserRejectsMissingAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)

	store.EXPECT().UpdateVisitorUser(gomock.Any(), gomock.Any()).Times(0)

	_, err := server.UpdateUser(context.Background(), &pb.UpdateUserRequest{Id: uuid.NewString()})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestUpdateUserMapsDuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		UpdateVisitorUser(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.User{}, &pgconn.PgError{Code: db.UniqueViolation})

	_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:              uuid.NewString(),
		FullName:        "Visitor",
		Email:           "visitor@example.com",
		IsEmailVerified: true,
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
}
```

Add an OK case that expects `UpdateVisitorUserParams` with trimmed `FullName` and `Email` and returns a `db.User`.

- [ ] **Step 3: Add disable/enable tests**

Create `gapi/rpc_disable_enable_user_test.go` with these cases:

```go
func TestDisableUserCallsTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	id := uuid.New()

	store.EXPECT().
		DisableVisitorUserTx(gomock.Any(), db.DisableVisitorUserTxParams{
			ID:             id,
			DisabledReason: "spam",
		}).
		Times(1).
		Return(db.DisableVisitorUserTxResult{User: db.User{
			ID:             id,
			Username:       "visitor",
			Role:           util.Visitor,
			DisabledReason: "spam",
			DisabledAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}}, nil)

	resp, err := server.DisableUser(ctx, &pb.DisableUserRequest{
		Id:     id.String(),
		Reason: " spam ",
	})

	require.NoError(t, err)
	require.Equal(t, id.String(), resp.GetUser().GetId())
	require.Equal(t, "spam", resp.GetUser().GetDisabledReason())
	require.NotNil(t, resp.GetUser().GetDisabledAt())
}

func TestEnableUserCallsQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	id := uuid.New()

	store.EXPECT().
		EnableVisitorUser(gomock.Any(), id).
		Times(1).
		Return(db.User{ID: id, Username: "visitor", Role: util.Visitor}, nil)

	resp, err := server.EnableUser(ctx, &pb.EnableUserRequest{Id: id.String()})

	require.NoError(t, err)
	require.Equal(t, id.String(), resp.GetUser().GetId())
	require.Nil(t, resp.GetUser().GetDisabledAt())
}
```

Add missing-auth and not-found cases for both methods.

- [ ] **Step 4: Guard public login**

In `api/user.go`, after `GetUserByUsername` succeeds and before `CheckPassword`, add:

```go
if user.DisabledAt.Valid {
	ctx.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("account disabled")))
	return
}
```

- [ ] **Step 5: Add login regression test**

Add a `DisabledUser` case to `TestUserLoginAPI` in `api/user_test.go`:

```go
{
	name: "DisabledUser",
	body: gin.H{
		"username": user.Username,
		"password": password,
	},
	buildStubs: func(store *mockdb.MockStore) {
		disabledUser := user
		disabledUser.DisabledAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}

		store.EXPECT().
			GetUserByUsername(gomock.Any(), gomock.Eq(user.Username)).
			Times(1).
			Return(disabledUser, nil)
		store.EXPECT().
			CreateSession(gomock.Any(), gomock.Any()).
			Times(0)
	},
	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		require.Equal(t, http.StatusForbidden, recorder.Code)
	},
},
```

- [ ] **Step 6: Run backend tests**

Run:

```bash
go test ./gapi ./api -count=1
```

Expected: PASS.

If this fails because generated mocks are stale, rerun:

```bash
make mock
go test ./gapi ./api -count=1
```

---

## Task 6: Frontend API Types, Route, and Source Contracts

**Files:**
- Modify: `web/frontend/src/admin/types.ts`
- Create: `web/frontend/src/admin/api/adminUserApi.ts`
- Modify: `web/frontend/src/router/index.ts`
- Modify: `web/frontend/src/views/admin/AdminLayout.vue`
- Create: `web/frontend/src/admin/adminUserManagement.test.ts`

- [ ] **Step 1: Add admin user types**

Append to `web/frontend/src/admin/types.ts`:

```ts
export type AdminUserStatusFilter = 'all' | 'enabled' | 'disabled'

export interface ManagedAdminUser {
  id: string
  username: string
  full_name: string
  email: string
  is_email_verified: boolean
  role: 'visitor'
  created_at: string
  updated_at?: string
  disabled_at?: string
  disabled_reason?: string
}

export interface AdminUserListResponse {
  users: ManagedAdminUser[]
  count: string | number
}

export interface UpdateAdminUserRequest {
  id: string
  full_name: string
  email: string
  is_email_verified: boolean
}

export interface DisableAdminUserRequest {
  reason?: string
}
```

- [ ] **Step 2: Add API helper**

Create `web/frontend/src/admin/api/adminUserApi.ts`:

```ts
import adminHttp from './adminHttp'
import type {
  AdminUserListResponse,
  AdminUserStatusFilter,
  DisableAdminUserRequest,
  ManagedAdminUser,
  UpdateAdminUserRequest,
} from '../types'

export interface ListAdminUsersParams {
  q?: string
  status: AdminUserStatusFilter
  page: number
  limit: number
}

export function listAdminUsers(params: ListAdminUsersParams) {
  return adminHttp.get<AdminUserListResponse>('/users', { params })
}

export function updateAdminUser(data: UpdateAdminUserRequest) {
  return adminHttp.patch<{ user: ManagedAdminUser }>(`/users/${data.id}`, data)
}

export function disableAdminUser(id: string, data: DisableAdminUserRequest) {
  return adminHttp.post<{ user: ManagedAdminUser }>(`/users/${id}/disable`, data)
}

export function enableAdminUser(id: string) {
  return adminHttp.post<{ user: ManagedAdminUser }>(`/users/${id}/enable`)
}
```

- [ ] **Step 3: Add route**

In `web/frontend/src/router/index.ts`, add the route after `adminCategories`:

```ts
{
  path: 'users',
  name: 'adminUsers',
  component: () => import('@/views/admin/AdminUserManagementView.vue')
},
```

- [ ] **Step 4: Add sidebar nav item**

In `web/frontend/src/views/admin/AdminLayout.vue`, import `Users` from `@lucide/vue` and insert this nav item after "分类":

```ts
{
  label: '用户',
  to: { name: 'adminUsers' },
  icon: Users,
  activeRoutes: ['adminUsers'],
},
```

- [ ] **Step 5: Add source contract tests**

Create `web/frontend/src/admin/adminUserManagement.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const src = (...parts: string[]) => resolve(import.meta.dir, '..', ...parts)
const read = (...parts: string[]) => readFileSync(src(...parts), 'utf8')

describe('admin user management source contract', () => {
  test('registers backend users route and sidebar item', () => {
    const router = read('router/index.ts')
    const layout = read('views/admin/AdminLayout.vue')

    expect(router).toContain("name: 'adminUsers'")
    expect(router).toContain("path: 'users'")
    expect(layout).toContain("label: '用户'")
    expect(layout).toContain("activeRoutes: ['adminUsers']")
  })

  test('admin user API uses expected /v1 endpoints', () => {
    const api = read('admin/api/adminUserApi.ts')

    expect(api).toContain("adminHttp.get<AdminUserListResponse>('/users'")
    expect(api).toContain('`/users/${data.id}`')
    expect(api).toContain('`/users/${id}/disable`')
    expect(api).toContain('`/users/${id}/enable`')
  })

  test('management view exposes required controls', () => {
    const view = read('views/admin/AdminUserManagementView.vue')

    expect(view).toContain('用户管理')
    expect(view).toContain('placeholder="搜索用户名、姓名或邮箱"')
    expect(view).toContain('pageSize')
    expect(view).toContain('jumpPage')
    expect(view).toContain('selectedStatus')
    expect(view).toContain('openEdit')
    expect(view).toContain('openDisable')
    expect(view).toContain('openEnable')
  })
})
```

This test will fail until Task 7 creates the view.

- [ ] **Step 6: Run frontend contract test and confirm the expected failure**

Run:

```bash
cd web/frontend && bun test src/admin/adminUserManagement.test.ts
```

Expected: fails because `AdminUserManagementView.vue` does not exist yet.

---

## Task 7: Frontend User Management View

**Files:**
- Create: `web/frontend/src/views/admin/AdminUserManagementView.vue`

- [ ] **Step 1: Create script state and handlers**

Create `web/frontend/src/views/admin/AdminUserManagementView.vue` with this script:

```vue
<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Check, Pencil, RotateCcw, Search, ShieldOff, X } from '@lucide/vue'
import {
  disableAdminUser,
  enableAdminUser,
  listAdminUsers,
  updateAdminUser,
} from '@/admin/api/adminUserApi'
import type { AdminUserStatusFilter, ManagedAdminUser } from '@/admin/types'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const users = ref<ManagedAdminUser[]>([])
const loading = ref(false)
const saving = ref(false)
const activeAction = ref('')
const total = ref(0)
const searchText = ref('')
const selectedStatus = ref<AdminUserStatusFilter>('all')
const page = ref(1)
const pageSize = ref(20)
const jumpPage = ref('1')
const editingUser = ref<ManagedAdminUser | null>(null)
const disablingUser = ref<ManagedAdminUser | null>(null)
const enablingUser = ref<ManagedAdminUser | null>(null)
const disableReason = ref('')

const editForm = reactive({
  full_name: '',
  email: '',
  is_email_verified: false,
})

const pageSizeOptions = [10, 20, 50]
const statusOptions: Array<{ label: string; value: AdminUserStatusFilter }> = [
  { label: '全部', value: 'all' },
  { label: '启用', value: 'enabled' },
  { label: '禁用', value: 'disabled' },
]

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const showingFrom = computed(() => (total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const showingTo = computed(() => Math.min(page.value * pageSize.value, total.value))

const fetchUsers = async () => {
  if (loading.value) return
  loading.value = true

  try {
    const response = await listAdminUsers({
      q: searchText.value.trim() || undefined,
      status: selectedStatus.value,
      page: page.value,
      limit: pageSize.value,
    })
    users.value = response.data.users ?? []
    total.value = Number(response.data.count || 0)

    if (users.value.length === 0 && page.value > 1) {
      page.value -= 1
      await fetchUsers()
    }
  } catch {
    users.value = []
  } finally {
    loading.value = false
  }
}

const resetToFirstPage = () => {
  page.value = 1
  jumpPage.value = '1'
  void fetchUsers()
}

const changePageSize = (event: Event) => {
  pageSize.value = Number((event.target as HTMLSelectElement).value)
  resetToFirstPage()
}

const changeStatus = (status: AdminUserStatusFilter) => {
  selectedStatus.value = status
  resetToFirstPage()
}

const goPage = (next: number) => {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  jumpPage.value = String(next)
  void fetchUsers()
}

const jumpToPage = () => {
  const next = Number(jumpPage.value)
  if (!Number.isFinite(next)) return
  goPage(Math.min(totalPages.value, Math.max(1, Math.floor(next))))
}

const openEdit = (user: ManagedAdminUser) => {
  editingUser.value = user
  editForm.full_name = user.full_name
  editForm.email = user.email
  editForm.is_email_verified = user.is_email_verified
}

const closeEdit = () => {
  if (saving.value) return
  editingUser.value = null
}

const saveEdit = async () => {
  if (!editingUser.value || saving.value) return
  const fullName = editForm.full_name.trim()
  const email = editForm.email.trim()

  if (!fullName || !email.includes('@')) {
    toast.add({
      severity: 'warning',
      summary: '用户信息不完整',
      detail: '请输入姓名和有效邮箱',
      life: 2400,
    })
    return
  }

  saving.value = true
  try {
    await updateAdminUser({
      id: editingUser.value.id,
      full_name: fullName,
      email,
      is_email_verified: editForm.is_email_verified,
    })
    editingUser.value = null
    await fetchUsers()
    toast.add({ severity: 'success', summary: '用户已更新', detail: email, life: 2400 })
  } catch {
    // Admin HTTP client already shows request errors.
  } finally {
    saving.value = false
  }
}

const openDisable = (user: ManagedAdminUser) => {
  disablingUser.value = user
  disableReason.value = user.disabled_reason || ''
}

const closeDisable = () => {
  if (activeAction.value) return
  disablingUser.value = null
}

const confirmDisable = async () => {
  if (!disablingUser.value) return
  activeAction.value = `disable:${disablingUser.value.id}`
  try {
    await disableAdminUser(disablingUser.value.id, { reason: disableReason.value.trim() || undefined })
    disablingUser.value = null
    await fetchUsers()
    toast.add({ severity: 'success', summary: '用户已禁用', detail: '该用户现有会话已阻断', life: 2600 })
  } catch {
    // Admin HTTP client already shows request errors.
  } finally {
    activeAction.value = ''
  }
}

const openEnable = (user: ManagedAdminUser) => {
  enablingUser.value = user
}

const closeEnable = () => {
  if (activeAction.value) return
  enablingUser.value = null
}

const confirmEnable = async () => {
  if (!enablingUser.value) return
  activeAction.value = `enable:${enablingUser.value.id}`
  try {
    await enableAdminUser(enablingUser.value.id)
    enablingUser.value = null
    await fetchUsers()
    toast.add({ severity: 'success', summary: '用户已恢复', detail: '用户可重新登录', life: 2600 })
  } catch {
    // Admin HTTP client already shows request errors.
  } finally {
    activeAction.value = ''
  }
}

const isDisabled = (user: ManagedAdminUser) => Boolean(user.disabled_at)
const isBusy = (key: string) => activeAction.value === key

const formatDate = (value?: string) => {
  if (!value) return '未记录'
  const date = new Date(value)
  if (!Number.isFinite(date.getTime()) || date.getFullYear() <= 1) return '未记录'
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

watch(page, (value) => {
  jumpPage.value = String(value)
})

onMounted(() => {
  void fetchUsers()
})
</script>
```

- [ ] **Step 2: Add dense admin table template**

Add this template to the same file:

```vue
<template>
  <section class="space-y-5">
    <header class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
      <div>
        <div class="flex items-center gap-3">
          <h1 class="m-0 text-2xl font-bold text-foreground">用户管理</h1>
          <AppBadge tone="neutral">{{ total }} 位用户</AppBadge>
        </div>
        <p class="mt-2 text-sm text-muted-foreground">管理前台注册的访客账号。</p>
      </div>
    </header>

    <div class="archive-surface rounded-archive p-4">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-1 items-center gap-2">
          <div class="relative flex-1">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <AppInput
              v-model="searchText"
              class="pl-10"
              placeholder="搜索用户名、姓名或邮箱"
              @keyup.enter="resetToFirstPage"
            />
          </div>
          <AppButton variant="secondary" @click="resetToFirstPage">搜索</AppButton>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <div class="inline-flex rounded-full border border-border bg-surface p-1">
            <button
              v-for="item in statusOptions"
              :key="item.value"
              type="button"
              class="h-9 rounded-full px-3 text-sm font-semibold transition"
              :class="selectedStatus === item.value ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
              @click="changeStatus(item.value)"
            >
              {{ item.label }}
            </button>
          </div>

          <select
            :value="pageSize"
            class="h-10 rounded-full border border-border bg-surface px-3 text-sm font-semibold text-foreground"
            @change="changePageSize"
          >
            <option v-for="size in pageSizeOptions" :key="size" :value="size">{{ size }} / 页</option>
          </select>
        </div>
      </div>

      <div class="mt-4 overflow-x-auto">
        <table class="min-w-full border-separate border-spacing-0 text-left text-sm">
          <thead class="text-xs uppercase text-muted-foreground">
            <tr>
              <th class="border-b border-border px-3 py-3 font-semibold">用户名</th>
              <th class="border-b border-border px-3 py-3 font-semibold">姓名</th>
              <th class="border-b border-border px-3 py-3 font-semibold">邮箱</th>
              <th class="border-b border-border px-3 py-3 font-semibold">邮箱状态</th>
              <th class="border-b border-border px-3 py-3 font-semibold">账号状态</th>
              <th class="border-b border-border px-3 py-3 font-semibold">注册时间</th>
              <th class="border-b border-border px-3 py-3 text-right font-semibold">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td class="px-3 py-8 text-center text-muted-foreground" colspan="7">加载中...</td>
            </tr>
            <tr v-else-if="users.length === 0">
              <td class="px-3 py-8 text-center text-muted-foreground" colspan="7">暂无用户</td>
            </tr>
            <tr v-for="user in users" v-else :key="user.id" class="border-b border-border/70">
              <td class="border-b border-border/70 px-3 py-3 font-semibold text-foreground">{{ user.username }}</td>
              <td class="border-b border-border/70 px-3 py-3 text-foreground">{{ user.full_name }}</td>
              <td class="border-b border-border/70 px-3 py-3 text-muted-foreground">{{ user.email }}</td>
              <td class="border-b border-border/70 px-3 py-3">
                <AppBadge :tone="user.is_email_verified ? 'accent' : 'warning'">
                  <Check v-if="user.is_email_verified" class="size-3" />
                  <X v-else class="size-3" />
                  {{ user.is_email_verified ? '已验证' : '未验证' }}
                </AppBadge>
              </td>
              <td class="border-b border-border/70 px-3 py-3">
                <AppBadge :tone="isDisabled(user) ? 'danger' : 'accent'">
                  {{ isDisabled(user) ? '已禁用' : '启用中' }}
                </AppBadge>
              </td>
              <td class="border-b border-border/70 px-3 py-3 text-muted-foreground">{{ formatDate(user.created_at) }}</td>
              <td class="border-b border-border/70 px-3 py-3">
                <div class="flex justify-end gap-2">
                  <AppButton size="sm" variant="ghost" @click="openEdit(user)">
                    <Pencil class="size-4" />
                    编辑
                  </AppButton>
                  <AppButton
                    v-if="isDisabled(user)"
                    size="sm"
                    variant="secondary"
                    :disabled="isBusy(`enable:${user.id}`)"
                    @click="openEnable(user)"
                  >
                    <RotateCcw class="size-4" />
                    恢复
                  </AppButton>
                  <AppButton
                    v-else
                    size="sm"
                    variant="danger"
                    :disabled="isBusy(`disable:${user.id}`)"
                    @click="openDisable(user)"
                  >
                    <ShieldOff class="size-4" />
                    禁用
                  </AppButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <footer class="mt-4 flex flex-col gap-3 text-sm text-muted-foreground lg:flex-row lg:items-center lg:justify-between">
        <span>显示 {{ showingFrom }} - {{ showingTo }}，共 {{ total }} 条</span>
        <div class="flex flex-wrap items-center gap-2">
          <AppButton size="sm" variant="ghost" :disabled="page <= 1" @click="goPage(page - 1)">上一页</AppButton>
          <span class="font-semibold text-foreground">{{ page }} / {{ totalPages }}</span>
          <AppButton size="sm" variant="ghost" :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</AppButton>
          <AppInput v-model="jumpPage" class="h-9 w-20 px-3 text-center" @keyup.enter="jumpToPage" />
          <AppButton size="sm" variant="secondary" @click="jumpToPage">跳转</AppButton>
        </div>
      </footer>
    </div>

    <Teleport to="body">
      <div v-if="editingUser" class="fixed inset-0 z-50 grid place-items-center bg-background/60 p-4 backdrop-blur-sm" role="dialog" aria-modal="true">
        <form class="archive-surface w-full max-w-lg rounded-archive p-5" @submit.prevent="saveEdit">
          <h2 class="m-0 text-lg font-bold text-foreground">编辑用户</h2>
          <div class="mt-4 space-y-3">
            <label class="block text-sm font-semibold text-foreground">
              姓名
              <AppInput v-model="editForm.full_name" class="mt-2" />
            </label>
            <label class="block text-sm font-semibold text-foreground">
              邮箱
              <AppInput v-model="editForm.email" class="mt-2" type="email" />
            </label>
            <label class="flex items-center gap-2 text-sm font-semibold text-foreground">
              <input v-model="editForm.is_email_verified" type="checkbox" class="size-4 rounded border-border accent-accent" />
              邮箱已验证
            </label>
          </div>
          <div class="mt-5 flex justify-end gap-2">
            <AppButton variant="ghost" :disabled="saving" @click="closeEdit">取消</AppButton>
            <AppButton type="submit" :disabled="saving">保存</AppButton>
          </div>
        </form>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="disablingUser" class="fixed inset-0 z-50 grid place-items-center bg-background/60 p-4 backdrop-blur-sm" role="dialog" aria-modal="true">
        <div class="archive-surface w-full max-w-md rounded-archive p-5">
          <h2 class="m-0 text-lg font-bold text-foreground">禁用用户</h2>
          <p class="mt-2 text-sm text-muted-foreground">禁用后该用户无法重新登录，现有刷新会话会被阻断。</p>
          <label class="mt-4 block text-sm font-semibold text-foreground">
            原因
            <textarea
              v-model="disableReason"
              class="mt-2 min-h-24 w-full rounded-2xl border border-border bg-surface px-4 py-3 text-sm text-foreground focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/18"
              placeholder="可选"
            />
          </label>
          <div class="mt-5 flex justify-end gap-2">
            <AppButton variant="ghost" :disabled="Boolean(activeAction)" @click="closeDisable">取消</AppButton>
            <AppButton variant="danger" :disabled="Boolean(activeAction)" @click="confirmDisable">确认禁用</AppButton>
          </div>
        </div>
      </div>
    </Teleport>

    <ConfirmDialog
      :open="Boolean(enablingUser)"
      title="恢复用户"
      description="恢复后用户可以重新登录，但旧会话不会自动恢复。"
      confirm-label="确认恢复"
      :danger="false"
      @cancel="closeEnable"
      @confirm="confirmEnable"
    />
  </section>
</template>
```

- [ ] **Step 3: Run frontend contract test**

Run:

```bash
cd web/frontend && bun test src/admin/adminUserManagement.test.ts
```

Expected: PASS.

- [ ] **Step 4: Run frontend type check**

Run:

```bash
cd web/frontend && bun run type-check
```

Expected: PASS.

---

## Task 8: Full Verification and Commit

**Files:**
- All files touched by Tasks 1-7.

- [ ] **Step 1: Run generated-code commands**

Run:

```bash
make sqlc
make proto
make mock
```

Expected:

- Commands complete successfully.
- Generated `pb/`, `db/sqlc/`, and `db/mock/` changes are present and intentional.

- [ ] **Step 2: Run backend verification**

Run:

```bash
go test ./db/sqlc ./gapi ./api -count=1
```

Expected: PASS.

If `./db/sqlc` fails due to unavailable PostgreSQL, run the compile-level fallback and record the database blocker:

```bash
go test ./gapi ./api -count=1
go test ./... -run '^$' -count=1
```

- [ ] **Step 3: Run frontend verification**

Run:

```bash
cd web/frontend && bun test src/admin/adminUserManagement.test.ts
cd web/frontend && bun run build
```

Expected: PASS.

- [ ] **Step 4: Check formatting and diff hygiene**

Run:

```bash
git diff --check
git status --short
```

Expected:

- `git diff --check` has no whitespace errors.
- `git status --short` only shows files related to admin user management.

- [ ] **Step 5: Commit**

Run:

```bash
git add db/migration db/query db/sqlc db/mock proto pb gapi api web/frontend/src docs/superpowers/plans/2026-06-18-admin-user-management.md
git commit -m "feat: add admin user management"
```

Expected:

- Commit succeeds on `feature/admin-user-management`.
- Do not push or open a PR until the user asks, unless this task is being executed under an explicit finish/PR instruction.

---

## Self-Review

- Spec coverage:
  - `/backend/users` route and "用户" sidebar item: Task 6 and Task 7.
  - Visitor-only list/search/status/pagination/page-size/page-jump: Task 1, Task 4, Task 7.
  - Edit full name/email/email verification state: Task 1, Task 4, Task 7.
  - Disable/restore account: Task 1, Task 2, Task 4, Task 7.
  - Block sessions on disable: Task 1 and Task 2.
  - Reject disabled login: Task 5.
  - No backend create/reset-password/role-management/admin-management: preserved by visitor-only SQL predicates and no extra UI actions.
  - Tests and verification: Tasks 2, 5, 6, 7, and 8.
- Placeholder scan:
  - No placeholder wording or unscoped test instructions remain.
- Type consistency:
  - API names match spec: `ListUsers`, `UpdateUser`, `DisableUser`, `EnableUser`.
  - Frontend route name is `adminUsers`.
  - Frontend API endpoint paths match `/v1/users`, `/v1/users/{id}`, `/v1/users/{id}/disable`, `/v1/users/{id}/enable`.
  - Nullable `disabled_at` is consistently represented as `pgtype.Timestamptz` with `.Valid` checks.
