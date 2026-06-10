package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthorizeAdmin(t *testing.T) {
	testCases := []struct {
		name          string
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, payload *token.Payload, err error)
	}{
		{
			name: "AdminRole",
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithUserBearerToken(t, tokenMaker, util.RandUserID(), util.RandomOwner(), util.Admin, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.NoError(t, err)
				require.NotNil(t, payload)
				require.Equal(t, util.Admin, payload.Role)
			},
		},
		{
			name: "VisitorRole",
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithUserBearerToken(t, tokenMaker, util.RandUserID(), util.RandomOwner(), util.Visitor, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Error(t, err)
				require.Nil(t, payload)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "MissingAuthorization",
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Error(t, err)
				require.Nil(t, payload)
			},
		},
		{
			name: "ExpiredToken",
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithUserBearerToken(t, tokenMaker, util.RandUserID(), util.RandomOwner(), util.Admin, -time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Error(t, err)
				require.Nil(t, payload)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			server := newTestServer(t, newGAPITestStore(store), nil, nil)
			ctx := tc.buildContext(t, server.tokenMaker)

			payload, _, err := server.authorizeAdmin(ctx)
			tc.checkResponse(t, payload, err)
		})
	}
}
