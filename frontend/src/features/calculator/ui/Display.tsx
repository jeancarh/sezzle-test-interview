interface DisplayProps {
  expression: string;
  value: string;
  error: string | null;
  busy: boolean;
}

export function Display({ expression, value, error, busy }: DisplayProps) {
  const size = value.length > 24 ? "small" : value.length > 14 ? "medium" : "large";
  return (
    <div className="display">
      <div className="expression" aria-label="Expression">
        {expression || " "}
      </div>
      <output
        className={`value value-${size}`}
        aria-label="Result"
        aria-live="polite"
        aria-busy={busy}
        aria-invalid={error !== null}
        aria-describedby={error ? "calculator-error" : undefined}
      >
        {value}
      </output>
      <p id="calculator-error" className="error" role="alert">
        {error ?? ""}
      </p>
    </div>
  );
}
