package entrypoint

import (
	"regexp"

	"github.com/stretchr/testify/assert"
	"testing"
)

var splitKeepDelimiterTests = []struct {
	in  string
	out []string
}{
	{"one two three", []string{"one ", "two ", "three"}},
	{"one\ttwo\tthree", []string{"one\t", "two\t", "three"}},
	{"one\n two    three\t\n ", []string{"one\n ", "two    ", "three\t\n "}},
	{"onetwothree\n", []string{"onetwothree\n"}},
	{"\nonetwothree", []string{"\n", "onetwothree"}},
}

func TestSplitKeepDelimiter(t *testing.T) {
	for _, tt := range splitKeepDelimiterTests {
		t.Run(tt.in, func(t *testing.T) {
			re := regexp.MustCompile("\\s+")
			assert.Equal(t, SplitKeepDelimiter(re, tt.in), tt.out)
		})
	}
}

var determineTabTests = []struct {
	in    string
	delim string
	out   string
}{
	{"   o three", "\t", "   "},
	{"\to  three", "\t", "\t"},
	{"o three", "\t", ""},
	{"  one\ttwo", "\t", "     \t"},
	{"  \ttwo", "\t", "  \t"},
	{"  hello|world", "\\|", "        "},
}

func TestHelpTableAwareDetermineIndent(t *testing.T) {
	for _, tt := range determineTabTests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, HelpTableAwareDetermineIndent(tt.in, tt.delim), tt.out)
		})
	}
}

var wrapTextTests = []struct {
	in  string
	out string
	tab string
}{
	{
		"    Great Scott!",
		"    Great\n    Scott!",
		"    ",
	},
	{
		"You made a time machine out of a Delorean!?",
		"You made a time\nmachine out of\na Delorean!?",
		"",
	},
	{
		"  fc\tThe box that",
		"  fc\tThe\n    \tbox\n    \tthat",
		"    \t",
	},
}

func TestTabAwareWrapText(t *testing.T) {
	for _, tt := range wrapTextTests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(
				t,
				TabAwareWrapText(tt.in, 15, tt.tab),
				tt.out)
		})
	}
}
