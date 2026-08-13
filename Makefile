COMPOSE := docker-compose
SERVICES := auth account transaction notifications gateway
MIGRATE_SERVICES := auth account transaction notifications

# Подхватываем DB_DSN из .env, если есть; иначе дефолт для local.
ifneq (,$(wildcard .env))
  include .env
  export
endif
DB_DSN ?= postgres://ledger:ledger@localhost:5432/ledgerpay?sslmode=disable
GOOSE ?= goose

.PHONY: help up up-infra down logs ps build tidy vendor test vet fmt \
	run-auth run-account run-transaction run-notifications run-gateway \
	proto migrate migrate-auth migrate-account migrate-transaction migrate-notifications \
	migrate-status migrate-down

help:
	@echo "LedgerPay"
	@echo ""
	@echo "  make up-infra              - поднять postgres + redis + redpanda"
	@echo "  make vendor                - go mod vendor во всех сервисах"
	@echo "  make up                    - инфраструктура + все сервисы"
	@echo "  make down                  - остановить всё"
	@echo "  make logs                  - логи"
	@echo "  make ps                    - статус контейнеров"
	@echo "  make build                 - собрать docker-образы сервисов"
	@echo "  make tidy                  - go mod tidy во всех модулях"
	@echo "  make test                  - go test во всех модулях"
	@echo "  make vet                   - go vet во всех модулях"
	@echo "  make fmt                   - gofmt во всех модулях"
	@echo "  make run-<svc>             - запустить сервис локально (нужен up-infra)"
	@echo "  make proto                 - генерация protobuf"
	@echo "  make migrate               - прогон миграций всех сервисов (порядок FK)"
	@echo "  make migrate-<svc>         - миграции одного сервиса (auth|account|transaction|notifications)"
	@echo "  make migrate-status        - статус миграций по сервисам"
	@echo "  make migrate-down SVC=auth - откатить последнюю миграцию сервиса"

up-infra:
	$(COMPOSE) up -d postgres redis redpanda

up: vendor
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f --tail=200

ps:
	$(COMPOSE) ps

build: vendor
	$(COMPOSE) build

tidy:
	@for s in contracts $(SERVICES); do \
		echo "==> tidy $$s"; \
		(cd $$s && GOWORK=off go mod tidy); \
	done

vendor:
	@for s in $(SERVICES); do \
		echo "==> vendor $$s"; \
		(cd $$s && GOWORK=off go mod vendor); \
	done

test:
	@for s in contracts $(SERVICES); do \
		echo "==> test $$s"; \
		(cd $$s && go test ./...); \
	done

vet:
	@for s in contracts $(SERVICES); do \
		echo "==> vet $$s"; \
		(cd $$s && go vet ./...); \
	done

fmt:
	@for s in contracts $(SERVICES); do \
		echo "==> fmt $$s"; \
		(cd $$s && gofmt -s -w .); \
	done

run-auth:
	cd auth && go run ./cmd

run-account:
	cd account && go run ./cmd

run-transaction:
	cd transaction && go run ./cmd

run-notifications:
	cd notifications && go run ./cmd

run-gateway:
	cd gateway && go run ./cmd

proto:
	@echo "добавьте .proto в contracts/, затем подключите генерацию здесь"

migrate-auth:
	@echo "==> migrate auth"
	$(GOOSE) -table goose_auth -dir ./auth/migrations postgres "$(DB_DSN)" up

migrate-account:
	@echo "==> migrate account"
	$(GOOSE) -table goose_account -dir ./account/migrations postgres "$(DB_DSN)" up

migrate-transaction:
	@echo "==> migrate transaction"
	$(GOOSE) -table goose_transaction -dir ./transaction/migrations postgres "$(DB_DSN)" up

migrate-notifications:
	@echo "==> migrate notifications"
	$(GOOSE) -table goose_notifications -dir ./notifications/migrations postgres "$(DB_DSN)" up

# Порядок важен из‑за FK: users → accounts → transfers → …
migrate: migrate-auth migrate-account migrate-transaction migrate-notifications

migrate-status:
	@for s in $(MIGRATE_SERVICES); do \
		echo "==> status $$s"; \
		$(GOOSE) -table goose_$$s -dir ./$$s/migrations postgres "$(DB_DSN)" status; \
	done

# Пример: make migrate-down SVC=auth
migrate-down:
	@test -n "$(SVC)" || (echo "usage: make migrate-down SVC=auth"; exit 1)
	@echo "==> down $(SVC)"
	$(GOOSE) -table goose_$(SVC) -dir ./$(SVC)/migrations postgres "$(DB_DSN)" down
