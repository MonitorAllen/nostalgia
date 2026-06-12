import type { AdminArticle } from './types'

export const isAutomationDraft = (article: AdminArticle) =>
  Boolean(article.created_by_automation) &&
  article.automation_status === 'pending_review' &&
  !article.is_publish
