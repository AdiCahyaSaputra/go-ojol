package middlewares

import (
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/utils"
	"github.com/gin-gonic/gin"
)

func Authorize(enforcer pkgcasbin.Enforcer, resource, action string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		emailVal, exists := ctx.Get("email")
		email, _ := emailVal.(string)
		if !exists || email == "" {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		allowed, err := enforcer.Enforce(email, resource, action)
		if err != nil || !allowed {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		ctx.Next()
	}
}
