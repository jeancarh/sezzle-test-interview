import { describe, expect, it } from "vitest";
import {
  calculatorReducer,
  initialState,
  maxInputLength,
  type CalculatorState,
  type Key,
} from "./keypad";

const digit = (d: string): Key => ({ kind: "digit", digit: d });
const op = (operation: Extract<Key, { kind: "operator" }>["operation"]): Key => ({ kind: "operator", operation });
const keys: Record<string, Key> = {
  ".": { kind: "decimal" },
  "±": { kind: "sign" },
  "⌫": { kind: "backspace" },
  C: { kind: "clear" },
  "=": { kind: "equals" },
  "√": { kind: "sqrt" },
  "+": op("add"),
  "-": op("subtract"),
  "*": op("multiply"),
  "/": op("divide"),
  "^": op("power"),
  "%": op("percentage"),
};

function press(state: CalculatorState, sequence: string): CalculatorState {
  return [...sequence.replace(/ /g, "")].reduce(
    (s, ch) => calculatorReducer(s, { type: "key", key: keys[ch] ?? digit(ch) }),
    state,
  );
}

function resolve(state: CalculatorState, value: string): CalculatorState {
  if (!state.inFlight) throw new Error("nothing in flight");
  return calculatorReducer(state, { type: "resolved", id: state.inFlight.id, value });
}

function fail(state: CalculatorState, message: string): CalculatorState {
  if (!state.inFlight) throw new Error("nothing in flight");
  return calculatorReducer(state, { type: "failed", id: state.inFlight.id, message });
}

describe("entering numbers", () => {
  it("starts at zero and replaces the leading zero", () => {
    expect(initialState.display).toBe("0");
    expect(press(initialState, "0 0 7").display).toBe("7");
  });

  it("accepts one decimal point", () => {
    expect(press(initialState, "1 . 5 . 2").display).toBe("1.52");
    expect(press(initialState, ". 5").display).toBe("0.5");
  });

  it("toggles the sign but never on zero", () => {
    expect(press(initialState, "4 2 ±").display).toBe("-42");
    expect(press(initialState, "4 2 ± ±").display).toBe("42");
    expect(press(initialState, "±").display).toBe("0");
  });

  it("deletes the last character and falls back to zero", () => {
    expect(press(initialState, "1 2 3 ⌫").display).toBe("12");
    expect(press(initialState, "5 ± ⌫").display).toBe("0");
    expect(press(initialState, "7 ⌫ ⌫").display).toBe("0");
  });

  it("caps the input length", () => {
    const long = press(initialState, "9".repeat(maxInputLength + 5));
    expect(long.display).toHaveLength(maxInputLength);
  });

  it("clears everything but the tape", () => {
    let s = press(initialState, "2 + 3 =");
    s = resolve(s, "5");
    s = press(s, "9 C");
    expect(s.display).toBe("0");
    expect(s.operator).toBeNull();
    expect(s.tape).toHaveLength(1);
  });
});

describe("binary operations", () => {
  it("sends the operation on equals and shows the result", () => {
    const s = press(initialState, "1 . 5 + 2 =");
    expect(s.inFlight?.request).toEqual({ operation: "add", operands: ["1.5", "2"] });
    const done = resolve(s, "3.5");
    expect(done.display).toBe("3.5");
    expect(done.expression).toBe("1.5 + 2 =");
    expect(done.inFlight).toBeNull();
    expect(done.tape[0]).toEqual({ expression: "1.5 + 2", result: "3.5" });
  });

  it("drops a trailing decimal point before sending", () => {
    const s = press(initialState, "5 . * 2 =");
    expect(s.inFlight?.request.operands).toEqual(["5", "2"]);
  });

  it("lets you change the operator before typing the next number", () => {
    const s = press(initialState, "8 + - * 2 =");
    expect(s.inFlight?.request).toEqual({ operation: "multiply", operands: ["8", "2"] });
  });

  it("chains operators by calculating the pending one first", () => {
    let s = press(initialState, "2 + 3 *");
    expect(s.inFlight?.request).toEqual({ operation: "add", operands: ["2", "3"] });
    s = resolve(s, "5");
    expect(s.display).toBe("5");
    expect(s.expression).toBe("5 ×");
    s = press(s, "4 =");
    expect(s.inFlight?.request).toEqual({ operation: "multiply", operands: ["5", "4"] });
  });

  it("repeats the last operation on repeated equals", () => {
    let s = resolve(press(initialState, "2 + 3 ="), "5");
    s = press(s, "=");
    expect(s.inFlight?.request).toEqual({ operation: "add", operands: ["5", "3"] });
    s = resolve(s, "8");
    expect(s.display).toBe("8");
  });

  it("uses the shown number twice when equals follows an operator", () => {
    const s = press(initialState, "5 * =");
    expect(s.inFlight?.request).toEqual({ operation: "multiply", operands: ["5", "5"] });
  });

  it("starts a fresh number after a result", () => {
    let s = resolve(press(initialState, "2 + 3 ="), "5");
    s = press(s, "7");
    expect(s.display).toBe("7");
    expect(s.expression).toBe("");
  });

  it("continues from a result with a new operator", () => {
    let s = resolve(press(initialState, "2 + 3 ="), "5");
    s = press(s, "- 1 =");
    expect(s.inFlight?.request).toEqual({ operation: "subtract", operands: ["5", "1"] });
  });

  it("reads percentage as a percent of b", () => {
    const s = press(initialState, "1 5 % 2 0 0 =");
    expect(s.inFlight?.request).toEqual({ operation: "percentage", operands: ["15", "200"] });
  });

  it("does nothing on equals with no pending operation", () => {
    const s = press(initialState, "4 =");
    expect(s.inFlight).toBeNull();
    expect(s.display).toBe("4");
  });
});

