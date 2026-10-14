# Coverage

Generated with `make coverage`. HTML reports are written to `coverage/backend.html` and `coverage/frontend/index.html` (not committed). CI uploads the same reports as build artifacts.

## Backend (Go, statement coverage)

| Package | Coverage |
| --- | --- |
| `internal/adapter/httpapi` | 95.7% |
| `internal/calculator` | 91.4% |
| `internal/config` | 100.0% |

`cmd/api` and the embed-only packages have no tests: they only wire things together, and the integration tests exercise the same router that `main.go` builds.

## Frontend (Vitest, v8)

| File | Statements | Branches | Functions | Lines |
| --- | --- | --- | --- | --- |
| `src/app/App.tsx` | 75.0% | 100.0% | 66.7% | 75.0% |
| `src/features/calculator/application/calculator.ts` | 100.0% | 100.0% | 100.0% | 100.0% |
| `src/features/calculator/domain/calculatorApi.ts` | 100.0% | 100.0% | 100.0% | 100.0% |
| `src/features/calculator/domain/keypad.ts` | 95.3% | 92.4% | 100.0% | 95.3% |
| `src/features/calculator/domain/operations.ts` | 100.0% | 100.0% | 100.0% | 100.0% |
| `src/features/calculator/infrastructure/httpCalculatorApi.ts` | 88.9% | 75.0% | 100.0% | 88.9% |
| `src/features/calculator/ui/CalculatorPage.tsx` | 100.0% | 88.9% | 100.0% | 100.0% |
| `src/features/calculator/ui/Display.tsx` | 100.0% | 80.0% | 100.0% | 100.0% |
| `src/features/calculator/ui/Keypad.tsx` | 100.0% | 100.0% | 100.0% | 100.0% |
| `src/features/calculator/ui/Tape.tsx` | 100.0% | 100.0% | 100.0% | 100.0% |
| `src/features/calculator/ui/keyboard.ts` | 100.0% | 100.0% | 100.0% | 100.0% |
| `src/shared/di/createPortContext.tsx` | 100.0% | 100.0% | 100.0% | 100.0% |
| `src/shared/http/client.ts` | 100.0% | 100.0% | 100.0% | 100.0% |
| **Total** | **96.4%** | **92.1%** | **97.8%** | **96.3%** |
