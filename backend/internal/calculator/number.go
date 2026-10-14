package calculator

import (
	"regexp"

	"github.com/cockroachdb/apd/v3"
)

const (
	precision         = 34
	maxExponent       = 9999
	minExponent       = -9999
	maxOperandLength  = 100
	plainNotationLow  = -7
	plainNotationHigh = precision
)

var numberPattern = regexp.MustCompile(`^-?(\d+(\.\d+)?|\.\d+)([eE][+-]?\d{1,4})?$`)

type Number struct {
	value apd.Decimal
}

func ParseNumber(s string) (Number, error) {
	if len(s) > maxOperandLength {
		return Number{}, fail(CodeInvalidNumber, "numbers can have at most 100 characters")
	}
	if !numberPattern.MatchString(s) {
		return Number{}, fail(CodeInvalidNumber, "not a valid number")
	}
	d, _, err := apd.NewFromString(s)
	if err != nil {
		return Number{}, fail(CodeInvalidNumber, "not a valid number")
	}
	n := normalize(d)
	if n.value.NumDigits() > precision {
		return Number{}, fail(CodeInvalidNumber, "numbers can have at most 34 significant digits")
	}
	if !n.inRange() {
		return Number{}, fail(CodeInvalidNumber, "number is out of the supported range (1E-9999 to 1E+9999)")
	}
	return n, nil
}

func (n Number) String() string {
	adjusted := n.adjustedExponent()
	if n.value.IsZero() || (adjusted >= plainNotationLow && adjusted < plainNotationHigh) {
		return n.value.Text('f')
	}
	return n.value.String()
}

func (n Number) decimal() *apd.Decimal {
	return &n.value
}

func (n Number) adjustedExponent() int64 {
	return int64(n.value.Exponent) + n.value.NumDigits() - 1
}

func (n Number) inRange() bool {
	if n.value.IsZero() {
		return true
	}
	adjusted := n.adjustedExponent()
	return adjusted >= minExponent && adjusted <= maxExponent
}

func (n Number) isInteger() bool {
	return n.value.Exponent >= 0
}

func normalize(d *apd.Decimal) Number {
	var n Number
	n.value.Reduce(d)
	if n.value.IsZero() {
		n.value.Negative = false
		n.value.Exponent = 0
	}
	return n
}
