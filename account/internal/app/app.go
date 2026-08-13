package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Kenny201/LedgerPay/account/internal/config"
	"github.com/Kenny201/LedgerPay/account/internal/handler"
	"github.com/Kenny201/LedgerPay/account/internal/repository"
	"github.com/Kenny201/LedgerPay/account/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

// App — composition root сервиса: DI + HTTP lifecycle.
type App struct {
	cfg      *config.Config
	pool     *pgxpool.Pool
	accounts *service.Account
	http     *http.Server
}

func New(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

// Run поднимает зависимости и блокируется до отмены ctx или ошибки.
func (a *App) Run(ctx context.Context) error {
	if err := a.init(ctx); err != nil {
		return err
	}
	defer a.Close()

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

func (a *App) init(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, a.cfg.DBDSN)
	if err != nil {
		return fmt.Errorf("подключение к БД: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("ping БД: %w", err)
	}

	a.pool = pool
	a.accounts = service.NewAccount(repository.NewAccountPostgres(pool))
	a.http = &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.HTTPPort),
		Handler:           a.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return nil
}

func (a *App) routes() http.Handler {
	h := handler.NewAccount(a.accounts)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := a.pool.Ping(r.Context()); err != nil {
			http.Error(w, "db not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.Create(w, r)
		case http.MethodGet:
			h.Get(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
		a.pool = nil
	}
}
