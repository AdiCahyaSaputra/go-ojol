package controller

import (
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/validation"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	DispatchController interface {
		CalculateArgo(ctx *gin.Context)
		FindDriver(ctx *gin.Context)
	}

	dispatchController struct {
		dispatchService    service.DispatchService
		dispatchValidation *validation.DispatchValidation
		db                 *gorm.DB
	}
)

func NewDispatchController(injector *do.Injector, s service.DispatchService) DispatchController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	dispatchValidation := validation.NewDispatchValidation()

	return &dispatchController{
		dispatchService:    s,
		dispatchValidation: dispatchValidation,
		db:                 db,
	}
}

func (c *dispatchController) CalculateArgo(ctx *gin.Context) {
	var req dto.CalculateArgoRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.dispatchValidation.ValidateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// Find best route OSRM, Calculate the distance, Multiply by fare per distance + platform percentage
}

func (c *dispatchController) FindDriver(ctx *gin.Context) {
	var req dto.FindDriverRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.dispatchValidation.ValidateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// TODO: find nearby driver lookup by redis geoloc
}
