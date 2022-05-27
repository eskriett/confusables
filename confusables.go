// Package confusables provides functions for identifying words that appear to
// be similar but use different characters.
package confusables

//go:generate go run scripts/build-tables.go > tables.go

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Confusables provides functions for identifying words that appear to be similar but use different characters.
type Confusables struct {
	removeMarks transform.Transformer
}

// New creates a new instance of Confusables.
func New() *Confusables {
	return &Confusables{
		removeMarks: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
	}
}

// ToASCII converts characters in a string to their ASCII equivalent if possible.
func (c *Confusables) ToASCII(s string) string {
	if isASCII(s) {
		return s
	}

	var ascii strings.Builder

	for _, r := range s {
		if num := handleNumber(r); num != "" {
			ascii.WriteString(num)

			continue
		}

		if r > unicode.MaxASCII {
			if v, ok := confusables[r]; ok {
				c.removeMarks.Reset()

				v, _, _ := transform.String(c.removeMarks, v)

				if isASCII(v) {
					ascii.WriteString(v)

					continue
				}
			}

			c.removeMarks.Reset()

			v, _, _ := transform.String(c.removeMarks, string(r))
			if isASCII(v) {
				ascii.WriteString(v)

				continue
			}
		}

		ascii.WriteRune(r)
	}

	return norm.NFKC.String(ascii.String())
}

// ToNumber converts a string to a number.
func (c *Confusables) ToNumber(s string) string {
	s = c.ToASCII(s)

	var number strings.Builder

	for _, r := range s {
		switch strings.ToLower(string(r)) {
		case "o":
			r = '0'
		case "i", "l", "!":
			r = '1'
		}

		number.WriteRune(r)
	}

	return number.String()
}

// AddMapping allows custom mappings to be defined for a rune.
func AddMapping(r rune, s string) {
	confusables[r] = s
}

// IsConfusable checks if two strings are confusable of one another.
func IsConfusable(s1, s2 string) bool {
	return ToSkeleton(s1) == ToSkeleton(s2)
}

// ToASCII converts characters in a string to their ASCII equivalent if possible.
func ToASCII(s string) string {
	return New().ToASCII(s)
}

// ToNumber converts characters in a string to their numeric values if possible.
func ToNumber(s string) string {
	return New().ToNumber(s)
}

// ToSkeleton converts a string to its skeleton form as defined by the skeleton
// algorithm in https://www.unicode.org/reports/tr39/#def-skeleton.
func ToSkeleton(s string) string {
	nfd := norm.NFD.String(s)

	var skeleton strings.Builder

	for _, r := range nfd {
		if c, ok := confusables[r]; ok {
			skeleton.WriteString(c)
		} else {
			skeleton.WriteRune(r)
		}
	}

	return skeleton.String()
}

// Diff details the mapping from a rune to its confusable if it exists.
type Diff struct {
	Rune       rune
	Confusable *string
}

// ToSkeletonDiff returns a slice of Diff detailing the changes that have been
// made within the string to reach its skeleton form.
func ToSkeletonDiff(s string) []Diff {
	nfd := norm.NFD.String(s)

	if len(nfd) == 0 {
		return nil
	}

	diffs := make([]Diff, len(nfd))

	for i, r := range nfd {
		var confusable *string
		if c, ok := confusables[r]; ok {
			confusable = &c
		}

		diffs[i] = Diff{
			Rune:       r,
			Confusable: confusable,
		}
	}

	return diffs
}

func handleNumber(r rune) string {
	switch {
	// Dingbat Negative Circled
	case r >= '❶' && r <= '❾':
		r -= '❶' - '1'
	// Dingbat Circled Sans-Serif
	case r >= '➀' && r <= '➈':
		r -= '➀' - '1'
	case r == '🄋':
		r = '0'
	// Dingbat Negative Circled Sans-Serif
	case r >= '➊' && r <= '➒':
		r -= '➊' - '1'
	case r == '🄌':
		r = '0'
	// Parenthesized
	case r >= '⑴' && r <= '⑼':
		r -= '⑴' - '1'
	case r >= '⑽' && r <= '⒆':
		r -= '⑾' - '1'

		return "1" + string(r)
	// Full Stop
	case r >= '⒈' && r <= '⒐':
		r -= '⒈' - '1'
	case r >= '⒑' && r <= '⒚':
		r -= '⒒' - '1'

		return "1" + string(r)
	case r == '🄀':
		r = '0'
	// Negative Circled
	case r >= '⓫' && r <= '⓳':
		r -= '⓫' - '1'

		return "1" + string(r)
	case r == '⓿':
		r = '0'
	// Double Circled
	case r >= '⓵' && r <= '⓽':
		r -= '⓵' - '1'
	// Mathematical Bold
	case r >= '𝟎' && r <= '𝟗':
		r -= '𝟏' - '1'
	// Mathematical Double-struck
	case r >= '𝟘' && r <= '𝟡':
		r -= '𝟙' - '1'
	// Mathematical Sans-serif
	case r >= '𝟢' && r <= '𝟫':
		r -= '𝟣' - '1'
	// Mathematical Sans-serif bold
	case r >= '𝟬' && r <= '𝟵':
		r -= '𝟭' - '1'
	// Mathematical Monospace
	case r >= '𝟶' && r <= '𝟿':
		r -= '𝟷' - '1'
	// Double digits various character classes
	case r == '❿', r == '➉', r == '➓', r == '⓾':
		return "10"
	case r == '⒇', r == '⒛', r == '⓴':
		return "20"
	default:
		return ""
	}

	return string(r)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
