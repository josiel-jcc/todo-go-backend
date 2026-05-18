package handlers

import (
	"net/http"
	"todo-go-backend/internal/config"
	"todo-go-backend/internal/errors"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var appConfig *config.Config

// SetResponseConfig sets global config for error sanitization.
func SetResponseConfig(cfg *config.Config) {
	appConfig = cfg
}

func handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, ErrorResponse{
			Error:   appErr.Error(),
			Message: appErr.Message,
		})
		return
	}

	message := "Internal server error"
	if appConfig == nil || !appConfig.IsProduction() {
		message = err.Error()
	}
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "Internal server error",
		Message: message,
	})
}

func handleSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	response := SuccessResponse{Message: message}
	if data != nil {
		response.Data = data
	}
	c.JSON(statusCode, response)
}

func handleValidationError(c *gin.Context, err error) {
	handleError(c, errors.NewInvalidInputError(err.Error()))
}
