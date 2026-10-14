import type { TapeEntry } from "../domain/keypad";

interface TapeProps {
  entries: TapeEntry[];
  onRecall: (value: string) => void;
}

export function Tape({ entries, onRecall }: TapeProps) {
  return (
    <section className="tape" aria-labelledby="tape-title">
      <h2 id="tape-title">Recent</h2>
      {entries.length === 0 ? (
        <p className="tape-empty">Your results will show up here.</p>
      ) : (
        <ol>
          {entries.map((entry, i) => (
            <li key={`${entries.length - i}`}>
              <button type="button" aria-label={`Use ${entry.result} (${entry.expression})`} onClick={() => onRecall(entry.result)}>
                <span className="tape-expression">{entry.expression} =</span>
                <span className="tape-result">{entry.result}</span>
              </button>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
