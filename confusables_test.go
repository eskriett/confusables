package confusables_test

import (
	"testing"

	"github.com/eskriett/confusables"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := confusables.New()

	assert.IsType(t, &confusables.Confusables{}, c)
}

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
		isConfuse := confusables.IsConfusable(d.s1, d.s2)
		if isConfuse != d.isConfusable {
			t.Errorf("Test[%d]: IsConfusable('%s','%s') returned %t, want %t",
				i, d.s1, d.s2, isConfuse, d.isConfusable)
		}
	}
}

func TestToASCII(t *testing.T) {
	tests := []struct {
		confusable, ascii string
	}{
		{"", ""},
		{"example", "example"},
		{"exαʍple", "example"},
		{"exαʍple", "example"},
		{"ɼecoɼd", "record"},
		{"exȧmple", "example"},
		{"newtòñ", "newton"},
	}

	// Allow custom mappings to be defined
	confusables.AddMapping('ʍ', "m")

	for _, test := range tests {
		assert.Equal(t, test.ascii, confusables.ToASCII(test.confusable))
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
		skeleton := confusables.ToSkeleton(d.s)
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
		diff []confusables.Diff
	}{
		{"", nil},
		{
			"tum",
			[]confusables.Diff{
				{Rune: 't'},
				{Rune: 'u'},
				{Rune: 'm', Confusable: &confusable},
			},
		},
	}
	for _, d := range tests {
		diff := confusables.ToSkeletonDiff(d.s)
		assert.EqualValues(t, d.diff, diff)
	}
}

func BenchmarkToASCII(b *testing.B) {
	b.Run("ToASCII", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			confusables.ToASCII("𝐞х⍺𝓂𝕡Іꬲ")
		}
	})
}

func BenchmarkToSkeleton(b *testing.B) {
	b.Run("ToSkeleton", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			confusables.ToSkeleton("𝐞х⍺𝓂𝕡Іꬲ")
		}
	})
}

func BenchmarkIsConfusable(b *testing.B) {
	b.Run("IsConfusable", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			confusables.IsConfusable("example", "𝐞х⍺𝓂𝕡Іꬲ")
		}
	})
}
