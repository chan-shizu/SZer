
.PHONY: help migrate-up migrate-up1 migrate-down1 migrate-down

# docker compose wrapper
COMPOSE ?= docker compose
BACKEND_SERVICE ?= backend

# dev DB URL (from memo.md)
DEV_DATABASE_URL ?= postgres://dev_user:dev_password@postgres:5432/szer_dev?sslmode=disable

# migrate runs inside the backend container. WORKDIR is /app, so db/migrations is valid.
MIGRATIONS_PATH ?= db/migrations

# Optional flags for `docker compose exec` (e.g. DOCKER_EXEC_FLAGS=-T for CI)
DOCKER_EXEC_FLAGS ?=

help:
	@echo "Targets:"
	@echo "  migrate-up     Run all migrations (inside backend container)"
	@echo "  migrate-up1    Run 1 migration step (inside backend container)"
	@echo "  migrate-down1  Rollback 1 migration step (inside backend container)"
	@echo "  migrate-down   Rollback all migrations (inside backend container)"
	@echo ""
	@echo "Variables:"
	@echo "  DEV_DATABASE_URL=... (override DB url)"
	@echo "  MIGRATIONS_PATH=... (default: db/migrations)"
	@echo "  DOCKER_EXEC_FLAGS=... (e.g. -T)"

# migrate
migrate-up:
	$(COMPOSE) exec $(DOCKER_EXEC_FLAGS) $(BACKEND_SERVICE) migrate -path $(MIGRATIONS_PATH) -database "$(DEV_DATABASE_URL)" up

migrate-up1:
	$(COMPOSE) exec $(DOCKER_EXEC_FLAGS) $(BACKEND_SERVICE) migrate -path $(MIGRATIONS_PATH) -database "$(DEV_DATABASE_URL)" up 1

migrate-down1:
	$(COMPOSE) exec $(DOCKER_EXEC_FLAGS) $(BACKEND_SERVICE) migrate -path $(MIGRATIONS_PATH) -database "$(DEV_DATABASE_URL)" down 1

migrate-down:
	$(COMPOSE) exec $(DOCKER_EXEC_FLAGS) $(BACKEND_SERVICE) migrate -path $(MIGRATIONS_PATH) -database "$(DEV_DATABASE_URL)" down

# sqlc
sqlc:
	$(COMPOSE) exec $(DOCKER_EXEC_FLAGS) $(BACKEND_SERVICE) sqlc generate

# cmd
seed:
	$(COMPOSE) exec $(DOCKER_EXEC_FLAGS) $(BACKEND_SERVICE) go run ./cmd/seed/main.go
