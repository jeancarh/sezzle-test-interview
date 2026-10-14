import type { CalculationRequest } from "./calculatorApi";
import { symbols, type BinaryOperation } from "./operations";

export const maxInputLength = 32;
export const maxTapeLength = 10;

export type Key =
  | { kind: "digit"; digit: string }
  | { kind: "decimal" }
  | { kind: "sign" }
  | { kind: "backspace" }
  | { kind: "clear" }
  | { kind: "equals" }
  | { kind: "sqrt" }
  | { kind: "operator"; operation: BinaryOperation };

export type Action =
  | { type: "key"; key: Key }
  | { type: "resolved"; id: number; value: string }
  | { type: "failed"; id: number; message: string }
  | { type: "recall"; value: string };

type Continuation =
  | { kind: "equals"; operation: BinaryOperation; left: string; right: string }
  | { kind: "chain"; operation: BinaryOperation; left: string; right: string; next: BinaryOperation }
  | { kind: "unary"; operand: string };

export interface InFlight {
  id: number;
  request: CalculationRequest;
  then: Continuation;
}

export interface TapeEntry {
  expression: string;
  result: string;
}

export interface CalculatorState {
  display: string;
  expression: string;
  accumulator: string | null;
  operator: BinaryOperation | null;
  overwrite: boolean;
  operandReady: boolean;
  repeat: { operation: BinaryOperation; operand: string } | null;
  inFlight: InFlight | null;
  queue: Key[];
  error: string | null;
  nextId: number;
  tape: TapeEntry[];
}

export const initialState: CalculatorState = {
  display: "0",
  expression: "",
  accumulator: null,
  operator: null,
  overwrite: true,
  operandReady: false,
  repeat: null,
  inFlight: null,
  queue: [],
  error: null,
  nextId: 1,
  tape: [],
};

export function calculatorReducer(state: CalculatorState, action: Action): CalculatorState {
  switch (action.type) {
    case "key":
      return onKey(state, action.key);
    case "resolved":
      return onResolved(state, action.id, action.value);
    case "failed":
      return onFailed(state, action.id, action.message);
    case "recall":
      if (state.inFlight) {
        return state;
      }
      return { ...state, display: action.value, overwrite: true, operandReady: true, error: null };
  }
}

function onKey(state: CalculatorState, key: Key): CalculatorState {
  if (key.kind === "clear") {
    return { ...initialState, nextId: state.nextId, tape: state.tape };
  }
  if (state.inFlight) {
    return { ...state, queue: [...state.queue, key] };
  }
  const s = state.error ? { ...initialState, nextId: state.nextId, tape: state.tape } : state;

  switch (key.kind) {
    case "digit":
      return typeDigit(s, key.digit);
    case "decimal":
      return typeDecimal(s);
    case "sign":
      return toggleSign(s);
    case "backspace":
      return backspace(s);
    case "operator":
      return pressOperator(s, key.operation);
    case "equals":
      return pressEquals(s);
    case "sqrt":
      return request(s, { operation: "sqrt", operands: [operand(s.display)] }, { kind: "unary", operand: operand(s.display) });
  }
}

function typeDigit(s: CalculatorState, digit: string): CalculatorState {
  if (s.overwrite) {
    return { ...s, display: digit, overwrite: false, operandReady: true, expression: s.operator ? s.expression : "" };
  }
  if (s.display === "0") {
    return { ...s, display: digit, operandReady: true };
  }
  if (s.display === "-0") {
    return { ...s, display: "-" + digit, operandReady: true };
  }
  if (s.display.length >= maxInputLength) {
    return s;
  }
  return { ...s, display: s.display + digit, operandReady: true };
}

function typeDecimal(s: CalculatorState): CalculatorState {
  if (s.overwrite) {
    return { ...s, display: "0.", overwrite: false, operandReady: true, expression: s.operator ? s.expression : "" };
  }
  if (s.display.includes(".") || s.display.length >= maxInputLength) {
    return s;
  }
  return { ...s, display: s.display + ".", operandReady: true };
}

