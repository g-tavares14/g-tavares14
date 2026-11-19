package main

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

// fixtureCard builds the card as it looked at 3 PM on Nov 19, 2026 in São Paulo.
func fixtureCard(t *testing.T) Card {
	t.Helper()
	var u User
	if err := json.Unmarshal(fixture(t, "user.json"), &u); err != nil {
		t.Fatal(err)
	}
	var page struct {
		Data []Kata `json:"data"`
	}
	if err := json.Unmarshal(fixture(t, "completed.json"), &page); err != nil {
		t.Fatal(err)
	}
	return NewCard(u, page.Data, time.Date(2026, 11, 19, 15, 0, 0, 0, saoPaulo(t)))
}

func saoPaulo(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestNewCardCountsKatasPerLocalDay(t *testing.T) {
	c := fixtureCard(t)
	if got := c.First.Format(time.DateOnly); got != "2026-10-21" {
		t.Fatalf("chart starts on %s, want 2026-10-21", got)
	}
	// The newest kata was completed at 01:53 UTC on Nov 19, which is still
	// Nov 18 in São Paulo, so Nov 19 has none.
	want := map[string]int{
		"2026-11-02": 11, "2026-11-03": 9, "2026-11-04": 3, "2026-11-05": 1,
		"2026-11-06": 7, "2026-11-07": 3, "2026-11-08": 1, "2026-11-09": 4,
		"2026-11-11": 8, "2026-11-12": 6, "2026-11-13": 3, "2026-11-15": 3,
		"2026-11-16": 2, "2026-11-17": 1, "2026-11-18": 1,
	}
	for i, n := range c.PerDay {
		day := c.First.AddDate(0, 0, i).Format(time.DateOnly)
		if n != want[day] {
			t.Errorf("%s: %d katas, want %d", day, n, want[day])
		}
	}
}

func TestNewCardTotals(t *testing.T) {
	c := fixtureCard(t)
	if c.Rank.Name != "5 kyu" || c.Solved != 63 || c.Honor != 306 {
		t.Errorf("rank %q, solved %d, honor %d", c.Rank.Name, c.Solved, c.Honor)
	}
	if !slices.Equal(c.Languages, []string{"go"}) {
		t.Errorf("languages %v, want [go]", c.Languages)
	}
	if got := c.RecentTotal(); got != 10 {
		t.Errorf("%d katas in the last %d days, want 10", got, Recent)
	}
	if day, n := c.Peak(); c.First.AddDate(0, 0, day).Format(time.DateOnly) != "2026-11-02" || n != 11 {
		t.Errorf("peak of %d on day %d, want 11 on Nov 2", n, day)
	}
}

func TestNewCardOrdersLanguagesByUse(t *testing.T) {
	at := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	katas := []Kata{
		{Languages: []string{"python", "go"}, CompletedAt: at},
		{Languages: []string{"go"}, CompletedAt: at},
		{Languages: []string{"sql"}, CompletedAt: at},
		{Languages: []string{"rust"}, CompletedAt: at},
	}
	c := NewCard(User{}, katas, at)
	if want := []string{"go", "python", "rust"}; !slices.Equal(c.Languages, want) {
		t.Errorf("languages %v, want %v", c.Languages, want)
	}
}

func TestPeakPicksTheLatestDayOnATie(t *testing.T) {
	var c Card
	c.PerDay[5], c.PerDay[9], c.PerDay[12] = 3, 3, 1
	if day, n := c.Peak(); day != 9 || n != 3 {
		t.Errorf("Peak() = %d, %d; want 9, 3", day, n)
	}
}

func TestDaysBetweenAcrossClockChanges(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		a, b time.Time
		want int
	}{
		{time.Date(2026, 3, 8, 0, 0, 0, 0, ny), time.Date(2026, 3, 9, 0, 0, 0, 0, ny), 1},    // 23 hours
		{time.Date(2026, 10, 31, 0, 0, 0, 0, ny), time.Date(2026, 11, 2, 0, 0, 0, 0, ny), 2}, // 49 hours
	}
	for _, tc := range cases {
		if got := daysBetween(tc.a, tc.b); got != tc.want {
			t.Errorf("daysBetween(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
