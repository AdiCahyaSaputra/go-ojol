package middlewares

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/pkg/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/pkg/utils"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v\n%s", recovered, debug.Stack())
				res := utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, dto.MESSAGE_INTERNAL_SERVER_ERROR, nil)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
			}
		}()
		ctx.Next()
	}
}
