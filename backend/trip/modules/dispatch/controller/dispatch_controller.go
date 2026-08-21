package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
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
		res := utils.BuildResponseFailed("Validation failed", err, nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	reqCtx := context.WithValue(
		ctx.Request.Context(),
		"customer",
		(ctx.MustGet("customer")).(entities.Customer),
	)

	result, err := c.dispatchService.CalculateArgo(reqCtx, req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNoRoute) || errors.Is(err, service.ErrInvalidLatLong) || errors.Is(err, service.ErrUnknownVehicle) {
			status = http.StatusBadRequest
		} else if errors.Is(err, service.ErrOSRMUnavailable) {
			status = http.StatusBadGateway
		}

		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CALCULATE_ARGO, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CALCULATE_ARGO, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *dispatchController) FindDriver(ctx *gin.Context) {
	var req dto.FindDriverRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.dispatchValidation.ValidateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.dispatchService.FindDriver(ctx.Request.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidLatLong) {
			status = http.StatusBadRequest
		}

		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_FIND_DRIVER, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_FIND_DRIVER, result)
	ctx.JSON(http.StatusOK, res)
}
