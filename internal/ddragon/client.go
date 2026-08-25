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
)

// Client fetches static Data Dragon payloads. No Riot API key is required.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient returns a client pointed at the official CDN.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
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
	url := fmt.Sprintf("%s/cdn/%s/data/%s/item.json", c.BaseURL, version, locale)
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

// IconURL is the CDN path for an item icon.
func IconURL(baseURL, version, imageFull string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf("%s/cdn/%s/img/item/%s", base, version, imageFull)
}
