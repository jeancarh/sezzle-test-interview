package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jeancarh/sezzle-test-interview/backend/internal/calculator"
)

func newServer(t *testing.T, calc calculator.Calculator) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(NewRouter(Deps{
		Calculator: calc,
		Assets: fstest.MapFS{
			"index.html":         {Data: []byte("<html>calculator</html>")},
			"assets/app-abc1.js": {Data: []byte("console.log(1)")},
		},
		Spec: []byte("openapi: 3.1.0"),
	}))
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, srv *httptest.Server, body string) (*http.Response, map[string]any) {
	t.Helper()
	res, err := http.Post(srv.URL+"/api/v1/calculations", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out := map[string]any{}
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res, out
}

func TestCalculationsHappyPath(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	tests := []struct {
		body   string
		result string
	}{
		{`{"operation":"add","operands":["0.1","0.2"]}`, "0.3"},
		{`{"operation":"subtract","operands":["10","4.25"]}`, "5.75"},
		{`{"operation":"multiply","operands":[1.5, 4]}`, "6"},
		{`{"operation":"divide","operands":["1","3"]}`, "0.3333333333333333333333333333333333"},
		{`{"operation":"power","operands":["2","10"]}`, "1024"},
		{`{"operation":"sqrt","operands":["2"]}`, "1.414213562373095048801688724209698"},
		{`{"operation":"percentage","operands":["15","200"]}`, "30"},
	}
	for _, tt := range tests {
		res, body := post(t, srv, tt.body)
		if res.StatusCode != http.StatusOK || body["result"] != tt.result {
			t.Fatalf("%s: got %d %v", tt.body, res.StatusCode, body)
		}
		if res.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content type %q", res.Header.Get("Content-Type"))
		}
	}
}

func TestCalculationsDomainErrors(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	tests := map[string]string{
		`{"operation":"divide","operands":["1","0"]}`:      "division_by_zero",
		`{"operation":"sqrt","operands":["-4"]}`:           "negative_square_root",
		`{"operation":"power","operands":["0","-1"]}`:      "zero_to_negative_power",
		`{"operation":"power","operands":["-8","0.5"]}`:    "complex_result",
		`{"operation":"power","operands":["10","1e4"]}`:    "result_too_large",
		`{"operation":"power","operands":["10","-99999"]}`: "result_too_small",
		`{"operation":"add","operands":["1","NaN"]}`:       "invalid_number",
		`{"operation":"add","operands":["1"]}`:             "wrong_operand_count",
		`{"operation":"add","operands":[]}`:                "wrong_operand_count",
		`{"operation":"add"}`:                              "wrong_operand_count",
		`{"operation":"modulo","operands":["5","2"]}`:      "unknown_operation",
		`{"operands":["5","2"]}`:                           "unknown_operation",
	}
	for body, code := range tests {
		res, problem := post(t, srv, body)
		if res.StatusCode != http.StatusUnprocessableEntity || problem["code"] != code {
			t.Fatalf("%s: got %d %v", body, res.StatusCode, problem)
		}
		if res.Header.Get("Content-Type") != "application/problem+json" || problem["detail"] == "" || problem["status"] != float64(422) {
			t.Fatalf("%s: not a proper problem: %v", body, problem)
		}
	}
}

func TestCalculationsMalformedRequests(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	tests := map[string]int{
		`not json`: http.StatusBadRequest,
		`{"operation":"add","operands":["1","2"],"extra":1}`:          http.StatusBadRequest,
		`{"operation":"add","operands":[true,"2"]}`:                   http.StatusBadRequest,
		`{"operation":"add","operands":"12"}`:                         http.StatusBadRequest,
		`{"operation":"add","operands":["1","2"]}{"operation":"add"}`: http.StatusBadRequest,
		`[]`: http.StatusBadRequest,
		`{"operation":"add","operands":["` + strings.Repeat("9", 5000) + `","1"]}`: http.StatusRequestEntityTooLarge,
	}
	for body, status := range tests {
		res, problem := post(t, srv, body)
		if res.StatusCode != status {
			t.Fatalf("%.60s: got %d %v", body, res.StatusCode, problem)
		}
		want := codeMalformedRequest
		if status == http.StatusRequestEntityTooLarge {
			want = codeRequestTooLarge
		}
		if problem["code"] != want {
			t.Fatalf("%.60s: code %v", body, problem["code"])
		}
	}
}

