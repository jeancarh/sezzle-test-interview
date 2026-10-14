import { describe, expect, it } from "vitest";
import { keyFromKeyboard } from "./keyboard";

describe("keyFromKeyboard", () => {
  it.each([
    ["7", { kind: "digit", digit: "7" }],
    [".", { kind: "decimal" }],
    [",", { kind: "decimal" }],
    ["*", { kind: "operator", operation: "multiply" }],
    ["x", { kind: "operator", operation: "multiply" }],
    ["/", { kind: "operator", operation: "divide" }],
    ["^", { kind: "operator", operation: "power" }],
    ["%", { kind: "operator", operation: "percentage" }],
    ["r", { kind: "sqrt" }],
    ["n", { kind: "sign" }],
    ["Enter", { kind: "equals" }],
    ["=", { kind: "equals" }],
    ["Backspace", { kind: "backspace" }],
    ["Escape", { kind: "clear" }],
  ])("maps %s", (key, expected) => {
    expect(keyFromKeyboard(key)).toEqual(expected);
  });

  it.each(["a", "Tab", "F5", "ArrowUp", "١", " "])("ignores %s", (key) => {
    expect(keyFromKeyboard(key)).toBeNull();
  });
});
