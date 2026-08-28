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

	"golol/internal/champions"
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

	snap, err := store.Load(ctx, locale)
	if err != nil {
		return err
	}
	cat, err := items.Parse(snap.Version, locale, ddragon.DefaultBaseURL, snap.Items)
	if err != nil {
		return err
	}
	champs, err := champions.Parse(snap.Version, locale, ddragon.DefaultBaseURL, snap.Champions)
	if err != nil {
		return err
	}
	if kits, kitErr := store.LoadMeraki(ctx); kitErr != nil {
		log.Printf("kits de habilidades: %v", kitErr)
	} else if err := champs.ApplyKits(kits); err != nil {
		log.Printf("kits de habilidades: %v", err)
	}

	app, err := httpx.New(cat, champs)
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
		log.Printf("golol en http://localhost%s — parche %s, %d objetos, %d campeones (%s)", addr, cat.Version, len(cat.Items), len(champs.Champions), locale)
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
			loadCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			snap, err := store.Load(loadCtx, locale)
			if err != nil {
				cancel()
				log.Printf("refresh Data Dragon: %v", err)
				continue
			}
			if cat, err := items.Parse(snap.Version, locale, ddragon.DefaultBaseURL, snap.Items); err != nil {
				log.Printf("refresh parse items: %v", err)
			} else {
				srv.SetCatalog(cat)
				log.Printf("catálogo de objetos actualizado: parche %s (%d objetos)", cat.Version, len(cat.Items))
			}
			if cat, err := champions.Parse(snap.Version, locale, ddragon.DefaultBaseURL, snap.Champions); err != nil {
				log.Printf("refresh parse champions: %v", err)
			} else {
				if kits, kitErr := store.RefreshMeraki(loadCtx); kitErr != nil {
					log.Printf("refresh kits de habilidades: %v", kitErr)
				} else if err := cat.ApplyKits(kits); err != nil {
					log.Printf("refresh kits de habilidades: %v", err)
				}
				srv.SetChampions(cat)
				log.Printf("catálogo de campeones actualizado: parche %s (%d campeones)", cat.Version, len(cat.Champions))
			}
			cancel()
		}
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
