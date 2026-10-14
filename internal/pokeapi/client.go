// Package pokeapi provides bounded, cacheable access to PokéAPI.
package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	pathpkg "path"
	"strings"
	"time"
)

const BaseURL = "https://pokeapi.co/api/v2/"
const maxResponseBytes = 4 << 20

type cache interface {
	Add(string, []byte)
	Get(string) ([]byte, bool)
}

type Client struct {
	base  *url.URL
	http  *http.Client
	cache cache
}

type Resource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Locations struct {
	Next     *string    `json:"next"`
	Previous *string    `json:"previous"`
	Results  []Resource `json:"results"`
}
type Area struct {
	PokemonEncounters []struct {
		Pokemon Resource `json:"pokemon"`
	} `json:"pokemon_encounters"`
}
type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int      `json:"base_stat"`
		Stat     Resource `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type Resource `json:"type"`
	} `json:"types"`
}

func NewClient(baseURL string, timeout time.Duration) (*Client, error) {
	base, err := url.Parse(baseURL)
	if err != nil || base.Host == "" || (base.Scheme != "https" && base.Scheme != "http") || !strings.HasSuffix(base.Path, "/") || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, fmt.Errorf("invalid API base URL")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}
	c := &Client{base: base, http: &http.Client{Timeout: timeout}}
	c.http.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return fmt.Errorf("API redirects are not supported")
	}
	return c, nil
}

func (c *Client) SetCache(store cache) { c.cache = store }

func (c *Client) get(ctx context.Context, path string, dest any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(path) > 4096 {
		return fmt.Errorf("API URL exceeds 4096 bytes")
	}
	ref, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}
	target := c.base.ResolveReference(ref)
	if target.Scheme != c.base.Scheme || target.Host != c.base.Host || target.User != nil || target.Fragment != "" || !strings.HasPrefix(pathpkg.Clean(target.Path)+"/", c.base.Path) {
		return fmt.Errorf("API URL outside configured origin or path")
	}
	key := target.String()
	if c.cache != nil {
		if data, ok := c.cache.Get(key); ok {
			if err := json.Unmarshal(data, dest); err != nil {
				return err
			}
			return validateResponse(dest)
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, key, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "dexrove/1.1")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request PokéAPI: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PokéAPI returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read PokéAPI: %w", err)
	}
	if len(data) > maxResponseBytes {
		return fmt.Errorf("PokéAPI response exceeds 4 MiB")
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("decode PokéAPI: %w", err)
	}
	if err := validateResponse(dest); err != nil {
		return err
	}
	if c.cache != nil {
		c.cache.Add(key, data)
	}
	return nil
}

func (c *Client) ListLocations(ctx context.Context, page *string) (Locations, error) {
	path := "location-area/?limit=20"
	if page != nil {
		path = *page
	}
	var result Locations
	err := c.get(ctx, path, &result)
	return result, err
}
func (c *Client) Explore(ctx context.Context, name string) (Area, error) {
	var result Area
	if !ValidName(name) {
		return result, fmt.Errorf("invalid area name")
	}
	err := c.get(ctx, "location-area/"+url.PathEscape(name)+"/", &result)
	return result, err
}
func (c *Client) GetPokemon(ctx context.Context, name string) (Pokemon, error) {
	var result Pokemon
	if !ValidName(name) {
		return result, fmt.Errorf("invalid Pokemon name")
	}
	err := c.get(ctx, "pokemon/"+url.PathEscape(name)+"/", &result)
	return result, err
}

// ValidName accepts PokéAPI identifiers and rejects paths and terminal controls.
func ValidName(name string) bool {
	if len(name) == 0 || len(name) > 100 || strings.HasPrefix(name, "-") {
		return false
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}

// ValidatePokemon verifies fields used in terminal output and saved collections.
func ValidatePokemon(p Pokemon) error {
	if !ValidName(p.Name) || p.Height < 0 || p.Weight < 0 || p.BaseExperience < 0 {
		return fmt.Errorf("invalid Pokemon fields")
	}
	for _, s := range p.Stats {
		if !ValidName(s.Stat.Name) || s.BaseStat < 0 {
			return fmt.Errorf("invalid Pokemon stat")
		}
	}
	for _, t := range p.Types {
		if !ValidName(t.Type.Name) {
			return fmt.Errorf("invalid Pokemon type")
		}
	}
	return nil
}

func validateResponse(dest any) error {
	switch value := dest.(type) {
	case *Pokemon:
		return ValidatePokemon(*value)
	case *Locations:
		for _, area := range value.Results {
			if !ValidName(area.Name) {
				return fmt.Errorf("invalid area name in API response")
			}
		}
	case *Area:
		for _, encounter := range value.PokemonEncounters {
			if !ValidName(encounter.Pokemon.Name) {
				return fmt.Errorf("invalid Pokemon name in API response")
			}
		}
	}
	return nil
}
