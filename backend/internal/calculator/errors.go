package calculator

type Code string

const (
	CodeInvalidNumber       Code = "invalid_number"
	CodeWrongOperandCount   Code = "wrong_operand_count"
	CodeUnknownOperation    Code = "unknown_operation"
	CodeDivisionByZero      Code = "division_by_zero"
	CodeNegativeSquareRoot  Code = "negative_square_root"
	CodeZeroToNegativePower Code = "zero_to_negative_power"
	CodeComplexResult       Code = "complex_result"
	CodeResultTooLarge      Code = "result_too_large"
	CodeResultTooSmall      Code = "result_too_small"
)

type Error struct {
	Code    Code
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func fail(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}
