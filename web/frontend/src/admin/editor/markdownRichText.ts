import { sanitizeHtml } from '../../util/sanitizeHtml'
import { hasHtmlTags, looksLikeMarkdown } from './markdownDetection'

export const getSuggestionPreviewHtml = (content: string, editor?: any) => {
  const trimmed = content.trim()
  if (!trimmed) return ''

  if (hasHtmlTags(trimmed)) {
    return sanitizeHtml(trimmed)
  }

  if (!looksLikeMarkdown(trimmed)) {
    return sanitizeHtml(trimmed.replace(/\n/g, '<br>'))
  }

  const html = convertMarkdownToHtml(trimmed, editor)
  return sanitizeHtml(html)
}

export const insertSuggestionContent = (editor: any, content: string, replacementRange: any) => {
  const source = content.trim()
  if (!source) return

  const viewFragment =
    looksLikeMarkdown(source) && !hasHtmlTags(source)
      ? getMarkdownViewFragment(editor, source)
      : editor.data.processor.toView(source)
  const modelFragment = editor.data.toModel(viewFragment)

  editor.model.insertContent(modelFragment, replacementRange)
}

const convertMarkdownToHtml = (content: string, editor?: any) => {
  if (!editor) return content.replace(/\n/g, '<br>')

  const viewFragment = getMarkdownViewFragment(editor, content)
  return editor.data.processor.toData(viewFragment)
}

const getMarkdownViewFragment = (editor: any, content: string) => {
  const pasteFromMarkdown = editor.plugins.get('PasteFromMarkdownExperimental') as
    | { _gfmDataProcessor?: { toView: (data: string) => unknown } }
    | undefined
  const viewFragment = pasteFromMarkdown?._gfmDataProcessor?.toView(content)

  if (!viewFragment) {
    return editor.data.processor.toView(content)
  }

  return viewFragment
}
