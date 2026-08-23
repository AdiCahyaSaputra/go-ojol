package dto

import (
	"errors"
	"time"
)

const (
	MESSAGE_FAILED_REFRESH_TOKEN        = "failed refresh token"
	MESSAGE_SUCCESS_REFRESH_TOKEN       = "success refresh token"
	MESSAGE_FAILED_LOGOUT               = "failed logout"
	MESSAGE_SUCCESS_LOGOUT              = "success logout"
	MESSAGE_FAILED_LOGOUT_ALL           = "failed logout all"
	MESSAGE_SUCCESS_LOGOUT_ALL          = "success logout all"
	MESSAGE_FAILED_LIST_SESSIONS        = "failed list sessions"
	MESSAGE_SUCCESS_LIST_SESSIONS       = "success list sessions"
	MESSAGE_FAILED_REVOKE_SESSION       = "failed revoke session"
	MESSAGE_SUCCESS_REVOKE_SESSION      = "success revoke session"
	MESSAGE_FAILED_SEND_PASSWORD_RESET  = "failed send password reset"
	MESSAGE_SUCCESS_SEND_PASSWORD_RESET = "success send password reset"
	MESSAGE_FAILED_RESET_PASSWORD       = "failed reset password"
	MESSAGE_SUCCESS_RESET_PASSWORD      = "success reset password"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrPasswordResetToken   = errors.New("password reset token invalid")
)

type (
	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	TokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Role         string `json:"role"`
	}

	LoginMeta struct {
		UserAgent string
		IP        string
	}

	SessionResponse struct {
		ID        string     `json:"id"`
		UserAgent *string    `json:"user_agent,omitempty"`
		IP        *string    `json:"ip,omitempty"`
		CreatedAt time.Time  `json:"created_at"`
		ExpiresAt time.Time  `json:"expires_at"`
		RevokedAt *time.Time `json:"revoked_at,omitempty"`
		IsCurrent bool       `json:"is_current"`
	}

	SendPasswordResetRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
)
