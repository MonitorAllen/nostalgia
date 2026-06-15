package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePromptTemplatesFillsDefaults(t *testing.T) {
	templates := NormalizePromptTemplates(map[string]string{
		ModeImprove: "custom improve {{text}}",
	})

	require.Equal(t, "custom improve {{text}}", templates[ModeImprove])
	require.NotEmpty(t, templates[ModeShorten])
	require.NotEmpty(t, templates[ModeExpand])
	require.NotEmpty(t, templates[ModeTitleCandidates])
	require.NotEmpty(t, templates[ModeSummaryCandidates])
}

func TestRenderPromptTemplateReplacesVariables(t *testing.T) {
	rendered := RenderPromptTemplate("{{mode}} {{text}} {{article_title}} {{max_suggestions}}", PromptRenderData{
		Mode:           ModeImprove,
		Text:           "原文",
		ArticleTitle:   "标题",
		MaxSuggestions: 3,
	})

	require.Equal(t, "improve 原文 标题 3", rendered)
}

func TestDefaultPromptTemplatesAskForSuggestionsJSON(t *testing.T) {
	for mode, template := range DefaultPromptTemplates() {
		require.Contains(t, template, `"suggestions"`, mode)
		require.Contains(t, strings.ToLower(template), "json", mode)
	}
}
