import { useEffect } from "react";
import { useCalculator } from "../application/calculator";
import { Display } from "./Display";
import { Keypad } from "./Keypad";
import { Tape } from "./Tape";
import { keyFromKeyboard } from "./keyboard";

export function CalculatorPage() {
  const { state, press, recall, busy } = useCalculator();

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.ctrlKey || event.metaKey || event.altKey) {
        return;
      }
      const key = keyFromKeyboard(event.key);
      if (key) {
        event.preventDefault();
        press(key);
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [press]);

  return (
    <main className="layout">
      <section className="calculator" aria-label="Calculator">
        <Display expression={state.expression} value={state.display} error={state.error} busy={busy} />
        <Keypad activeOperator={state.operandReady ? null : state.operator} onPress={press} />
      </section>
      <Tape entries={state.tape} onRecall={recall} />
    </main>
  );
}
