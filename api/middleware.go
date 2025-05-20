package api

import (
	"errors"
	"fmt"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"github.com/rs/zerolog/log"
	"net/http"
	"slices"
	"strings"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationTypeBearer)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func uploadFileMiddleware(config util.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Info().Msg("uploadFileMiddleware")
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "上传文件失败"}})
			return
		}

		if file.Size > 0 && file.Size > config.UploadFileSizeLimit {
			ctx.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "文件大小超过限制"}})
			return
		}

		openFile, err := file.Open()

		// 读取文件的前 261 个字节
		head := make([]byte, 261)
		_, err = openFile.Read(head)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "解析文件失败"}})
			return
		}

		// 通过文件内容检测类型
		kind, err := filetype.Match(head)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "未知的文件类型"}})
			return
		}

		if !slices.Contains(config.UploadFileAllowedMime, kind.MIME.Value) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": gin.H{"message": "不支持的文件类型"}})
			return
		}

		ctx.Next()
	}
}
