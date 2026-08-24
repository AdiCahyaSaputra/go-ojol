package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/middlewares"
	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginRateLimit_BlocksAfterFourFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := middlewares.NewLoginRateLimiter()
	router := gin.New()
	router.POST("/login", middlewares.LoginRateLimitWith(limiter), func(ctx *gin.Context) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": false})
	})

	body, err := json.Marshal(userDto.UserLoginRequest{
		Email:    "user@example.com",
		Password: "wrong",
	})
	require.NoError(t, err)

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestLoginRateLimit_ResetClearsBlockedState(t *testing.T) {
	limiter := middlewares.NewLoginRateLimiter()
	key := limiter.Key("127.0.0.1", "reset@example.com")

	for i := 0; i < 4; i++ {
		limiter.RecordFailure(key)
	}
	assert.True(t, limiter.IsBlocked(key))

	limiter.Reset(key)
	assert.False(t, limiter.IsBlocked(key))
}
