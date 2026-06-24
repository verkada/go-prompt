package prompt

import (
	"syscall"
	"testing"
)

// makeAutoscaleRender builds a Render wired to a harmless writer (stdin) with a
// fixed configured max suggestion cap, mirroring how vtb configures the prompt
// via OptionMaxSuggestion.
func makeAutoscaleRender(row, col, maxSuggestion uint16, header func() string) *Render {
	return &Render{
		prefix:                       "> ",
		out:                          &PosixWriter{fd: syscall.Stdin},
		livePrefixCallback:           func() (string, bool) { return "", false },
		headerCallback:               header,
		maxSuggestion:                maxSuggestion,
		row:                          row,
		col:                          col,
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
	}
}

// completionWith returns a CompletionManager already populated with n suggestions.
func completionWith(n int, max uint16) *CompletionManager {
	suggests := make([]Suggest, n)
	for i := range suggests {
		suggests[i] = Suggest{Text: "cmd", Description: "desc"}
	}
	c := NewCompletionManager(func(Document) []Suggest { return suggests }, max)
	c.Update(*NewDocument())
	return c
}

func TestRenderAutoscalesMaxSuggestion(t *testing.T) {
	const configured uint16 = 20

	cases := []struct {
		name   string
		row    uint16
		header func() string
		want   uint16
	}{
		// Tall window: full configured cap is honored.
		{"tall window keeps configured cap", 50, nil, 20},
		// Short window, no header: input line (1) leaves row-1 for suggestions.
		{"short window clamps to fit", 10, nil, 9},
		// Very short window still shows at least one suggestion.
		{"tiny window floors at 1", 2, nil, 1},
		// Header eats into the available height.
		{"single-line header reduces room", 10, func() string { return "ENV:prod | ORG:x" }, 8},
		// Exactly enough room for the full cap.
		{"exact fit keeps cap", 21, nil, 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := completionWith(30, configured)
			r := makeAutoscaleRender(tc.row, 80, configured, tc.header)
			r.Render(NewBuffer(), c)
			if c.max != tc.want {
				t.Errorf("row=%d header=%v: completion.max = %d, want %d",
					tc.row, tc.header != nil, c.max, tc.want)
			}
		})
	}
}

// TestRenderNoAutoscaleWhenUnset verifies behavior is unchanged for consumers
// that never call OptionMaxSuggestion (maxSuggestion == 0).
func TestRenderNoAutoscaleWhenUnset(t *testing.T) {
	c := completionWith(30, 6)
	r := makeAutoscaleRender(10, 80, 0 /* unset */, nil)
	r.Render(NewBuffer(), c)
	if c.max != 6 {
		t.Errorf("with maxSuggestion unset, completion.max should stay 6, got %d", c.max)
	}
}
