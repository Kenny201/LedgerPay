# contracts

Общие protobuf-контракты и сгенерированные Go-пакеты для сервисов LedgerPay.

## Планируемая структура

```text
contracts/
  auth/
  account/
  transaction/
  notifications/
  Makefile
```

После добавления `.proto` файлов генерация запускается командой `make proto` из корня репозитория.
