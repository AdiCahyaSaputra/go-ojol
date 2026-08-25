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
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	DispatchController interface {
		// Customer
		CalculateArgo(ctx *gin.Context)
		FindDriver(ctx *gin.Context)

		// Driver
		SetDriverMode(ctx *gin.Context)
		RespondOffer(ctx *gin.Context)
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
	if err := ctx.ShouldBindQuery(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.dispatchValidation.ValidateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err, nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.dispatchService.CalculateArgo(ctx.Request.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNoRoute) || errors.Is(err, service.ErrInvalidLatLong) {
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.dispatchValidation.ValidateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err, nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	customerVal, exists := ctx.Get("customer")
	if !exists {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_FIND_DRIVER, dto.MESSAGE_CUSTOMER_NOT_FOUND_CTX, nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	customer, ok := customerVal.(entities.Customer)
	if !ok {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_FIND_DRIVER, dto.MESSAGE_CUSTOMER_NOT_FOUND_CTX, nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}

	reqCtx := context.WithValue(ctx.Request.Context(), "customer", customer)

	result, err := c.dispatchService.FindDriver(reqCtx, req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidLatLong) || errors.Is(err, service.ErrNoRoute) || errors.Is(err, service.ErrCustomerNotInCtx) {
			status = http.StatusBadRequest
		} else if errors.Is(err, service.ErrOSRMUnavailable) {
			status = http.StatusBadGateway
		}

		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_FIND_DRIVER, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_FIND_DRIVER, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *dispatchController) SetDriverMode(ctx *gin.Context) {
	var req dto.SetDriverModeRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
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
		"user_id",
		(ctx.MustGet("user_id")).(string),
	)

	err := c.dispatchService.SetDriverMode(reqCtx, req)
	if err != nil {
		status := http.StatusInternalServerError

		res := utils.BuildResponseFailed(dto.MESSAGE_SET_DRIVER_MODE_FAILED, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SET_DRIVER_MODE_SUCCESS, gin.H{})
	ctx.JSON(http.StatusOK, res)
}

func (c *dispatchController) RespondOffer(ctx *gin.Context) {
	transactionID, err := uuid.Parse(ctx.Param("transaction_id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_RESPOND_OFFER, "invalid transaction_id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	var req dto.RespondOfferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.dispatchValidation.ValidateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err, nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	driverVal, exists := ctx.Get("driver")
	if !exists {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_RESPOND_OFFER, dto.MESSAGE_DRIVER_NOT_FOUND_CTX, nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	driver, ok := driverVal.(entities.Driver)
	if !ok {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_RESPOND_OFFER, dto.MESSAGE_DRIVER_NOT_FOUND_CTX, nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}

	reqCtx := context.WithValue(ctx.Request.Context(), "driver", driver)

	result, err := c.dispatchService.RespondOffer(reqCtx, transactionID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrInvalidOfferAction),
			errors.Is(err, service.ErrDriverNotInCtx):
			status = http.StatusBadRequest
		case errors.Is(err, service.ErrOfferNotFound),
			errors.Is(err, service.ErrNotOfferedDriver):
			status = http.StatusNotFound
		case errors.Is(err, service.ErrOfferUnavailable):
			status = http.StatusConflict
		}

		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_RESPOND_OFFER, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_RESPOND_OFFER, result)
	ctx.JSON(http.StatusOK, res)
}
