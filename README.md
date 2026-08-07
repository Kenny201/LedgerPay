# LedgerPay

Микросервисная платформа кошелька / ledger на Go.

## Сервисы

| Сервис | HTTP (health) | gRPC | Назначение |
|--------|---------------|------|------------|
| `gateway` | `:8080` | — | публичный HTTP API |
| `auth` | `:8081` | `:50051` | регистрация / JWT |
| `account` | `:8082` | `:50052` | счета / баланс |
| `transaction` | `:8083` | `:50053` | пополнение / перевод / outbox |
| `notifications` | `:8084` | `:50054` | consumer событий / уведомления |
| `contracts` | — | — | общий protobuf API |

Инфраструктура: **PostgreSQL**, **Redis**, **Redpanda** (Kafka API).

## Быстрый старт

```bash
cp .env.example .env
cp go.work.example go.work

make up-infra
make vendor
make up

curl http://localhost:8081/healthz
curl http://localhost:8080/healthz

make logs
make down
```

Запуск сервисов на хосте (инфраструктура в Docker):

```bash
make up-infra
make run-auth
make run-account
make run-transaction
make run-notifications
make run-gateway
```

## Модули

Каждый сервис — отдельный Go-модуль. Для локальной разработки в monorepo:

```bash
cp go.work.example go.work
```

Файлы `go.work` / `go.work.sum` в git не коммитим. В репозитории лежит только `go.work.example`.

```bash
make tidy
make vendor
make test
make vet
```

## Заметки

- `.env` в `.gitignore` — шаблон берите из `.env.example`.
- Образы сервисов собираются с `-mod=vendor`, поэтому Docker не зависит от доступа к `proxy.golang.org`.
- После добавления Go-зависимостей: `make tidy && make vendor`.
