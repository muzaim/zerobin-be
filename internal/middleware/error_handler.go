package middleware

import (
	"net/http"

	"zerobin-be/pkg/apperror"
	"zerobin-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		appErr, ok := err.(*apperror.AppError)
		if !ok {
			appErr = apperror.New(apperror.TypeInternal, "An unexpected error occurred", err)
		}

		var statusCode int
		switch appErr.Type {
		case apperror.TypeNotFound:
			statusCode = http.StatusNotFound
		case apperror.TypeUnauthorized:
			statusCode = http.StatusUnauthorized
		case apperror.TypeForbidden:
			statusCode = http.StatusForbidden
		case apperror.TypeBadRequest:
			statusCode = http.StatusBadRequest
		case apperror.TypeConflict:
			statusCode = http.StatusConflict
		case apperror.TypeValidation:
			statusCode = http.StatusUnprocessableEntity
		default:
			statusCode = http.StatusInternalServerError
		}

		reqID, _ := c.Get(RequestIDKey)
		reqIDStr, _ := reqID.(string)

		if statusCode == http.StatusInternalServerError {
			log.Error().
				Str("request_id", reqIDStr).
				Err(appErr.Err).
				Msgf("Internal error occurred: %s", appErr.Message)

			response.Error(c, statusCode, "An unexpected internal server error occurred", nil)
		} else {
			if appErr.Err != nil {
				log.Debug().
					Str("request_id", reqIDStr).
					Err(appErr.Err).
					Msgf("Client error [%s]: %s", appErr.Type, appErr.Message)
			}
			response.JSON(c, statusCode, false, appErr.Message, nil, appErr.Detail)
		}
	}
}
