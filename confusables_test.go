package confusables_test

import (
	"testing"

	"github.com/eskriett/confusables"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	c := confusables.New()

	assert.IsType(t, &confusables.Confusables{}, c)
}

func TestIsConfusable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s1, s2       string
		isConfusable bool
	}{
		{"", "", true},
		{"", "testing", false},
		{"Ａ", "Α", true},
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
	t.Parallel()

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
		{"❶,❷,❸,❹,❺,❻,❼,❽,❾,❿", "1,2,3,4,5,6,7,8,9,10"},
		{"➀,➁,➂,➃,➄,➅,➆,➇,➈,➉", "1,2,3,4,5,6,7,8,9,10"},
		{"➊,➋,➌,➍,➎,➏,➐,➑,➒,➓", "1,2,3,4,5,6,7,8,9,10"},
		{"⓪,①,②,③,④,⑤,⑥,⑦,⑧,⑨,⑩,⑪,⑫,⑬,⑭,⑮,⑯,⑰,⑱,⑲,⑳", "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"},
		{"⑴,⑵,⑶,⑷,⑸,⑹,⑺,⑻,⑼,⑽,⑾,⑿,⒀,⒁,⒂,⒃,⒄,⒅,⒆,⒇", "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"},
		{"🄀,⒈,⒉,⒊,⒋,⒌,⒍,⒎,⒏,⒐,⒑,⒒,⒓,⒔,⒕,⒖,⒗,⒘,⒙,⒚,⒛", "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"},
		{"⓿,⓫,⓬,⓭,⓮,⓯,⓰,⓱,⓲,⓳,⓴", "0,11,12,13,14,15,16,17,18,19,20"},
		{"⓵,⓶,⓷,⓸,⓹,⓺,⓻,⓼,⓽,⓾", "1,2,3,4,5,6,7,8,9,10"},
		{"𝟎,𝟏,𝟐,𝟑,𝟒,𝟓,𝟔,𝟕,𝟖,𝟗", "0,1,2,3,4,5,6,7,8,9"},
		{"𝟘,𝟙,𝟚,𝟛,𝟜,𝟝,𝟞,𝟟,𝟠,𝟡", "0,1,2,3,4,5,6,7,8,9"},
		{"𝟢,𝟣,𝟤,𝟥,𝟦,𝟧,𝟨,𝟩,𝟪,𝟫", "0,1,2,3,4,5,6,7,8,9"},
		{"𝟬,𝟭,𝟮,𝟯,𝟰,𝟱,𝟲,𝟳,𝟴,𝟵", "0,1,2,3,4,5,6,7,8,9"},
		{"𝟶,𝟷,𝟸,𝟹,𝟺,𝟻,𝟼,𝟽,𝟾,𝟿", "0,1,2,3,4,5,6,7,8,9"},
		{"０,１,２,３,４,５,６,７,８,９", "0,1,2,3,4,5,6,7,8,9"},
	}

	// Allow custom mappings to be defined
	confusables.AddMappingWithDesc('ʍ', "m",
		"LATIN SMALL LETTER TURNED W", "LATIN SMALL LETTER M")

	for _, test := range tests {
		assert.Equal(t, test.ascii, confusables.ToASCII(test.confusable))
	}
}

func TestToASCIIDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		confusable, ascii string
		diff              []confusables.Diff
	}{
		{"", "", []confusables.Diff{}},
		{"tòñ", "ton", []confusables.Diff{
			{
				Confusable:  nil,
				Description: nil,
				Rune:        't',
			},
			{
				Confusable: strPtr("o"),
				Description: &confusables.Description{
					From: "LATIN SMALL LETTER O, COMBINING GRAVE ACCENT",
					To:   "LATIN SMALL LETTER O",
				},
				Rune: 'ò',
			},
			{
				Confusable: strPtr("n"),
				Description: &confusables.Description{
					From: "LATIN SMALL LETTER N, COMBINING TILDE",
					To:   "LATIN SMALL LETTER N",
				},
				Rune: 'ñ',
			},
		}},
		{"❶", "1", []confusables.Diff{
			{
				Confusable: strPtr("1"),
				Description: &confusables.Description{
					From: "DINGBAT NEGATIVE CIRCLED DIGIT ONE",
					To:   "DIGIT ONE",
				},
				Rune: '❶',
			},
		}},
	}

	for _, test := range tests {
		ascii, diff := confusables.ToASCIIDiff(test.confusable)

		assert.Equal(t, test.ascii, ascii)
		assert.EqualValues(t, test.diff, diff)
	}
}

func TestToNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		confusable, number string
	}{
		{"", ""},
		{"foobar", "f00bar"},
		{"O12", "012"},
		{"I23", "123"},
		{"!23", "123"},
		{"𝘖l2", "012"},
	}

	for _, test := range tests {
		assert.Equal(t, test.number, confusables.ToNumber(test.confusable))
	}
}

func TestToSkeleton(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
				{
					Confusable: strPtr("rn"),
					Description: &confusables.Description{
						From: "LATIN SMALL LETTER M",
						To:   "LATIN SMALL LETTER R, LATIN SMALL LETTER N",
					},
					Rune: 'm',
				},
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

func strPtr(s string) *string {
	return &s
}
