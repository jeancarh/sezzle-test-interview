package calculator

import (
	"errors"
	"strings"
	"testing"
)

func TestParseNumberAccepts(t *testing.T) {
	tests := map[string]string{
		"0":                                  "0",
		"42":                                 "42",
		"-7":                                 "-7",
		"0.5":                                "0.5",
		".5":                                 "0.5",
		"-0":                                 "0",
		"1.50":                               "1.5",
		"00012":                              "12",
		"1e3":                                "1000",
		"1.5E+40":                            "1.5E+40",
		"2.5e-3":                             "0.0025",
		"1e-20":                              "1E-20",
		"123456789012345678901234567890.5":   "123456789012345678901234567890.5",
		"1234567890123456789012345678901234": "1234567890123456789012345678901234",
		"0.0000000000000000000000000000000000001":  "1E-37",
		"1234567890123456789012345678901234000000": "1.234567890123456789012345678901234E+39",
	}
	for in, want := range tests {
		n, err := ParseNumber(in)
		if err != nil {
			t.Fatalf("ParseNumber(%q): %v", in, err)
		}
		if got := n.String(); got != want {
			t.Fatalf("ParseNumber(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestParseNumberRejects(t *testing.T) {
	inputs := []string{
		"", " ", " 1", "1 ", "+3", "5.", "1.2.3", "--1", "abc", "NaN", "nan", "Infinity", "-Infinity",
		"0x10", "1_000", "1,5", "١٢", "１２", "1e", "1e+", "1e99999", "1e-99999",
		strings.Repeat("9", maxOperandLength+1),
		"12345678901234567890123456789012345",
		"1.2345678901234567890123456789012345",
	}
	for _, in := range inputs {
		_, err := ParseNumber(in)
		var calcErr *Error
		if !errors.As(err, &calcErr) || calcErr.Code != CodeInvalidNumber {
			t.Fatalf("ParseNumber(%q): want invalid_number, got %v", in, err)
		}
	}
}

func TestTooManySignificantDigitsSaysWhy(t *testing.T) {
	_, err := ParseNumber("12345678901234567890123456789012345")
	if err == nil || !strings.Contains(err.Error(), "34 significant digits") {
		t.Fatalf("got %v", err)
	}
}
