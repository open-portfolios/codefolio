package components

import "strings"

const codefolioGlyphGap = "  "

var codefolioGlyphs = map[rune][]string{
	'C': {"██████", "██    ", "██    ", "██    ", "██████"},
	'D': {"█████ ", "██  ██", "██  ██", "██  ██", "█████ "},
	'E': {"██████", "██    ", "█████ ", "██    ", "██████"},
	'F': {"██████", "██    ", "█████ ", "██    ", "██    "},
	'I': {"██████", "  ██  ", "  ██  ", "  ██  ", "██████"},
	'L': {"██    ", "██    ", "██    ", "██    ", "██████"},
	'O': {"██████", "██  ██", "██  ██", "██  ██", "██████"},
}

// CodefolioBanner is shared by the Timeline welcome state and the restored
// terminal after exit, so both surfaces use the same recognizable mark.
func CodefolioBanner() []string {
	rows := make([]string, 5, 6)
	for _, letter := range "CODEFOLIO" {
		glyph := codefolioGlyphs[letter]
		for row := range rows {
			if rows[row] != "" {
				rows[row] += codefolioGlyphGap
			}
			rows[row] += glyph[row]
		}
	}
	return append(rows, "CODEFOLIO Coding Agent")
}

func CodefolioBannerText() string {
	return strings.Join(CodefolioBanner(), "\n")
}
