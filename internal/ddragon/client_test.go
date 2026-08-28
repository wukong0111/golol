package ddragon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLatestVersionAndFetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/versions.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != UserAgent {
			t.Errorf("user-agent %s", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte(`["16.16.1","16.15.1"]`))
	})
	mux.HandleFunc("/cdn/16.16.1/data/es_ES/item.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"type":"item","version":"16.16.1","data":{}}`))
	})
	mux.HandleFunc("/cdn/16.16.1/data/es_ES/championFull.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"type":"champion","version":"16.16.1","data":{}}`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := &Client{BaseURL: ts.URL, HTTPClient: ts.Client()}
	ver, err := c.LatestVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ver != "16.16.1" {
		t.Fatalf("version %s", ver)
	}
	body, err := c.FetchItems(context.Background(), ver, "es_ES")
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("empty items body")
	}
	champs, err := c.FetchChampions(context.Background(), ver, "es_ES")
	if err != nil {
		t.Fatal(err)
	}
	if len(champs) == 0 {
		t.Fatal("empty champions body")
	}
}

func TestStoreUsesDiskCache(t *testing.T) {
	dir := t.TempDir()
	itemsPath := filepath.Join(dir, "16.16.1", "es_ES", ItemsFile)
	champsPath := filepath.Join(dir, "16.16.1", "es_ES", ChampionsFile)
	if err := os.MkdirAll(filepath.Dir(itemsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	itemsPayload := []byte(`{"type":"item","version":"16.16.1","data":{}}`)
	champsPayload := []byte(`{"type":"champion","version":"16.16.1","data":{}}`)
	if err := os.WriteFile(itemsPath, itemsPayload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(champsPath, champsPayload, 0o644); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/versions.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`["16.16.1"]`))
	})
	mux.HandleFunc("/cdn/", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not hit CDN when cache exists")
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	store := &Store{
		Client: &Client{BaseURL: ts.URL, HTTPClient: ts.Client()},
		Dir:    dir,
	}
	snap, err := store.Load(context.Background(), "es_ES")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Version != "16.16.1" || string(snap.Items) != string(itemsPayload) {
		t.Fatalf("items cache miss: %s %s", snap.Version, snap.Items)
	}
	if string(snap.Champions) != string(champsPayload) {
		t.Fatalf("champions cache miss: %s", snap.Champions)
	}
}

func TestStoreFallsBackWhenChampionsMissing(t *testing.T) {
	dir := t.TempDir()
	oldItems := filepath.Join(dir, "16.15.1", "es_ES", ItemsFile)
	oldChamps := filepath.Join(dir, "16.15.1", "es_ES", ChampionsFile)
	if err := os.MkdirAll(filepath.Dir(oldItems), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldItems, []byte(`{"type":"item","version":"16.15.1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldChamps, []byte(`{"type":"champion","version":"16.15.1"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/versions.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`["16.16.1"]`))
	})
	mux.HandleFunc("/cdn/16.16.1/data/es_ES/item.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"type":"item","version":"16.16.1"}`))
	})
	mux.HandleFunc("/cdn/16.16.1/data/es_ES/championFull.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	store := &Store{
		Client: &Client{BaseURL: ts.URL, HTTPClient: ts.Client()},
		Dir:    dir,
	}
	snap, err := store.Load(context.Background(), "es_ES")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Version != "16.15.1" {
		t.Fatalf("expected cached pair 16.15.1, got %s", snap.Version)
	}
}

func TestIconURL(t *testing.T) {
	got := IconURL(DefaultBaseURL, "16.16.1", "1029.png")
	want := "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/item/1029.png"
	if got != want {
		t.Fatalf("%s != %s", got, want)
	}
}

func TestChampionAssetURLs(t *testing.T) {
	icon := ChampionIconURL(DefaultBaseURL, "16.16.1", "Aatrox.png")
	if icon != "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/champion/Aatrox.png" {
		t.Fatalf("icon: %s", icon)
	}
	splash := SplashURL(DefaultBaseURL, "Aatrox")
	if splash != "https://ddragon.leagueoflegends.com/cdn/img/champion/splash/Aatrox_0.jpg" {
		t.Fatalf("splash: %s", splash)
	}
	spell := SpellIconURL(DefaultBaseURL, "16.16.1", "AatroxQ.png")
	if spell != "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/spell/AatroxQ.png" {
		t.Fatalf("spell: %s", spell)
	}
	passive := PassiveIconURL(DefaultBaseURL, "16.16.1", "Aatrox_Passive.png")
	if passive != "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/passive/Aatrox_Passive.png" {
		t.Fatalf("passive: %s", passive)
	}
}

func TestLoadMerakiUsesCache(t *testing.T) {
	dir := t.TempDir()
	payload := []byte(`{"Aatrox":{"key":"Aatrox"}}`)
	if err := os.WriteFile(filepath.Join(dir, MerakiFile), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	store := &Store{
		Client: &Client{HTTPClient: http.DefaultClient, MerakiURL: "http://127.0.0.1:1/nope"},
		Dir:    dir,
	}
	got, err := store.LoadMeraki(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("cache miss: %s", got)
	}
}

func TestLoadMerakiFetchesOnMiss(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Ahri":{"key":"Ahri"}}`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	dir := t.TempDir()
	store := &Store{
		Client: &Client{HTTPClient: ts.Client(), MerakiURL: ts.URL},
		Dir:    dir,
	}
	got, err := store.LoadMeraki(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(got) || string(got) != `{"Ahri":{"key":"Ahri"}}` {
		t.Fatalf("fetched: %s", got)
	}
	cached, err := os.ReadFile(filepath.Join(dir, MerakiFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(cached) != string(got) {
		t.Fatalf("not written: %s", cached)
	}
}
