package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type parseSuggestionsEnvelope struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func ParseSuggestions(raw string, max int) ([]Suggestion, error) {
	content := strings.TrimSpace(raw)
	if content == "" {
		return nil, fmt.Errorf("%w: empty provider content", ErrMalformedResponse)
	}

	var envelope parseSuggestionsEnvelope
	if err := json.Unmarshal([]byte(content), &envelope); err == nil {
		suggestions := normalizeSuggestions(envelope.Suggestions, max)
		if len(suggestions) > 0 {
			return suggestions, nil
		}
	}

	return []Suggestion{{Content: content}}, nil
}
