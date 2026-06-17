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
	rendered := RenderPromptTemplate("{{mode}} {{text}} {{rich_text}} {{input_format}} {{article_title}} {{max_suggestions}}", PromptRenderData{
		Mode:           ModeImprove,
		Text:           "原文",
		RichText:       "<p>原文</p>",
		InputFormat:    "html",
		ArticleTitle:   "标题",
		MaxSuggestions: 3,
	})

	require.Equal(t, "improve 原文 <p>原文</p> html 标题 3", rendered)
}

func TestDefaultPromptTemplatesAskForSuggestionsJSON(t *testing.T) {
	for mode, template := range DefaultPromptTemplates() {
		require.Contains(t, template, `"suggestions"`, mode)
		require.Contains(t, strings.ToLower(template), "json", mode)
	}
}

func TestDefaultContentPromptTemplatesAskForHTMLFragments(t *testing.T) {
	defaults := DefaultPromptTemplates()
	for _, mode := range []string{ModeImprove, ModeShorten, ModeExpand} {
		template := strings.ToLower(defaults[mode])
		require.Contains(t, template, "html fragment", mode)
		require.Contains(t, defaults[mode], "{{rich_text}}", mode)
		require.Contains(t, defaults[mode], "{{input_format}}", mode)
		require.Contains(t, defaults[mode], "不要扁平化", mode)
		require.Contains(t, defaults[mode], "必须以 rich_text 作为主要输入", mode)
	}
}

func TestDefaultTitleAndSummaryPromptTemplatesStayPlainText(t *testing.T) {
	defaults := DefaultPromptTemplates()
	for _, mode := range []string{ModeTitleCandidates, ModeSummaryCandidates} {
		require.Contains(t, defaults[mode], "纯文本", mode)
		require.Contains(t, defaults[mode], "不要包含 HTML 或 Markdown", mode)
	}
}
