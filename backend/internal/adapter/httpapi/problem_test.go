package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jeancarh/sezzle-test-interview/backend/internal/calculator"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"domain error", &calculator.Error{Code: calculator.CodeDivisionByZero, Message: "can't divide by zero"}, http.StatusUnprocessableEntity, "division_by_zero"},
		{"wrapped domain error", fmt.Errorf("wrap: %w", &calculator.Error{Code: calculator.CodeInvalidNumber, Message: "x"}), http.StatusUnprocessableEntity, "invalid_number"},
		{"body too large", &http.MaxBytesError{Limit: 4096}, http.StatusRequestEntityTooLarge, codeRequestTooLarge},
		{"malformed", fmt.Errorf("%w: bad json", errMalformed), http.StatusBadRequest, codeMalformedRequest},
		{"anything else", errors.New("boom"), http.StatusInternalServerError, codeInternalError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, detail := classify(tt.err)
			if status != tt.status || code != tt.code || detail == "" {
				t.Fatalf("got %d %s %q", status, code, detail)
			}
		})
	}
}

func TestInternalErrorsDontLeakDetails(t *testing.T) {
	_, _, detail := classify(errors.New("runtime error: index out of range [3] with length 2"))
	if detail != "something went wrong on our side" {
		t.Fatalf("leaked %q", detail)
	}
}
