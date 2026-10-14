import type { components } from "../../../shared/api/schema";
import { ApiError, postJSON } from "../../../shared/http/client";
import { CalculationError, type CalculatorApi } from "../domain/calculatorApi";

type CalculationRequestDto = components["schemas"]["CalculationRequest"];
type CalculationResultDto = components["schemas"]["CalculationResult"];

const friendlyMessages: Record<string, string> = {
  division_by_zero: "Can't divide by zero",
  negative_square_root: "Can't take the square root of a negative number",
  zero_to_negative_power: "Zero can't be raised to a negative power",
  complex_result: "A negative number can't be raised to a fractional power",
  result_too_large: "The result is too large",
  result_too_small: "The result is too small",
  invalid_number: "That number isn't valid",
  network_error: "Can't reach the calculator service. Is the backend running?",
};

export const httpCalculatorApi: CalculatorApi = {
  async calculate(request) {
    const body: CalculationRequestDto = { operation: request.operation, operands: request.operands };
    try {
      const data = await postJSON<CalculationResultDto>("/api/v1/calculations", body);
      return data.result;
    } catch (error) {
      if (error instanceof ApiError) {
        throw new CalculationError(error.code, friendlyMessages[error.code] ?? "Something went wrong, please try again");
      }
      throw error;
    }
  },
};
