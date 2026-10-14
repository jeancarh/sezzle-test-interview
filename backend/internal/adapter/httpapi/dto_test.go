package httpapi

import (
	"encoding/json"
	"testing"

	"github.com/jeancarh/sezzle-test-interview/backend/internal/calculator"
)

func TestOperandAcceptsStringsAndNumbers(t *testing.T) {
	var req calculationRequest
	body := `{"operation":"add","operands":["0.1", 0.2, 12345678901234567890.123456789, -3e5]}`
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	got := req.operandStrings()
	want := []string{"0.1", "0.2", "12345678901234567890.123456789", "-3e5"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("operand %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestOperandRejectsOtherTypes(t *testing.T) {
	for _, value := range []string{`true`, `null`, `{}`, `[1]`} {
		var req calculationRequest
		body := `{"operation":"add","operands":[` + value + `, "1"]}`
		if err := json.Unmarshal([]byte(body), &req); err == nil {
			t.Fatalf("%s: expected an error", value)
		}
	}
}

func TestToResponse(t *testing.T) {
	res := toResponse(calculator.Result{Operation: "add", Operands: []string{"1", "2"}, Value: "3"})
	b, _ := json.Marshal(res)
	if string(b) != `{"operation":"add","operands":["1","2"],"result":"3"}` {
		t.Fatalf("got %s", b)
	}
}

func TestToOperationList(t *testing.T) {
	list := toOperationList([]calculator.OperationInfo{{Name: "sqrt", Symbol: "√", Arity: 1}})
	b, _ := json.Marshal(list)
	if string(b) != `{"items":[{"name":"sqrt","symbol":"√","arity":1}]}` {
		t.Fatalf("got %s", b)
	}
}
