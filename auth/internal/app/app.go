package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Kenny201/LedgerPay/auth/internal/config"
	"github.com/Kenny201/LedgerPay/auth/internal/handler"
	"github.com/Kenny201/LedgerPay/auth/internal/repository"
	"github.com/Kenny201/LedgerPay/auth/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

// App — composition root сервиса: DI + HTTP lifecycle.
type App struct {
	cfg  *config.Config
	pool *pgxpool.Pool
	auth *service.Auth
	http *http.Server
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

	repo := repository.NewUserPostgres(pool)
	a.pool = pool
	a.auth = service.NewAuth(repo, a.cfg.JWTSecret)
	a.http = &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.HTTPPort),
		Handler:           a.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return nil
}

func (a *App) routes() http.Handler {
	authHandler := handler.NewAuth(a.auth)

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
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	return mux
}

// Close освобождает ресурсы (пул БД).
func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
		a.pool = nil
	}
}
