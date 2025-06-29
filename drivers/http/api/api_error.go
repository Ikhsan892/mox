package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"goodin/exceptions"
)

// ApiError defines the struct for a basic api error response.
type ApiError struct {
	Code      int       `json:"code"`
	RequestId string    `json:"request_id"`
	Message   string    `json:"message"`
	Status    string    `json:"event_code,omitempty"`
	Time      time.Time `json:"time"`

	// stores unformatted error data (could be an internal error, text, etc.)
	rawData any
}

// Error makes it compatible with the `error` interface.
func (e *ApiError) Error() string {
	return e.Message
}

// RawData returns the unformatted error data (could be an internal error, text, etc.)
func (e *ApiError) RawData() any {
	return e.rawData
}

func NewServiceError(err error) *ApiError {
	var srvErr *exceptions.BaseException

	if errors.As(err, &srvErr) {
		apiError := NewApiError(srvErr.Code, srvErr.Message, nil).WithStatus(srvErr.Status)

		return apiError
	}

	return NewInternalServerError(err)
}

// NewNotFoundError creates and returns 404 `ApiError`.
func NewNotFoundError(message string, data any) *ApiError {
	if message == "" {
		message = "The requested resource wasn't found."
	}

	return NewApiError(http.StatusNotFound, message, data)
}

// NewBadRequestError creates and returns 400 `ApiError`.
func NewBadRequestError(message string, data any) *ApiError {
	if message == "" {
		message = "Something went wrong while processing your request."
	}

	return NewApiError(http.StatusBadRequest, message, data).WithStatus(exceptions.Error)
}

// NewForbiddenError creates and returns 403 `ApiError`.
func NewForbiddenError(message string, data any) *ApiError {
	if message == "" {
		message = "You are not allowed to perform this request."
	}

	return NewApiError(http.StatusForbidden, message, data)
}

// NewUnauthorizedError creates and returns 401 `ApiError`.
func NewUnauthorizedError(message string, data any) *ApiError {
	if message == "" {
		message = "Missing or invalid authentication token."
	}

	return NewApiError(http.StatusUnauthorized, message, data).WithStatus(exceptions.Error)
}

// NewApiError creates and returns new normalized `ApiError` instance.
func NewApiError(status int, message string, data any) *ApiError {
	return &ApiError{
		rawData: data,
		Code:    status,
		Time:    time.Now(),
		Message: strings.TrimSpace(message),
	}
}

func (b *ApiError) WithStatus(status exceptions.Status) *ApiError {
	b.Status = status.String()

	return b
}

func NewInternalServerError(err error) *ApiError {
	return NewApiError(http.StatusInternalServerError, err.Error(), nil).WithStatus(exceptions.Error)
}

// handle for ozzo validation
func NewValidationErrorV2(err error) *ApiError {
	b, errParse := json.Marshal(err)
	if errParse != nil {
		return NewApiError(http.StatusInternalServerError, errParse.Error(), nil).WithStatus(exceptions.Error)
	}

	return NewApiError(http.StatusUnprocessableEntity, err.Error(), b).WithStatus(exceptions.Error)
}

// func NewValidationError(err error) *ApiError {
// 	message := ""

// 	if castedObject, ok := err.(validator.ValidationErrors); ok {
// 		for _, err := range castedObject {
// 			switch err.Tag() {
// 			case "required":
// 				message = fmt.Sprintf("%s is required",
// 					err.Field())
// 			case "email":
// 				message = fmt.Sprintf("%s is not valid email",
// 					err.Field())
// 			case "gte":
// 				message = fmt.Sprintf("%s value must be greater than %s",
// 					err.Field(), err.Param())
// 			case "lte":
// 				message = fmt.Sprintf("%s value must be lower than %s",
// 					err.Field(), err.Param())
// 			case "min":
// 				message = fmt.Sprintf("%s value at least %s data", err.Field(), err.Param())
// 			}
// 		}
// 	}

// 	return NewApiError(http.StatusUnprocessableEntity, message, err)
// }
