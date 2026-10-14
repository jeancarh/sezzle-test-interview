import type { Key } from "../domain/keypad";
import type { BinaryOperation } from "../domain/operations";

interface Button {
  label: string;
  name: string;
  key: Key;
  tone: "digit" | "operator" | "function" | "equals";
  wide?: boolean;
}

const digit = (d: string): Button => ({ label: d, name: d, key: { kind: "digit", digit: d }, tone: "digit" });
const operator = (label: string, name: string, operation: BinaryOperation, wide = false): Button => ({
  label,
  name,
  key: { kind: "operator", operation },
  tone: "operator",
  wide,
});

const layout: Button[] = [
  { label: "C", name: "Clear", key: { kind: "clear" }, tone: "function" },
  { label: "⌫", name: "Backspace", key: { kind: "backspace" }, tone: "function" },
  { label: "√", name: "Square root", key: { kind: "sqrt" }, tone: "function" },
  operator("xʸ", "Power", "power"),
  digit("7"),
  digit("8"),
  digit("9"),
  operator("÷", "Divide", "divide"),
  digit("4"),
  digit("5"),
  digit("6"),
  operator("×", "Multiply", "multiply"),
  digit("1"),
  digit("2"),
  digit("3"),
  operator("−", "Subtract", "subtract"),
  { label: "±", name: "Toggle sign", key: { kind: "sign" }, tone: "digit" },
  digit("0"),
  { label: ".", name: "Decimal point", key: { kind: "decimal" }, tone: "digit" },
  operator("+", "Add", "add"),
  operator("% of", "Percent of", "percentage", true),
  { label: "=", name: "Equals", key: { kind: "equals" }, tone: "equals", wide: true },
];

interface KeypadProps {
  activeOperator: BinaryOperation | null;
  onPress: (key: Key) => void;
}

export function Keypad({ activeOperator, onPress }: KeypadProps) {
  return (
    <div className="keypad" role="group" aria-label="Keypad">
      {layout.map((button) => {
        const pressed = button.key.kind === "operator" ? button.key.operation === activeOperator : undefined;
        return (
          <button
            key={button.name}
            type="button"
            className={`key key-${button.tone}${button.wide ? " key-wide" : ""}`}
            aria-label={button.name}
            aria-pressed={pressed}
            onClick={() => onPress(button.key)}
          >
            {button.label}
          </button>
        );
      })}
    </div>
  );
}
