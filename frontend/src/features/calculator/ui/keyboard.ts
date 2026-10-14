import type { Key } from "../domain/keypad";

const keyMap: Record<string, Key> = {
  ".": { kind: "decimal" },
  ",": { kind: "decimal" },
  "+": { kind: "operator", operation: "add" },
  "-": { kind: "operator", operation: "subtract" },
  "*": { kind: "operator", operation: "multiply" },
  x: { kind: "operator", operation: "multiply" },
  X: { kind: "operator", operation: "multiply" },
  "/": { kind: "operator", operation: "divide" },
  "^": { kind: "operator", operation: "power" },
  "%": { kind: "operator", operation: "percentage" },
  r: { kind: "sqrt" },
  R: { kind: "sqrt" },
  n: { kind: "sign" },
  N: { kind: "sign" },
  Enter: { kind: "equals" },
  "=": { kind: "equals" },
  Backspace: { kind: "backspace" },
  Escape: { kind: "clear" },
  Delete: { kind: "clear" },
};

export function keyFromKeyboard(key: string): Key | null {
  if (/^[0-9]$/.test(key)) {
    return { kind: "digit", digit: key };
  }
  return keyMap[key] ?? null;
}
