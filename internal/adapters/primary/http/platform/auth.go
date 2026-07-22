package platform

import (
	"errors"
	"net/http"

	domain "go-crud-db-p2/internal/core/domain/platform"
	"go-crud-db-p2/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *PlatformHandler) Register(ctx *gin.Context) {
	var request domain.RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	authResponse, err := h.authSvc.Register(ctx.Request.Context(), request)
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response.Created(authResponse))
}

func (h *PlatformHandler) Login(ctx *gin.Context) {
	var request domain.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	authResponse, err := h.authSvc.Login(ctx.Request.Context(), request)
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(authResponse, "signed in"))
}

func (h *PlatformHandler) GoogleLogin(ctx *gin.Context) {
	loginURL, err := h.authSvc.GoogleLoginURL(ctx.Request.Context())
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, loginURL)
}

func (h *PlatformHandler) GoogleLoginURL(ctx *gin.Context) {
	loginURL, err := h.authSvc.GoogleLoginURL(ctx.Request.Context())
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(map[string]string{"url": loginURL}, "google login url"))
}

func (h *PlatformHandler) GoogleCallback(ctx *gin.Context) {
	authResponse, err := h.authSvc.GoogleCallback(
		ctx.Request.Context(),
		ctx.Query("state"),
		ctx.Query("code"),
	)
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(authResponse, "signed in with google"))
}

func (h *PlatformHandler) UpdateProfile(ctx *gin.Context) {
	userID, ok := authUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "missing authenticated user"))
		return
	}

	var request domain.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	parsedID, err := domain.ParseUserID(userID)
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	user, err := h.authSvc.UpdateProfile(ctx.Request.Context(), parsedID, request)
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(user, "profile updated"))
}

func (h *PlatformHandler) ChangePassword(ctx *gin.Context) {
	userID, ok := authUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "missing authenticated user"))
		return
	}

	var request domain.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	parsedID, err := domain.ParseUserID(userID)
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	if err := h.authSvc.ChangePassword(ctx.Request.Context(), parsedID, request); err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(nil, "password changed"))
}

func (h *PlatformHandler) ForgotPassword(ctx *gin.Context) {
	var request domain.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	if err := h.authSvc.ForgotPassword(ctx.Request.Context(), request); err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(nil, "if the email exists, an otp has been sent"))
}

func (h *PlatformHandler) ResetPassword(ctx *gin.Context) {
	var request domain.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	if err := h.authSvc.ResetPassword(ctx.Request.Context(), request); err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(nil, "password has been reset"))
}

func (h *PlatformHandler) Me(ctx *gin.Context) {
	userID, ok := authUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "missing authenticated user"))
		return
	}

	user, err := h.authSvc.CurrentUser(ctx.Request.Context(), userID)
	if err != nil {
		handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OK(user))
}

func handleAuthError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidAuthRequest):
		ctx.JSON(http.StatusBadRequest, response.Error("INVALID_AUTH_REQUEST", err.Error()))
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		ctx.JSON(http.StatusConflict, response.Error("EMAIL_ALREADY_EXISTS", "email already exists"))
	case errors.Is(err, domain.ErrInvalidCredentials):
		ctx.JSON(http.StatusUnauthorized, response.Error("INVALID_CREDENTIALS", "invalid email or password"))
	case errors.Is(err, domain.ErrUnauthorized):
		ctx.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "unauthorized"))
	case errors.Is(err, domain.ErrInvalidPassword):
		ctx.JSON(http.StatusBadRequest, response.Error("INVALID_PASSWORD", "current password is incorrect"))
	case errors.Is(err, domain.ErrAccountLocked):
		ctx.JSON(http.StatusForbidden, response.Error("ACCOUNT_LOCKED", "account is locked, reset password to unlock"))
	case errors.Is(err, domain.ErrInvalidOTP):
		ctx.JSON(http.StatusBadRequest, response.Error("INVALID_OTP", "invalid otp"))
	case errors.Is(err, domain.ErrOTPExpired):
		ctx.JSON(http.StatusBadRequest, response.Error("OTP_EXPIRED", "otp has expired"))
	case errors.Is(err, domain.ErrGoogleAuthUnavailable):
		ctx.JSON(http.StatusServiceUnavailable, response.Error("GOOGLE_AUTH_UNAVAILABLE", "google sign in is not configured"))
	case errors.Is(err, domain.ErrGoogleInvalidOAuthState):
		ctx.JSON(http.StatusBadRequest, response.Error("INVALID_GOOGLE_OAUTH_STATE", "invalid google oauth state"))
	case errors.Is(err, domain.ErrGoogleEmailNotVerified):
		ctx.JSON(http.StatusBadRequest, response.Error("GOOGLE_EMAIL_NOT_VERIFIED", "google email is not verified"))
	case errors.Is(err, domain.ErrGoogleProfileUnavailable):
		ctx.JSON(http.StatusBadGateway, response.Error("GOOGLE_PROFILE_UNAVAILABLE", "could not read google profile"))
	default:
		ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, response.ErrorWithDetails("INTERNAL_SERVER_ERROR", "internal server error", err.Error()))
	}
}
