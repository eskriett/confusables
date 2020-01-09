package confusables

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
