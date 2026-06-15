import type { AdminAIPolishMode, AdminAIPolishRequest, AdminAIPolishTarget } from '../types'

interface BuildAIPolishRequestInput {
  mode: AdminAIPolishMode
  target: AdminAIPolishTarget
  text: string
  articleId?: string
  articleTitle?: string
  articleSummary?: string
  articleExcerpt?: string
  locale?: string
  maxContextChars?: number
}

const DEFAULT_CONTEXT_CHARS = 4000

export type AIPolishSessionStatus = 'idle' | 'loading' | 'ready' | 'error'

export interface AIPolishSession {
  mode: AdminAIPolishMode
  target: AdminAIPolishTarget
  sourceText: string
  status: AIPolishSessionStatus
  selectedSuggestionIndex: number
}

const modeLabels: Record<AdminAIPolishMode, string> = {
  improve: '润色',
  shorten: '精简',
  expand: '扩写',
  title_candidates: '标题候选',
  summary_candidates: '摘要候选'
}

export function normalizeSelectedText(value: string) {
  return value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .join('\n')
    .trim()
}

export function truncateForAIPolish(value = '', maxChars = DEFAULT_CONTEXT_CHARS) {
  if (maxChars <= 0) return ''
  return Array.from(value).slice(0, maxChars).join('')
}

export function getAIPolishModeLabel(mode: AdminAIPolishMode) {
  return modeLabels[mode]
}

export function createAIPolishSession(input: {
  mode: AdminAIPolishMode
  target: AdminAIPolishTarget
  sourceText: string
}): AIPolishSession {
  return {
    mode: input.mode,
    target: input.target,
    sourceText: input.sourceText,
    status: 'loading',
    selectedSuggestionIndex: -1
  }
}

export function getAIPolishApplyLabel(target: AdminAIPolishTarget) {
  if (target === 'title') return '应用到标题'
  if (target === 'summary') return '应用到摘要'
  return '替换选区'
}

export function buildAIPolishRequest(input: BuildAIPolishRequestInput): AdminAIPolishRequest {
  const maxContextChars = input.maxContextChars ?? DEFAULT_CONTEXT_CHARS

  return {
    mode: input.mode,
    target: input.target,
    text: normalizeSelectedText(input.text),
    ...(input.articleId ? { article_id: input.articleId } : {}),
    article_title: truncateForAIPolish(input.articleTitle || '', maxContextChars),
    article_summary: truncateForAIPolish(input.articleSummary || '', maxContextChars),
    article_excerpt: truncateForAIPolish(input.articleExcerpt || '', maxContextChars),
    locale: input.locale || 'zh-CN'
  }
}
