# Prompts

The assignment asks for the prompts I used, so here they are.

## How I used them

Before writing any code, I wrote the plan as these eight prompts, one per phase. Each one states the architecture, the decisions and the reason behind each decision, so the tool never had to guess. I reviewed and approved them first, then the project was built phase by phase from them, in order, and a phase only started when the previous one was green.

What was mine and what was the AI's:

- **My decisions:** the scope and what's left out on purpose, hexagonal on the backend without a fake database, layered React with lint-enforced boundaries, exact decimals over floats, one endpoint with a strategy registry, string numbers in JSON, stable error codes, abuse limits as part of correctness, unit plus integration tests on both sides, one container, and the review checklist.
- **The AI's work:** writing the code, tests and docs from these prompts, probing the decimal library's edge cases, and running the reviews.
- **What I checked:** every phase's tests and coverage, the running container, real browser screenshots on desktop and mobile in both themes, and that every claim in the README matches the code.

## Prompt 1 — Scope, quality bar and how we'll work

You're helping me build a take-home for Sezzle: a full-stack calculator with a React + TypeScript frontend and a Go REST microservice. The budget is 2–4 hours, so correctness, clarity and maintainability beat features.

In scope: add, subtract, multiply, divide, plus power, square root and percentage. Out of scope, on purpose: auth, persistence, accounts and server-side history. A calculator is stateless, and adding a database would be scope creep, not engineering.

Rules for the whole project:
- Idiomatic Go and TypeScript.
- No code comments anywhere, generated code included. The README and the tests explain the code.
- Tests written before or together with the code, unit and integration on both sides, plus a coverage report.
- One command to run everything, needing only Docker.

We go phase by phase: domain, contract, backend adapters, frontend, packaging, adversarial review, docs. A phase starts only when the previous one is green.

## Prompt 2 — Backend architecture and domain

Build the backend with hexagonal architecture (ports and adapters):
- `internal/calculator` is the core. It holds the math rules and imports nothing from HTTP or any framework. It defines the interfaces it offers. The driving port is a `Calculator` interface with `Calculate(ctx, operation, operands)` and `Operations()`, implemented by a service. Interfaces live with the code that owns them, not in a shared "interfaces" package.
- Adapters live in `internal/adapter/...`. There's only a driving adapter (HTTP), because nothing is stored. Don't invent a repository or a database just to show the pattern: the boundary goes where a real dependency is.
- `cmd/api/main.go` is the composition root and the only file that knows every piece. Config comes from environment variables with safe defaults (`PORT`). The math limits are constants in the domain, because they're rules, not deployment knobs.

Decisions, and why:
- **Exact decimal arithmetic with `cockroachdb/apd/v3` at 34 significant digits, never float64.** `0.1 + 0.2` must be exactly `0.3`, because a calculator that's wrong at the 17th digit is wrong. If apd turns out unsuitable, fall back to `math/big`.
- **`Number` is a value object.** The only way to get one is a factory, `ParseNumber`, that enforces every input rule. Nothing downstream re-validates.
- **Each operation is a Strategy** (name, symbol, arity, apply) registered in one registry. Adding an operation means one new type and one line, with no change to the service or the HTTP layer (open/closed).
- **Percentage means "a percent of b"** (`percentage(15, 200) = 30`). It's an assumption to confirm with the product owner.
- **Typed domain errors with stable codes**, so clients never parse messages: `division_by_zero`, `negative_square_root`, `zero_to_negative_power`, `complex_result` (negative base with a fractional exponent), `result_too_large`, `result_too_small`, `invalid_number`, `wrong_operand_count`, `unknown_operation`.
- **Abuse limits are part of correctness.** Operands are at most 100 characters. Reject NaN, Infinity, hex, unicode digits and stray whitespace. Bound exponents so `10^999999999` can't eat the CPU.
- **Results are normalized:** no trailing zeros, plain notation for normal magnitudes, `-0` becomes `0`.
- **Calculations are pure,** so retries are harmless and no idempotency key is needed. Say so in the docs instead of adding one by reflex.

