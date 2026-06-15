package ai

import (
	"sort"
	"strconv"
	"strings"
)

type PromptRenderData struct {
	Mode           string
	Target         string
	Text           string
	ArticleTitle   string
	ArticleSummary string
	ArticleExcerpt string
	Locale         string
	MaxSuggestions int
}

func DefaultPromptTemplates() map[string]string {
	return map[string]string{
		ModeImprove: `请润色以下内容。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
mode={{mode}}
target={{target}}
locale={{locale}}
article_title={{article_title}}
article_summary={{article_summary}}
text:
{{text}}`,
		ModeShorten: `请精简以下内容。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
locale={{locale}}
text:
{{text}}`,
		ModeExpand: `请扩写以下内容。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
locale={{locale}}
text:
{{text}}`,
		ModeTitleCandidates: `请基于文章上下文生成标题候选。最多 {{max_suggestions}} 个。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
article_title={{article_title}}
article_summary={{article_summary}}
article_excerpt:
{{article_excerpt}}`,
		ModeSummaryCandidates: `请基于文章上下文生成摘要候选。最多 {{max_suggestions}} 个。返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。
article_title={{article_title}}
article_summary={{article_summary}}
article_excerpt:
{{article_excerpt}}`,
	}
}

func PromptTemplateKeys() []string {
	keys := []string{ModeImprove, ModeShorten, ModeExpand, ModeTitleCandidates, ModeSummaryCandidates}
	sort.Strings(keys)
	return keys
}

func NormalizePromptTemplates(values map[string]string) map[string]string {
	defaults := DefaultPromptTemplates()
	normalized := make(map[string]string, len(defaults))
	for key, value := range defaults {
		normalized[key] = value
		if custom := strings.TrimSpace(values[key]); custom != "" {
			normalized[key] = custom
		}
	}
	return normalized
}

func RenderPromptTemplate(template string, data PromptRenderData) string {
	replacements := map[string]string{
		"{{mode}}":            data.Mode,
		"{{target}}":          data.Target,
		"{{text}}":            data.Text,
		"{{article_title}}":   data.ArticleTitle,
		"{{article_summary}}": data.ArticleSummary,
		"{{article_excerpt}}": data.ArticleExcerpt,
		"{{locale}}":          data.Locale,
		"{{max_suggestions}}": strconv.Itoa(data.MaxSuggestions),
	}
	rendered := template
	for token, value := range replacements {
		rendered = strings.ReplaceAll(rendered, token, value)
	}
	return rendered
}
