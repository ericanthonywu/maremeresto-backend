package apperror

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound               = errors.New("resource not found")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrForbidden              = errors.New("forbidden")
	ErrBadRequest             = errors.New("bad request")
	ErrConflict               = errors.New("resource conflict")
	ErrInvalidPhone           = errors.New("invalid phone number format, must be +628... with 10-15 digits")
	ErrOrderAlreadyPaid       = errors.New("order has already been paid")
	ErrOrderNotPayable        = errors.New("order cannot be paid in current status")
	ErrConcurrentModification = errors.New("concurrent modification detected, please refresh and retry")
	ErrInvalidStatusTransition= errors.New("invalid order status transition")
	ErrInvalidPromo           = errors.New("invalid or expired promo code")
	ErrStoreClosed            = errors.New("branch is currently closed")
	ErrMidtransFailed         = errors.New("payment gateway error")
)

type AppError struct {
	Err     error  `json:"-"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func New(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func FromError(err error) *AppError {
	switch {
	case errors.Is(err, ErrNotFound):
		return New(http.StatusNotFound, err.Error(), err)
	case errors.Is(err, ErrUnauthorized):
		return New(http.StatusUnauthorized, err.Error(), err)
	case errors.Is(err, ErrForbidden):
		return New(http.StatusForbidden, err.Error(), err)
	case errors.Is(err, ErrConflict), errors.Is(err, ErrOrderAlreadyPaid), errors.Is(err, ErrConcurrentModification):
		return New(http.StatusConflict, err.Error(), err)
	case errors.Is(err, ErrBadRequest), errors.Is(err, ErrInvalidPhone), errors.Is(err, ErrOrderNotPayable), errors.Is(err, ErrInvalidStatusTransition), errors.Is(err, ErrInvalidPromo), errors.Is(err, ErrStoreClosed):
		return New(http.StatusBadRequest, err.Error(), err)
	default:
		return New(http.StatusInternalServerError, "internal server error", err)
	}
}
