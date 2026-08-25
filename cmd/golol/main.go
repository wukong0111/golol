package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golol/internal/ddragon"
	"golol/internal/httpx"
	"golol/internal/items"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	locale := env("DDRAGON_LOCALE", "es_ES")
	cacheDir := env("CACHE_DIR", ".cache/ddragon")
	addr := listenAddr()

	store := &ddragon.Store{
		Client: ddragon.NewClient(),
		Dir:    cacheDir,
	}

	version, raw, err := store.Load(ctx, locale)
	if err != nil {
		return err
	}
	cat, err := items.Parse(version, locale, ddragon.DefaultBaseURL, raw)
	if err != nil {
		return err
	}

	app, err := httpx.New(cat)
	if err != nil {
		return err
	}

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go refresh(ctx, store, app, locale)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("golol en http://localhost%s — parche %s, %d objetos (%s)", addr, cat.Version, len(cat.Items), locale)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			stop()
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Printf("señal de apagado, drenando conexiones…")
	}

	app.SetReady(false)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout())
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		_ = httpSrv.Close()
		return err
	}
	log.Printf("apagado completo")
	return <-errCh
}

func refresh(ctx context.Context, store *ddragon.Store, srv *httpx.Server, locale string) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			loadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			version, raw, err := store.Load(loadCtx, locale)
			cancel()
			if err != nil {
				log.Printf("refresh Data Dragon: %v", err)
				continue
			}
			cat, err := items.Parse(version, locale, ddragon.DefaultBaseURL, raw)
			if err != nil {
				log.Printf("refresh parse: %v", err)
				continue
			}
			srv.SetCatalog(cat)
			log.Printf("catálogo actualizado: parche %s (%d objetos)", cat.Version, len(cat.Items))
		}
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
