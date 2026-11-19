package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultAPI = "https://www.codewars.com/api/v1"

// Client reads the public Codewars API. It needs no token.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Retries int           // extra attempts after a network error, a 429 or a 5xx
	Backoff time.Duration // wait before the first retry, doubled after each one
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 20 * time.Second},
		Retries: 3,
		Backoff: 2 * time.Second,
	}
}

// User is the part of a Codewars profile that the card uses.
type User struct {
	Username string `json:"username"`
	Honor    int    `json:"honor"`
	Ranks    struct {
		Overall Rank `json:"overall"`
	} `json:"ranks"`
	CodeChallenges struct {
		TotalCompleted int `json:"totalCompleted"`
	} `json:"codeChallenges"`
}

type Rank struct {
	Name  string `json:"name"`  // "5 kyu"
	Color string `json:"color"` // "white", "yellow", "blue", "purple", "black" or "red"
}

// Kata is one entry of the completed katas list.
type Kata struct {
	Name        string    `json:"name"`
	Languages   []string  `json:"completedLanguages"`
	CompletedAt time.Time `json:"completedAt"`
}

func (c *Client) User(ctx context.Context, username string) (User, error) {
	var u User
	err := c.get(ctx, "/users/"+url.PathEscape(username), &u)
	return u, err
}

// Completed returns every kata the user has completed. The API sends them
// 200 per page, newest first.
func (c *Client) Completed(ctx context.Context, username string) ([]Kata, error) {
	var all []Kata
	for page := 0; ; page++ {
		var p struct {
			TotalPages int    `json:"totalPages"`
			Data       []Kata `json:"data"`
		}
		path := fmt.Sprintf("/users/%s/code-challenges/completed?page=%d", url.PathEscape(username), page)
		if err := c.get(ctx, path, &p); err != nil {
			return nil, err
		}
		all = append(all, p.Data...)
		if page+1 >= p.TotalPages || len(p.Data) == 0 {
			return all, nil
		}
	}
}

func (c *Client) get(ctx context.Context, path string, v any) error {
	wait := c.Backoff
	for attempt := 0; ; attempt++ {
		retry, err := c.try(ctx, path, v)
		if err == nil || !retry || attempt == c.Retries {
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w (gave up: %v)", err, ctx.Err())
		case <-time.After(wait):
		}
		wait *= 2
	}
}

// try makes one request and says whether a failure is worth retrying.
func (c *Client) try(ctx context.Context, path string, v any) (retry bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "codewars-card")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		retry := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return retry, fmt.Errorf("GET %s: %s", path, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return false, fmt.Errorf("GET %s: %w", path, err)
	}
	return false, nil
}
