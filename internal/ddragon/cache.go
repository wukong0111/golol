package ddragon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Store loads Data Dragon JSON from disk when present, otherwise from the CDN.
type Store struct {
	Client *Client
	Dir    string
}

// Snapshot is one patch's item.json + championFull.json.
type Snapshot struct {
	Version   string
	Items     []byte
	Champions []byte
}

// Load returns the latest items and champions JSON for the same patch.
// On a CDN failure it falls back to the newest patch already cached on disk
// that has both files.
func (s *Store) Load(ctx context.Context, locale string) (Snapshot, error) {
	version, err := s.Client.LatestVersion(ctx)
	if err != nil {
		snap, cacheErr := s.latestCached(locale)
		if cacheErr != nil {
			return Snapshot{}, fmt.Errorf("ddragon version: %w (cache: %v)", err, cacheErr)
		}
		return snap, nil
	}

	items, err := s.loadPatch(ctx, version, locale, ItemsFile)
	if err != nil {
		snap, cacheErr := s.latestCached(locale)
		if cacheErr != nil {
			return Snapshot{}, fmt.Errorf("ddragon items %s: %w (cache: %v)", version, err, cacheErr)
		}
		return snap, nil
	}
	champions, err := s.loadPatch(ctx, version, locale, ChampionsFile)
	if err != nil {
		snap, cacheErr := s.latestCached(locale)
		if cacheErr != nil {
			return Snapshot{}, fmt.Errorf("ddragon champions %s: %w (cache: %v)", version, err, cacheErr)
		}
		return snap, nil
	}
	return Snapshot{Version: version, Items: items, Champions: champions}, nil
}

func (s *Store) loadPatch(ctx context.Context, version, locale, filename string) ([]byte, error) {
	path := s.filePath(version, locale, filename)
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 && json.Valid(data) {
		return data, nil
	}

	data, err := s.fetch(ctx, version, locale, filename)
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("%s for %s is not valid JSON", filename, version)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}
	return data, nil
}

func (s *Store) fetch(ctx context.Context, version, locale, filename string) ([]byte, error) {
	switch filename {
	case ItemsFile:
		return s.Client.FetchItems(ctx, version, locale)
	case ChampionsFile:
		return s.Client.FetchChampions(ctx, version, locale)
	default:
		return nil, fmt.Errorf("unknown ddragon file %s", filename)
	}
}

func (s *Store) latestCached(locale string) (Snapshot, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return Snapshot{}, err
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	for _, version := range versions {
		items, err := os.ReadFile(s.filePath(version, locale, ItemsFile))
		if err != nil || len(items) == 0 || !json.Valid(items) {
			continue
		}
		champions, err := os.ReadFile(s.filePath(version, locale, ChampionsFile))
		if err != nil || len(champions) == 0 || !json.Valid(champions) {
			continue
		}
		return Snapshot{Version: version, Items: items, Champions: champions}, nil
	}
	return Snapshot{}, fmt.Errorf("no cached item.json+championFull.json for %s", locale)
}

func (s *Store) filePath(version, locale, filename string) string {
	return filepath.Join(s.Dir, version, locale, filename)
}

func (s *Store) merakiFile(name string) string {
	return filepath.Join(s.Dir, name)
}

// LoadMeraki returns cached ability numbers, or downloads them on a cache miss.
func (s *Store) LoadMeraki(ctx context.Context) ([]byte, error) {
	return s.loadMerakiDump(ctx, MerakiFile, false, s.Client.FetchMeraki)
}

// RefreshMeraki downloads a fresh dump and replaces the cache.
func (s *Store) RefreshMeraki(ctx context.Context) ([]byte, error) {
	return s.loadMerakiDump(ctx, MerakiFile, true, s.Client.FetchMeraki)
}

// LoadMerakiItems returns cached shop classes, or downloads them on a cache miss.
func (s *Store) LoadMerakiItems(ctx context.Context) ([]byte, error) {
	return s.loadMerakiDump(ctx, MerakiItemsFile, false, s.Client.FetchMerakiItems)
}

// RefreshMerakiItems downloads a fresh item dump and replaces the cache.
func (s *Store) RefreshMerakiItems(ctx context.Context) ([]byte, error) {
	return s.loadMerakiDump(ctx, MerakiItemsFile, true, s.Client.FetchMerakiItems)
}

func (s *Store) loadMerakiDump(ctx context.Context, filename string, refresh bool, fetch func(context.Context) ([]byte, error)) ([]byte, error) {
	path := s.merakiFile(filename)
	if !refresh {
		if data, err := os.ReadFile(path); err == nil && len(data) > 0 && json.Valid(data) {
			return data, nil
		}
	}
	data, err := fetch(ctx)
	if err != nil {
		cached, cacheErr := os.ReadFile(path)
		if cacheErr == nil && len(cached) > 0 && json.Valid(cached) {
			return cached, nil
		}
		return nil, fmt.Errorf("meraki %s: %w", filename, err)
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("meraki %s is not valid JSON", filename)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}
	return data, nil
}
