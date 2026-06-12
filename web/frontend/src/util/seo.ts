import type { Article } from '@/types/article'

export const SITE_NAME = 'Nostalgia'
export const DEFAULT_SITE_TITLE = 'Nostalgia | 技术文章与开发笔记'
export const DEFAULT_SITE_DESCRIPTION =
  'Nostalgia 记录 Go、前端、架构与个人项目实践，沉淀可公开阅读的技术文章。'
export const SEO_MANAGED_ATTR = 'data-managed'
export const SEO_MANAGED_VALUE = 'nostalgia-seo'

const PRIVATE_ROUTE_NAMES = new Set([
  'login',
  'register',
  'verifyEmail',
  'Forbidden',
  'NotFound',
  'setup',
  'adminLogin',
  'adminArticles',
  'adminArticleNew',
  'adminArticleEdit',
  'adminCategories',
  'adminNotFound'
])

export interface SeoOpenGraph {
  type: 'website' | 'article'
  title: string
  description: string
  url: string
  image?: string
}

export interface SeoMetadata {
  title: string
  description: string
  canonicalPath?: string
  canonicalUrl?: string
  robots?: string
  openGraph?: SeoOpenGraph
  twitterCard?: 'summary' | 'summary_large_image'
  jsonLd?: Record<string, unknown>
}

export type SeoHeadDescriptor =
  | { kind: 'title'; value: string }
  | { kind: 'meta'; keyAttr: 'name' | 'property'; key: string; content: string }
  | { kind: 'link'; rel: 'canonical'; href: string }
  | { kind: 'jsonLd'; value: Record<string, unknown> }

interface BuildSeoOptions {
  siteOrigin?: string
}

interface RouteSeoInput {
  name?: string | symbol | null
  path: string
  fullPath?: string
  query?: Record<string, unknown>
}

const fallbackOrigin = () => {
  if (typeof window !== 'undefined' && window.location?.origin) return window.location.origin
  return 'http://localhost'
}

export function resolveSiteOrigin(origin?: string, fallback = fallbackOrigin()): string {
  const source = origin?.trim() || fallback

  try {
    const url = new URL(source)
    return url.origin
  } catch {
    return fallback.replace(/\/+$/, '')
  }
}

export function getConfiguredSiteOrigin(): string {
  const env = (import.meta as unknown as { env?: Record<string, string | undefined> }).env
  return resolveSiteOrigin(env?.VITE_PUBLIC_SITE_URL)
}

export function toAbsoluteUrl(pathOrUrl: string | undefined, siteOrigin: string): string | undefined {
  if (!pathOrUrl) return undefined

  try {
    return new URL(pathOrUrl).toString()
  } catch {
    return new URL(pathOrUrl.startsWith('/') ? pathOrUrl : `/${pathOrUrl}`, resolveSiteOrigin(siteOrigin))
      .toString()
  }
}

