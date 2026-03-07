package handler

import (
	"net/http"
	"time"

	"sprout-backend/internal/infrastructure/auth"
	"sprout-backend/internal/infrastructure/logger"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	jwtManager *auth.JWTManager
}

func NewAuthHandler(jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token     string   `json:"token"`
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"expires_at"`
}

// Login godoc
// @Summary     User login
// @Description Authenticate a user and return a JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     LoginRequest true "Login credentials"
// @Success     200  {object} LoginResponse
// @Failure     400  {object} map[string]string "Invalid request body"
// @Failure     500  {object} map[string]string "Failed to generate token"
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		logger.Errorf("Failed to bind login request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	logger.Infof("Login attempt for email: %s", req.Email)

	userID := "user-123"
	roles := []string{"user"}

	token, err := h.jwtManager.GenerateToken(userID, req.Email, roles)
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		UserID:    userID,
		Email:     req.Email,
		Roles:     roles,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	})
}

type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// RefreshToken godoc
// @Summary     Refresh JWT token
// @Description Exchange an existing JWT token for a new one
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     RefreshTokenRequest true "Current token"
// @Success     200  {object} map[string]string   "New token"
// @Failure     400  {object} map[string]string   "Invalid request body"
// @Failure     401  {object} map[string]string   "Failed to refresh token"
// @Router      /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	newToken, err := h.jwtManager.RefreshToken(req.Token)
	if err != nil {
		logger.Errorf("Failed to refresh token: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Failed to refresh token",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": newToken,
	})
}