describe("square root", () => {
  it("applies to the shown number and keeps it usable as an operand", () => {
    let s = press(initialState, "9 + 1 6 √");
    expect(s.inFlight?.request).toEqual({ operation: "sqrt", operands: ["16"] });
    s = resolve(s, "4");
    expect(s.display).toBe("4");
    expect(s.expression).toBe("9 + √(16)");
    s = press(s, "=");
    expect(s.inFlight?.request).toEqual({ operation: "add", operands: ["9", "4"] });
  });
});

describe("requests in flight", () => {
  it("queues keys pressed while waiting and replays them in order", () => {
    let s = press(initialState, "2 + 3 * 4 =");
    expect(s.inFlight?.request.operation).toBe("add");
    expect(s.queue).toHaveLength(2);
    s = resolve(s, "5");
    expect(s.inFlight?.request).toEqual({ operation: "multiply", operands: ["5", "4"] });
    expect(s.queue).toHaveLength(0);
    s = resolve(s, "20");
    expect(s.display).toBe("20");
  });

  it("ignores responses for requests that are no longer current", () => {
    const s = press(initialState, "2 + 3 =");
    const stale = s.inFlight!.id;
    const cleared = press(s, "C");
    const after = calculatorReducer(cleared, { type: "resolved", id: stale, value: "5" });
    expect(after.display).toBe("0");
  });

  it("ignores a duplicate response for the same request", () => {
    const s = press(initialState, "2 + 3 =");
    const id = s.inFlight!.id;
    const once = calculatorReducer(s, { type: "resolved", id, value: "5" });
    const twice = calculatorReducer(once, { type: "resolved", id, value: "5" });
    expect(twice.tape).toHaveLength(1);
  });
});

describe("errors", () => {
  it("shows the message and resets the calculation", () => {
    const s = fail(press(initialState, "1 / 0 ="), "Can't divide by zero");
    expect(s.error).toBe("Can't divide by zero");
    expect(s.display).toBe("0");
    expect(s.operator).toBeNull();
    expect(s.inFlight).toBeNull();
  });

  it("drops queued keys when a request fails", () => {
    let s = press(initialState, "1 / 0 = + 5");
    s = fail(s, "Can't divide by zero");
    expect(s.queue).toHaveLength(0);
  });

  it("clears the error on the next key", () => {
    const s = press(fail(press(initialState, "1 / 0 ="), "Can't divide by zero"), "7");
    expect(s.error).toBeNull();
    expect(s.display).toBe("7");
  });
});

describe("tape", () => {
  it("keeps the ten most recent results, newest first", () => {
    let s = initialState;
    for (let i = 1; i <= 12; i++) {
      s = resolve(press(s, `${i % 10} + 1 =`), String((i % 10) + 1));
    }
    expect(s.tape).toHaveLength(10);
    expect(s.tape[0]?.expression).toBe("2 + 1");
  });

  it("recalls a value as the current operand", () => {
    let s = resolve(press(initialState, "2 + 3 ="), "5");
    s = press(s, "* ");
    s = calculatorReducer(s, { type: "recall", value: "5" });
    expect(s.display).toBe("5");
    s = press(s, "=");
    expect(s.inFlight?.request).toEqual({ operation: "multiply", operands: ["5", "5"] });
  });
});
