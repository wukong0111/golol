package ddragon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Store loads item.json from disk when present, otherwise from the CDN.
type Store struct {
	Client *Client
	Dir    string
}

// Load returns the latest catalog JSON. On a CDN failure it falls back to the
// newest patch already cached on disk, if any.
func (s *Store) Load(ctx context.Context, locale string) (version string, raw []byte, err error) {
	version, err = s.Client.LatestVersion(ctx)
	if err != nil {
		cachedVersion, cached, cacheErr := s.latestCached(locale)
		if cacheErr != nil {
			return "", nil, fmt.Errorf("ddragon version: %w (cache: %v)", err, cacheErr)
		}
		return cachedVersion, cached, nil
	}

	raw, err = s.loadPatch(ctx, version, locale)
	if err != nil {
		cachedVersion, cached, cacheErr := s.latestCached(locale)
		if cacheErr != nil {
			return "", nil, fmt.Errorf("ddragon items %s: %w (cache: %v)", version, err, cacheErr)
		}
		return cachedVersion, cached, nil
	}
	return version, raw, nil
}

func (s *Store) loadPatch(ctx context.Context, version, locale string) ([]byte, error) {
	path := s.itemPath(version, locale)
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 && json.Valid(data) {
		return data, nil
	}

	data, err := s.Client.FetchItems(ctx, version, locale)
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("item.json for %s is not valid JSON", version)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}
	return data, nil
}

func (s *Store) latestCached(locale string) (string, []byte, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return "", nil, err
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	for _, version := range versions {
		data, err := os.ReadFile(s.itemPath(version, locale))
		if err == nil && len(data) > 0 && json.Valid(data) {
			return version, data, nil
		}
	}
	return "", nil, fmt.Errorf("no cached item.json for %s", locale)
}

func (s *Store) itemPath(version, locale string) string {
	return filepath.Join(s.Dir, version, locale, "item.json")
}
