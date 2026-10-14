package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/jeancarh/sezzle-test-interview/backend/internal/calculator"
)

var errOperandType = errors.New("operands must be strings or numbers")

type operand string

func (o *operand) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*o = operand(s)
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return err
	}
	n, ok := v.(json.Number)
	if !ok {
		return errOperandType
	}
	*o = operand(n.String())
	return nil
}

type calculationRequest struct {
	Operation string    `json:"operation"`
	Operands  []operand `json:"operands"`
}

func (r calculationRequest) operandStrings() []string {
	out := make([]string, len(r.Operands))
	for i, o := range r.Operands {
		out[i] = string(o)
	}
	return out
}

type calculationResponse struct {
	Operation string   `json:"operation"`
	Operands  []string `json:"operands"`
	Result    string   `json:"result"`
}

func toResponse(r calculator.Result) calculationResponse {
	return calculationResponse{Operation: r.Operation, Operands: r.Operands, Result: r.Value}
}

type operationInfo struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Arity  int    `json:"arity"`
}

type operationList struct {
	Items []operationInfo `json:"items"`
}

func toOperationList(infos []calculator.OperationInfo) operationList {
	items := make([]operationInfo, 0, len(infos))
	for _, info := range infos {
		items = append(items, operationInfo{Name: info.Name, Symbol: info.Symbol, Arity: info.Arity})
	}
	return operationList{Items: items}
}
