const HTML_TAG_PATTERN = /<\/?[a-z][\s\S]*>/i
const MARKDOWN_BLOCK_PATTERN =
  /(^|\n)\s{0,3}(#{1,6}\s+\S|[-*+]\s+\S|\d+\.\s+\S|>\s+\S|```|~~~|\|.+\|)/
const MARKDOWN_INLINE_PATTERN = /(\*\*[^*\n]+?\*\*|__[^_\n]+?__|`[^`\n]+?`|\[[^\]\n]+?\]\([^)]+?\))/

export const hasHtmlTags = (content: string) => HTML_TAG_PATTERN.test(content)

export const looksLikeMarkdown = (content: string) => {
  const trimmed = content.trim()
  if (!trimmed || hasHtmlTags(trimmed)) return false

  return MARKDOWN_BLOCK_PATTERN.test(trimmed) || MARKDOWN_INLINE_PATTERN.test(trimmed)
}
