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

func TestParseSuggestionsReadsJSONEnvelopeFromMarkdownFence(t *testing.T) {
	raw := "```json\n" +
		`{"suggestions":[{"content":"扩写候选","reason":"信息更完整"},{"content":"精简候选"}]}` +
		"\n```"

	suggestions, err := ParseSuggestions(raw, 3)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{
		{Content: "扩写候选", Reason: "信息更完整"},
		{Content: "精简候选"},
	}, suggestions)
}

func TestParseSuggestionsReadsJSONEnvelopeFromMarkdownFenceWithSurroundingText(t *testing.T) {
	raw := "下面是候选：\n```json\n" +
		`{"suggestions":[{"content":"候选正文","reason":"符合要求"}]}` +
		"\n```\n请确认。"

	suggestions, err := ParseSuggestions(raw, 3)

	require.NoError(t, err)
	require.Equal(t, []Suggestion{{Content: "候选正文", Reason: "符合要求"}}, suggestions)
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
