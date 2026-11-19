package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func wellFormed(t *testing.T, svg []byte) {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(svg))
	for {
		_, err := d.Token()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			t.Fatalf("SVG is not well-formed XML: %v", err)
		}
	}
}

func TestSVGShowsTheNumbers(t *testing.T) {
	svg := fixtureCard(t).SVG()
	wellFormed(t, svg)
	s := string(svg)

	for _, want := range []string{
		">5 kyu<", ">63<", ">306<", ">Go<", ">11<", ">Oct 21<",
		">10 katas in the last 7 days<",
		">Updated Nov 19, 2026 from the Codewars API<",
		"font/woff2;base64,",
		"prefers-reduced-motion",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("SVG is missing %q", want)
		}
	}
	// 15 days with katas, 5 of them in the last 7 days.
	if n := strings.Count(s, `class="bar"`); n != 15 {
		t.Errorf("%d bars, want 15", n)
	}
	if n := strings.Count(s, `fill="`+accent+`" style=`); n != 5 {
		t.Errorf("%d highlighted bars, want 5", n)
	}
}

func TestSVGWithoutKatas(t *testing.T) {
	now := time.Date(2026, 11, 19, 15, 0, 0, 0, time.UTC)
	svg := NewCard(User{}, nil, now).SVG()
	wellFormed(t, svg)
	s := string(svg)
	for _, want := range []string{">Unranked<", ">No katas in the last 30 days<"} {
		if !strings.Contains(s, want) {
			t.Errorf("SVG is missing %q", want)
		}
	}
	if strings.Contains(s, `class="bar"`) {
		t.Error("an empty chart has bars")
	}
}

func TestSVGEscapesText(t *testing.T) {
	c := fixtureCard(t)
	c.Rank.Name = `<b>"1 dan" & up</b>`
	svg := c.SVG()
	wellFormed(t, svg)
	if !strings.Contains(string(svg), "&lt;b&gt;&quot;1 dan&quot; &amp; up&lt;/b&gt;") {
		t.Error("rank name is not escaped")
	}
}

func TestFormatCount(t *testing.T) {
	for n, want := range map[int]string{
		0: "0", 306: "306", 1000: "1,000", 9999: "9,999", 10000: "10K",
		12345: "12.3K", 999999: "999.9K", 1234567: "1.2M",
	} {
		if got := formatCount(n); got != want {
			t.Errorf("formatCount(%d) = %q, want %q", n, got, want)
		}
	}
}
