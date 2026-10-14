export type BinaryOperation = "add" | "subtract" | "multiply" | "divide" | "power" | "percentage";
export type UnaryOperation = "sqrt";
export type OperationName = BinaryOperation | UnaryOperation;

export const symbols: Record<OperationName, string> = {
  add: "+",
  subtract: "−",
  multiply: "×",
  divide: "÷",
  power: "^",
  percentage: "% of",
  sqrt: "√",
};
