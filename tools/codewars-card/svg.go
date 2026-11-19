package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Colors and fonts match the other images in the README.
const (
	surface  = "#1B1E25"
	border   = "#2C3039"
	ink      = "#EEECE8"
	body     = "#C9CCD1"
	muted    = "#9AA0A9"
	faint    = "#454B56"
	accent   = "#FFB224"
	green    = "#3DD68C"
	chipFill = "#252932"
	chipInk  = "#D8D5CF"
	older    = "#5D6470" // bars before the highlighted days
)

// rankColors follows the colors Codewars gives its ranks, adjusted to read
// on a dark card.
var rankColors = map[string]string{
	"white":  "#E4E2DD",
	"yellow": accent,
	"blue":   "#5AA2F0",
	"purple": "#A98BF0",
	"black":  "#C9CCD1",
	"red":    "#F26D5B",
}

var languageNames = map[string]string{
	"c": "C", "clojure": "Clojure", "coffeescript": "CoffeeScript", "cpp": "C++",
	"csharp": "C#", "dart": "Dart", "elixir": "Elixir", "fsharp": "F#", "go": "Go",
	"haskell": "Haskell", "java": "Java", "javascript": "JavaScript", "kotlin": "Kotlin",
	"lua": "Lua", "php": "PHP", "python": "Python", "r": "R", "ruby": "Ruby",
	"rust": "Rust", "scala": "Scala", "shell": "Shell", "sql": "SQL", "swift": "Swift",
	"typescript": "TypeScript",
}

// Latin subsets of Inter and JetBrains Mono (SIL Open Font License, see
// fonts/OFL.txt). The SVG carries them because GitHub shows README images
// through a proxy, and an SVG loaded as an image can't fetch fonts.
var (
	//go:embed fonts/inter-400.woff2
	inter400 []byte
	//go:embed fonts/inter-700.woff2
	inter700 []byte
	//go:embed fonts/jetbrains-mono-700.woff2
	mono700 []byte
)

// Layout, in pixels.
const (
	width, height = 860, 176

	chartLeft, chartRight = 440.0, 836.0
	chartTop, chartBase   = 56.0, 128.0 // top of the tallest bar, and the baseline
	barWidth              = 9.0
	bottomRow             = 153.0 // baseline of the text under the chart
	hexRadius             = 19.0
)

