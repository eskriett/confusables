// Package confusables provides functions for identifying words that appear to
// be similar but use different characters
package confusables

//go:generate go run scripts/build-tables.go > tables.go

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// IsConfusable checks if two strings are confusable of one another
func IsConfusable(s1, s2 string) bool {
	return ToSkeleton(s1) == ToSkeleton(s2)
}

// ToSkeleton converts a string to its skeleton form as defined by the skeleton
// algorithm in https://www.unicode.org/reports/tr39/#def-skeleton
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

// Diff details the mapping from a rune to its confusable if it exists
type Diff struct {
	Rune       rune
	Confusable *string
}

// ToSkeletonDiff returns a slice of Diff detailing the changes that have been
// made within the string to reach its skeleton form
func ToSkeletonDiff(s string) []Diff {
	nfd := norm.NFD.String(s)

	var diffs []Diff
	for _, r := range nfd {
		var confusable *string
		if c, ok := confusables[r]; ok {
			confusable = &c
		}
		diffs = append(diffs, Diff{
			Rune:       r,
			Confusable: confusable,
		})
	}

	return diffs
}
