package controller

import (
	"errors"
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type TripController interface {
	Protected(ctx *gin.Context)
	GetActive(ctx *gin.Context)
	GetByID(ctx *gin.Context)
	StartTrip(ctx *gin.Context)
	CompleteTrip(ctx *gin.Context)
	CancelTrip(ctx *gin.Context)
}

type tripController struct {
	tripService service.TripService
}

func NewTripController(injector *do.Injector, tripService service.TripService) TripController {
	return &tripController{tripService: tripService}
}

func (c *tripController) Protected(ctx *gin.Context) {
	response := utils.BuildResponseSuccess("success", gin.H{
		"user_id": ctx.MustGet("user_id"),
		"email":   ctx.MustGet("email"),
		"role":    ctx.MustGet("role"),
	})
	ctx.JSON(http.StatusOK, response)
}

func (c *tripController) GetActive(ctx *gin.Context) {
	role, _ := ctx.Get("role")
	reqCtx := ctx.Request.Context()

	switch role {
	case constants.ENUM_ROLE_CUSTOMER:
		customer, ok := ctxMustCustomer(ctx)
		if !ok {
			return
		}
		result, err := c.tripService.GetActiveForCustomer(reqCtx, customer)
		c.writeTransactionResult(ctx, dto.MESSAGE_SUCCESS_GET_ACTIVE_TRANSACTION, dto.MESSAGE_FAILED_GET_ACTIVE_TRANSACTION, result, err)
	case constants.ENUM_ROLE_DRIVER:
		driver, ok := ctxMustDriver(ctx)
		if !ok {
			return
		}
		result, err := c.tripService.GetActiveForDriver(reqCtx, driver)
		c.writeTransactionResult(ctx, dto.MESSAGE_SUCCESS_GET_ACTIVE_TRANSACTION, dto.MESSAGE_FAILED_GET_ACTIVE_TRANSACTION, result, err)
	default:
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_ACTIVE_TRANSACTION, "unsupported role", nil)
		ctx.AbortWithStatusJSON(http.StatusForbidden, res)
	}
}

func (c *tripController) GetByID(ctx *gin.Context) {
	txID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_TRANSACTION, "invalid transaction id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	role, _ := ctx.Get("role")
	reqCtx := ctx.Request.Context()

	switch role {
	case constants.ENUM_ROLE_CUSTOMER:
		customer, ok := ctxMustCustomer(ctx)
		if !ok {
			return
		}
		result, err := c.tripService.GetByIDForCustomer(reqCtx, customer, txID)
		c.writeTransactionResult(ctx, dto.MESSAGE_SUCCESS_GET_TRANSACTION, dto.MESSAGE_FAILED_GET_TRANSACTION, result, err)
	case constants.ENUM_ROLE_DRIVER:
		driver, ok := ctxMustDriver(ctx)
		if !ok {
			return
		}
		result, err := c.tripService.GetByIDForDriver(reqCtx, driver, txID)
		c.writeTransactionResult(ctx, dto.MESSAGE_SUCCESS_GET_TRANSACTION, dto.MESSAGE_FAILED_GET_TRANSACTION, result, err)
	default:
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_TRANSACTION, "unsupported role", nil)
		ctx.AbortWithStatusJSON(http.StatusForbidden, res)
	}
}

func (c *tripController) StartTrip(ctx *gin.Context) {
	txID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_START_TRIP, "invalid transaction id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	driver, ok := ctxMustDriver(ctx)
	if !ok {
		return
	}

	result, err := c.tripService.StartTrip(ctx.Request.Context(), driver, txID)
	if err != nil {
		status := tripErrorStatus(err)
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_START_TRIP, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_START_TRIP, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *tripController) CompleteTrip(ctx *gin.Context) {
	txID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_COMPLETE_TRIP, "invalid transaction id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	driver, ok := ctxMustDriver(ctx)
	if !ok {
		return
	}

	result, err := c.tripService.CompleteTrip(ctx.Request.Context(), driver, txID)
	if err != nil {
		status := tripErrorStatus(err)
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_COMPLETE_TRIP, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_COMPLETE_TRIP, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *tripController) CancelTrip(ctx *gin.Context) {
	txID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CANCEL_TRIP, "invalid transaction id", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	role, _ := ctx.Get("role")
	reqCtx := ctx.Request.Context()

	switch role {
	case constants.ENUM_ROLE_CUSTOMER:
		customer, ok := ctxMustCustomer(ctx)
		if !ok {
			return
		}
		result, err := c.tripService.CancelTripAsCustomer(reqCtx, customer, txID)
		c.writeCancelResult(ctx, result, err)
	case constants.ENUM_ROLE_DRIVER:
		driver, ok := ctxMustDriver(ctx)
		if !ok {
			return
		}
		result, err := c.tripService.CancelTripAsDriver(reqCtx, driver, txID)
		c.writeCancelResult(ctx, result, err)
	default:
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CANCEL_TRIP, "unsupported role", nil)
		ctx.AbortWithStatusJSON(http.StatusForbidden, res)
	}
}

func (c *tripController) writeTransactionResult(ctx *gin.Context, successMsg, failMsg string, result dto.TransactionResponse, err error) {
	if err != nil {
		status := tripErrorStatus(err)
		res := utils.BuildResponseFailed(failMsg, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}
	res := utils.BuildResponseSuccess(successMsg, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *tripController) writeCancelResult(ctx *gin.Context, result dto.CancelTripResponse, err error) {
	if err != nil {
		status := tripErrorStatus(err)
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CANCEL_TRIP, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CANCEL_TRIP, result)
	ctx.JSON(http.StatusOK, res)
}

func tripErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrTransactionNotFound),
		errors.Is(err, service.ErrNoActiveTransaction):
		return http.StatusNotFound
	case errors.Is(err, service.ErrNotTransactionParticipant):
		return http.StatusForbidden
	case errors.Is(err, service.ErrInvalidStatusTransition):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func ctxMustCustomer(ctx *gin.Context) (entities.Customer, bool) {
	customerVal, exists := ctx.Get("customer")
	if !exists {
		res := utils.BuildResponseFailed("request failed", "customer not found in context", nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return entities.Customer{}, false
	}
	customer, ok := customerVal.(entities.Customer)
	if !ok {
		res := utils.BuildResponseFailed("request failed", "customer not found in context", nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return entities.Customer{}, false
	}
	return customer, true
}

func ctxMustDriver(ctx *gin.Context) (entities.Driver, bool) {
	driverVal, exists := ctx.Get("driver")
	if !exists {
		res := utils.BuildResponseFailed("request failed", "driver not found in context", nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return entities.Driver{}, false
	}
	driver, ok := driverVal.(entities.Driver)
	if !ok {
		res := utils.BuildResponseFailed("request failed", "driver not found in context", nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return entities.Driver{}, false
	}
	return driver, true
}
