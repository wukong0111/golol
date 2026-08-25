package ddragon

import (
	"context"
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
		t.Fatal("empty body")
	}
}

func TestStoreUsesDiskCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "16.16.1", "es_ES", "item.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"type":"item","version":"16.16.1","data":{}}`)
	if err := os.WriteFile(path, payload, 0o644); err != nil {
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
	ver, raw, err := store.Load(context.Background(), "es_ES")
	if err != nil {
		t.Fatal(err)
	}
	if ver != "16.16.1" || string(raw) != string(payload) {
		t.Fatalf("cache miss: %s %s", ver, raw)
	}
}

func TestIconURL(t *testing.T) {
	got := IconURL(DefaultBaseURL, "16.16.1", "1029.png")
	want := "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/item/1029.png"
	if got != want {
		t.Fatalf("%s != %s", got, want)
	}
}
