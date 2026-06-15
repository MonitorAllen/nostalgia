package ai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSuggestionsReadsJSONEnvelope(t *testing.T) {
	suggestions, err := ParseSuggestions(`{"suggestions":[{"content":"A","reason":"R"},{"content":"B"}]}`, 1)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: "A", Reason: "R"}}, suggestions)
}

func TestParseSuggestionsFallsBackToRawText(t *testing.T) {
	suggestions, err := ParseSuggestions("普通文本候选", 3)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: "普通文本候选"}}, suggestions)
}

func TestParseSuggestionsFallsBackWhenJSONHasNoUsableSuggestions(t *testing.T) {
	suggestions, err := ParseSuggestions(`{"suggestions":[{"content":"   "}],"note":"raw"}`, 3)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: `{"suggestions":[{"content":"   "}],"note":"raw"}`}}, suggestions)
}

func TestParseSuggestionsRejectsEmptyOutput(t *testing.T) {
	_, err := ParseSuggestions("  ", 3)

	require.ErrorIs(t, err, ErrMalformedResponse)
}
