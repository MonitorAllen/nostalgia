import DOMPurify from 'dompurify'

const allowedUri = /^(?:(?:https?|mailto|tel):|\/|#)/i

DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (!(node instanceof HTMLElement)) return

  if (node.hasAttribute('target')) {
    node.setAttribute('rel', 'noopener noreferrer')
  }

  for (const attr of ['href', 'src']) {
    const value = node.getAttribute(attr)
    if (value && !allowedUri.test(value)) node.removeAttribute(attr)
  }
})

export const sanitizeHtml = (html: string) =>
  DOMPurify.sanitize(html, {
    ADD_ATTR: ['target']
  })