export function stripHtmlToPlainText(value: string): string {
  return value
    .replace(/<script[\s\S]*?<\/script>/gi, ' ')
    .replace(/<style[\s\S]*?<\/style>/gi, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/gi, ' ')
    .replace(/&amp;/gi, '&')
    .replace(/&lt;/gi, '<')
    .replace(/&gt;/gi, '>')
    .replace(/&quot;/gi, '"')
    .replace(/&#39;/g, "'")
    .replace(/\s+/g, ' ')
    .trim()
}

export function truncateText(value: string, maxLength: number): string {
  const text = stripHtmlToPlainText(value)
  if (text.length <= maxLength) return text
  return `${text.slice(0, Math.max(0, maxLength - 1)).trimEnd()}…`
}

export function getCanonicalArticlePath(article: Pick<Article, 'id' | 'slug'>): string {
  const slug = article.slug?.trim()
  return `/article/${slug || article.id}`
}

function titleWithSiteName(title: string): string {
  const normalized = truncateText(title, 64)
  return normalized === SITE_NAME ? SITE_NAME : `${normalized} | ${SITE_NAME}`
}

function canonicalFromPath(path: string, siteOrigin: string): string {
  return toAbsoluteUrl(path, siteOrigin) || resolveSiteOrigin(siteOrigin)
}

export function buildArticleSeoMetadata(
  article: Article,
  options: BuildSeoOptions = {}
): SeoMetadata {
  const siteOrigin = resolveSiteOrigin(options.siteOrigin || getConfiguredSiteOrigin())
  const canonicalPath = getCanonicalArticlePath(article)
  const canonicalUrl = canonicalFromPath(canonicalPath, siteOrigin)
  const title = titleWithSiteName(article.title)
  const description = truncateText(article.summary || article.content, 150) || DEFAULT_SITE_DESCRIPTION
  const image = toAbsoluteUrl(article.cover, siteOrigin)

  return {
    title,
    description,
    canonicalPath,
    canonicalUrl,
    robots: article.is_publish ? 'index,follow' : 'noindex,nofollow',
    openGraph: {
      type: 'article',
      title,
      description,
      url: canonicalUrl,
      image
    },
    twitterCard: image ? 'summary_large_image' : 'summary',
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'BlogPosting',
      headline: article.title,
      description,
      image,
      author: {
        '@type': 'Person',
        name: article.username || SITE_NAME
      },
      datePublished: article.created_at,
      dateModified: article.updated_at || article.last_updated || article.created_at,
      articleSection: article.category_name,
      mainEntityOfPage: canonicalUrl
    }
  }
}

export function buildRouteSeoMetadata(route: RouteSeoInput, options: BuildSeoOptions = {}): SeoMetadata {
  const siteOrigin = resolveSiteOrigin(options.siteOrigin || getConfiguredSiteOrigin())
  const routeName = typeof route.name === 'string' ? route.name : ''
  const path = route.fullPath || route.path || '/'
  const canonicalPath = path.split('#')[0]
  const isPrivate = PRIVATE_ROUTE_NAMES.has(routeName) || canonicalPath.startsWith('/backend')
  const robots = isPrivate ? 'noindex,nofollow' : 'index,follow'

  if (routeName === 'home') {
    const canonicalUrl = canonicalFromPath('/', siteOrigin)
    return {
      title: DEFAULT_SITE_TITLE,
      description: DEFAULT_SITE_DESCRIPTION,
      canonicalPath: '/',
      canonicalUrl,
      robots,
      openGraph: {
        type: 'website',
        title: DEFAULT_SITE_TITLE,
        description: DEFAULT_SITE_DESCRIPTION,
        url: canonicalUrl
      },
      twitterCard: 'summary'
    }
  }

  const title = isPrivate ? `私有页面 | ${SITE_NAME}` : `文章索引 | ${SITE_NAME}`
  const canonicalUrl = canonicalFromPath(canonicalPath || '/', siteOrigin)
  return {
    title,
    description: DEFAULT_SITE_DESCRIPTION,
    canonicalPath: canonicalPath || '/',
    canonicalUrl,
    robots,
    openGraph: {
      type: 'website',
      title,
      description: DEFAULT_SITE_DESCRIPTION,
      url: canonicalUrl
    },
    twitterCard: 'summary'
  }
}

