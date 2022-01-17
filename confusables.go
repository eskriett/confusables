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

	return ascii.String()
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

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
