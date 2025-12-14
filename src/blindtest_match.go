package server

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func normalizeGuess(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = norm.NFD.String(s)

	var b strings.Builder
	b.Grow(len(s))
	space := false

	for _, r := range s {
		// retire les accents/diacritiques
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		// garde lettres/chiffres
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			space = false
			continue
		}
		// tout le reste -> espace unique
		if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

func isCorrectGuess(guess, title, artist string) bool {
	g := normalizeGuess(guess)
	if g == "" {
		return false
	}
	return g == normalizeGuess(title) || g == normalizeGuess(artist)
}
