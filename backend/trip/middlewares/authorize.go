package middlewares

import (
	"net/http"

	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/utils"
	"github.com/gin-gonic/gin"
)

func Authorize(enforcer pkgcasbin.Enforcer, role, resource, action string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		emailVal, exists := ctx.Get("email")
		email, _ := emailVal.(string)
		if !exists || email == "" {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		if role != "" {
			roleFromCtx, ok := ctx.Get("role")
			userRole := roleFromCtx.(string)
			if !ok || userRole == "" || userRole != role {
				response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
				ctx.AbortWithStatusJSON(http.StatusForbidden, response)
				return
			}
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
