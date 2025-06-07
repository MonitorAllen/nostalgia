package gapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if len(req.GetContent()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "file is empty")
	}

	if int64(len(req.GetContent())) > server.config.UploadFileSizeLimit {
		return nil, status.Errorf(codes.ResourceExhausted, "file size exceeds limit")
	}

	head := req.GetContent()
	if len(head) > 261 {
		head = head[:261]
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to detect file type: %v", err)
	}

	if !slices.Contains(server.config.UploadFileAllowedMime, kind.MIME.Value) {
		return nil, status.Errorf(codes.PermissionDenied, "unsupported file type: %s", kind.MIME.Value)
	}

	root := server.config.ResourcePath
	path := ""
	if req.GetArticleId() != "" {
		path = filepath.Join(root, "articles", req.GetArticleId())
	} else {
		path = filepath.Join(root, "temp")
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create dir: %v", err)
	}

	newFileName, err := uuid.NewRandom()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成文件名失败: %v", err)
	}

	saveFileName := fmt.Sprintf("/%s.%s", newFileName.String(), kind.Extension)

	filePath := path + saveFileName

	err = os.WriteFile(filePath, req.GetContent(), 0664)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save file: %v", err)
	}

	filePath = strings.TrimLeft(filePath, ".")

	resp := &pb.UploadFileResponse{
		Url:      fmt.Sprintf("%s/%s", server.config.Domain, filePath),
		Filename: saveFileName,
	}

	return resp, nil
}