// SVG draws the card.
func (c Card) SVG() []byte {
	var s canvas
	s.printf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-labelledby="t d">`,
		width, height, width, height)
	s.printf(`<title id="t">Codewars</title><desc id="d">%s</desc>`, escape(c.Summary()))
	s.printf(`<style>%s%s</style>`, fontFaces(), css())
	s.printf(`<rect x=".5" y=".5" width="%d" height="%d" rx="13.5" fill="%s" stroke="%s"/>`, width-1, height-1, surface, border)

	c.drawHeading(&s)
	c.drawStats(&s)
	c.drawChart(&s)
	s.text(24, bottomRow, "sans", 11, 400, muted, "", "Updated "+c.Updated.Format("Jan 2, 2006")+" from the Codewars API")

	s.printf(`</svg>`)
	return []byte(s.String())
}

// Summary is the card in words, for screen readers.
func (c Card) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Rank %s, %s katas solved, %s honor.", c.rankName(), formatCount(c.Solved), formatCount(c.Honor))
	if len(c.Languages) > 0 {
		names := make([]string, len(c.Languages))
		for i, id := range c.Languages {
			names[i] = languageName(id)
		}
		fmt.Fprintf(&b, " Solving in %s.", strings.Join(names, ", "))
	}
	last := c.First.AddDate(0, 0, Days-1)
	fmt.Fprintf(&b, " Katas per day from %s to %s: %s", c.First.Format("Jan 2"), last.Format("Jan 2, 2006"), c.recentLabel())
	if day, n := c.Peak(); n > 0 {
		fmt.Fprintf(&b, ", and %d on the busiest day, %s", n, c.First.AddDate(0, 0, day).Format("Jan 2"))
	}
	fmt.Fprintf(&b, ". Updated %s from the Codewars API.", c.Updated.Format("Jan 2, 2006"))
	return b.String()
}

// drawHeading draws the top row: the label, one chip per language and the
// link arrow used on the project cards.
func (c Card) drawHeading(s *canvas) {
	s.printf(`<circle class="dot" cx="24" cy="27" r="3.6" fill="%s"/>`, green)
	s.text(34, 31, "mono", 10.5, 700, accent, ` letter-spacing="1.8"`, "CODEWARS")

	x := 34 + monoWidth("CODEWARS", 10.5, 1.8) + 12
	for _, id := range c.Languages {
		name := languageName(id)
		w := monoWidth(name, 10.5, 0) + 18
		if x+w > chartLeft-20 {
			break
		}
		s.printf(`<rect x="%s" y="16" width="%s" height="22" rx="11" fill="%s"/>`, num(x), num(w), chipFill)
		s.text(x+9, 31, "mono", 10.5, 700, chipInk, "", name)
		x += w + 6
	}

	s.printf(`<path d="M817 33l9-9M819 24h7v7" fill="none" stroke="%s" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>`, muted)
}

// drawStats draws the rank, the katas solved and the honor.
func (c Card) drawStats(s *canvas) {
	const y = 99.0 // baseline of the numbers

	color, ok := rankColors[c.Rank.Color]
	if !ok {
		color = muted
	}
	s.printf(`<path class="hex" d="%s" fill="%s" fill-opacity=".16" stroke="%s" stroke-width="2" stroke-linejoin="round"/>`,
		hexagon(24+hexRadius, y, hexRadius), color, color)

	stats := []struct {
		x            float64
		value, label string
	}{
		{76, c.rankName(), "Rank"},
		{196, formatCount(c.Solved), "Katas solved"},
		{318, formatCount(c.Honor), "Honor"},
	}
	for i, st := range stats {
		s.text(st.x, y, "sans rise", 28, 700, ink, fmt.Sprintf(` letter-spacing="-0.4" style="animation-delay:%dms"`, 100+90*i), st.value)
		s.text(st.x, y+21, "sans", 13, 400, muted, "", st.label)
	}
}

// drawChart draws katas per day as bars, with the last Recent days in the
// accent color, a label on the busiest day and a bracket that says how many
// katas the highlighted days add up to.
func (c Card) drawChart(s *canvas) {
	s.text(chartLeft, 31, "sans", 12, 400, muted, "", fmt.Sprintf("Katas per day, last %d days", Days))

	slot := (chartRight - chartLeft) / Days
	pad := (slot - barWidth) / 2
	s.printf(`<path d="M%s %sH%s" stroke="%s"/>`, num(chartLeft+pad), num(chartBase+0.5), num(chartRight-pad), faint)

	peakDay, peak := c.Peak()
	if peak == 0 {
		s.text((chartLeft+chartRight)/2, 96, "sans", 12, 400, muted, ` text-anchor="middle"`, fmt.Sprintf("No katas in the last %d days", Days))
	}
	for i, n := range c.PerDay {
		if n == 0 {
			continue
		}
		x := chartLeft + float64(i)*slot + pad
		h := max(3, (chartBase-chartTop)*float64(n)/float64(peak))
		fill := older
		if i >= Days-Recent {
			fill = accent
		}
		delay := 250 + 22*i
		s.printf(`<path class="bar" d="%s" fill="%s" style="animation-delay:%dms"/>`, bar(x, chartBase, barWidth, h), fill, delay)
		if i == peakDay {
			s.text(x+barWidth/2, chartBase-h-6, "sans fade", 11, 700, body,
				fmt.Sprintf(` text-anchor="middle" style="animation-delay:%dms"`, delay+450), strconv.Itoa(n))
		}
	}

	s.text(chartLeft+pad, bottomRow, "sans", 11, 400, muted, "", c.First.Format("Jan 2"))
	if peak > 0 {
		from := chartLeft + float64(Days-Recent)*slot + pad
		to := chartRight - pad
		s.printf(`<path d="M%s %sv4H%sv-4" fill="none" stroke="%s"/>`, num(from), num(chartBase+8), num(to), muted)
		s.text(to, bottomRow, "sans", 11, 400, body, ` text-anchor="end"`, c.recentLabel())
	}
}

func (c Card) rankName() string {
	if c.Rank.Name == "" {
		return "Unranked"
	}
	return c.Rank.Name
}

func (c Card) recentLabel() string {
	switch n := c.RecentTotal(); n {
	case 0:
		return fmt.Sprintf("No katas in the last %d days", Recent)
	case 1:
		return fmt.Sprintf("1 kata in the last %d days", Recent)
	default:
		return fmt.Sprintf("%d katas in the last %d days", n, Recent)
	}
}

func css() string {
	perimeter := num(6 * hexRadius)
	return ".sans{font-family:'GTSans',-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif}" +
		".mono{font-family:'GTMono',ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}" +
		"@keyframes pulse{50%{opacity:.3}}.dot{animation:pulse 2.2s ease-in-out infinite}" +
		"@keyframes rise{from{opacity:0;transform:translateY(6px)}}.rise{animation:rise .6s ease-out backwards}" +
		"@keyframes draw{from{stroke-dashoffset:" + perimeter + "}}" +
		".hex{stroke-dasharray:" + perimeter + ";animation:draw 1.2s ease-out backwards}" +
		"@keyframes grow{from{transform:scaleY(0)}}" +
		".bar{transform-box:fill-box;transform-origin:50% 100%;animation:grow .6s cubic-bezier(.2,.8,.2,1) backwards}" +
		"@keyframes fade{from{opacity:0}}.fade{animation:fade .4s ease-out backwards}" +
		"@media (prefers-reduced-motion:reduce){*{animation:none!important}}"
}

func fontFaces() string {
	face := func(family string, weight int, woff2 []byte) string {
		return fmt.Sprintf("@font-face{font-family:'%s';font-weight:%d;src:url(data:font/woff2;base64,%s) format('woff2')}",
			family, weight, base64.StdEncoding.EncodeToString(woff2))
	}
	return face("GTSans", 400, inter400) + face("GTSans", 700, inter700) + face("GTMono", 700, mono700)
}

type canvas struct{ strings.Builder }

func (s *canvas) printf(format string, a ...any) { fmt.Fprintf(&s.Builder, format, a...) }

// text writes one line of text. class picks the font ("sans" or "mono") and
// any animation; attrs, when set, starts with a space.
func (s *canvas) text(x, y float64, class string, size float64, weight int, fill, attrs, str string) {
	s.printf(`<text x="%s" y="%s" class="%s" font-size="%s" font-weight="%d" fill="%s"%s>%s</text>`,
		num(x), num(y), class, num(size), weight, fill, attrs, escape(str))
}

// bar is a bar standing on the baseline y, with rounded top corners.
func bar(x, y, w, h float64) string {
	r := min(4, w/2, h)
	top := y - h
	return fmt.Sprintf("M%s %sV%sA%s %s 0 0 1 %s %sH%sA%s %s 0 0 1 %s %sV%sZ",
		num(x), num(y), num(top+r), num(r), num(r), num(x+r), num(top),
		num(x+w-r), num(r), num(r), num(x+w), num(top+r), num(y))
}

// hexagon is a regular hexagon with a corner at the top.
func hexagon(cx, cy, r float64) string {
	var b strings.Builder
	for i := range 6 {
		a := (60*float64(i) - 90) * math.Pi / 180
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&b, "%s%s %s", cmd, num(cx+r*math.Cos(a)), num(cy+r*math.Sin(a)))
	}
	b.WriteString("Z")
	return b.String()
}

// monoWidth is the width of s in JetBrains Mono, where every character is 0.6 em.
func monoWidth(s string, size, letterSpacing float64) float64 {
	return float64(utf8.RuneCountInString(s)) * (0.6*size + letterSpacing)
}

// formatCount writes 1234 as "1,234", and from 10,000 on shortens to "12.3K".
func formatCount(n int) string {
	switch {
	case n >= 1_000_000:
		return num(math.Floor(float64(n)/100_000)/10) + "M"
	case n >= 10_000:
		return num(math.Floor(float64(n)/100)/10) + "K"
	case n >= 1_000:
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return strconv.Itoa(n)
}

func languageName(id string) string {
	if name, ok := languageNames[id]; ok {
		return name
	}
	if id == "" {
		return id
	}
	return strings.ToUpper(id[:1]) + id[1:]
}

// num rounds to one decimal and drops a trailing ".0".
func num(f float64) string {
	return strconv.FormatFloat(math.Round(f*10)/10, 'f', -1, 64)
}

var escape = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;").Replace
