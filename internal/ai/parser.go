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

	jsonContent := extractMarkdownJSONFence(content)
	var envelope parseSuggestionsEnvelope
	if err := json.Unmarshal([]byte(jsonContent), &envelope); err == nil {
		suggestions := normalizeSuggestions(envelope.Suggestions, max)
		if len(suggestions) > 0 {
			return suggestions, nil
		}
	}

	return []Suggestion{{Content: content}}, nil
}

func extractMarkdownJSONFence(content string) string {
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		fenceInfo, ok := markdownFenceInfo(line)
		if !ok || (fenceInfo != "" && !strings.EqualFold(fenceInfo, "json")) {
			continue
		}

		for end := index + 1; end < len(lines); end++ {
			if strings.TrimSpace(lines[end]) == "```" {
				return strings.TrimSpace(strings.Join(lines[index+1:end], "\n"))
			}
		}

		return content
	}

	return content
}

func markdownFenceInfo(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "```") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, "```")), true
}

func normalizeSuggestions(values []Suggestion, max int) []Suggestion {
	if max <= 0 {
		max = 3
	}
	suggestions := make([]Suggestion, 0, min(len(values), max))
	for _, suggestion := range values {
		content := strings.TrimSpace(suggestion.Content)
		if content == "" {
			continue
		}
		suggestions = append(suggestions, Suggestion{
			Content: content,
			Reason:  strings.TrimSpace(suggestion.Reason),
		})
		if len(suggestions) == max {
			break
		}
	}
	return suggestions
}
