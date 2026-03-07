package handler

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"sprout-backend/internal/config"
	"sprout-backend/internal/infrastructure/auth"
	"sprout-backend/internal/infrastructure/logger"
)

func SetupMiddleware(e *echo.Echo, cfg *config.Config, jwtManager *auth.JWTManager) {
	e.Use(middleware.Recover())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogError:     true,
		LogMethod:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogRequestID: true,
		LogHeaders:   []string{"User-Agent"},
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			logger.Infof("Request: %s %s | Status: %d | Latency: %v | IP: %s",
				values.Method,
				values.URI,
				values.Status,
				values.Latency,
				values.RemoteIP,
			)
			return nil
		},
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           int(12 * time.Hour.Seconds()),
	}))

	e.Use(middleware.Gzip())

	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
}

func SetupAuthMiddleware(jwtManager *auth.JWTManager) echo.MiddlewareFunc {
	secret := jwtManager.GetSecret()
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    []byte(secret),
		TokenLookup:   "header:Authorization:Bearer ",
		AuthScheme:    "Bearer",
		SigningMethod: "HS256",
		ContextKey:    "user",
		ErrorHandlerWithContext: func(err error, c echo.Context) error {
			logger.Warnf("Authentication failed: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
		},
	})
}

func RequireRoles(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user")
			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized",
				})
			}

			token := user.(*jwt.Token)
			claims := token.Claims.(*auth.Claims)

			if !hasRole(claims.Roles, roles) {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Forbidden",
				})
			}

			return next(c)
		}
	}
}

func hasRole(userRoles []string, requiredRoles []string) bool {
	roleMap := make(map[string]bool)
	for _, role := range userRoles {
		roleMap[role] = true
	}

	for _, required := range requiredRoles {
		if !roleMap[required] {
			return false
		}
	}

	return true
}
