package confusables

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsConfusable(t *testing.T) {
	tests := []struct {
		s1, s2       string
		isConfusable bool
	}{
		{"", "", true},
		{"", "testing", false},
		{"example", "𝐞х⍺𝓂𝕡Іꬲ", true},
		{"example", "𝐞х⍺𝓂𝕡І", false},
		{"example", "𝐞х⍺𝓂𝕡Іe", true},
	}
	for i, d := range tests {
		isConfuse := IsConfusable(d.s1, d.s2)
		if isConfuse != d.isConfusable {
			t.Errorf("Test[%d]: IsConfusable('%s','%s') returned %t, want %t",
				i, d.s1, d.s2, isConfuse, d.isConfusable)
		}
	}
}

func TestToSkeleton(t *testing.T) {
	tests := []struct {
		s, skeleton string
	}{
		{"", ""},
		{"example", "exarnple"},
		{"𝐞х⍺𝓂𝕡Іꬲ", "exarnple"},
	}
	for i, d := range tests {
		skeleton := ToSkeleton(d.s)
		if skeleton != d.skeleton {
			t.Errorf("Test[%d]: ToSkeleton('%s') returned %s, want %s",
				i, d.s, skeleton, d.skeleton)
		}
	}
}

func TestToSkeletonDiff(t *testing.T) {
	confusable := "rn"
	tests := []struct {
		s    string
		diff []Diff
	}{
		{"", nil},
		{
			"tum",
			[]Diff{
				{Rune: 't'},
				{Rune: 'u'},
				{Rune: 'm', Confusable: &confusable},
			},
		},
	}
	for _, d := range tests {
		diff := ToSkeletonDiff(d.s)
		assert.EqualValues(t, d.diff, diff)
	}
}

func BenchmarkToSkeleton(b *testing.B) {
	b.Run("ToSkeleton", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ToSkeleton("𝐞х⍺𝓂𝕡Іꬲ")
		}
	})
}

func BenchmarkIsConfusable(b *testing.B) {
	b.Run("IsConfusable", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			IsConfusable("example", "𝐞х⍺𝓂𝕡Іꬲ")
		}
	})
}
