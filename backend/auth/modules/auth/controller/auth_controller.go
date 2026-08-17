package controller

import (
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/validation"
	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	AuthController interface {
		Register(ctx *gin.Context)
		Login(ctx *gin.Context)
		Logout(ctx *gin.Context)
	}

	authController struct {
		authService    service.AuthService
		authValidation *validation.AuthValidation
		db             *gorm.DB
	}
)

func NewAuthController(injector *do.Injector, as service.AuthService) AuthController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	authValidation := validation.NewAuthValidation()
	return &authController{
		authService:    as,
		authValidation: authValidation,
		db:             db,
	}
}

func (c *authController) Register(ctx *gin.Context) {
	var req userDto.UserCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// Validate request
	if err := c.authValidation.ValidateRegisterRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Register(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_REGISTER_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(userDto.MESSAGE_SUCCESS_REGISTER_USER, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *authController) Login(ctx *gin.Context) {
	var req userDto.UserLoginRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Validate request
	if err := c.authValidation.ValidateLoginRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Login(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_LOGIN, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(userDto.MESSAGE_SUCCESS_LOGIN, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *authController) Logout(ctx *gin.Context) {
	userId := ctx.MustGet("user_id").(string)

	err := c.authService.Logout(ctx.Request.Context(), userId)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_LOGOUT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_LOGOUT, nil)
	ctx.JSON(http.StatusOK, res)
}
