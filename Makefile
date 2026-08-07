.PHONY: all run install install-backend install-frontend run-backend run-frontend test clean help

.DEFAULT_GOAL := run

# File target for frontend node_modules (only reinstalls when package.json or package-lock.json changes)
frontend/node_modules: frontend/package.json frontend/package-lock.json
	@echo "==> Installing frontend dependencies (npm)..."
	cd frontend && npm install
	@touch frontend/node_modules

## install-backend: Download Go dependencies for backend
install-backend:
	@echo "==> Checking backend dependencies..."
	go mod download

## install-frontend: Install frontend dependencies if needed
install-frontend: frontend/node_modules

## install: Install dependencies for both backend and frontend if needed
install: install-backend install-frontend

## run-backend: Run backend Fiber server
run-backend:
	@echo "==> Starting backend server on http://localhost:9000..."
	go run main.go

## run-frontend: Run Next.js frontend dev server
run-frontend:
	@echo "==> Starting frontend dev server on http://localhost:3000..."
	cd frontend && npm run dev

## run: Check/install dependencies and run both backend and frontend concurrently
run: install
	@echo "==> Starting both backend and frontend servers concurrently..."
	@bash -c "trap 'trap - SIGINT SIGTERM EXIT; kill 0' SIGINT SIGTERM EXIT; \
		(go run main.go) & \
		(cd frontend && npm run dev) & \
		wait"

## test: Run backend tests
test:
	@echo "==> Running backend tests..."
	go test ./...

## help: Display available Makefile targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  run               Install dependencies (if missing/updated) and run both frontend and backend (Default)"
	@echo "  install           Check/install dependencies for both frontend and backend"
	@echo "  install-backend   Check/install Go backend dependencies"
	@echo "  install-frontend  Check/install npm frontend dependencies"
	@echo "  run-backend       Run backend server only"
	@echo "  run-frontend      Run frontend dev server only"
	@echo "  test              Run backend tests"
	@echo "  help              Show this help message"
