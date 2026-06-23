package gapi

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUploadCoverReturnsPublicResourceURL(t *testing.T) {
	articleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	pngBytes, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
	)
	require.NoError(t, err)

	absoluteResourcePath := filepath.Join(t.TempDir(), "resources")
	testCases := []struct {
		name             string
		resourcePath     string
		expectedDiskRoot string
	}{
		{
			name:             "AbsoluteResourcePath",
			resourcePath:     absoluteResourcePath,
			expectedDiskRoot: absoluteResourcePath,
		},
		{
			name:             "RelativeResourcePath",
			resourcePath:     "./resources",
			expectedDiskRoot: "./resources",
		},
		{
			name:             "BlankResourcePathDefaultsToResourcesDirectory",
			resourcePath:     "",
			expectedDiskRoot: "./resources",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			workingDir := t.TempDir()
			previousWorkingDir, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(workingDir))
			t.Cleanup(func() {
				require.NoError(t, os.Chdir(previousWorkingDir))
			})

			store := mockdb.NewMockStore(ctrl)
			server := newTestServer(t, newGAPITestStore(store), nil, nil)
			server.config.ResourcePath = testCase.resourcePath
			server.config.UploadFileSizeLimit = 1024
			server.config.UploadFileAllowedMime = []string{"image/png"}

			ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
			articleIDString := articleID.String()
			resp, err := server.UploadFile(ctx, &pb.UploadFileRequest{
				ArticleId: &articleIDString,
				Content:   pngBytes,
				Type:      "cover",
			})

			require.NoError(t, err)
			require.Equal(t, "cover.png", resp.GetFilename())
			require.Equal(t, "/resources/articles/"+articleID.String()+"/cover.png", resp.GetUrl())

			savedBytes, err := os.ReadFile(filepath.Join(testCase.expectedDiskRoot, "articles", articleID.String(), "cover.png"))
			require.NoError(t, err)
			require.Equal(t, pngBytes, savedBytes)
		})
	}
}
