package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	maxLoginFailures = 4
	loginBlockWindow = 15 * time.Minute
)

type loginAttempt struct {
	count     int
	blockedAt time.Time
}

type LoginRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*loginAttempt
}

func NewLoginRateLimiter() *LoginRateLimiter {
	return &LoginRateLimiter{
		entries: make(map[string]*loginAttempt),
	}
}

func (l *LoginRateLimiter) key(ip, email string) string {
	return ip + "|" + strings.ToLower(strings.TrimSpace(email))
}

func (l *LoginRateLimiter) IsBlocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[key]
	if !ok {
		return false
	}
	if entry.count < maxLoginFailures {
		return false
	}
	if time.Since(entry.blockedAt) > loginBlockWindow {
		delete(l.entries, key)
		return false
	}
	return true
}

func (l *LoginRateLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[key]
	if !ok {
		entry = &loginAttempt{}
		l.entries[key] = entry
	}
	entry.count++
	if entry.count >= maxLoginFailures {
		entry.blockedAt = time.Now()
	}
}

func (l *LoginRateLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

var defaultLoginRateLimiter = NewLoginRateLimiter()

func LoginRateLimit() gin.HandlerFunc {
	return LoginRateLimitWith(defaultLoginRateLimiter)
}

func LoginRateLimitWith(l *LoginRateLimiter) gin.HandlerFunc {
	return loginRateLimit(l)
}

func (l *LoginRateLimiter) Key(ip, email string) string {
	return l.key(ip, email)
}

func loginRateLimit(l *LoginRateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Next()
			return
		}
		ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		email := loginEmailFromBody(bodyBytes)
		key := l.key(ctx.ClientIP(), email)

		if email != "" && l.IsBlocked(key) {
			res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_LOGIN, "too many failed login attempts", nil)
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, res)
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		ctx.Next()

		if ctx.Writer.Status() == http.StatusOK {
			if email != "" {
				l.Reset(key)
			}
			return
		}
		if ctx.GetBool("login_failed") && email != "" {
			l.RecordFailure(key)
		}
	}
}

func loginEmailFromBody(body []byte) string {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}
	return strings.TrimSpace(req.Email)
}
