package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/jeancarh/sezzle-test-interview/backend/internal/calculator"
)

const maxBodyBytes = 4 << 10

type api struct {
	calc calculator.Calculator
}

func (a *api) calculate(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		writeProblem(w, http.StatusUnsupportedMediaType, codeUnsupportedMediaType, "send the body as application/json")
		return
	}
	var req calculationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	res, err := a.calc.Calculate(r.Context(), req.Operation, req.operandStrings())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(res))
}

func (a *api) listOperations(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, toOperationList(a.calc.Operations()))
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func isJSON(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return err
		}
		return fmt.Errorf("%w: %v", errMalformed, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: the body must contain a single JSON object", errMalformed)
	}
	return nil
}
