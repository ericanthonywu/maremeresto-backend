package apperror

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound                = errors.New("resource not found")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrBadRequest              = errors.New("bad request")
	ErrConflict                = errors.New("resource conflict")
	ErrInvalidPhone            = errors.New("format nomor telepon tidak valid, gunakan 08... atau +628...")
	ErrOrderAlreadyPaid        = errors.New("pesanan ini sudah dibayar")
	ErrOrderNotPayable         = errors.New("pesanan tidak dapat dibayar pada status saat ini")
	ErrConcurrentModification  = errors.New("data berubah di perangkat lain, silakan muat ulang dan coba lagi")
	ErrInvalidStatusTransition = errors.New("perubahan status pesanan tidak diizinkan")
	ErrInvalidPromo            = errors.New("kode promo tidak valid atau sudah kedaluwarsa")
	ErrStoreClosed             = errors.New("outlet sedang tutup, silakan pilih outlet lain atau coba lagi nanti")
	ErrMidtransFailed          = errors.New("gerbang pembayaran sedang tidak dapat dihubungi, silakan coba lagi")
	ErrLocationRequired        = errors.New("lokasi pengantaran belum ditentukan")
	ErrOutOfDeliveryRange      = errors.New("alamat berada di luar jangkauan pengantaran outlet ini")
	ErrBelowMinimumOrder       = errors.New("total pesanan belum memenuhi minimum order outlet ini")
	ErrItemUnavailable         = errors.New("salah satu item pesanan sedang tidak tersedia")
	ErrGeocoderUnavailable     = errors.New("layanan pencarian alamat sedang tidak tersedia")
	ErrTooManyRequests         = errors.New("terlalu banyak permintaan, silakan tunggu sebentar")
)

type AppError struct {
	Err     error  `json:"-"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return http.StatusText(e.Code)
}

// Unwrap lets errors.Is match the sentinel an AppError was built from, so
// callers can branch on the category while the client still sees the specific
// human-readable message.
func (e *AppError) Unwrap() error { return e.Err }

func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Invalid builds a 400 carrying a message written for the end user. Prefer it
// over wrapping ErrBadRequest with fmt.Errorf, which leaks the sentinel's own
// text ("bad request: ...") into the API response.
func Invalid(message string) *AppError {
	return New(http.StatusBadRequest, message, ErrBadRequest)
}

// Conflict builds a 409 with a user-facing message.
func Conflict(message string) *AppError {
	return New(http.StatusConflict, message, ErrConflict)
}

func FromError(err error) *AppError {
	// An AppError already carries the status and the message to show.
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return New(http.StatusNotFound, err.Error(), err)
	case errors.Is(err, ErrUnauthorized):
		return New(http.StatusUnauthorized, err.Error(), err)
	case errors.Is(err, ErrForbidden):
		return New(http.StatusForbidden, err.Error(), err)
	case errors.Is(err, ErrConflict), errors.Is(err, ErrOrderAlreadyPaid), errors.Is(err, ErrConcurrentModification):
		return New(http.StatusConflict, err.Error(), err)
	case errors.Is(err, ErrTooManyRequests):
		return New(http.StatusTooManyRequests, err.Error(), err)
	case errors.Is(err, ErrGeocoderUnavailable), errors.Is(err, ErrMidtransFailed):
		return New(http.StatusServiceUnavailable, err.Error(), err)
	case errors.Is(err, ErrBadRequest), errors.Is(err, ErrInvalidPhone), errors.Is(err, ErrOrderNotPayable),
		errors.Is(err, ErrInvalidStatusTransition), errors.Is(err, ErrInvalidPromo), errors.Is(err, ErrStoreClosed),
		errors.Is(err, ErrLocationRequired), errors.Is(err, ErrOutOfDeliveryRange), errors.Is(err, ErrBelowMinimumOrder),
		errors.Is(err, ErrItemUnavailable):
		return New(http.StatusBadRequest, err.Error(), err)
	default:
		// Never surface an unmapped internal error verbatim: it can leak SQL,
		// hostnames or credentials into the client response.
		return New(http.StatusInternalServerError, "terjadi kesalahan pada server, silakan coba lagi", err)
	}
}
