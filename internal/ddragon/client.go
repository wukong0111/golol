package ddragon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://ddragon.leagueoflegends.com"
	UserAgent      = "golol/1.0"
	ItemsFile      = "item.json"
	ChampionsFile  = "championFull.json"
	MerakiFile     = "meraki-champions.json"
	MerakiURL      = "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/champions.json"
)

// Client fetches static Data Dragon payloads. No Riot API key is required.
type Client struct {
	BaseURL    string
	MerakiURL  string
	HTTPClient *http.Client
}

// NewClient returns a client pointed at the official CDN.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// LatestVersion returns the first entry of versions.json (newest patch).
func (c *Client) LatestVersion(ctx context.Context) (string, error) {
	body, err := c.get(ctx, c.BaseURL+"/api/versions.json")
	if err != nil {
		return "", err
	}
	var versions []string
	if err := json.Unmarshal(body, &versions); err != nil {
		return "", fmt.Errorf("decode versions.json: %w", err)
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("versions.json is empty")
	}
	return versions[0], nil
}

// FetchItems downloads item.json for a patch and locale.
func (c *Client) FetchItems(ctx context.Context, version, locale string) ([]byte, error) {
	url := fmt.Sprintf("%s/cdn/%s/data/%s/%s", c.BaseURL, version, locale, ItemsFile)
	return c.get(ctx, url)
}

// FetchChampions downloads championFull.json for a patch and locale.
func (c *Client) FetchChampions(ctx context.Context, version, locale string) ([]byte, error) {
	url := fmt.Sprintf("%s/cdn/%s/data/%s/%s", c.BaseURL, version, locale, ChampionsFile)
	return c.get(ctx, url)
}

// FetchMeraki downloads the community ability-number dump (rank tables, ratios).
func (c *Client) FetchMeraki(ctx context.Context) ([]byte, error) {
	url := MerakiURL
	if c != nil && strings.TrimSpace(c.MerakiURL) != "" {
		url = c.MerakiURL
	}
	return c.get(ctx, url)
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	res, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, res.Status)
	}
	return body, nil
}

func (c *Client) http() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func cdnBase(baseURL string) string {
	return strings.TrimRight(baseURL, "/")
}

// IconURL is the CDN path for an item icon.
func IconURL(baseURL, version, imageFull string) string {
	return fmt.Sprintf("%s/cdn/%s/img/item/%s", cdnBase(baseURL), version, imageFull)
}

// ChampionIconURL is the square portrait used in the grid.
func ChampionIconURL(baseURL, version, imageFull string) string {
	return fmt.Sprintf("%s/cdn/%s/img/champion/%s", cdnBase(baseURL), version, imageFull)
}

// SplashURL is the default-skin splash. It is not versioned on the CDN.
func SplashURL(baseURL, championID string) string {
	return fmt.Sprintf("%s/cdn/img/champion/splash/%s_0.jpg", cdnBase(baseURL), championID)
}

// SpellIconURL is the CDN path for a Q/W/E/R icon.
func SpellIconURL(baseURL, version, imageFull string) string {
	return fmt.Sprintf("%s/cdn/%s/img/spell/%s", cdnBase(baseURL), version, imageFull)
}

// PassiveIconURL is the CDN path for the innate ability icon.
func PassiveIconURL(baseURL, version, imageFull string) string {
	return fmt.Sprintf("%s/cdn/%s/img/passive/%s", cdnBase(baseURL), version, imageFull)
}
