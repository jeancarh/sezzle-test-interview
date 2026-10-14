# Calculator

A full-stack calculator: a React + TypeScript frontend that talks to a small Go REST service. It does add, subtract, multiply, divide, power, square root and percentage, with exact decimal math, so `0.1 + 0.2` is `0.3`.

The whole thing ships as one Docker container. You only need Git and Docker to run it and to run the tests.

- [Requirements](#requirements)
- [Quick start](#quick-start)
- [Running the backend and frontend separately](#running-the-backend-and-frontend-separately)
- [Tests and coverage](#tests-and-coverage)
- [Using the API](#using-the-api)
- [Project layout](#project-layout)
- [Approach](#approach)
- [Design decisions](#design-decisions)
- [Assumptions I'd confirm with the product owner](#assumptions-id-confirm-with-the-product-owner)
- [Alternatives considered](#alternatives-considered)
- [Deploying to AWS](#deploying-to-aws)
- [Known limitations](#known-limitations)

Prompts used to build this are in [PROMPTS.md](PROMPTS.md), and the coverage report is in [COVERAGE.md](COVERAGE.md).

## Requirements

To **run the app** and **run the tests** you only need Git and Docker.

| | macOS | Windows 10/11 | Linux |
| --- | --- | --- | --- |
| Git | `xcode-select --install` or `brew install git` | `winget install Git.Git` or [git-scm.com](https://git-scm.com/download/win) | your package manager, e.g. `sudo apt install git` |
| Docker | [Docker Desktop](https://docs.docker.com/desktop/setup/install/mac-install/) or `brew install --cask docker-desktop` | [Docker Desktop](https://docs.docker.com/desktop/setup/install/windows-install/) or `winget install Docker.DockerDesktop` (it asks to enable WSL 2, say yes and reboot) | [Docker Engine](https://docs.docker.com/engine/install/) with the Compose plugin |
| `make` (optional) | already there with the Xcode command line tools | not needed, every `make` command below has a plain Docker version | usually there, or `sudo apt install make` |

Only if you want to **work on the code without Docker**:

| | macOS | Windows | Linux |
| --- | --- | --- | --- |
| Go 1.26+ | `brew install go` | `winget install GoLang.Go` | [go.dev/dl](https://go.dev/dl/) |
| Node 20.19+ (24 recommended) | `brew install node` | `winget install OpenJS.NodeJS.LTS` | [nodejs.org](https://nodejs.org/) or nvm |

A few things to know:

- The first build downloads the Node, Go and runtime images, so keep about 2 GB of free disk space.
- On Linux, if Docker only works with `sudo`, run `sudo usermod -aG docker $USER`, then log out and back in.
- Apple Silicon and Intel Macs both work, since all images are multi-architecture.
- On Windows, use PowerShell. Commands are the same unless a Windows version is shown.

Check what you have:

```bash
git --version
docker --version
docker compose version
```

## Quick start

**1. Get the code**

```bash
git clone https://github.com/jeancarh/sezzle-test-interview.git
cd sezzle-test-interview
```

**2. Make sure Docker is running**

Open Docker Desktop (or start the Docker service on Linux) and check that it answers:

```bash
docker info
```

If you see "Cannot connect to the Docker daemon", Docker isn't running yet. On Windows, wait until Docker Desktop says "Engine running".

**3. Start the app**

```bash
make up
```

No `make`? Run the same thing directly:

```bash
docker compose up --build -d --wait
```

The first run builds the image and takes a few minutes. After that it takes seconds. The command returns once the app passes its health check.

**4. Open it**

| What | URL |
| --- | --- |
| Calculator | http://localhost:8080 |
| API docs (Swagger UI) | http://localhost:8080/docs |
| OpenAPI spec | http://localhost:8080/api/openapi.yaml |
| Health check | http://localhost:8080/healthz |

**5. Try it (2 minutes)**

1. Type `0.1 + 0.2 =` with the keyboard. You get `0.3`, not `0.30000000000000004`.
2. Type `2 + 3 * 4 =`. Like a pocket calculator, it works left to right: `2 + 3` is calculated when you press `*`, then `5 × 4 = 20`.
3. Type `1 / 0 =`. You get a clear error, and the next key starts over.
4. Click **√** after typing `2`, or **% of** between `15` and `200` (15% of 200 = 30).
5. Click any line under **Recent** to reuse that result.

Keyboard shortcuts: digits, `.`, `+ - * /`, `^` for power, `%` for percent of, `r` for square root, `n` to toggle the sign, `Enter` or `=` to calculate, `Backspace` to delete, `Esc` to clear.

**6. Stop it**

```bash
make down    # stop the app
make logs    # see the logs while it runs
```

Without `make`: `docker compose down` and `docker compose logs -f calculator`.

**Port 8080 already in use?** Pick another one and use it in the URLs above:

```bash
make up HOST_PORT=9090
# or, macOS / Linux
HOST_PORT=9090 docker compose up --build -d --wait
```

```powershell
# Windows PowerShell
$env:HOST_PORT=9090; docker compose up --build -d --wait
```

## Running the backend and frontend separately

This is for working on the code, and needs Go and Node. Stop the container first (`make down`), since both use port 8080.

```bash
make dev
```

That starts the API on http://localhost:8080 and the Vite dev server on http://localhost:5173. Open **http://localhost:5173**: it reloads on changes and forwards `/api` to the backend. `Ctrl+C` stops both.

`make dev` needs a Unix shell. On Windows, or if you prefer, run each side in its own terminal:

```bash
# terminal 1: backend
cd backend
go run ./cmd/api
```

```bash
# terminal 2: frontend
cd frontend
npm install
npm run dev
```

The backend reads one environment variable, `PORT` (default `8080`). Run it on another port with `PORT=9000 go run ./cmd/api`, or `$env:PORT=9000; go run ./cmd/api` in PowerShell. The frontend proxy expects 8080.

## Tests and coverage

**With Docker only**

```bash
make test-docker   # all tests
make coverage      # coverage reports + COVERAGE.md
```

Both run inside the official Go and Node images, so nothing gets installed on your machine. The first run downloads the images, which takes about a minute. Without `make`, from the repo root:

```bash
# macOS / Linux
docker run --rm -v "$PWD:/repo" -v calculator-go-cache:/go/pkg/mod -w /repo/backend golang:1.26 sh -c "go vet ./... && go test -race ./..."
docker run --rm -v "$PWD:/repo" -v /repo/frontend/node_modules -w /repo/frontend node:24-alpine sh -c "npm ci && npm run typecheck && npm run lint && npm test"
```

```powershell
# Windows PowerShell
docker run --rm -v "${PWD}:/repo" -v calculator-go-cache:/go/pkg/mod -w /repo/backend golang:1.26 sh -c "go vet ./... && go test -race ./..."
docker run --rm -v "${PWD}:/repo" -v /repo/frontend/node_modules -w /repo/frontend node:24-alpine sh -c "npm ci && npm run typecheck && npm run lint && npm test"
```

**With Go and Node installed**

```bash
make test
```

Or one side at a time:

```bash
cd backend && go vet ./... && go test -race -cover ./...
cd frontend && npm install && npm run typecheck && npm run lint && npm run coverage
```

Every Go package should print `ok`, and Vitest should end with `67 passed`. The current numbers are in [COVERAGE.md](COVERAGE.md): above 90% on every tested Go package and 96% on the frontend. CI uploads the HTML reports as build artifacts on every push.

**What the tests cover**

| Level | Backend | Frontend |
| --- | --- | --- |
| Unit | Number parsing and formatting, every operation, every error code (`internal/calculator`), DTO and error mapping (`internal/adapter/httpapi`) | The keypad state machine, every transition (`domain/keypad.test.ts`), keyboard mapping, the HTTP adapter with a stubbed `fetch` |
| Integration | The real router on a real HTTP server wired to the real domain (`router_test.go`): every operation, every error status and code, malformed JSON, wrong content type, body limit, routing, SPA fallback, caching headers, request ids, panics | The hook against a fake port, and the full page with Testing Library: clicking, typing, chaining, errors, keyboard-only use, history, accessibility names |

## Using the API

Swagger UI at http://localhost:8080/docs has every endpoint with examples. Open **POST /api/v1/calculations**, click **Try it out**, pick an example from the dropdown and click **Execute**.

The same with curl:

```bash
curl -X POST http://localhost:8080/api/v1/calculations \
  -H "Content-Type: application/json" \
  -d '{"operation": "add", "operands": ["0.1", "0.2"]}'
```

```json
{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}
```

```bash
curl -X POST http://localhost:8080/api/v1/calculations \
  -H "Content-Type: application/json" \
  -d '{"operation": "sqrt", "operands": ["2"]}'
```

```json
{"operation":"sqrt","operands":["2"],"result":"1.414213562373095048801688724209698"}
```

Errors use RFC 9457 `application/problem+json` with a stable `code`:

```bash
curl -X POST http://localhost:8080/api/v1/calculations \
  -H "Content-Type: application/json" \
  -d '{"operation": "divide", "operands": ["1", "0"]}'
```

```json
{"type":"about:blank","title":"Unprocessable Entity","status":422,"code":"division_by_zero","detail":"can't divide by zero"}
```

List the operations:

```bash
curl http://localhost:8080/api/v1/operations
```

| Operation | Operands | Example | Result |
| --- | --- | --- | --- |
| `add` | 2 | `["0.1", "0.2"]` | `"0.3"` |
| `subtract` | 2 | `["10", "4.25"]` | `"5.75"` |
| `multiply` | 2 | `["1.5", "4"]` | `"6"` |
| `divide` | 2 | `["1", "3"]` | `"0.3333333333333333333333333333333333"` |
| `power` | 2 | `["2", "10"]` | `"1024"` |
| `sqrt` | 1 | `["16"]` | `"4"` |
| `percentage` | 2 | `["15", "200"]` | `"30"` (15% of 200) |

| Status | `code` | When |
| --- | --- | --- |
| 400 | `malformed_request` | Invalid JSON, unknown fields, operands that aren't strings or numbers |
| 413 | `request_too_large` | Body over 4 KB |
| 415 | `unsupported_media_type` | Body not sent as `application/json` |
| 422 | `invalid_number` | Not a plain decimal, over 34 significant digits, or outside 1E-9999 to 1E+9999 |
| 422 | `wrong_operand_count`, `unknown_operation` | Wrong number of operands, or an operation that doesn't exist |
| 422 | `division_by_zero`, `negative_square_root`, `zero_to_negative_power`, `complex_result` | Math that has no real answer |
| 422 | `result_too_large`, `result_too_small` | The result is outside the supported range |

**Send numbers as strings.** JSON numbers are accepted, but JavaScript turns them into doubles before they leave the browser, so `0.1` is already wrong by then. Strings keep every digit. Results always come back as strings.

## Project layout

```
backend/
  cmd/api/                    composition root: config, wiring, server, graceful shutdown, healthcheck
  internal/calculator/        the core: Number value object, operations, service, error codes
  internal/adapter/httpapi/   driving adapter: handlers, DTOs, problem+json, middleware, SPA serving
  internal/config/            environment variables
  api/openapi.yaml            the contract, embedded and served at /docs
  web/                        the built frontend, embedded into the binary
frontend/src/
  app/                        composition root: providers, picks the HTTP adapter
  features/calculator/
    domain/                   keypad state machine, operation types, the CalculatorApi port
    application/              useCalculator hook (TanStack Query mutation + reducer)
    infrastructure/           HTTP adapter implementing the port
    ui/                       CalculatorPage (container), Display, Keypad, Tape (presentational)
  shared/                     HTTP client, the port-context factory, generated API types
```

## Approach

I wrote the plan as prompts before writing any code. Each prompt holds the architecture and the decisions for one phase, with the reason behind each. They're in [PROMPTS.md](PROMPTS.md). Then I built it phase by phase, and each phase had to be green before the next one started:

1. **Domain first.** Before designing anything on top of the decimal library, I probed how it behaves on edge cases. For example, it returns `Infinity` for `0 ^ -1` without an error flag, so the domain checks that case and the result range explicitly, with tests for each.
2. **Contract.** The OpenAPI spec was written by hand before any handler, with an example for every success and error.
3. **Adapters.** The HTTP layer was tested at two levels: unit tests for the mapping, and integration tests on a real server.
4. **Frontend.** The keypad logic went in first as a pure state machine with its own tests. The UI came after.
5. **Packaging.** One container, plus tests and coverage that run with Docker only.
6. **Adversarial review.** I tried to break it on purpose: precision, huge exponents, odd inputs, wrong content types, fast typing, double clicks, Enter on a focused button, narrow screens, dark mode. Every one of those behaviors has a test.

## Design decisions

**Hexagonal architecture, applied honestly.** `internal/calculator` is the core. It has the math rules, imports nothing from HTTP, and defines the port it offers: `Calculator` (`service.go`). The HTTP adapter depends on that interface, and `cmd/api/main.go` is the only place that wires things together. There's no repository or database, because a calculator has nothing to store, and inventing one to show the pattern would be ceremony, not design. If persistence ever shows up, say a history of calculations, it becomes a driven port next to the service.

**Exact decimal math.** Calculations use `cockroachdb/apd/v3` with 34 significant digits, never `float64`. A calculator that's wrong at the 17th digit is wrong. Numbers travel as strings in JSON for the same reason.

**`Number` is a value object.** `ParseNumber` (`number.go`) is the only way to build one, and it enforces every input rule: plain ASCII decimal, optional exponent, at most 100 characters, at most 34 significant digits, range 1E-9999 to 1E+9999. Nothing after it validates again.

**Operations are strategies in one registry.** Each operation is a name, a symbol, an arity and an apply function (`operations.go`). Adding one means one function and one line in the registry. The service, the HTTP layer and the contract's shape don't change.

**One endpoint, not one per operation.** `POST /api/v1/calculations` with `{operation, operands}` keeps the contract uniform, so a new operation is data, not a new route. `GET /api/v1/operations` lets clients discover them.

**Stable error codes.** Every domain error has a code (`errors.go`) that goes out in problem+json. Clients branch on `code` and never parse messages. One function in the adapter maps errors to status codes (`problem.go`).

**Abuse limits are correctness.** The operand length, significant digits, exponent range and 4 KB body limit together make sure no single request can be expensive. The worst case I found, `1.0000001 ^ 1e9999`, answers in about a millisecond.

**No idempotency keys.** Calculations are pure, so retrying a request can't do any harm. Adding a key would be a pattern applied by reflex.

**Hand-written OpenAPI, not swaggo.** swaggo builds the spec from code comments, and this code has none. The frontend generates its types from the same file (`npm run gen:api`), and the generator reprints them without comments too.

**Standard library HTTP.** Go's router handles methods and path params, so a framework would only add a dependency. Middleware covers request ids (`X-Request-ID`), JSON logs with `slog` and panic recovery. The server has timeouts and shuts down gracefully.

**Frontend layers that mirror the backend.** `domain` is plain TypeScript: the keypad state machine and the `CalculatorApi` port. `application` is the `useCalculator` hook. `infrastructure` is the HTTP adapter. `ui` holds a container and presentational components. `app/App.tsx` injects the HTTP adapter through a small port context (`shared/di`), and tests inject a fake. ESLint's `no-restricted-imports`, configured per folder, fails the build if a layer imports something it shouldn't. I didn't add use-case classes, because the hook is the application layer and classes would only forward calls.

**The keypad is a pure state machine.** `domain/keypad.ts` handles entry, chaining, repeated equals, errors and requests in flight. Keys pressed while a request is pending are queued and replayed in order, so fast typing never loses input. Stale or duplicate responses are ignored by request id.

**The backend does all the math.** The frontend never calculates. It only limits input shape: 32 characters and one decimal point.

**Accessibility is part of done.** Every key is a real button with an accessible name. The result is announced through `aria-live`, errors use `role="alert"` and `aria-invalid`, the operator waiting for its second number shows `aria-pressed`, and everything works from the keyboard. Light and dark themes come from CSS tokens, and the layout is mobile first.

**One container.** A multi-stage Dockerfile builds the frontend, embeds it in a static Go binary and runs it on distroless as non-root, with a `healthcheck` subcommand since there's no curl in the image. The image is about 16 MB.

**Patterns and principles actually used**

- Ports and adapters: the `Calculator` port in the core, the HTTP adapter around it, `CalculatorApi` in the frontend.
- Strategy and registry: `operations.go`.
- Value object and factory: `Number` and `ParseNumber`.
- State machine: `calculatorReducer`.
- Container and presentational components: `CalculatorPage` versus `Display`, `Keypad` and `Tape`.
- Dependency injection by hand in `main.go` and `App.tsx`, no container.
- SRP: parsing, operations, orchestration and transport each live in their own place. OCP: new operations don't touch existing code. DIP: handlers and hooks depend on ports, not implementations.

## Assumptions I'd confirm with the product owner

1. **What does `%` mean?** I went with "a percent of b" (`15 % of 200 = 30`). Pocket calculators often use "divide by 100" or "add a percentage" instead.
2. **What is `0 ^ 0`?** I return `1`, like most languages and calculators.
3. **How precise is enough?** 34 significant digits. Longer inputs are rejected rather than rounded.
4. **Chaining order.** `2 + 3 × 4` gives `20`, left to right like a pocket calculator, not `14`. A scientific calculator would respect operator precedence, and that would mean sending whole expressions to the API.
5. **Should history persist?** It's in the browser tab only. Saving it would add the first driven port (storage) and a user concept.

## Alternatives considered

| Topic | Chosen | Also considered | Why not |
| --- | --- | --- | --- |
| Number type | `apd` decimals as strings | `float64`, `math/big.Float` | Floats can't represent `0.1`. `big.Float` is binary floating point too, so `0.1` still isn't exact. |
| API shape | One `calculations` endpoint | `/add`, `/divide`, ... | One route per operation multiplies handlers and docs for no gain. |
| Expression parsing | Binary operations | `POST {"expression": "2+3*4"}` | Needs a parser and precedence rules, which is more scope than the brief asks for. Listed as an assumption. |
| HTTP | `net/http` | Gin, Echo, Chi | The stdlib router covers it. |
| API docs | Hand-written OpenAPI | swaggo annotations | Built from comments, and the code has none. |
| Frontend state | Reducer + TanStack Query mutation | Redux, Zustand | Only one screen's worth of state. |
| Layer rules | ESLint `no-restricted-imports` | eslint-plugin-boundaries | The built-in rule is enough and keeps `npm audit` clean. |
| Deployment shape | One container | nginx + API containers | Two artifacts to run and version for no benefit at this size. |
| TypeScript | 5.9 | 7.x | typescript-eslint and the type generator don't support 7 yet. |

## Deploying to AWS

The service is stateless and cheap per request, so the right AWS shape is pay-per-use:

- **AWS Lambda with the Lambda Web Adapter** runs this same container image. The only change is one extra `COPY` line in the Dockerfile that adds the adapter as a Lambda extension. It turns Lambda events into plain HTTP requests, so the Go server doesn't change. The image lives in **ECR**.
- **CloudFront** in front, with a free ACM certificate. It caches the hashed static assets at the edge, and throttling or **AWS WAF** rate limits go here too.
- **CloudWatch** gets the JSON logs as they are, plus an alarm on errors and latency.

At low traffic, that costs close to nothing: the Lambda and CloudFront free tiers cover millions of requests a month, and there's no server sitting idle.

If traffic becomes steady enough that per-request pricing costs more than a running service, the same image moves to **App Runner** or **ECS Fargate** behind an ALB with no code changes. The CI workflow would get one more job that pushes the image to ECR on merge to `main`.

## Known limitations

These are deliberate limits, not open bugs:

- **34 significant digits.** Results that can't be exact, like `1 / 3` or `sqrt(2)`, are rounded to 34 significant digits (half-even), the same precision as IEEE 754 decimal128. Inputs over 34 significant digits are rejected rather than rounded, so the calculator never changes a number you typed.
- **Range 1E-9999 to 1E+9999.** Anything outside returns `result_too_large` or `result_too_small`. The bound is what keeps a request like `10 ^ 999999999` cheap.
- **`0 ^ 0` is `1`,** like Go's `math.Pow` and most calculators.
- **Percentage is "a percent of b".** See the assumptions above.
- **Swagger UI loads from a CDN** (pinned to 5.17.14), so `/docs` needs internet access. The API and the calculator don't.
- **No rate limiting inside the service.** A calculation takes about a millisecond and has no side effects. A public deployment would throttle at the edge (CloudFront/WAF or API Gateway).
- **Same origin only.** The binary serves the UI, so there's no CORS setup. Hosting the frontend elsewhere would need it.
- **History lives in the browser tab** and is gone on reload, because the service is stateless.
