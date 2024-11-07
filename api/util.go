package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"net/http"
)

type uploadFileResponse struct {
	Url      string `json:"url"`
	Filename string `json:"filename"`
}

func (server *Server) uploadFile(ctx *gin.Context) {
	uploadFile, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	newFileName, err := uuid.NewRandom()
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "生成文件名失败"}})
		return
	}

	openFile, err := uploadFile.Open()

	head := make([]byte, 261)
	_, err = openFile.Read(head)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "解析文件失败"}})
		return
	}

	kind, _ := filetype.Match(head)
	saveFileName := fmt.Sprintf("%s.%s", newFileName.String(), kind.Extension)
	filePath := fmt.Sprintf("./temp/upload/%s", saveFileName)

	err = ctx.SaveUploadedFile(uploadFile, filePath)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "文件保存失败"}})
		return
	}

	resp := uploadFileResponse{
		Url:      fmt.Sprintf("http://%s/%s/%s", server.config.HTTPServerAddress, "/temp/upload/", saveFileName),
		Filename: saveFileName,
	}

	ctx.JSON(http.StatusOK, resp)
}
