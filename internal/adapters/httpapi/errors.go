package httpapi

import (
	"errors"
	"net/http"

	"github.com/forest-shadow/hm-cached-products/internal/application"
	"github.com/forest-shadow/hm-cached-products/internal/product"
)

var (
	// Request errors
	errInvalidID        = errors.New("id must be a positive integer")
	errNameRequired     = errors.New("name is required")
	errPriceRequired    = errors.New("price is required")
	errInvalidVersion   = errors.New("expected_version must be a positive integer")
	errMultipleJSONBody = errors.New("expected one JSON object")
)

type errorCode string

const (
	errorCodeInvalidJSON        errorCode = "INVALID_JSON"
	errorCodeBodyTooLarge       errorCode = "REQUEST_BODY_TOO_LARGE"
	errorCodeMultipleJSON       errorCode = "MULTIPLE_JSON_VALUES"
	errorCodeInvalidID          errorCode = "INVALID_ID"
	errorCodeMissingFields      errorCode = "MISSING_REQUIRED_FIELDS"
	errorCodeInvalidVersion     errorCode = "INVALID_EXPECTED_VERSION"
	errorCodeInvalidProduct     errorCode = "INVALID_PRODUCT"
	errorCodeInvalidStatsWindow errorCode = "INVALID_STATS_WINDOW"
	errorCodeProductNotFound    errorCode = "PRODUCT_NOT_FOUND"
	errorCodeVersionConflict    errorCode = "PRODUCT_VERSION_CONFLICT"
	errorCodeInternal           errorCode = "INTERNAL_ERROR"
)

type errorResponse struct {
	Code    errorCode `json:"code"`
	Message string    `json:"message"`
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	switch {
	case errors.As(err, &maxBytesError):
		writeError(w, http.StatusRequestEntityTooLarge, errorCodeBodyTooLarge, "request body is too large")
	case errors.Is(err, errMultipleJSONBody):
		writeError(w, http.StatusBadRequest, errorCodeMultipleJSON, errMultipleJSONBody.Error())
	default:
		writeError(w, http.StatusBadRequest, errorCodeInvalidJSON, "invalid JSON body")
	}
}

func writeRequestError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errInvalidID):
		writeError(w, http.StatusBadRequest, errorCodeInvalidID, errInvalidID.Error())
	case errors.Is(err, errNameRequired):
		writeError(w, http.StatusBadRequest, errorCodeMissingFields, errNameRequired.Error())
	case errors.Is(err, errPriceRequired):
		writeError(w, http.StatusBadRequest, errorCodeMissingFields, errPriceRequired.Error())
	case errors.Is(err, errInvalidVersion):
		writeError(w, http.StatusBadRequest, errorCodeInvalidVersion, errInvalidVersion.Error())
	default:
		writeError(w, http.StatusBadRequest, errorCodeInvalidJSON, "invalid request")
	}
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, product.ErrInvalid):
		writeError(w, http.StatusBadRequest, errorCodeInvalidProduct, "invalid product")
	case errors.Is(err, application.ErrInvalidWindow):
		writeError(w, http.StatusBadRequest, errorCodeInvalidStatsWindow, "invalid statistics window")
	case errors.Is(err, application.ErrNotFound):
		writeError(w, http.StatusNotFound, errorCodeProductNotFound, "product not found")
	case errors.Is(err, application.ErrConflict):
		writeError(w, http.StatusConflict, errorCodeVersionConflict, "product version conflict")
	default:
		h.logger.Error("request failed",
			"method", r.Method,
			"route", r.Pattern,
			"error", err,
		)
		writeError(w, http.StatusInternalServerError, errorCodeInternal, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, code errorCode, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}
