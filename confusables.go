// Package confusables provides functions for identifying words that appear to
// be similar but use different characters.
package confusables

//go:generate go run scripts/build-tables.go > tables.go

import (
	"fmt"
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
	a, _ := c.toASCII(s)

	return a
}

func (c *Confusables) ToASCIIDiff(s string) (string, []Diff) {
	return c.toASCII(s)
}

func noDiff(s string) []Diff {
	diff := make([]Diff, len(s))

	for i, r := range s {
		diff[i] = Diff{
			Rune: r,
		}
	}

	return diff
}

func (c *Confusables) toASCII(s string) (string, []Diff) {
	if isASCII(s) {
		return s, noDiff(s)
	}

	var ascii strings.Builder

	diffs := make([]Diff, 0, len(s))

	for _, r := range s {
		diff := c.processRune(r)
		diffs = append(diffs, *diff)

		if diff.Confusable != nil {
			ascii.WriteString(*diff.Confusable)
		} else {
			ascii.WriteRune(r)
		}
	}

	return norm.NFKC.String(ascii.String()), diffs
}

func (c *Confusables) processRune(r rune) *Diff {
	diff := &Diff{}

	diff.Rune = r

	if r <= unicode.MaxASCII {
		return diff
	}

	if v, ok := confusables[r]; ok {
		c.removeMarks.Reset()

		v, _, _ := transform.String(c.removeMarks, v)

		if isASCII(v) {
			diff.Confusable = &v
			diff.Description = getDescriptionMapping(r, &v)

			return diff
		}
	}

	c.removeMarks.Reset()

	v, _, _ := transform.String(c.removeMarks, string(r))
	if isASCII(v) {
		diff.Confusable = &v
		diff.Description = getDescriptionMapping(r, &v)
	}

	return diff
}

// ToNumber converts characters in a string that look like numbers into numbers.
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
func AddMapping(r rune, confusable string) {
	confusables[r] = confusable
}

// AddMappingWithDesc allows a custom mapping to be defined between a rune and its confusable and for a description to
// be provided for that mapping.
func AddMappingWithDesc(r rune, confusable, runeDesc, confusableDesc string) {
	AddMapping(r, confusable)

	descriptions[runeDesc] = confusableDesc
}

// IsConfusable checks if two strings are confusable of one another.
func IsConfusable(s1, s2 string) bool {
	return ToSkeleton(s1) == ToSkeleton(s2)
}

// ToASCII converts characters in a string to their ASCII equivalent if possible.
func ToASCII(s string) string {
	return New().ToASCII(s)
}

// ToASCIIDiff converts characters in a string to their ASCII equivalent if possible.
func ToASCIIDiff(s string) (string, []Diff) {
	return New().ToASCIIDiff(s)
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
	Confusable  *string
	Description string
	Rune        rune
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
			Confusable:  confusable,
			Description: getDescriptionMapping(r, confusable),
			Rune:        r,
		}
	}

	return diffs
}

// Get the mapping between a rune and its confusable.
func getDescriptionMapping(r rune, confusable *string) string {
	if confusable == nil {
		return ""
	}

	rDesc := descriptions[string(r)]
	if rDesc == "" {
		nfd := norm.NFD.String(string(r))
		parts := make([]string, 0, len(nfd))

		for _, c := range nfd {
			cDesc := descriptions[string(c)]
			if cDesc == "" {
				return ""
			}

			parts = append(parts, cDesc)
		}

		rDesc = strings.Join(parts, ", ")
	}

	confusableDesc := descriptions[*confusable]
	if confusableDesc == "" {
		return ""
	}

	return fmt.Sprintf("%s → %s", rDesc, confusableDesc)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