function toggleSign(s: CalculatorState): CalculatorState {
  if (operand(s.display) === "0") {
    return s;
  }
  const display = s.display.startsWith("-") ? s.display.slice(1) : "-" + s.display;
  return { ...s, display, operandReady: true };
}

function backspace(s: CalculatorState): CalculatorState {
  if (s.overwrite) {
    return s;
  }
  const trimmed = s.display.slice(0, -1);
  return { ...s, display: trimmed === "" || trimmed === "-" ? "0" : trimmed };
}

function pressOperator(s: CalculatorState, next: BinaryOperation): CalculatorState {
  if (s.operator && s.accumulator !== null && s.operandReady) {
    const left = s.accumulator;
    const right = operand(s.display);
    return request(
      s,
      { operation: s.operator, operands: [left, right] },
      { kind: "chain", operation: s.operator, left, right, next },
    );
  }
  if (s.operator && s.accumulator !== null) {
    return { ...s, operator: next, expression: `${s.accumulator} ${symbols[next]}` };
  }
  const accumulator = operand(s.display);
  return {
    ...s,
    accumulator,
    operator: next,
    overwrite: true,
    operandReady: false,
    repeat: null,
    expression: `${accumulator} ${symbols[next]}`,
  };
}

function pressEquals(s: CalculatorState): CalculatorState {
  if (s.operator && s.accumulator !== null) {
    const left = s.accumulator;
    const right = operand(s.display);
    return request(s, { operation: s.operator, operands: [left, right] }, { kind: "equals", operation: s.operator, left, right });
  }
  if (s.repeat) {
    const left = operand(s.display);
    const { operation, operand: right } = s.repeat;
    return request(s, { operation, operands: [left, right] }, { kind: "equals", operation, left, right });
  }
  return s;
}

function request(s: CalculatorState, req: CalculationRequest, then: Continuation): CalculatorState {
  return { ...s, inFlight: { id: s.nextId, request: req, then }, nextId: s.nextId + 1 };
}

function onResolved(state: CalculatorState, id: number, value: string): CalculatorState {
  if (state.inFlight?.id !== id) {
    return state;
  }
  const then = state.inFlight.then;
  const base: CalculatorState = { ...state, inFlight: null, queue: [], display: value, overwrite: true, error: null };
  let next: CalculatorState;

  switch (then.kind) {
    case "equals": {
      const expression = `${then.left} ${symbols[then.operation]} ${then.right}`;
      next = {
        ...base,
        expression: `${expression} =`,
        accumulator: null,
        operator: null,
        operandReady: false,
        repeat: { operation: then.operation, operand: then.right },
        tape: addToTape(state.tape, expression, value),
      };
      break;
    }
    case "chain": {
      const expression = `${then.left} ${symbols[then.operation]} ${then.right}`;
      next = {
        ...base,
        expression: `${value} ${symbols[then.next]}`,
        accumulator: value,
        operator: then.next,
        operandReady: false,
        repeat: null,
        tape: addToTape(state.tape, expression, value),
      };
      break;
    }
    case "unary": {
      const root = `${symbols.sqrt}(${then.operand})`;
      const prefix = state.operator && state.accumulator !== null ? `${state.accumulator} ${symbols[state.operator]} ` : "";
      next = {
        ...base,
        expression: prefix + root,
        operandReady: true,
        tape: addToTape(state.tape, root, value),
      };
      break;
    }
  }

  return state.queue.reduce((s, key) => onKey(s, key), next);
}

function onFailed(state: CalculatorState, id: number, message: string): CalculatorState {
  if (state.inFlight?.id !== id) {
    return state;
  }
  return { ...initialState, nextId: state.nextId, tape: state.tape, error: message };
}

function addToTape(tape: TapeEntry[], expression: string, result: string): TapeEntry[] {
  return [{ expression, result }, ...tape].slice(0, maxTapeLength);
}

function operand(display: string): string {
  const trimmed = display.endsWith(".") ? display.slice(0, -1) : display;
  return trimmed === "-0" || trimmed === "" ? "0" : trimmed;
}
