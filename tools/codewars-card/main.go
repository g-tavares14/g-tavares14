// Command codewars-card draws the Codewars card shown in the profile README.
//
// It reads a user's public profile and completed katas from the Codewars API
// and writes an SVG with the rank, a few totals and a chart of katas per day.
// The codewars-card workflow runs it a few times a day and publishes the SVG
// to the output branch, which the README links to.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "time/tzdata" // so -tz works even where the system has no zone files
)

func main() {
	user := flag.String("user", "g-tavares14", "Codewars username")
	out := flag.String("out", "codewars.svg", "file to write the SVG to")
	zone := flag.String("tz", "America/Sao_Paulo", "time zone that decides which day a kata counts for")
	api := flag.String("api", defaultAPI, "base URL of the Codewars API")
	flag.Parse()

	if err := run(*user, *out, *zone, *api); err != nil {
		fmt.Fprintln(os.Stderr, "codewars-card:", err)
		os.Exit(1)
	}
}

func run(user, out, zone, api string) error {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := NewClient(api)
	profile, err := client.User(ctx, user)
	if err != nil {
		return err
	}
	katas, err := client.Completed(ctx, user)
	if err != nil {
		return err
	}

	card := NewCard(profile, katas, time.Now().In(loc))
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, card.SVG(), 0o644)
}
