package ai

import "fmt"

func BuildMessages(req PolishRequest) []ChatMessage {
	req = req.normalized()
	system := `You are Nostalgia's owner-only writing assistant. Return strict JSON only, shaped as {"suggestions":[{"content":"...","reason":"..."}]}. Keep the original language, preserve technical terms, code identifiers, commands, URLs, and quoted code. Do not invent facts, benchmark numbers, dependency versions, citations, or links.`

	user := fmt.Sprintf(
		"locale: %s\nmode: %s\ntarget: %s\narticle_title: %s\narticle_summary: %s\narticle_excerpt: %s\ntext:\n%s",
		req.Locale,
		req.Mode,
		req.Target,
		req.ArticleTitle,
		req.ArticleSummary,
		req.ArticleExcerpt,
		req.Text,
	)

	return []ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}
