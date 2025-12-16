package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	baseResourceDir := "./resources"
	var folderPath string
	if articleID != "" {
		folderPath = filepath.Join(baseResourceDir, articleID)
	} else {
		folderPath = filepath.Join(baseResourceDir, "temp")
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
		}
		saveFileName = fmt.Sprintf("%s.%s", newUUID.String(), fileExt)
	}

	fullSavePath := filepath.Join(folderPath, saveFileName)

	if err := ctx.SaveUploadedFile(fileHeader, fullSavePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("文件保存失败")))
		return
	}

	relativePath := "/" + filepath.ToSlash(fullSavePath)
	relativePath = strings.Replace(relativePath, "/./", "/", 1)

	resp := uploadFileResponse{
		Url:      relativePath,
		Default:  relativePath,
		Filename: saveFileName,
	}

	ctx.JSON(http.StatusOK, resp)
}
