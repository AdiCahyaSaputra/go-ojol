package middlewares

import (
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ResolveProfileId(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, userIdOk := ctx.Get("user_id")
		role, roleOk := ctx.Get("role")

		if !userIdOk || !roleOk {
			response := utils.BuildResponseFailed(dto.MESSAGE_PROFILE_CONTEXT_NOT_FOUND, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		switch role {
		case "driver":
			var driver entities.Driver
			result := db.Where("user_id = ?", userId).First(&driver)

			if result.Error != nil {
				response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
				return
			}

			ctx.Set("driver", driver)
		case "customer":
			var customer entities.Customer
			result := db.Where("user_id = ?", userId).First(&customer)

			if result.Error != nil {
				response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
				return
			}

			ctx.Set("customer", customer)
		default:
			response := utils.BuildResponseFailed(dto.MESSAGE_ROLE_INVALID, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		ctx.Next()
	}
}
