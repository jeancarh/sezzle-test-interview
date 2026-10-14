package calculator

import (
	"context"
	"errors"
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

type Calculator interface {
	Calculate(ctx context.Context, operation string, operands []string) (Result, error)
	Operations() []OperationInfo
}

type Result struct {
	Operation string
	Operands  []string
	Value     string
}

type OperationInfo struct {
	Name   string
	Symbol string
	Arity  int
}

type Service struct {
	byName map[string]Operation
	infos  []OperationInfo
}

func NewService() *Service {
	s := &Service{byName: make(map[string]Operation, len(registry))}
	for _, op := range registry {
		s.byName[op.Name] = op
		s.infos = append(s.infos, OperationInfo{Name: op.Name, Symbol: op.Symbol, Arity: op.Arity})
	}
	return s
}

func (s *Service) Operations() []OperationInfo {
	return append([]OperationInfo(nil), s.infos...)
}

func (s *Service) Calculate(_ context.Context, operation string, operands []string) (Result, error) {
	op, ok := s.byName[operation]
	if !ok {
		return Result{}, fail(CodeUnknownOperation, fmt.Sprintf("unknown operation %q", operation))
	}
	if len(operands) != op.Arity {
		return Result{}, fail(CodeWrongOperandCount, fmt.Sprintf("%s takes %d operand(s), got %d", op.Name, op.Arity, len(operands)))
	}

	args := make([]Number, len(operands))
	echoed := make([]string, len(operands))
	for i, raw := range operands {
		n, err := ParseNumber(raw)
		if err != nil {
			var calcErr *Error
			if errors.As(err, &calcErr) {
				calcErr.Message = fmt.Sprintf("%s operand: %s", ordinal(i), calcErr.Message)
			}
			return Result{}, err
		}
		args[i] = n
		echoed[i] = n.String()
	}

	value, err := evaluate(op, args)
	if err != nil {
		return Result{}, err
	}
	return Result{Operation: op.Name, Operands: echoed, Value: value.String()}, nil
}

func evaluate(op Operation, args []Number) (Number, error) {
	c := apd.BaseContext.WithPrecision(precision)
	c.MaxExponent = maxExponent
	c.MinExponent = minExponent
	c.Traps = 0

	var z apd.Decimal
	cond, err := op.apply(c, &z, args)
	var calcErr *Error
	switch {
	case errors.As(err, &calcErr):
		return Number{}, err
	case cond.Overflow():
		return Number{}, tooLarge()
	case cond.Underflow() || cond.Subnormal():
		return Number{}, tooSmall()
	case err != nil:
		return Number{}, tooLarge()
	case z.Form != apd.Finite:
		return Number{}, tooLarge()
	}

	result := normalize(&z)
	if !result.inRange() {
		if result.adjustedExponent() > 0 {
			return Number{}, tooLarge()
		}
		return Number{}, tooSmall()
	}
	return result, nil
}

func tooLarge() error {
	return fail(CodeResultTooLarge, "the result is too large (over 1E+9999)")
}

func tooSmall() error {
	return fail(CodeResultTooSmall, "the result is too small to represent (under 1E-9999)")
}

func ordinal(i int) string {
	switch i {
	case 0:
		return "first"
	case 1:
		return "second"
	default:
		return fmt.Sprintf("#%d", i+1)
	}
}
