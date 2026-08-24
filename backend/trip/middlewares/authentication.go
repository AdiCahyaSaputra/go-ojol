package middlewares

import (
	"net/http"
	"strings"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Authenticate(verifier jwks.Verifier, sessions session.Checker) gin.HandlerFunc {
	return authenticate(verifier, sessions)
}

func AuthenticateWS(verifier jwks.Verifier, sessions session.Checker) gin.HandlerFunc {
	return authenticate(verifier, sessions)
}

func authenticate(verifier jwks.Verifier, sessions session.Checker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		raw, ok := bearerToken(ctx)
		if !ok {
			if raw == "" {
				response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_FOUND, nil)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
				return
			}
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_VALID, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claims, err := verifier.Verify(raw)
		if err != nil {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_VALID, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		if claims.UserID == "" {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		sessionUUID, err := uuid.Parse(claims.SessionID)
		if err != nil {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_VALID, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		active, err := sessions.IsActive(ctx.Request.Context(), sessionUUID)
		if err != nil || !active {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		ctx.Set("token", raw)
		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)
		ctx.Set("role", claims.Role)
		ctx.Set("session_id", claims.SessionID)
		ctx.Next()
	}
}

func bearerToken(ctx *gin.Context) (raw string, ok bool) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader, false
	}
	raw = strings.TrimPrefix(authHeader, "Bearer ")
	return raw, raw != ""
}
