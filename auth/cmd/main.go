package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kenny201/LedgerPay/auth/internal/app"
	"github.com/Kenny201/LedgerPay/auth/internal/config"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.New(cfg).Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cfg.ServiceName, err)
		os.Exit(1)
	}
}
