package calculator

import "github.com/cockroachdb/apd/v3"

type Operation struct {
	Name   string
	Symbol string
	Arity  int
	apply  func(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error)
}

var hundred = apd.New(100, 0)

var registry = []Operation{
	{Name: "add", Symbol: "+", Arity: 2, apply: add},
	{Name: "subtract", Symbol: "−", Arity: 2, apply: subtract},
	{Name: "multiply", Symbol: "×", Arity: 2, apply: multiply},
	{Name: "divide", Symbol: "÷", Arity: 2, apply: divide},
	{Name: "power", Symbol: "^", Arity: 2, apply: power},
	{Name: "sqrt", Symbol: "√", Arity: 1, apply: squareRoot},
	{Name: "percentage", Symbol: "%", Arity: 2, apply: percentage},
}

func add(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error) {
	return c.Add(z, args[0].decimal(), args[1].decimal())
}

func subtract(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error) {
	return c.Sub(z, args[0].decimal(), args[1].decimal())
}

func multiply(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error) {
	return c.Mul(z, args[0].decimal(), args[1].decimal())
}

func divide(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error) {
	if args[1].decimal().IsZero() {
		return 0, fail(CodeDivisionByZero, "can't divide by zero")
	}
	return c.Quo(z, args[0].decimal(), args[1].decimal())
}

func power(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error) {
	base, exponent := args[0], args[1]
	switch {
	case base.decimal().IsZero() && exponent.decimal().IsZero():
		z.SetInt64(1)
		return 0, nil
	case base.decimal().IsZero() && exponent.decimal().Negative:
		return 0, fail(CodeZeroToNegativePower, "zero can't be raised to a negative power")
	case base.decimal().Negative && !exponent.isInteger():
		return 0, fail(CodeComplexResult, "a negative number raised to a fractional power isn't a real number")
	}
	return c.Pow(z, base.decimal(), exponent.decimal())
}

func squareRoot(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error) {
	if args[0].decimal().Negative {
		return 0, fail(CodeNegativeSquareRoot, "can't take the square root of a negative number")
	}
	return c.Sqrt(z, args[0].decimal())
}

func percentage(c *apd.Context, z *apd.Decimal, args []Number) (apd.Condition, error) {
	var product apd.Decimal
	cond, err := c.Mul(&product, args[0].decimal(), args[1].decimal())
	if err != nil {
		return cond, err
	}
	quoCond, err := c.Quo(z, &product, hundred)
	return cond | quoCond, err
}
