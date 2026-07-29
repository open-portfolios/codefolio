package components

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCodefolioBannerIsFiveAlignedRows(t *testing.T) {
	banner := CodefolioBanner()
	if len(banner) != 6 {
		t.Fatalf("banner rows = %d, want five glyph rows plus label", len(banner))
	}
	width := utf8.RuneCountInString(banner[0])
	for row, line := range banner[:5] {
		if got := utf8.RuneCountInString(line); got != width {
			t.Fatalf("banner row %d width = %d, want %d", row, got, width)
		}
	}
	if banner[5] != "CODEFOLIO Coding Agent" {
		t.Fatalf("banner label = %q, want CODEFOLIO Coding Agent", banner[5])
	}
	if !strings.Contains(CodefolioBannerText(), "█") {
		t.Fatal("banner must use block glyphs")
	}
}

func TestCodefolioBannerUsesTwoColumnsBetweenEveryGlyph(t *testing.T) {
	letters := "CODEFOLIO"
	for row, line := range CodefolioBanner()[:5] {
		var want strings.Builder
		for i, letter := range letters {
			if i > 0 {
				want.WriteString(codefolioGlyphGap)
			}
			want.WriteString(codefolioGlyphs[letter][row])
		}
		if line != want.String() {
			t.Fatalf("banner row %d does not preserve two-column glyph gaps: %q", row, line)
		}
	}
}

func TestEmptyTranscriptShowsCodefolioBanner(t *testing.T) {
	lines := (&Transcript{}).lines(80)
	banner := CodefolioBanner()
	if len(lines) < len(banner)+2 {
		t.Fatalf("welcome lines = %d, want banner and caption", len(lines))
	}
	for i, row := range banner {
		if lines[i].text != row || lines[i].fg != Theme.Primary {
			t.Fatalf("welcome line %d = %#v, want banner row %q", i, lines[i], row)
		}
	}
	if lines[len(lines)-1].text != "Ready when you are." {
		t.Fatalf("welcome caption = %q", lines[len(lines)-1].text)
	}
}
