package api

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type setupStatusResponse struct {
	Initialized    bool `json:"initialized"`
	SetupAvailable bool `json:"setup_available"`
}

func (server *Server) setupStatus(ctx *gin.Context) {
	adminCount, err := server.store.CountAdminUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	initialized := adminCount > 0
	ctx.JSON(http.StatusOK, setupStatusResponse{
		Initialized:    initialized,
		SetupAvailable: !initialized,
	})
}

type createSetupAdminRequest struct {
	SetupToken string `json:"setup_token" binding:"required"`
	Username   string `json:"username" binding:"required,alphanum"`
	Password   string `json:"password" binding:"required,min=6"`
	FullName   string `json:"full_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
}

func (server *Server) createSetupAdmin(ctx *gin.Context) {
	var req createSetupAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	adminCount, err := server.store.CountAdminUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if adminCount > 0 {
		ctx.JSON(http.StatusConflict, errorResponse(fmt.Errorf("setup already initialized")))
		return
	}

	if server.config.SetupToken == "" {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("setup token is not configured")))
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.SetupToken), []byte(server.config.SetupToken)) != 1 {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("invalid setup token")))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	userID, err := uuid.NewRandom()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	user, err := server.store.CreateUserWithRole(ctx, db.CreateUserWithRoleParams{
		ID:              userID,
		Username:        req.Username,
		HashedPassword:  hashedPassword,
		FullName:        req.FullName,
		Email:           req.Email,
		IsEmailVerified: true,
		Role:            util.Admin,
	})
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusConflict, errorResponse(fmt.Errorf("setup already initialized")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}
