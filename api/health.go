package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	healthStatusOK        = "ok"
	readinessStatusReady  = "ready"
	readinessStatusFailed = "not_ready"
	readinessCheckOK      = "ok"
	readinessCheckFailed  = "unavailable"
	readinessTimeout      = 2 * time.Second
)

func (server *Server) healthz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": healthStatusOK})
}

func (server *Server) readyz(ctx *gin.Context) {
	checkCtx, cancel := context.WithTimeout(ctx.Request.Context(), readinessTimeout)
	defer cancel()

	checks := gin.H{
		"database": readinessCheckOK,
		"redis":    readinessCheckOK,
	}
	statusCode := http.StatusOK
	status := readinessStatusReady

	if server.store == nil || server.store.Ping(checkCtx) != nil {
		checks["database"] = readinessCheckFailed
		statusCode = http.StatusServiceUnavailable
		status = readinessStatusFailed
	}

	if server.cache == nil || server.cache.Ping(checkCtx) != nil {
		checks["redis"] = readinessCheckFailed
		statusCode = http.StatusServiceUnavailable
		status = readinessStatusFailed
	}

	ctx.JSON(statusCode, gin.H{
		"status": status,
		"checks": checks,
	})
}
