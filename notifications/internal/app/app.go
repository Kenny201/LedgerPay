package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Kenny201/LedgerPay/notifications/internal/config"

	"golang.org/x/sync/errgroup"
)

// App — composition root сервиса notifications.
type App struct {
	cfg  *config.Config
	http *http.Server
}

func New(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run(ctx context.Context) error {
	a.http = &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.HTTPPort),
		Handler:           a.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		fmt.Printf("%s: HTTP на %s (gRPC порт %d)\n", a.cfg.ServiceName, a.http.Addr, a.cfg.GRPCPort)
		if err := a.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		fmt.Printf("%s: завершение работы\n", a.cfg.ServiceName)
		return a.http.Shutdown(shutdownCtx)
	})

	return g.Wait()
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	return mux
}
