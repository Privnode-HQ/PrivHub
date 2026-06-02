FRONTEND_DIR = ./web
BACKEND_DIR = .
USER_DOCS_DIR = ./user-docs

.PHONY: all build-frontend build-user-docs start-backend

all: build-user-docs build-frontend start-backend

build-user-docs:
	@echo "Building user docs..."
	@cd $(USER_DOCS_DIR) && pnpm install --frozen-lockfile && pnpm run build

build-frontend:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && bun install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) bun run build

start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go &
