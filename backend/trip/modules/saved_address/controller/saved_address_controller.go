package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/validation"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	SavedAddressController interface {
		List(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Create(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}

	savedAddressController struct {
		savedAddressService    service.SavedAddressService
		savedAddressValidation *validation.SavedAddressValidation
		db                     *gorm.DB
	}
)

func NewSavedAddressController(injector *do.Injector, s service.SavedAddressService) SavedAddressController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	savedAddressValidation := validation.NewSavedAddressValidation()
	return &savedAddressController{
		savedAddressService:    s,
		savedAddressValidation: savedAddressValidation,
		db:                     db,
	}
}

func (c *savedAddressController) List(ctx *gin.Context) {
	reqCtx, ok := c.withCustomer(ctx, dto.MESSAGE_FAILED_LIST_SAVED_ADDRESS)
	if !ok {
		return
	}

	result, err := c.savedAddressService.List(reqCtx)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrCustomerNotInCtx) {
			status = http.StatusBadRequest
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_LIST_SAVED_ADDRESS, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_LIST_SAVED_ADDRESS, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *savedAddressController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_SAVED_ADDRESS, "invalid id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	reqCtx, ok := c.withCustomer(ctx, dto.MESSAGE_FAILED_GET_SAVED_ADDRESS)
	if !ok {
		return
	}

	result, err := c.savedAddressService.GetByID(reqCtx, id)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, service.ErrCustomerNotInCtx):
			status = http.StatusBadRequest
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_SAVED_ADDRESS, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_SAVED_ADDRESS, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *savedAddressController) Create(ctx *gin.Context) {
	var req dto.SavedAddressCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.savedAddressValidation.ValidateCreateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err, nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	reqCtx, ok := c.withCustomer(ctx, dto.MESSAGE_FAILED_CREATE_SAVED_ADDRESS)
	if !ok {
		return
	}

	result, err := c.savedAddressService.Create(reqCtx, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrInvalidLatLong), errors.Is(err, service.ErrCustomerNotInCtx):
			status = http.StatusBadRequest
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_SAVED_ADDRESS, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_SAVED_ADDRESS, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *savedAddressController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_SAVED_ADDRESS, "invalid id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	var req dto.SavedAddressUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.savedAddressValidation.ValidateUpdateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err, nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	reqCtx, ok := c.withCustomer(ctx, dto.MESSAGE_FAILED_UPDATE_SAVED_ADDRESS)
	if !ok {
		return
	}

	result, err := c.savedAddressService.Update(reqCtx, id, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, service.ErrInvalidLatLong), errors.Is(err, service.ErrCustomerNotInCtx):
			status = http.StatusBadRequest
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_SAVED_ADDRESS, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_SAVED_ADDRESS, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *savedAddressController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_SAVED_ADDRESS, "invalid id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	reqCtx, ok := c.withCustomer(ctx, dto.MESSAGE_FAILED_DELETE_SAVED_ADDRESS)
	if !ok {
		return
	}

	if err := c.savedAddressService.Delete(reqCtx, id); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, service.ErrCustomerNotInCtx):
			status = http.StatusBadRequest
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_SAVED_ADDRESS, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_SAVED_ADDRESS, gin.H{})
	ctx.JSON(http.StatusOK, res)
}

func (c *savedAddressController) withCustomer(ctx *gin.Context, failMessage string) (context.Context, bool) {
	customerVal, exists := ctx.Get("customer")
	if !exists {
		res := utils.BuildResponseFailed(failMessage, dto.MESSAGE_CUSTOMER_NOT_FOUND_CTX, nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return nil, false
	}
	customer, ok := customerVal.(entities.Customer)
	if !ok {
		res := utils.BuildResponseFailed(failMessage, dto.MESSAGE_CUSTOMER_NOT_FOUND_CTX, nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return nil, false
	}
	return context.WithValue(ctx.Request.Context(), "customer", customer), true
}
