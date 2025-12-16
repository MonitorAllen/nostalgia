package api

import (
	"errors"
	"fmt"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
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
		// 获取文件
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(errors.New("请上传文件")))
			return
		}

		// 检查大小
		if file.Size > config.UploadFileSizeLimit {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(errors.New("文件大小超过限制")))
			return
		}

		openFile, err := file.Open()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(errors.New("无法读取上传的文件")))
		}

		// 读取文件的前 261 个字节
		head := make([]byte, 261)
		if _, err = openFile.Read(head); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(errors.New("解析文件失败")))
			return
		}

		// 通过文件头检测类型
		kind, err := filetype.Match(head)
		if err != nil || kind == filetype.Unknown {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(errors.New("未知的文件类型")))
			return
		}

		if !slices.Contains(config.UploadFileAllowedMime, kind.MIME.Value) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(errors.New("不支持的文件类型")))
			return
		}

		ctx.Set("file_header", file)
		ctx.Set("file_ext", kind.Extension)

		ctx.Next()
	}
}
