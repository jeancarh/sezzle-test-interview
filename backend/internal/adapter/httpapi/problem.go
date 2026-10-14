package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jeancarh/sezzle-test-interview/backend/internal/calculator"
)

const (
	codeMalformedRequest     = "malformed_request"
	codeUnsupportedMediaType = "unsupported_media_type"
	codeRequestTooLarge      = "request_too_large"
	codeNotFound             = "not_found"
	codeMethodNotAllowed     = "method_not_allowed"
	codeInternalError        = "internal_error"
)

var errMalformed = errors.New("malformed request")

type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Code:   code,
		Detail: detail,
	})
}

func classify(err error) (int, string, string) {
	var (
		calcErr  *calculator.Error
		tooLarge *http.MaxBytesError
	)
	switch {
	case errors.As(err, &calcErr):
		return http.StatusUnprocessableEntity, string(calcErr.Code), calcErr.Message
	case errors.As(err, &tooLarge):
		return http.StatusRequestEntityTooLarge, codeRequestTooLarge, fmt.Sprintf("the request body is larger than %d bytes", tooLarge.Limit)
	case errors.Is(err, errMalformed):
		return http.StatusBadRequest, codeMalformedRequest, err.Error()
	default:
		return http.StatusInternalServerError, codeInternalError, "something went wrong on our side"
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, detail := classify(err)
	if status == http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request failed", "error", err, "request_id", requestID(r.Context()))
	}
	writeProblem(w, status, code, detail)
}
