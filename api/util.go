package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"net/http"
	"strings"
)

type uploadFileRequest struct {
	ID string `uri:"id"`
}

type uploadFileResponse struct {
	Url      string `json:"url"`
	Filename string `json:"filename"`
}

func (server *Server) uploadFile(ctx *gin.Context) {
	var req uploadFileRequest

	_ = ctx.BindUri(&req)

	uploadFile, err := ctx.FormFile("upload")
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

	filePath := ""
	if req.ID != "" {
		filePath = fmt.Sprintf("./resources/%s/%s", req.ID, saveFileName)
	} else {
		filePath = fmt.Sprintf("./temp/upload/%s", saveFileName)
	}

	err = ctx.SaveUploadedFile(uploadFile, filePath)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "文件保存失败"}})
		return
	}

	filePath = strings.TrimLeft(filePath, ".")

	resp := uploadFileResponse{
		Url:      fmt.Sprintf("http://%s%s", "172.19.228.63:8080", filePath),
		Filename: saveFileName,
	}

	ctx.JSON(http.StatusOK, resp)
}
