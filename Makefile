COMPOSE := docker-compose
SERVICES := auth account transaction notifications gateway

.PHONY: help up up-infra down logs ps build tidy vendor test vet fmt \
	run-auth run-account run-transaction run-notifications run-gateway \
	proto migrate

help:
	@echo "LedgerPay"
	@echo ""
	@echo "  make up-infra     - поднять postgres + redis + redpanda"
	@echo "  make vendor       - go mod vendor во всех сервисах"
	@echo "  make up           - инфраструктура + все сервисы"
	@echo "  make down         - остановить всё"
	@echo "  make logs         - логи"
	@echo "  make ps           - статус контейнеров"
	@echo "  make build        - собрать docker-образы сервисов"
	@echo "  make tidy         - go mod tidy во всех модулях"
	@echo "  make test         - go test во всех модулях"
	@echo "  make vet          - go vet во всех модулях"
	@echo "  make fmt          - gofmt во всех модулях"
	@echo "  make run-<svc>    - запустить сервис локально (нужен up-infra)"
	@echo "  make proto        - генерация protobuf"
	@echo "  make migrate      - прогон миграций БД"

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

migrate:
	@echo "добавьте миграции goose в migrations/ соответствующих сервисов"
