# codewars-card

Draws the Codewars card in the profile README from the public Codewars API:
rank, katas solved, honor and katas per day over the last 30 days.

The [codewars-card workflow](../../.github/workflows/codewars-card.yml) runs it
every 6 hours and pushes `codewars.svg` to the `output` branch, which the
README links to. On a pull request it draws the card without publishing it.
If the API fails, the run fails and the last card stays up.

To draw the card locally:

```sh
go run . -out codewars.svg
```

| Flag | Default | |
|---|---|---|
| `-user` | `g-tavares14` | Codewars username |
| `-tz` | `America/Sao_Paulo` | time zone that decides which day a kata counts for |
| `-out` | `codewars.svg` | where to write the SVG |
| `-api` | `https://www.codewars.com/api/v1` | API base URL |

It uses only the standard library. The tests run against real API responses
saved in `testdata/`, so `go test ./...` needs no network.

`fonts/` has Latin subsets of Inter and JetBrains Mono, under the SIL Open Font
License ([OFL.txt](fonts/OFL.txt)). The SVG carries them because GitHub shows
README images through a proxy, and an SVG loaded as an image can't fetch fonts.
