package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// The fixtures are real API responses for g-tavares14, saved on Nov 19, 2026.
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// fakeAPI serves the fixtures the way Codewars does, with the completed katas
// split into pages of perPage.
func fakeAPI(t *testing.T, perPage int) *httptest.Server {
	t.Helper()
	user := fixture(t, "user.json")
	var completed struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(fixture(t, "completed.json"), &completed); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/g-tavares14", func(w http.ResponseWriter, r *http.Request) {
		w.Write(user)
	})
	mux.HandleFunc("GET /users/g-tavares14/code-challenges/completed", func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		all := completed.Data
		lo := min(page*perPage, len(all))
		hi := min(lo+perPage, len(all))
		json.NewEncoder(w).Encode(map[string]any{
			"totalPages": (len(all) + perPage - 1) / perPage,
			"totalItems": len(all),
			"data":       all[lo:hi],
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestClientUser(t *testing.T) {
	srv := fakeAPI(t, 200)
	u, err := NewClient(srv.URL).User(context.Background(), "g-tavares14")
	if err != nil {
		t.Fatal(err)
	}
	if u.Honor != 306 || u.CodeChallenges.TotalCompleted != 63 {
		t.Errorf("honor %d, completed %d; want 306 and 63", u.Honor, u.CodeChallenges.TotalCompleted)
	}
	if u.Ranks.Overall != (Rank{Name: "5 kyu", Color: "yellow"}) {
		t.Errorf("rank %+v, want 5 kyu, yellow", u.Ranks.Overall)
	}
}

func TestClientReadsEveryPage(t *testing.T) {
	srv := fakeAPI(t, 25) // 63 katas, so 3 pages
	katas, err := NewClient(srv.URL).Completed(context.Background(), "g-tavares14")
	if err != nil {
		t.Fatal(err)
	}
	if len(katas) != 63 {
		t.Fatalf("got %d katas, want 63", len(katas))
	}
	newest := time.Date(2026, 11, 19, 1, 53, 9, 336e6, time.UTC)
	if !katas[0].CompletedAt.Equal(newest) || katas[0].Name != "Is he gonna survive?" {
		t.Errorf("first kata %q at %v, want %q at %v", katas[0].Name, katas[0].CompletedAt, "Is he gonna survive?", newest)
	}
}

func TestClientRetriesWhenRateLimited(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"honor":306}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Backoff = time.Millisecond
	u, err := c.User(context.Background(), "g-tavares14")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 || u.Honor != 306 {
		t.Errorf("%d calls, honor %d; want 3 calls and 306", calls, u.Honor)
	}
}

func TestClientDoesNotRetryNotFound(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Backoff = time.Millisecond
	if _, err := c.User(context.Background(), "nobody"); err == nil {
		t.Error("want an error for a 404")
	}
	if calls != 1 {
		t.Errorf("%d calls, want 1", calls)
	}
}
