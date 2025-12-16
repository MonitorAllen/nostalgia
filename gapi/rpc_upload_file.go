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

	content := req.GetContent()
	if len(content) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "文件为空")
	}

	if int64(len(content)) > server.config.UploadFileSizeLimit {
		return nil, status.Errorf(codes.ResourceExhausted, "文件大小超过限制")
	}

	head := content
	if len(head) > 261 {
		head = head[:261]
	}

	kind, err := filetype.Match(head)
	if err != nil || kind == filetype.Unknown {
		return nil, status.Errorf(codes.InvalidArgument, "未知的文件类型")
	}

	if !slices.Contains(server.config.UploadFileAllowedMime, kind.MIME.Value) {
		return nil, status.Errorf(codes.Unimplemented, "不支持的文件类型: %s", kind.MIME.Value)
	}

	root := server.config.ResourcePath
	folderPath := ""
	if req.GetArticleId() != "" {
		folderPath = filepath.Join(root, "articles", req.GetArticleId())
	} else {
		folderPath = filepath.Join(root, "temp")
	}

	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建文件夹失败: %v", err)
	}

	var saveFileName string
	if req.GetType() == "cover" {
		saveFileName = fmt.Sprintf("cover.%s", kind.Extension)
	} else {
		newUUID, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "生成文件名失败: %v", err)
		}
		saveFileName = fmt.Sprintf("%s.%s", newUUID.String(), kind.Extension)
	}

	fullSavePath := filepath.Join(folderPath, saveFileName)

	err = os.WriteFile(fullSavePath, content, 0664)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "保存文件失败: %v", err)
	}

	relativePath := "/" + filepath.ToSlash(fullSavePath)

	relativePath = strings.Replace(relativePath, "/./", "/", 1)

	resp := &pb.UploadFileResponse{
		Url:      relativePath,
		Filename: saveFileName,
	}

	return resp, nil
}
