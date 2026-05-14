// +build !windows

package prompt

import (
	"reflect"
	"syscall"
	"testing"
)

func TestFormatCompletion(t *testing.T) {
	scenarioTable := []struct {
		scenario      string
		completions   []Suggest
		prefix        string
		suffix        string
		expected      []Suggest
		maxWidth      int
		expectedWidth int
	}{
		{
			scenario: "",
			completions: []Suggest{
				{Text: "select"},
				{Text: "from"},
				{Text: "insert"},
				{Text: "where"},
			},
			prefix: " ",
			suffix: " ",
			expected: []Suggest{
				{Text: " select "},
				{Text: " from   "},
				{Text: " insert "},
				{Text: " where  "},
			},
			maxWidth:      20,
			expectedWidth: 8,
		},
		{
			scenario: "",
			completions: []Suggest{
				{Text: "select", Description: "select description"},
				{Text: "from", Description: "from description"},
				{Text: "insert", Description: "insert description"},
				{Text: "where", Description: "where description"},
			},
			prefix: " ",
			suffix: " ",
			expected: []Suggest{
				{Text: " select ", Description: " select description "},
				{Text: " from   ", Description: " from description   "},
				{Text: " insert ", Description: " insert description "},
				{Text: " where  ", Description: " where description  "},
			},
			maxWidth:      40,
			expectedWidth: 28,
		},
	}

	for _, s := range scenarioTable {
		ac, width := formatSuggestions(s.completions, s.maxWidth)
		if !reflect.DeepEqual(ac, s.expected) {
			t.Errorf("Should be %#v, but got %#v", s.expected, ac)
		}
		if width != s.expectedWidth {
			t.Errorf("Should be %#v, but got %#v", s.expectedWidth, width)
		}
	}
}

func TestBreakLineCallback(t *testing.T) {
	var i int
	r := &Render{
		prefix: "> ",
		out: &PosixWriter{
			fd: syscall.Stdin, // "write" to stdin just so we don't mess with the output of the tests
		},
		livePrefixCallback:           func() (string, bool) { return "", false },
		prefixTextColor:              Blue,
		prefixBGColor:                DefaultColor,
		inputTextColor:               DefaultColor,
		inputBGColor:                 DefaultColor,
		previewSuggestionTextColor:   Green,
		previewSuggestionBGColor:     DefaultColor,
		suggestionTextColor:          White,
		suggestionBGColor:            Cyan,
		selectedSuggestionTextColor:  Black,
		selectedSuggestionBGColor:    Turquoise,
		descriptionTextColor:         Black,
		descriptionBGColor:           Turquoise,
		selectedDescriptionTextColor: White,
		selectedDescriptionBGColor:   Cyan,
		scrollbarThumbColor:          DarkGray,
		scrollbarBGColor:             Cyan,
		col:                          1,
	}
	b := NewBuffer()
	r.BreakLine(b)

	if i != 0 {
		t.Errorf("i should initially be 0, before applying a break line callback")
	}

	r.breakLineCallback = func(doc *Document) {
		i++
	}
	r.BreakLine(b)
	r.BreakLine(b)
	r.BreakLine(b)

	if i != 3 {
		t.Errorf("BreakLine callback not called, i should be 3")
	}
}

func TestHeaderCallback(t *testing.T) {
	makeRender := func(header func() string) *Render {
		return &Render{
			prefix: "> ",
			out: &PosixWriter{
				fd: syscall.Stdin,
			},
			livePrefixCallback:           func() (string, bool) { return "", false },
			headerCallback:               header,
			prefixTextColor:              Blue,
			prefixBGColor:                DefaultColor,
			inputTextColor:               DefaultColor,
			inputBGColor:                 DefaultColor,
			previewSuggestionTextColor:   Green,
			previewSuggestionBGColor:     DefaultColor,
			suggestionTextColor:          White,
			suggestionBGColor:            Cyan,
			selectedSuggestionTextColor:  Black,
			selectedSuggestionBGColor:    Turquoise,
			descriptionTextColor:         Black,
			descriptionBGColor:           Turquoise,
			selectedDescriptionTextColor: White,
			selectedDescriptionBGColor:   Cyan,
			scrollbarThumbColor:          DarkGray,
			scrollbarBGColor:             Cyan,
			col:                          80,
		}
	}

	t.Run("BreakLine resets headerLines to 0 for single-line header", func(t *testing.T) {
		r := makeRender(func() string { return "ENV:prod1 | ORG:xxxx" })
		r.headerLines = 1
		r.BreakLine(NewBuffer())
		if r.headerLines != 0 {
			t.Errorf("expected headerLines=0 after BreakLine, got %d", r.headerLines)
		}
	})

	t.Run("BreakLine resets headerLines to 0 for multi-line header", func(t *testing.T) {
		r := makeRender(func() string { return "ENV:prod1\nORG:xxxx\nUSR:yyyy" })
		r.headerLines = 3
		r.BreakLine(NewBuffer())
		if r.headerLines != 0 {
			t.Errorf("expected headerLines=0 after BreakLine, got %d", r.headerLines)
		}
	})

	t.Run("BreakLine is a no-op when headerLines is already 0", func(t *testing.T) {
		r := makeRender(nil)
		r.headerLines = 0
		r.BreakLine(NewBuffer())
		if r.headerLines != 0 {
			t.Errorf("expected headerLines to remain 0, got %d", r.headerLines)
		}
	})
}
