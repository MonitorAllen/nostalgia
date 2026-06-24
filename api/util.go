package api

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type uploadFileResponse struct {
	Url      string `json:"url"`
	Default  string `json:"default"`
	Filename string `json:"filename"`
}

func (server *Server) uploadFile(ctx *gin.Context) {
	fileHeader := ctx.MustGet("file_header").(*multipart.FileHeader)
	fileExt := ctx.MustGet("file_ext").(string)

	articleID := ctx.PostForm("article_id")
	uploadType := ctx.PostForm("upload_type")

	baseResourceDir := util.ResolveResourcePath(server.config.ResourcePath)
	absBaseResourceDir, err := filepath.Abs(baseResourceDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("无法解析资源路径: %v", err)))
		return
	}

	var folderPath string
	if articleID != "" {
		// Validate articleID is a valid UUID to prevent path traversal
		validID, err := util.ValidateArticleID(articleID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("无效的文章ID: %v", err)))
			return
		}
		safePath, err := util.SafeJoinPath(absBaseResourceDir, validID.String())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("路径不合法: %v", err)))
			return
		}
		folderPath = safePath
	} else {
		folderPath = filepath.Join(absBaseResourceDir, "temp")
	}

	if err := os.MkdirAll(folderPath, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("创建目录失败：%v, folder: %s", err, folderPath)))
		return
	}

	var saveFileName string

	if uploadType == "cover" {
		saveFileName = fmt.Sprintf("cover.%s", fileExt)
	} else {
		newUUID, err := uuid.NewUUID()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("生成文件名失败")))
			return
		}
		saveFileName = fmt.Sprintf("%s.%s", newUUID.String(), fileExt)
	}

	fullSavePath := filepath.Join(folderPath, saveFileName)

	if err := ctx.SaveUploadedFile(fileHeader, fullSavePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("文件保存失败")))
		return
	}

	resourceRelativePath, err := filepath.Rel(absBaseResourceDir, fullSavePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("生成文件访问路径失败：%v", err)))
		return
	}
	relativePath := "/resources/" + filepath.ToSlash(resourceRelativePath)

	resp := uploadFileResponse{
		Url:      relativePath,
		Default:  relativePath,
		Filename: saveFileName,
	}

	ctx.JSON(http.StatusOK, resp)
}
