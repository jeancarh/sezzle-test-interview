package calculator

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		operation string
		operands  []string
		want      string
	}{
		{"add", []string{"2", "3"}, "5"},
		{"add", []string{"0.1", "0.2"}, "0.3"},
		{"add", []string{"-1.5", "1.5"}, "0"},
		{"add", []string{"1e30", "1"}, "1000000000000000000000000000001"},
		{"subtract", []string{"10", "4.25"}, "5.75"},
		{"subtract", []string{"0.3", "0.1"}, "0.2"},
		{"multiply", []string{"1.5", "4"}, "6"},
		{"multiply", []string{"-0", "5"}, "0"},
		{"multiply", []string{"1.1", "1.1"}, "1.21"},
		{"divide", []string{"10", "4"}, "2.5"},
		{"divide", []string{"1", "3"}, "0.3333333333333333333333333333333333"},
		{"divide", []string{"-9", "3"}, "-3"},
		{"divide", []string{"0", "5"}, "0"},
		{"power", []string{"2", "10"}, "1024"},
		{"power", []string{"-2", "3"}, "-8"},
		{"power", []string{"2", "-2"}, "0.25"},
		{"power", []string{"4", "0.5"}, "2"},
		{"power", []string{"0", "0"}, "1"},
		{"power", []string{"0", "5"}, "0"},
		{"power", []string{"10", "40"}, "1E+40"},
		{"sqrt", []string{"16"}, "4"},
		{"sqrt", []string{"2"}, "1.414213562373095048801688724209698"},
		{"sqrt", []string{"0"}, "0"},
		{"sqrt", []string{"0.25"}, "0.5"},
		{"percentage", []string{"15", "200"}, "30"},
		{"percentage", []string{"12.5", "80"}, "10"},
		{"percentage", []string{"200", "50"}, "100"},
	}
	svc := NewService()
	for _, tt := range tests {
		t.Run(tt.operation+" "+strings.Join(tt.operands, " "), func(t *testing.T) {
			res, err := svc.Calculate(context.Background(), tt.operation, tt.operands)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Value != tt.want {
				t.Fatalf("got %s, want %s", res.Value, tt.want)
			}
			if res.Operation != tt.operation {
				t.Fatalf("operation echoed as %q", res.Operation)
			}
		})
	}
}

func TestCalculateErrors(t *testing.T) {
	tests := []struct {
		operation string
		operands  []string
		code      Code
	}{
		{"divide", []string{"1", "0"}, CodeDivisionByZero},
		{"divide", []string{"0", "0"}, CodeDivisionByZero},
		{"divide", []string{"5", "-0"}, CodeDivisionByZero},
		{"sqrt", []string{"-4"}, CodeNegativeSquareRoot},
		{"power", []string{"0", "-1"}, CodeZeroToNegativePower},
		{"power", []string{"-8", "0.5"}, CodeComplexResult},
		{"power", []string{"10", "999999999"}, CodeResultTooLarge},
		{"power", []string{"10", "1e4"}, CodeResultTooLarge},
		{"power", []string{"10", "-99999"}, CodeResultTooSmall},
		{"multiply", []string{"1e9000", "1e9000"}, CodeResultTooLarge},
		{"divide", []string{"1e-9000", "1e9000"}, CodeResultTooSmall},
		{"add", []string{"1", "abc"}, CodeInvalidNumber},
		{"add", []string{"NaN", "1"}, CodeInvalidNumber},
		{"add", []string{"1"}, CodeWrongOperandCount},
		{"add", []string{"1", "2", "3"}, CodeWrongOperandCount},
		{"sqrt", []string{"4", "2"}, CodeWrongOperandCount},
		{"add", nil, CodeWrongOperandCount},
		{"modulo", []string{"5", "2"}, CodeUnknownOperation},
		{"", []string{"5", "2"}, CodeUnknownOperation},
		{"ADD", []string{"5", "2"}, CodeUnknownOperation},
	}
	svc := NewService()
	for _, tt := range tests {
		t.Run(string(tt.code)+" "+tt.operation+" "+strings.Join(tt.operands, " "), func(t *testing.T) {
			_, err := svc.Calculate(context.Background(), tt.operation, tt.operands)
			var calcErr *Error
			if !errors.As(err, &calcErr) {
				t.Fatalf("want *Error, got %v", err)
			}
			if calcErr.Code != tt.code {
				t.Fatalf("got code %s (%s), want %s", calcErr.Code, calcErr.Message, tt.code)
			}
			if calcErr.Message == "" {
				t.Fatal("error message is empty")
			}
		})
	}
}

func TestInvalidOperandMessageSaysWhich(t *testing.T) {
	_, err := NewService().Calculate(context.Background(), "add", []string{"1", "x"})
	if err == nil || !strings.Contains(err.Error(), "second") {
		t.Fatalf("got %v", err)
	}
}

func TestCalculateEchoesNormalizedOperands(t *testing.T) {
	res, err := NewService().Calculate(context.Background(), "add", []string{"1.50", "-0"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(res.Operands, ",") != "1.5,0" {
		t.Fatalf("got %v", res.Operands)
	}
}

func TestOperationsAreListedInOrder(t *testing.T) {
	ops := NewService().Operations()
	var names []string
	for _, op := range ops {
		names = append(names, op.Name)
		if op.Symbol == "" || op.Arity < 1 || op.Arity > 2 {
			t.Fatalf("bad operation info %+v", op)
		}
	}
	want := "add,subtract,multiply,divide,power,sqrt,percentage"
	if strings.Join(names, ",") != want {
		t.Fatalf("got %s", strings.Join(names, ","))
	}
}

func TestServiceIsSafeForConcurrentUse(t *testing.T) {
	svc := NewService()
	done := make(chan error)
	for i := 0; i < 50; i++ {
		go func() {
			res, err := svc.Calculate(context.Background(), "divide", []string{"1", "3"})
			if err == nil && !strings.HasPrefix(res.Value, "0.333") {
				err = errors.New(res.Value)
			}
			done <- err
		}()
	}
	for i := 0; i < 50; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
}
