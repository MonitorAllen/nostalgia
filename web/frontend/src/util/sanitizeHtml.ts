import DOMPurify from 'dompurify'

import {
  getSafeTargetRel,
  getSanitizedHtmlConfig,
  isAllowedSanitizedUri,
  type SanitizedHtmlProfile
} from './sanitizeHtmlPolicy'

export interface SanitizeHtmlOptions {
  profile?: SanitizedHtmlProfile
}

DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (!(node instanceof HTMLElement)) return

  if (node.hasAttribute('target')) {
    node.setAttribute('rel', getSafeTargetRel(node.getAttribute('rel') || ''))
  }

  for (const attr of ['href', 'src']) {
    const value = node.getAttribute(attr)
    if (value && !isAllowedSanitizedUri(value)) node.removeAttribute(attr)
  }
})

export const sanitizeHtml = (html: string, options: SanitizeHtmlOptions = {}) =>
  DOMPurify.sanitize(html, getSanitizedHtmlConfig(options.profile))
