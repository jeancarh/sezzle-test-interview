import type { OperationName } from "./operations";

export interface CalculationRequest {
  operation: OperationName;
  operands: string[];
}

export interface CalculatorApi {
  calculate(request: CalculationRequest): Promise<string>;
}

export class CalculationError extends Error {
  readonly code: string;

  constructor(code: string, message: string) {
    super(message);
    this.name = "CalculationError";
    this.code = code;
  }
}
