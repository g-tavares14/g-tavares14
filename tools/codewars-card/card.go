package main

import (
	"cmp"
	"math"
	"slices"
	"time"
)

const (
	Days   = 30 // days in the chart, ending today
	Recent = 7  // days at the end of the chart that are highlighted
)

// Card holds what the SVG shows.
type Card struct {
	Rank      Rank
	Honor     int
	Solved    int
	Languages []string  // Codewars language ids, most used first, at most three
	PerDay    [Days]int // katas completed per day, oldest first; the last day is today
	First     time.Time // the day PerDay[0] counts
	Updated   time.Time
}

// NewCard counts the katas by the day they were completed, in now's time zone.
func NewCard(u User, katas []Kata, now time.Time) Card {
	today := startOfDay(now)
	c := Card{
		Rank:    u.Ranks.Overall,
		Honor:   u.Honor,
		Solved:  u.CodeChallenges.TotalCompleted,
		First:   today.AddDate(0, 0, 1-Days),
		Updated: now,
	}

	uses := map[string]int{}
	for _, k := range katas {
		for _, lang := range k.Languages {
			uses[lang]++
		}
		day := startOfDay(k.CompletedAt.In(now.Location()))
		if i := daysBetween(c.First, day); i >= 0 && i < Days {
			c.PerDay[i]++
		}
	}

	for lang := range uses {
		c.Languages = append(c.Languages, lang)
	}
	slices.SortFunc(c.Languages, func(a, b string) int {
		return cmp.Or(cmp.Compare(uses[b], uses[a]), cmp.Compare(a, b))
	})
	if len(c.Languages) > 3 {
		c.Languages = c.Languages[:3]
	}
	return c
}

// RecentTotal is how many katas were completed in the last Recent days.
func (c Card) RecentTotal() int {
	n := 0
	for _, v := range c.PerDay[Days-Recent:] {
		n += v
	}
	return n
}

// Peak returns the busiest day in the chart. On a tie it picks the latest
// day. With no katas at all, count is 0.
func (c Card) Peak() (day, count int) {
	for i, n := range c.PerDay {
		if n > 0 && n >= count {
			day, count = i, n
		}
	}
	return day, count
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// daysBetween counts calendar days from a to b, both at the start of a day.
// It rounds, so a 23 or 25 hour day around a clock change still counts as one.
func daysBetween(a, b time.Time) int {
	return int(math.Round(b.Sub(a).Hours() / 24))
}
