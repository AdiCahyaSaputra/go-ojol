package controller

import (
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/utils"
	"github.com/gin-gonic/gin"
)

type TripController interface {
	Protected(ctx *gin.Context)
}

type tripController struct{}

func NewTripController() TripController {
	return &tripController{}
}

func (c *tripController) Protected(ctx *gin.Context) {
	response := utils.BuildResponseSuccess("success", gin.H{
		"user_id": ctx.MustGet("user_id"),
		"email":   ctx.MustGet("email"),
		"role":    ctx.MustGet("role"),
	})
	ctx.JSON(http.StatusOK, response)
}
