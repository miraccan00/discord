.PHONY: help backend-run frontend-dev backend-test backend-lint frontend-build helm-lint helm-template docker-build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

backend-run: ## Run the Go signaling server (:8080)
	cd backend && ALLOWED_ORIGINS=http://localhost:5173 go run ./cmd/server

frontend-dev: ## Run the Vite dev server (:5173)
	cd frontend && npm run dev

backend-test: ## Run backend tests with race detector
	cd backend && go test -race ./...

backend-lint: ## Lint the backend
	cd backend && golangci-lint run ./...

frontend-build: ## Typecheck, lint and build the frontend
	cd frontend && npm run typecheck && npm run lint && npm run build

helm-lint: ## Strict-lint the Helm chart
	helm lint deploy/helm/discord --strict

helm-template: ## Render the Helm chart
	helm template discord deploy/helm/discord

docker-build: ## Build both images locally
	docker build -t miraccan/discord-backend:dev ./backend
	docker build -t miraccan/discord-frontend:dev ./frontend
