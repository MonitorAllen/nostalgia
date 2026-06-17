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
	RichText       string
	InputFormat    string
	ArticleTitle   string
	ArticleSummary string
	ArticleExcerpt string
	Locale         string
	MaxSuggestions int
}

func DefaultPromptTemplates() map[string]string {
	return map[string]string{
		ModeImprove: `你是专业中文内容编辑，请润色以下内容。必须只返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。

【正文富文本输出硬性要求】
- 当 target=content_selection 时，suggestions[].content 必须是可直接插入 CKEditor 的 HTML fragment 字符串。
- 不要返回完整 HTML 文档，不要返回 Markdown 代码围栏，不要把 HTML 当作可见源码展示。
- 如果 input_format=html 或 rich_text 非空，必须以 rich_text 作为主要输入，理解并保留/优化其中的结构。
- 如果 rich_text 包含 h1-h6、ul、ol、li、blockquote、table、pre、code、strong、em、u、a 等结构，不要扁平化成普通纯文本段落。
- 允许为了表达更清晰调整结构或拆分层级，例如把纯文本改写为列表、表格、引用，或增加加粗、斜体、下划线、行内代码等强调。
- 如果原文是列表，候选通常也应返回 <ul>/<ol><li>...</li></ul>；如果原文是标题加列表，候选通常也应返回标题加列表或更合适的富文本结构。
- content 字段中的 HTML 必须是片段，例如 <h2>标题</h2><ul><li>要点</li></ul>，不要只返回去掉标签后的纯文本，除非输入本身完全没有结构且纯文本就是最佳表达。

mode={{mode}}
target={{target}}
locale={{locale}}
article_title={{article_title}}
article_summary={{article_summary}}
input_format={{input_format}}
plain_text:
{{text}}
rich_text:
{{rich_text}}`,
		ModeShorten: `你是专业中文内容编辑，请精简以下内容。必须只返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。

【正文富文本输出硬性要求】
- 当 target=content_selection 时，suggestions[].content 必须是可直接插入 CKEditor 的 HTML fragment 字符串。
- 精简时保留必要的信息层级，不要扁平化标题、列表、引用、表格等富文本结构为普通纯文本。
- 如果 input_format=html 或 rich_text 非空，必须以 rich_text 作为主要输入，输出仍使用 HTML fragment。
- 如果原文包含 <ul>/<ol>/<li>，候选应继续使用列表结构，除非改成表格或分段确实更清晰。
- 允许删减冗余文字、合并重复项、压缩句子，但保留能提升阅读效率的加粗、斜体、下划线、行内代码等格式。
- 不要返回完整 HTML 文档，不要返回 Markdown 代码围栏，不要只返回去掉标签后的纯文本。

locale={{locale}}
target={{target}}
input_format={{input_format}}
plain_text:
{{text}}
rich_text:
{{rich_text}}`,
		ModeExpand: `你是专业中文内容编辑，请扩写以下内容。必须只返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。

【正文富文本输出硬性要求】
- 当 target=content_selection 时，suggestions[].content 必须是可直接插入 CKEditor 的 HTML fragment 字符串。
- 扩写时优先延续 rich_text 的原始结构；如果原文是标题、列表、引用、表格或代码块，不要扁平化或扩写成一整段纯文本。
- 允许补充解释、拆分层级、增加列表项、把内容组织成表格，或用加粗、斜体、下划线、行内代码突出重点。
- 如果 input_format=html 或 rich_text 非空，必须以 rich_text 作为主要输入，输出仍使用 HTML fragment。
- content 字段示例：<h2>标题</h2><ul><li><strong>要点</strong>：说明</li></ul>。
- 不要返回完整 HTML 文档，不要返回 Markdown 代码围栏，不要只返回去掉标签后的纯文本。

locale={{locale}}
target={{target}}
input_format={{input_format}}
plain_text:
{{text}}
rich_text:
{{rich_text}}`,
		ModeTitleCandidates: `请基于文章上下文生成标题候选。最多 {{max_suggestions}} 个。必须只返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。标题候选 content 必须是纯文本，不要包含 HTML 或 Markdown。
article_title={{article_title}}
article_summary={{article_summary}}
article_excerpt:
{{article_excerpt}}`,
		ModeSummaryCandidates: `请基于文章上下文生成摘要候选。最多 {{max_suggestions}} 个。必须只返回 JSON：{"suggestions":[{"content":"...","reason":"..."}]}。摘要候选 content 必须是纯文本，不要包含 HTML 或 Markdown。
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
		"{{rich_text}}":       data.RichText,
		"{{input_format}}":    data.InputFormat,
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
