package controller

import (
	"errors"
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
		Refresh(ctx *gin.Context)
		Logout(ctx *gin.Context)
		LogoutAll(ctx *gin.Context)
		ListSessions(ctx *gin.Context)
		RevokeSession(ctx *gin.Context)
		JWKS(ctx *gin.Context)
	}

	authController struct {
		authService    service.AuthService
		jwtService     service.JWTService
		authValidation *validation.AuthValidation
		db             *gorm.DB
	}
)

func NewAuthController(injector *do.Injector, as service.AuthService) AuthController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)
	authValidation := validation.NewAuthValidation()
	return &authController{
		authService:    as,
		jwtService:     jwtService,
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

	if err := c.authValidation.ValidateRegisterRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Register(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_REGISTER_USER, utils.ClientErrorMessage(err,
			userDto.ErrCreateUser,
			userDto.ErrEmailAlreadyExists,
			userDto.ErrInvalidRole,
			userDto.ErrRoleNotAssigned,
			userDto.ErrVehicleRequired,
			userDto.ErrLicenseNumberExists,
			userDto.ErrInvalidProfilePicture,
			userDto.ErrUploadProfilePicture,
		), nil)
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

	if err := c.authValidation.ValidateLoginRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err, nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Login(ctx.Request.Context(), req, dto.LoginMeta{
		UserAgent: ctx.GetHeader("User-Agent"),
		IP:        ctx.ClientIP(),
	})
	if err != nil {
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_LOGIN, utils.ClientErrorMessage(err,
			dto.ErrInvalidCredentials,
			userDto.ErrEmailNotFound,
			userDto.ErrRoleNotAssigned,
		), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(userDto.MESSAGE_SUCCESS_LOGIN, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *authController) Refresh(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.authValidation.ValidateRefreshTokenRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Refresh(ctx.Request.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, dto.ErrRefreshTokenNotFound) ||
			errors.Is(err, dto.ErrRefreshTokenExpired) ||
			errors.Is(err, dto.ErrSessionRevoked) {
			status = http.StatusUnauthorized
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_REFRESH_TOKEN, utils.ClientErrorMessage(err,
			dto.ErrRefreshTokenNotFound,
			dto.ErrRefreshTokenExpired,
			dto.ErrSessionRevoked,
		), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_REFRESH_TOKEN, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *authController) JWKS(ctx *gin.Context) {
	ctx.Header("Cache-Control", "public, max-age=300")
	ctx.JSON(http.StatusOK, c.jwtService.JWKS())
}

func (c *authController) Logout(ctx *gin.Context) {
	sessionID := ctx.MustGet("session_id").(string)

	err := c.authService.Logout(ctx.Request.Context(), sessionID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_LOGOUT, utils.ClientErrorMessage(err), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_LOGOUT, nil)
	ctx.JSON(http.StatusOK, res)
}

func (c *authController) LogoutAll(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	err := c.authService.LogoutAll(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_LOGOUT_ALL, utils.ClientErrorMessage(err), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_LOGOUT_ALL, nil)
	ctx.JSON(http.StatusOK, res)
}

func (c *authController) ListSessions(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	sessionID := ctx.MustGet("session_id").(string)

	result, err := c.authService.ListSessions(ctx.Request.Context(), userID, sessionID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_LIST_SESSIONS, utils.ClientErrorMessage(err), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_LIST_SESSIONS, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *authController) RevokeSession(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	sessionID := ctx.Param("id")

	err := c.authService.RevokeSession(ctx.Request.Context(), userID, sessionID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, dto.ErrSessionNotFound) {
			status = http.StatusNotFound
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_REVOKE_SESSION, utils.ClientErrorMessage(err, dto.ErrSessionNotFound), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_REVOKE_SESSION, nil)
	ctx.JSON(http.StatusOK, res)
}
