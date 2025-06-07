package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const DATA_USER_CONTRIBUTIONS = "user:contributions"

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	FullName        string    `json:"full_name"`
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreateAt        time.Time `json:"create_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		ID:              user.ID,
		Username:        user.Username,
		FullName:        user.FullName,
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		CreateAt:        user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	UID, err := uuid.NewRandom()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			ID:             UID,
			Username:       req.Username,
			HashedPassword: hashedPassword,
			FullName:       req.FullName,
			Email:          req.Email,
		},
		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{UserID: user.ID}
			opts := []asynq.Option{
				asynq.MaxRetry(5),
				asynq.Timeout(5), // 谷歌API部分CDN无法连接，避免长时间等待
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	result, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		code, _ := db.ErrorCode(err)
		if code == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := newUserResponse(result.User)

	ctx.JSON(http.StatusOK, resp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("incorrect password")))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(user.ID, user.Username, user.Role, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(user.ID, user.Username, user.Role, server.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpireAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpireAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpireAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, resp)
}

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		log.Error().Err(err).Msg("验证token失败")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 检查session状态
	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		log.Error().Err(err).Msg("用户被拉黑")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if session.UserID != refreshPayload.UserID {
		err := fmt.Errorf("incorrect session user")
		log.Error().Err(err).Msg("不匹配的用户")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		log.Error().Err(err).Msg("refresh token不匹配")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		log.Error().Err(err).Msg("session过期")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(refreshPayload.UserID, refreshPayload.Username, refreshPayload.Role, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpireAt,
	}

	ctx.JSON(http.StatusOK, resp)
}

type verifyEmailRequest struct {
	EmailID    int64  `form:"email_id" binding:"required,min=1"`
	SecretCode string `form:"secret_code" binding:"required,min=32,max=128"`
}

type verifyEmailResponse struct {
	IsEmailVerified bool `json:"is_email_verified"`
}

func (server *Server) verifyEmail(ctx *gin.Context) {
	var req verifyEmailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.VerifyEmailTxParams{
		EmailId:    req.EmailID,
		SecretCode: req.SecretCode,
	}

	verifyEmailResult, err := server.store.VerifyEmailTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("failed to verify email")))
		return
	}

	resp := verifyEmailResponse{
		IsEmailVerified: verifyEmailResult.User.IsEmailVerified,
	}

	ctx.JSON(http.StatusOK, resp)
}

type githubContributions struct {
	Years []struct {
		Year  string `json:"year"`
		Total int64  `json:"total"`
		Range struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"range"`
	} `json:"years"`
	Contributions []struct {
		Date      string `json:"date"`
		Count     int64  `json:"count"`
		Color     string `json:"color"`
		Intensity string `json:"intensity"`
	} `json:"contributions"`
}

func contributionResponse(contributions githubContributions) githubContributions {
	currentDate := time.Now().Format(time.DateOnly)
	var prefixPoint int
	for i, contribution := range contributions.Contributions {
		if contribution.Date == currentDate {
			prefixPoint = i
			break
		}
	}
	return githubContributions{
		Years:         contributions.Years[:1],
		Contributions: contributions.Contributions[prefixPoint : prefixPoint+90],
	}
}

func (server *Server) contributions(ctx *gin.Context) {
	// 先在 redis 获取
	contributionsJsonStr, err := server.redisService.Get(DATA_USER_CONTRIBUTIONS)
	if err != nil && !errors.Is(err, redis.Nil) {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if contributionsJsonStr != "" {
		var contributions githubContributions
		err := json.Unmarshal([]byte(contributionsJsonStr), &contributions)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, contributions)
		return
	}

	reqUrl := "https://github-contributions.vercel.app/api/v1/MonitorAllen"

	proxyUrl, err := url.Parse(server.config.HTTPProxyAddr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Get(reqUrl)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("执行请求活动数据失败：%s", err.Error())))
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var contributions githubContributions
	err = json.Unmarshal(body, &contributions)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	contributions = contributionResponse(contributions)

	contributionsBytes, err := json.Marshal(contributions)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = server.redisService.Set(DATA_USER_CONTRIBUTIONS, string(contributionsBytes), 12*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, contributions)
}
