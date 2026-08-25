package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golol/internal/ddragon"
	"golol/internal/httpx"
	"golol/internal/items"
)

func main() {
	addr := env("ADDR", ":8080")
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	locale := env("DDRAGON_LOCALE", "es_ES")
	cacheDir := env("CACHE_DIR", ".cache/ddragon")

	ctx := context.Background()
	store := &ddragon.Store{
		Client: ddragon.NewClient(),
		Dir:    cacheDir,
	}

	version, raw, err := store.Load(ctx, locale)
	if err != nil {
		log.Fatalf("cargar Data Dragon: %v", err)
	}
	cat, err := items.Parse(version, locale, ddragon.DefaultBaseURL, raw)
	if err != nil {
		log.Fatalf("parsear objetos: %v", err)
	}

	srv, err := httpx.New(cat)
	if err != nil {
		log.Fatalf("http: %v", err)
	}

	go refresh(store, srv, locale)

	log.Printf("golol en http://localhost%s — parche %s, %d objetos (%s)", addr, cat.Version, len(cat.Items), locale)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

func refresh(store *ddragon.Store, srv *httpx.Server, locale string) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		version, raw, err := store.Load(ctx, locale)
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

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
