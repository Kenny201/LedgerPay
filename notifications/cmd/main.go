package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kenny201/LedgerPay/notifications/internal/config"

	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	httpAddr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := &http.Server{
		Addr:              httpAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		fmt.Printf("%s: HTTP health на %s (gRPC порт %d)\n", cfg.ServiceName, httpAddr, cfg.GRPCPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		fmt.Printf("%s: завершение работы\n", cfg.ServiceName)
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cfg.ServiceName, err)
		os.Exit(1)
	}
}