Tests: table-driven **unit tests** for the value object, every strategy and every error code, nasty inputs included, written before the implementation, run with `-race`.

## Prompt 3 — API contract

Write the OpenAPI 3.1 spec by hand before writing any handler, and embed it in the binary. It's the contract for both sides.

Decisions, and why:
- **Hand-written instead of swaggo.** swaggo builds the spec from code comments, and the code has none.
- **One endpoint, `POST /api/v1/calculations`, with `{operation, operands}`,** instead of one URL per operation. The contract stays uniform, and a new operation is data, not a route. Also `GET /api/v1/operations` (name, symbol, arity) and `GET /healthz`.
- **Operands and result are decimal strings.** JSON numbers become IEEE doubles in JavaScript and lose precision. The API also accepts JSON numbers for convenience, read with `json.Number` so nothing is lost on the way in.
- **Errors are RFC 9457 `application/problem+json` plus a `code` field.** 400 for malformed JSON or unknown fields, 413 for oversized bodies, 422 for valid requests that break a math or validation rule, 404 problem+json for unknown `/api` routes.
- **Every request and error has an example,** so Swagger UI at `/docs` explains itself. Swagger UI loads from a pinned CDN version, and that's a known limitation, since `/docs` needs internet.
- **The frontend generates its TypeScript types from this file** with a small script that prints them without comments, so generated code follows the same rule.

## Prompt 4 — Backend adapters, wiring and integration tests

Implement the HTTP driving adapter with the standard library router only:
- Handlers and DTO mapping.
- One function that maps domain error codes to HTTP statuses.
- Middleware: request id echoed in `X-Request-ID`, slog JSON logging and panic recovery.
- A 4 KB body limit and `DisallowUnknownFields`.
- Server read/write/idle timeouts.
- The built React app served from an embedded filesystem: SPA fallback for client routes, immutable cache headers for hashed assets, and a placeholder so `go build` works before the frontend exists.

Decisions, and why:
- **Standard library over Gin/Echo/Chi.** Go's router handles methods and path params, so a framework would only add a dependency.
- **Handlers depend on the `Calculator` port, not the concrete service,** so a test can force an internal error with a stub.
- **Graceful shutdown on SIGTERM and a `healthcheck` subcommand,** because the runtime image is distroless and has no curl.

Tests, two levels:
- **Unit tests** for DTO and error mapping.
- **Integration tests** that start the real router with `httptest.NewServer`, wired to the real domain, and hit it over HTTP: every operation, every error code with its status and `code`, malformed JSON, unknown fields, an oversized body, the wrong method, unknown API routes, the spec and docs endpoints, and the SPA fallback.

## Prompt 5 — Frontend architecture

Build the frontend in React + TypeScript with a layered, feature-first architecture that mirrors the backend:
- `src/features/calculator/domain`: plain TypeScript, no React. It holds the keypad state machine as a pure reducer, the operation types and the `CalculatorApi` port interface.
- `application`: hooks that run use cases against the port (TanStack Query `useMutation` for the API call) and feed results into the reducer. There are no use-case classes, because the hooks are the application layer and extra classes would only forward calls.
- `infrastructure`: the HTTP adapter that implements the port. It maps the generated OpenAPI DTOs to domain types, and problem+json codes to friendly messages.
- `ui`: a container that uses the hooks, and presentational components (display, keypad, tape) that only take props.
- `src/app` is the composition root. It injects the HTTP adapter through a tiny port-context factory, and tests inject a fake instead.
- Shared pieces live in `src/shared`: an HTTP client that turns problem+json into a typed error, and small UI parts.
- ESLint `no-restricted-imports` per folder enforces the direction: domain imports nothing outer, application never imports infrastructure or ui, and ui never touches HTTP. Use the built-in rule rather than a boundaries plugin, to keep `npm audit` clean. Pin TypeScript to a version that typescript-eslint and the type generator support.

