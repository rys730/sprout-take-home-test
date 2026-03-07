package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sprout-backend/internal/config"
	"sprout-backend/internal/handler"
	"sprout-backend/internal/infrastructure/auth"
	"sprout-backend/internal/infrastructure/database"
	"sprout-backend/internal/infrastructure/logger"
	"sprout-backend/internal/repository"
	"sprout-backend/internal/usecase"

	_ "sprout-backend/docs"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title           Sprout Accounting API
// @version         1.0
// @description     Backend API for Sprout Accounting — Chart of Accounts, General Journal, and AR Management.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Sprout Engineering
// @contact.email  engineering@sprout.co

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format **Bearer &lt;token&gt;**

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Env)

	logger.Info("Starting Sprout Backend Server")
	logger.Infof("Environment: %s", cfg.Env)

	db, err := database.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		logger.Fatalf("Failed to ping database: %v", err)
	}

	migrator := database.NewMigrator(db.GetDB())
	migrationsDir := "./migrations"
	if err := migrator.RunMigrations(migrationsDir); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	// Initialize repositories
	accountRepo := repository.NewAccountRepository(db.GetPool())

	// Initialize use cases
	accountUseCase := usecase.NewAccountUseCase(accountRepo)

	e := echo.New()

	e.HideBanner = true
	e.HidePort = false

	handler.SetupMiddleware(e, cfg, jwtManager)

	healthHandler := handler.NewHealthHandler()
	e.GET("/health", healthHandler.Check)
	e.GET("/docs/*", echoSwagger.WrapHandler)

	authHandler := handler.NewAuthHandler(jwtManager)
	accountHandler := handler.NewAccountHandler(accountUseCase)
	api := e.Group("/api")

	v1 := api.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// protected := v1.Group("", handler.SetupAuthMiddleware(jwtManager))
		{
			// Chart of Accounts routes
			// accounts := protected.Group("/accounts")
			{
				// accounts.GET("", accountHandler.ListAccounts)
				// accounts.GET("/tree", accountHandler.GetAccountTree)
				// accounts.GET("/:id", accountHandler.GetAccount)
				// accounts.POST("", accountHandler.CreateAccount)
				// accounts.PUT("/:id", accountHandler.UpdateAccount)
				// accounts.DELETE("/:id", accountHandler.DeleteAccount)
			}
		}
	}
	account := v1.Group("/accounts")
	account.GET("", accountHandler.ListAccounts)
	account.GET("/tree", accountHandler.GetAccountTree)
	account.GET("/:id", accountHandler.GetAccount)
	account.POST("", accountHandler.CreateAccount)
	account.PUT("/:id", accountHandler.UpdateAccount)
	account.DELETE("/:id", accountHandler.DeleteAccount)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Server starting on %s", addr)

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	logger.Info("Server stopped")
}