export function createSeoHeadDescriptors(metadata: SeoMetadata): SeoHeadDescriptor[] {
  const canonicalUrl =
    metadata.canonicalUrl ||
    (metadata.canonicalPath ? canonicalFromPath(metadata.canonicalPath, getConfiguredSiteOrigin()) : undefined)
  const descriptors: SeoHeadDescriptor[] = [
    { kind: 'title', value: metadata.title },
    { kind: 'meta', keyAttr: 'name', key: 'description', content: metadata.description },
    { kind: 'meta', keyAttr: 'name', key: 'robots', content: metadata.robots || 'index,follow' }
  ]

  if (canonicalUrl) {
    descriptors.push({ kind: 'link', rel: 'canonical', href: canonicalUrl })
  }

  if (metadata.openGraph) {
    descriptors.push(
      { kind: 'meta', keyAttr: 'property', key: 'og:type', content: metadata.openGraph.type },
      { kind: 'meta', keyAttr: 'property', key: 'og:title', content: metadata.openGraph.title },
      {
        kind: 'meta',
        keyAttr: 'property',
        key: 'og:description',
        content: metadata.openGraph.description
      },
      { kind: 'meta', keyAttr: 'property', key: 'og:url', content: metadata.openGraph.url }
    )
    if (metadata.openGraph.image) {
      descriptors.push({
        kind: 'meta',
        keyAttr: 'property',
        key: 'og:image',
        content: metadata.openGraph.image
      })
    }
  }

  descriptors.push(
    {
      kind: 'meta',
      keyAttr: 'name',
      key: 'twitter:card',
      content: metadata.twitterCard || 'summary'
    },
    { kind: 'meta', keyAttr: 'name', key: 'twitter:title', content: metadata.title },
    { kind: 'meta', keyAttr: 'name', key: 'twitter:description', content: metadata.description }
  )

  if (metadata.openGraph?.image) {
    descriptors.push({
      kind: 'meta',
      keyAttr: 'name',
      key: 'twitter:image',
      content: metadata.openGraph.image
    })
  }

  if (metadata.jsonLd) {
    descriptors.push({ kind: 'jsonLd', value: metadata.jsonLd })
  }

  return descriptors
}

function upsertMeta(documentRef: Document, keyAttr: 'name' | 'property', key: string, content: string) {
  const selector = `meta[${keyAttr}="${key}"][${SEO_MANAGED_ATTR}="${SEO_MANAGED_VALUE}"]`
  let element = documentRef.head.querySelector<HTMLMetaElement>(selector)
  if (!element) {
    element = documentRef.createElement('meta')
    element.setAttribute(keyAttr, key)
    element.setAttribute(SEO_MANAGED_ATTR, SEO_MANAGED_VALUE)
    documentRef.head.appendChild(element)
  }
  element.setAttribute('content', content)
}

function upsertCanonical(documentRef: Document, href: string) {
  const selector = `link[rel="canonical"][${SEO_MANAGED_ATTR}="${SEO_MANAGED_VALUE}"]`
  let element = documentRef.head.querySelector<HTMLLinkElement>(selector)
  if (!element) {
    element = documentRef.createElement('link')
    element.setAttribute('rel', 'canonical')
    element.setAttribute(SEO_MANAGED_ATTR, SEO_MANAGED_VALUE)
    documentRef.head.appendChild(element)
  }
  element.setAttribute('href', href)
}

function upsertJsonLd(documentRef: Document, value: Record<string, unknown>) {
  const selector = `script[type="application/ld+json"][${SEO_MANAGED_ATTR}="${SEO_MANAGED_VALUE}"]`
  let element = documentRef.head.querySelector<HTMLScriptElement>(selector)
  if (!element) {
    element = documentRef.createElement('script')
    element.setAttribute('type', 'application/ld+json')
    element.setAttribute(SEO_MANAGED_ATTR, SEO_MANAGED_VALUE)
    documentRef.head.appendChild(element)
  }
  element.textContent = JSON.stringify(value)
}

export function applySeoMetadata(metadata: SeoMetadata, documentRef = globalThis.document) {
  if (!documentRef?.head) return

  for (const descriptor of createSeoHeadDescriptors(metadata)) {
    if (descriptor.kind === 'title') {
      documentRef.title = descriptor.value
    } else if (descriptor.kind === 'meta') {
      upsertMeta(documentRef, descriptor.keyAttr, descriptor.key, descriptor.content)
    } else if (descriptor.kind === 'link') {
      upsertCanonical(documentRef, descriptor.href)
    } else if (descriptor.kind === 'jsonLd') {
      upsertJsonLd(documentRef, descriptor.value)
    }
  }
}
