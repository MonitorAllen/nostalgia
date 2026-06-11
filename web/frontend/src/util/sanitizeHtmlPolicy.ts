import type { Config } from 'dompurify'

export type SanitizedHtmlProfile = 'article' | 'comment'

export const SANITIZED_HTML_CONFIG: Config = {
  ADD_ATTR: ['target'],
  FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed']
}

const COMMENT_SANITIZED_HTML_CONFIG: Config = {
  ...SANITIZED_HTML_CONFIG,
  FORBID_ATTR: ['style']
}

const cloneConfig = (config: Config): Config => ({
  ...config,
  ADD_ATTR: Array.isArray(config.ADD_ATTR) ? [...config.ADD_ATTR] : config.ADD_ATTR,
  FORBID_TAGS: config.FORBID_TAGS ? [...config.FORBID_TAGS] : undefined,
  FORBID_ATTR: config.FORBID_ATTR ? [...config.FORBID_ATTR] : undefined
})

export const getSanitizedHtmlConfig = (profile: SanitizedHtmlProfile = 'article') =>
  cloneConfig(profile === 'comment' ? COMMENT_SANITIZED_HTML_CONFIG : SANITIZED_HTML_CONFIG)

const allowedProtocols = new Set(['http:', 'https:', 'mailto:', 'tel:'])

const normalizeUriForProtocolCheck = (uri: string) =>
  uri.trim().replace(/[\u0000-\u001F\u007F\s]+/g, '')

export const isAllowedSanitizedUri = (uri: string) => {
  const normalizedUri = normalizeUriForProtocolCheck(uri)

  if (normalizedUri.startsWith('/') || normalizedUri.startsWith('#')) return true

  const protocolMatch = normalizedUri.match(/^([a-z][a-z0-9+.-]*:)/i)
  if (!protocolMatch) return false

  return allowedProtocols.has(protocolMatch[1].toLowerCase())
}

export const getSafeTargetRel = (rel = '') => {
  const values = new Set(rel.split(/\s+/).filter(Boolean))

  values.add('noopener')
  values.add('noreferrer')

  return [...values].join(' ')
}
