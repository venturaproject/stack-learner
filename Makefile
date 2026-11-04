.PHONY: up down restart build logs ps \
        shell-backend shell-frontend shell-db \
        tidy swag test test-integration frontend-test lint seed console \
        clean-volumes arch-lint check setup config prod-up prod-down prod-build migrate help

DC      = docker compose -f compose.dev.yml
DC_PROD = docker compose -p boilerplate-learner-prod -f compose.yml

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Servicios ─────────────────────────────────────────────────────────────────

setup: ## Crea .env local desde .env.example si aún no existe
	@if [ ! -f .env ]; then cp .env.example .env && echo ".env creado para desarrollo local (cambia JWT_SECRET en producción)."; fi

config: setup ## Valida la configuración Docker Compose de desarrollo y producción
	$(DC) config --quiet
	$(DC_PROD) config --quiet

up: setup ## Construye y levanta el entorno de desarrollo
	$(DC) up -d --build

down: ## Detiene y elimina todos los contenedores
	$(DC) down

restart: ## Reinicia todos los servicios
	$(DC) restart

build: setup ## Reconstruye las imágenes de desarrollo
	$(DC) build

build-no-cache: setup ## Reconstruye las imágenes de desarrollo sin caché
	$(DC) build --no-cache

prod-build: setup ## Construye las imágenes de producción
	$(DC_PROD) build

prod-up: setup ## Levanta el stack con frontend estático de producción
	$(DC_PROD) up -d --build

prod-down: ## Detiene el stack de producción
	$(DC_PROD) down

logs: ## Logs en tiempo real (todos)
	$(DC) logs -f

logs-backend: ## Logs solo del backend Go
	$(DC) logs -f backend

ps: ## Estado de los contenedores
	$(DC) ps

# ── Shells ────────────────────────────────────────────────────────────────────

shell-backend: ## Shell en el contenedor backend
	$(DC) exec backend sh

shell-frontend: ## Shell en el contenedor frontend
	$(DC) exec frontend sh

shell-db: ## psql directo en postgres
	$(DC) exec postgres psql -U $${POSTGRES_USER:-postgres} -d $${POSTGRES_DB:-boilerplate_learner}

# ── Go ────────────────────────────────────────────────────────────────────────

tidy: ## go mod tidy dentro del contenedor
	$(DC) exec backend go mod tidy

swag: ## Genera documentación Swagger (requiere swag instalado)
	cd backend && ~/go/bin/swag init -g cmd/server/main.go --output docs --parseDependency --parseInternal

test: ## go test en el contenedor
	$(DC) exec backend go test ./...

test-integration: ## go test -tags=integration (crea la BD de test si falta)
	-$(DC) exec -T postgres psql -U postgres -c 'CREATE DATABASE boilerplate_learner_test'
	-$(DC) exec -T postgres psql -U postgres -c 'CREATE DATABASE boilerplate_learner_seed_test'
	$(DC) exec -e TEST_DATABASE_DSN="host=postgres port=5432 user=postgres password=postgres dbname=boilerplate_learner_test sslmode=disable" -e SEED_TEST_DATABASE_DSN="host=postgres port=5432 user=postgres password=postgres dbname=boilerplate_learner_seed_test sslmode=disable" backend go test -tags=integration ./...

lint: ## golangci-lint en el contenedor
	$(DC) exec backend golangci-lint run ./...
	$(MAKE) arch-lint

build-release: ## Compilar binario de producción localmente
	cd backend && CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/server ./cmd/server

# ── Frontend ──────────────────────────────────────────────────────────────────

frontend-test: ## Ejecuta las pruebas frontend en el contenedor
	$(DC) exec frontend pnpm test

install-frontend: ## pnpm install en el frontend
	$(DC) exec frontend pnpm install

tsc: ## TypeScript check en el frontend
	$(DC) exec frontend pnpm exec tsc --noEmit

# ── BD ────────────────────────────────────────────────────────────────────────

seed: ## Crea el usuario administrador de prueba
	$(DC) exec backend go run ./cmd/console seed

migrate: ## Aplica las migraciones pendientes
	$(DC) exec backend go run ./cmd/console migrate

console: ## Ejecuta un comando de consola (dev). Ej: make console CMD="seed --fresh"
	$(DC) exec backend console $(CMD)

vet: ## go vet en todos los paquetes
	$(DC) exec backend go vet ./...

staticcheck: ## Análisis estático avanzado (equivalente a clippy)
	$(DC) exec backend staticcheck ./...

arch-lint: ## Verifica dependencias entre capas Go con go-arch-lint
	docker run --rm -v "$(PWD)/backend:/app" fe3dback/go-arch-lint:release-v1.19.0 check --project-path /app

check: test vet staticcheck arch-lint frontend-test ## Tests y análisis estático/arquitectónico del proyecto

clean-volumes: ## Elimina contenedores y volúmenes ⚠️ borra datos
	$(DC) down -v
