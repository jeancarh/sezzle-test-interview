HOST_PORT ?= 8080
export HOST_PORT

GO_IMAGE = golang:1.26
NODE_IMAGE = node:24-alpine

.PHONY: up down logs dev test test-backend test-frontend test-docker coverage api-types

up:
	docker compose up --build -d --wait
	@echo "Calculator is running at http://localhost:$(HOST_PORT) (API docs at http://localhost:$(HOST_PORT)/docs)"

down:
	docker compose down

logs:
	docker compose logs -f calculator

dev:
	@cd frontend && npm install --no-audit --no-fund
	@trap 'kill 0' INT TERM; \
	(cd backend && go run ./cmd/api) & \
	(cd frontend && npm run dev) & \
	wait

test: test-backend test-frontend

test-backend:
	cd backend && go vet ./... && go test -race ./...

test-frontend:
	cd frontend && npm install --no-audit --no-fund && npm run typecheck && npm run lint && npm test

test-docker:
	docker run --rm -v "$(CURDIR):/repo" -v calculator-go-cache:/go/pkg/mod -w /repo/backend $(GO_IMAGE) \
		sh -c "go vet ./... && go test -race ./..."
	docker run --rm -v "$(CURDIR):/repo" -v /repo/frontend/node_modules -w /repo/frontend $(NODE_IMAGE) \
		sh -c "npm ci --no-audit --no-fund && npm run typecheck && npm run lint && npm test"

coverage:
	docker run --rm -v "$(CURDIR):/repo" -v calculator-go-cache:/go/pkg/mod -w /repo/backend $(GO_IMAGE) \
		sh -c "mkdir -p ../coverage && go test -coverprofile=../coverage/backend.out ./... > ../coverage/backend-packages.txt && go tool cover -html=../coverage/backend.out -o ../coverage/backend.html"
	docker run --rm -v "$(CURDIR):/repo" -v /repo/frontend/node_modules -w /repo/frontend $(NODE_IMAGE) \
		sh -c "npm ci --no-audit --no-fund && npx vitest run --coverage --coverage.reportsDirectory=../coverage/frontend && node scripts/coverage-report.mjs"
	@echo "Coverage written to COVERAGE.md, HTML reports in coverage/backend.html and coverage/frontend/index.html"

api-types:
	cd frontend && npm run gen:api
