// Package confusables provides functions for identifying words that appear to
// be similar but use different characters.
package confusables

//go:generate go run scripts/build-tables.go > tables.go

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ErrIgnoreLine is raised when processing a line which should be ignored.
var ErrIgnoreLine = errors.New("line should be ignored")

const (
	base    = 16
	bitsize = 64
)

// ConfusableEntry defines a parsed entry from a confusable mapping file.
type ConfusableEntry struct {
	Description Description
	Source      rune
	Target      string
}

// Confusables provides functions for identifying words that appear to be
// similar but use different characters.
type Confusables interface {
	ToASCII(s string) string
	ToASCIIDiff(s string) (string, []Diff)
	ToNumber(s string) string
}

// UnsafeConfusables is a single-threaded implementation of Confusables.
//
// Not safe to use in a concurrent context. Use [NewSafe] to return a
// thread-safe implementation that can be used for concurrent access.
type UnsafeConfusables struct {
	removeMarks transform.Transformer
}

// Ensure [UnsafeConfusables] implements [Confusables].
var _ Confusables = &UnsafeConfusables{}

// SafeConfusables is a thread-safe implementation of Confusables.
type SafeConfusables struct {
	UnsafeConfusables
	Lock sync.Mutex
}

// Ensure [SafeConfusables] implements [Confusables].
var _ Confusables = &SafeConfusables{}

// Description describes a mapping for a confusable.
type Description struct {
	From string
	To   string
}

// Diff details the mapping from a rune to its confusable if it exists.
type Diff struct {
	Confusable  *string
	Description *Description
	Rune        rune
}

// New creates a new instance of Confusables.
func New() *UnsafeConfusables {
	return &UnsafeConfusables{
		removeMarks: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
	}
}

// NewSafe creates a new instance of SafeConfusables.
func NewSafe() *SafeConfusables {
	return &SafeConfusables{
		UnsafeConfusables: *New(),
	}
}

// ToASCII converts characters in a string to their ASCII equivalent if possible.
func (c *UnsafeConfusables) ToASCII(s string) string {
	a, _ := c.toASCII(s)

	return a
}

func (c *UnsafeConfusables) ToASCIIDiff(s string) (string, []Diff) {
	return c.toASCII(s)
}

// ToNumber converts characters in a string that look like numbers into numbers.
func (c *UnsafeConfusables) ToNumber(s string) string {
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

func (c *UnsafeConfusables) processRune(r rune) *Diff {
	diff := &Diff{}

	diff.Rune = r

	if r <= unicode.MaxASCII {
		return diff
	}

	if v, ok := confusables[r]; ok {
		c.removeMarks.Reset()

		v, _, _ = transform.String(c.removeMarks, v)

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

func (c *UnsafeConfusables) toASCII(s string) (string, []Diff) {
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

// ToASCII converts characters in a string to their ASCII equivalent if
// possible.
//
// Thread-safe version.
func (c *SafeConfusables) ToASCII(s string) string {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	return c.UnsafeConfusables.ToASCII(s)
}

// ToASCIIDiff converts characters in a string to their ASCII equivalent if
// possible.
//
// Thread-safe version.
func (c *SafeConfusables) ToASCIIDiff(s string) (string, []Diff) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	return c.UnsafeConfusables.ToASCIIDiff(s)
}

// ToNumber converts characters in a string to their numeric values if possible.
//
// Thread-safe version.
func (c *SafeConfusables) ToNumber(s string) string {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	return c.UnsafeConfusables.ToNumber(s)
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

// LoadMappings reads r and loads in confusable mappings. Where a confusable already exists, this will override the
// mapping.
func LoadMappings(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		confusableEntry, err := ParseLine(scanner.Text())
		if err != nil {
			if errors.Is(err, ErrIgnoreLine) {
				continue
			}

			return err
		}

		AddMappingWithDesc(confusableEntry.Source, confusableEntry.Target, confusableEntry.Description.From,
			confusableEntry.Description.To)
	}

	return nil
}

// ParseLine takes a confusable line and returns a ConfusableEntry.
// If a line should be skipped an ErrIgnoreLine error is raised.
func ParseLine(line string) (*ConfusableEntry, error) {
	// Remove BOM, skip comments and blank lines
	line = strings.TrimPrefix(line, string([]byte{0xEF, 0xBB, 0xBF}))
	if strings.HasPrefix(line, "#") || line == "" {
		return nil, ErrIgnoreLine
	}

	// Extract source -> target mapping
	fields := strings.Split(line, " ;\t")

	sourceRunes, err := codepointsToRunes(fields[0])
	if err != nil {
		return nil, err
	}

	target, err := codepointsToRunes(fields[1])
	if err != nil {
		return nil, err
	}

	return &ConfusableEntry{
		Description: Description{
			From: strings.TrimSpace(strings.Split(strings.Split(fields[2], " → ")[1], " ) ")[1]),
			To:   strings.TrimSpace(strings.Split(strings.Split(fields[2], " → ")[2], "#")[0]),
		},
		Source: sourceRunes[0],
		Target: string(target),
	}, nil
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

func codepointsToRunes(s string) ([]rune, error) {
	codePoints := strings.Split(s, " ")

	runes := make([]rune, 0, len(codePoints))

	for _, unicodeCodePoint := range codePoints {
		codePoint, err := strconv.ParseUint(unicodeCodePoint, base, bitsize)
		if err != nil {
			return nil, err
		}

		runes = append(runes, rune(codePoint))
	}

	return runes, nil
}

// Get the mapping between a rune and its confusable.
func getDescriptionMapping(r rune, confusable *string) *Description {
	if confusable == nil {
		return nil
	}

	rDesc := descriptions[string(r)]
	if rDesc == "" {
		nfd := norm.NFD.String(string(r))
		parts := make([]string, 0, len(nfd))

		for _, c := range nfd {
			cDesc := descriptions[string(c)]
			if cDesc == "" {
				return nil
			}

			parts = append(parts, cDesc)
		}

		rDesc = strings.Join(parts, ", ")
	}

	confusableDesc := descriptions[*confusable]
	if confusableDesc == "" {
		return nil
	}

	return &Description{
		From: rDesc,
		To:   confusableDesc,
	}
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
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
