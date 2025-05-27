package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/user/go-git-test/pkg/logger"
	"go.uber.org/zap"
)

// CustomValidator is a custom validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the provided struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// NewValidator creates a new validator
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// LoggerMiddleware logs request information
func LoggerMiddleware(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			start := c.Get("start").(int64)

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			log.Info("Request",
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", res.Status),
				zap.String("remote_ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
				zap.Int64("start", start),
			)

			return err
		}
	}
}

// ErrorHandler is a custom error handler for Echo
func ErrorHandler(log *logger.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var (
			code    = http.StatusInternalServerError
			message interface{}
		)

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = he.Message
		} else {
			message = http.StatusText(code)
		}

		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				err = c.NoContent(code)
			} else {
				err = c.JSON(code, map[string]interface{}{
					"error": message,
				})
			}
			if err != nil {
				log.Error("Error handling error", zap.Error(err))
			}
		}

		log.Error("Request error",
			zap.Error(err),
			zap.Int("status", code),
			zap.String("uri", c.Request().RequestURI),
		)
	}
}