Decisions, and why:
- **A pure reducer state machine for the keypad:** digit entry, one decimal point, sign toggle, backspace, clear, operator chaining (`2 + 3 ×` shows 5 and continues), equals, repeated equals and an error state that the next key clears. It's the trickiest logic in the UI, so it must be testable without rendering.
- **The backend does all the math.** The frontend never computes results, so there's one source of truth. It only checks input shape: max length and one decimal point.
- **No global state library.** A reducer plus a mutation is all the state there is.
- **Accessibility is part of done:** real buttons with aria labels, `aria-live` on the result, `aria-invalid`/`aria-describedby` for errors, and full keyboard support.
- **Mobile-first grid, and light and dark themes from CSS tokens.**

Tests: **unit tests** for every reducer transition, **hook tests** with a fake port, and **integration tests** with Testing Library that click and type real sequences (`1 . 5 + 2 =`, division by zero, chaining, keyboard only) against the fake adapter.

## Prompt 6 — Packaging, CI and coverage

Package it as one container:
- A multi-stage Dockerfile builds the frontend, embeds it in a static Go binary (`CGO_ENABLED=0`, `-trimpath`) and runs on distroless as non-root with a HEALTHCHECK. One process serves the UI, the API and the docs.
- docker-compose with a configurable `HOST_PORT`, started with `--wait` so the command returns only when the app is healthy.
- Makefile targets: `up`, `down`, `logs`, `dev`, `test`, `test-docker` (both suites in the official Go and Node images, with the whole repo mounted, so a reviewer needs only Docker) and `coverage`.
- Coverage: Go `-coverprofile` with `go tool cover`, and Vitest with v8. A `COVERAGE.md` with the real numbers.
- `.gitattributes` forces LF, so a Windows checkout doesn't break scripts.
- GitHub Actions runs the backend and frontend jobs in parallel and uploads the coverage HTML as artifacts. The image build runs only if both pass.

Decision: **a single container over separate frontend and nginx containers.** At this size, it's one artifact to run and version.

## Prompt 7 — Adversarial review and verification

Review the whole repo as a skeptical senior reviewer and try to break it:
- Precision: `0.1 + 0.2`, `1 / 3`, huge and tiny numbers, negative zero.
- Inputs: huge exponents, whitespace and unicode digits in operands, malformed JSON, the wrong content type.
- The UI: double clicks, fast typing, keyboard-only use, narrow screens, dark mode.

Then verify like a reviewer would: run the built container, hit every error code with curl, and check the UI in a real browser at desktop and phone width, in both themes.

For every finding, give a reproduction. Fix the real bugs test-first. Write down the limits we keep on purpose, with the reason, in the README.

## Prompt 8 — Documentation

Write the README in plain, human English, no hype, and keep every claim true to the code:
1. What it is.
2. Requirements per OS (macOS, Windows, Linux) with `brew`/`winget` commands. Only Git and Docker are needed to run and test it, and Go and Node only to develop without Docker.
3. A numbered quick start: clone, check Docker with `docker info`, `make up` (or the plain `docker compose` command), open the URLs, a two-minute "try it" tour, how to stop, and what to do if the port is taken, with PowerShell variants where the syntax differs.
4. How to run the backend and the frontend separately.
5. Tests and coverage, with Docker and with local toolchains.
6. API usage with curl examples for success and errors, and how to use Swagger.
7. Project layout.
8. Approach (how I worked, phase by phase).
9. Design decisions, and the patterns and principles actually used, with where.
10. Assumptions I'd confirm with the product owner.
11. An alternatives-considered table.
12. Deploying to AWS cheaply. It's stateless, so Lambda with Lambda Web Adapter behind CloudFront costs close to nothing at low traffic. Mention App Runner or Fargate as the step up.
13. Known limitations, written as deliberate limits with their reason.
