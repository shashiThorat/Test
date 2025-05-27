package api

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"
	_ "github.com/user/go-git-test/docs" // Import generated docs
	"github.com/user/go-git-test/pkg/api/handlers"
	customMiddleware "github.com/user/go-git-test/pkg/api/middleware"
	"github.com/user/go-git-test/pkg/api/routes"
	"github.com/user/go-git-test/pkg/logger"
	"github.com/user/go-git-test/pkg/provider"
	"go.uber.org/zap"
)

// Config represents the API server configuration
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Version         string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		Version:         "1.0.0",
	}
}

// Server represents the API server
type Server struct {
	echo            *echo.Echo
	config          *Config
	logger          *logger.Logger
	providerFactory *provider.Factory
}

// NewServer creates a new API server
func NewServer(config *Config, logger *logger.Logger, providerFactory *provider.Factory) *Server {
	e := echo.New()
	
	// Setup validator
	e.Validator = customMiddleware.NewValidator()
	
	// Setup error handler
	e.HTTPErrorHandler = customMiddleware.ErrorHandler(logger)
	
	// Setup middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("start", time.Now().UnixNano())
			return next(c)
		}
	})
	e.Use(customMiddleware.LoggerMiddleware(logger))
	
	// Setup Swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	
	server := &Server{
		echo:            e,
		config:          config,
		logger:          logger,
		providerFactory: providerFactory,
	}
	
	// Create handlers
	h := handlers.NewHandlers(logger, providerFactory, config.Version)
	
	// Register routes
	routes.RegisterRoutes(e, h)
	
	return server
}

// Start starts the API server
func (s *Server) Start() error {
	s.logger.Info("Starting API server", 
		zap.Int("port", s.config.Port),
		zap.String("swagger_url", fmt.Sprintf("http://localhost:%d/swagger/index.html", s.config.Port)))
	return s.echo.Start(fmt.Sprintf(":%d", s.config.Port))
}

// Stop stops the API server
func (s *Server) Stop() error {
	s.logger.Info("Stopping API server")
	return s.echo.Close()
}