func TestOperationsEndpoint(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	res, err := http.Get(srv.URL + "/api/v1/operations")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var list struct {
		Items []struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
			Arity  int    `json:"arity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 7 || list.Items[5].Name != "sqrt" || list.Items[5].Arity != 1 {
		t.Fatalf("got %+v", list.Items)
	}
}

func TestRouting(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	tests := []struct {
		method, path string
		status       int
		contains     string
	}{
		{"GET", "/healthz", 200, `"status":"ok"`},
		{"GET", "/api/openapi.yaml", 200, "openapi: 3.1.0"},
		{"GET", "/docs", 200, "swagger-ui"},
		{"GET", "/", 200, "calculator"},
		{"GET", "/some/client/route", 200, "calculator"},
		{"GET", "/assets/app-abc1.js", 200, "console.log"},
		{"GET", "/api/v1/nope", 404, "not_found"},
		{"GET", "/api/v1/calculations", 405, "method_not_allowed"},
		{"DELETE", "/api/v1/operations", 405, "method_not_allowed"},
		{"POST", "/", 405, "method_not_allowed"},
	}
	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, srv.URL+tt.path, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != tt.status || !strings.Contains(string(body), tt.contains) {
			t.Fatalf("%s %s: got %d %s", tt.method, tt.path, res.StatusCode, body)
		}
	}
}

func TestHashedAssetsAreCachedForever(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	res, err := http.Get(srv.URL + "/assets/app-abc1.js")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if !strings.Contains(res.Header.Get("Cache-Control"), "immutable") {
		t.Fatalf("cache-control %q", res.Header.Get("Cache-Control"))
	}
	index, _ := http.Get(srv.URL + "/")
	index.Body.Close()
	if index.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("index cache-control %q", index.Header.Get("Cache-Control"))
	}
}

func TestRequestIDIsEchoedOrGenerated(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	req, _ := http.NewRequest("GET", srv.URL+"/healthz", nil)
	req.Header.Set("X-Request-ID", "trace-123")
	res, _ := http.DefaultClient.Do(req)
	res.Body.Close()
	if res.Header.Get("X-Request-ID") != "trace-123" {
		t.Fatalf("got %q", res.Header.Get("X-Request-ID"))
	}
	res, _ = http.Get(srv.URL + "/healthz")
	res.Body.Close()
	if len(res.Header.Get("X-Request-ID")) != 16 {
		t.Fatalf("generated id %q", res.Header.Get("X-Request-ID"))
	}
}

type brokenCalculator struct {
	panics bool
}

func (b brokenCalculator) Calculate(context.Context, string, []string) (calculator.Result, error) {
	if b.panics {
		panic("unexpected")
	}
	return calculator.Result{}, errors.New("disk on fire")
}

func (brokenCalculator) Operations() []calculator.OperationInfo { return nil }

func TestUnexpectedErrorsBecome500WithoutDetails(t *testing.T) {
	for _, calc := range []brokenCalculator{{panics: false}, {panics: true}} {
		srv := newServer(t, calc)
		res, problem := post(t, srv, `{"operation":"add","operands":["1","2"]}`)
		if res.StatusCode != http.StatusInternalServerError || problem["code"] != codeInternalError {
			t.Fatalf("panics=%v: got %d %v", calc.panics, res.StatusCode, problem)
		}
		if strings.Contains(problem["detail"].(string), "fire") {
			t.Fatal("internal detail leaked")
		}
	}
}

func TestCalculationsRequireJSONContentType(t *testing.T) {
	srv := newServer(t, calculator.NewService())
	body := `{"operation":"add","operands":["1","2"]}`
	tests := map[string]int{
		"application/json":                  http.StatusOK,
		"application/json; charset=utf-8":   http.StatusOK,
		"Application/JSON":                  http.StatusOK,
		"text/plain":                        http.StatusUnsupportedMediaType,
		"application/x-www-form-urlencoded": http.StatusUnsupportedMediaType,
		"":                                  http.StatusUnsupportedMediaType,
	}
	for contentType, status := range tests {
		req, _ := http.NewRequest("POST", srv.URL+"/api/v1/calculations", strings.NewReader(body))
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var problem map[string]any
		_ = json.NewDecoder(res.Body).Decode(&problem)
		res.Body.Close()
		if res.StatusCode != status {
			t.Fatalf("%q: got %d", contentType, res.StatusCode)
		}
		if status == http.StatusUnsupportedMediaType && problem["code"] != codeUnsupportedMediaType {
			t.Fatalf("%q: code %v", contentType, problem["code"])
		}
	}
}
