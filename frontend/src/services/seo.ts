import type { RouteLocationNormalizedLoaded } from 'vue-router'

import type { Locale } from '../i18n'

const descriptions = {
  en: {
    portal: 'The official home for the Wave programming language, documentation, releases, community, questions, source, and mail.',
	blog: 'Official Wave programming language news, engineering updates, and release articles.',
    docs: 'Official Wave programming language guides and reference documentation.',
    community: 'Wave programming language community posts and technical discussions.',
    questions: 'Technical questions and answers about the Wave programming language.',
	rfc: 'Public design proposals and decisions for significant changes to the Wave language and platform.',
    source: 'Read-only source browser for official Wave Git mirrors.',
    mail: 'Wave Mail.', account: 'Wave account security and authentication.', admin: 'Wave Platform administration.',
  },
  ko: {
    portal: 'Wave 프로그래밍 언어의 공식 문서, 릴리즈, 커뮤니티, 질문, 소스와 메일을 제공하는 플랫폼입니다.',
	blog: 'Wave 프로그래밍 언어의 공식 소식, 개발 이야기와 릴리즈 글입니다.',
    docs: 'Wave 프로그래밍 언어의 공식 안내서와 레퍼런스 문서입니다.',
    community: 'Wave 프로그래밍 언어 커뮤니티 글과 기술 토론입니다.',
    questions: 'Wave 프로그래밍 언어에 관한 기술 질문과 답변입니다.',
	rfc: 'Wave 언어와 플랫폼의 중요한 변경을 검토하고 결정하는 공개 제안서입니다.',
    source: 'Wave 공식 Git 미러를 탐색하는 읽기 전용 소스 브라우저입니다.',
    mail: 'Wave Mail입니다.', account: 'Wave 계정 보안과 인증입니다.', admin: 'Wave Platform 관리입니다.',
  },
} as const

const labels = { portal: 'Wave', blog: 'Blog', docs: 'Documentation', community: 'Community', questions: 'Questions', rfc: 'RFC', source: 'Source', mail: 'Mail', account: 'Account', admin: 'Administration', personal: 'LunaStev' } as const
type Service = keyof typeof labels

const nonIndexableRoutes = new Set([
  'login', 'register', 'account-recover', 'verify-recovery', 'account-security', 'admin',
  'search', 'not-found', 'community-new', 'personal-space-new', 'question-new',
	'rfc-new', 'rfc-edit',
])

interface PageSEO {
  title: string
  description: string
  locale: string
  path?: string
  schema?: Record<string, unknown>
  breadcrumbs?: Array<{ name: string, path: string }>
  image?: string
  imageAlt?: string
  noIndex?: boolean
}

function upsertMeta(selector: string, attributes: Record<string, string>) {
  let element = document.head.querySelector<HTMLMetaElement>(selector)
  if (!element) { element = document.createElement('meta'); document.head.append(element) }
  for (const [name, value] of Object.entries(attributes)) element.setAttribute(name, value)
}

function upsertLink(selector: string, attributes: Record<string, string>) {
  let element = document.head.querySelector<HTMLLinkElement>(selector)
  if (!element) { element = document.createElement('link'); document.head.append(element) }
  for (const [name, value] of Object.entries(attributes)) element.setAttribute(name, value)
}

function optionalMeta(selector: string, attributes?: Record<string, string>) {
  const existing = document.head.querySelector<HTMLMetaElement>(selector)
  if (!attributes) { existing?.remove(); return }
  upsertMeta(selector, attributes)
}

function schemaType(service: Service, detail: boolean): string {
  if (service === 'docs') return detail ? 'TechArticle' : 'CollectionPage'
	if (service === 'source') return detail ? 'SoftwareSourceCode' : 'CollectionPage'
	if (service === 'blog') return detail ? 'BlogPosting' : 'CollectionPage'
  if (['community', 'questions', 'rfc', 'personal'].includes(service)) return detail ? 'WebPage' : 'CollectionPage'
  return 'WebPage'
}

function canonicalURL(pathname: string): string {
  const value = new URL(pathname || '/', window.location.origin)
  value.search = ''
  value.hash = ''
  return value.toString()
}

function absoluteURL(value: string): string {
  if (!value.trim()) return ''
  try {
    const parsed = new URL(value, window.location.origin)
    return ['http:', 'https:'].includes(parsed.protocol) ? parsed.toString() : ''
  } catch {
    return ''
  }
}

function schemaImage(schema?: Record<string, unknown>): string {
  const image = schema?.image
  if (typeof image === 'string') return absoluteURL(image)
  if (Array.isArray(image)) {
    const first = image.find((value) => typeof value === 'string')
    return typeof first === 'string' ? absoluteURL(first) : ''
  }
  if (image && typeof image === 'object' && typeof (image as Record<string, unknown>).url === 'string') {
    return absoluteURL(String((image as Record<string, unknown>).url))
  }
  return ''
}

export function applyPageSEO(options: PageSEO) {
  const canonical = canonicalURL(options.path ?? window.location.pathname)
  const robots = options.noIndex ? 'noindex, nofollow, noarchive' : 'index, follow, max-image-preview:large'
  const schemaName = String(options.schema?.['@type'] ?? '')
  const openGraphType = ['Article', 'TechArticle', 'BlogPosting', 'DiscussionForumPosting'].includes(schemaName) ? 'article' : 'website'
  const image = absoluteURL(options.image ?? '') || schemaImage(options.schema)
  document.title = options.title
  document.documentElement.lang = options.locale
  upsertMeta('meta[name="description"]', { name: 'description', content: options.description })
  upsertMeta('meta[name="robots"]', { name: 'robots', content: robots })
  upsertMeta('meta[property="og:title"]', { property: 'og:title', content: options.title })
  upsertMeta('meta[property="og:description"]', { property: 'og:description', content: options.description })
  upsertMeta('meta[property="og:type"]', { property: 'og:type', content: openGraphType })
  upsertMeta('meta[property="og:url"]', { property: 'og:url', content: canonical })
  upsertMeta('meta[property="og:site_name"]', { property: 'og:site_name', content: 'Wave' })
  const openGraphLocales: Record<string, string> = {
    en: 'en_US', ko: 'ko_KR', ja: 'ja_JP', zh: 'zh_CN', es: 'es_ES',
    de: 'de_DE', ru: 'ru_RU', id: 'id_ID', vi: 'vi_VN',
  }
  upsertMeta('meta[property="og:locale"]', { property: 'og:locale', content: openGraphLocales[options.locale] ?? 'en_US' })
  optionalMeta('meta[property="og:image"]', image ? { property: 'og:image', content: image } : undefined)
  optionalMeta('meta[property="og:image:alt"]', image ? { property: 'og:image:alt', content: options.imageAlt || options.title } : undefined)
  upsertMeta('meta[name="twitter:card"]', { name: 'twitter:card', content: image ? 'summary_large_image' : 'summary' })
  upsertMeta('meta[name="twitter:title"]', { name: 'twitter:title', content: options.title })
  upsertMeta('meta[name="twitter:description"]', { name: 'twitter:description', content: options.description })
  optionalMeta('meta[name="twitter:image"]', image ? { name: 'twitter:image', content: image } : undefined)
  optionalMeta('meta[name="twitter:image:alt"]', image ? { name: 'twitter:image:alt', content: options.imageAlt || options.title } : undefined)
  const published = typeof options.schema?.datePublished === 'string' ? options.schema.datePublished : ''
  const modified = typeof options.schema?.dateModified === 'string' ? options.schema.dateModified : ''
  const section = typeof options.schema?.articleSection === 'string' ? options.schema.articleSection : ''
  const author = options.schema?.author && typeof options.schema.author === 'object'
    ? String((options.schema.author as Record<string, unknown>).name ?? '') : ''
  optionalMeta('meta[property="article:published_time"]', published ? { property: 'article:published_time', content: published } : undefined)
  optionalMeta('meta[property="article:modified_time"]', modified ? { property: 'article:modified_time', content: modified } : undefined)
  optionalMeta('meta[property="article:section"]', section ? { property: 'article:section', content: section } : undefined)
  optionalMeta('meta[property="article:author"]', author ? { property: 'article:author', content: author } : undefined)
  upsertLink('link[rel="canonical"]', { rel: 'canonical', href: canonical })

  let script = document.head.querySelector<HTMLScriptElement>('script[data-wave-schema]')
  if (options.noIndex || !options.schema) {
    script?.remove()
    return
  }
  if (!script) { script = document.createElement('script'); script.type = 'application/ld+json'; script.dataset.waveSchema = 'true'; document.head.append(script) }
  const organization = {
    '@type': 'Organization', '@id': `${window.location.origin}/#organization`, name: 'Wave Foundation', url: `${window.location.origin}/`,
    logo: { '@type': 'ImageObject', url: new URL('/img/wave-logo.ico', window.location.origin).toString(), width: 256, height: 256 },
    sameAs: ['https://github.com/wavefnd', 'https://discord.gg/3nev5nHqq9', 'https://opencollective.com/wave-lang'],
  }
  const website = {
    '@type': 'WebSite', '@id': `${window.location.origin}/#website`, name: 'Wave Programming Language', url: `${window.location.origin}/`,
    publisher: { '@id': organization['@id'] }, inLanguage: ['en', 'ko', 'ja', 'zh', 'es', 'de', 'ru', 'id', 'vi'],
  }
  const articleSchema = ['Article', 'TechArticle', 'BlogPosting'].includes(schemaName)
  const page: Record<string, unknown> = {
    ...(articleSchema ? {} : options.schema), '@type': articleSchema ? 'WebPage' : schemaName || 'WebPage',
    '@id': `${canonical}#webpage`, url: canonical, name: options.title, description: options.description,
    isPartOf: { '@id': website['@id'] }, inLanguage: options.locale,
  }
  const graph: Array<Record<string, unknown>> = [organization, website, page]
  if (options.breadcrumbs && options.breadcrumbs.length >= 2) {
    const breadcrumb = {
      '@type': 'BreadcrumbList', '@id': `${canonical}#breadcrumb`,
      itemListElement: options.breadcrumbs.map((item, index) => ({
        '@type': 'ListItem', position: index + 1, name: item.name, item: canonicalURL(item.path),
      })),
    }
    graph.push(breadcrumb)
    page.breadcrumb = { '@id': breadcrumb['@id'] }
  }
  if (articleSchema && options.schema) {
    const article = {
      ...options.schema, '@id': `${canonical}#article`, url: canonical,
      mainEntityOfPage: { '@id': page['@id'] },
      publisher: options.schema.publisher ?? { '@id': organization['@id'] },
      image: image ? [image] : options.schema.image,
    }
    if (!article.image) delete article.image
    graph.push(article)
    page.mainEntity = { '@id': article['@id'] }
  }
  script.textContent = JSON.stringify({ '@context': 'https://schema.org', '@graph': graph })
}

export function updateSEO(route: RouteLocationNormalizedLoaded, locale: Locale, service: Service) {
  const localDescriptions = descriptions[locale]
  const description = localDescriptions[service === 'personal' ? 'community' : service] ?? localDescriptions.portal
  const suffix = labels[service]
  const title = service === 'portal' && route.name === 'home' ? 'Wave Programming Language' : `${suffix} · Wave`
  const privateService = ['mail', 'account', 'admin'].includes(service)
  const noIndex = privateService || nonIndexableRoutes.has(String(route.name ?? ''))
  const detail = Boolean(route.params.pathMatch || route.params.thread || route.params.question || route.params.number || route.params.repository || route.params.slug)
  applyPageSEO({ title, description, locale, path: route.path, noIndex, schema: { '@type': schemaType(service, detail) } })
}

export function plainTextDescription(value: string, fallback: string): string {
  const text = plainText(value)
  return (text || fallback).slice(0, 180)
}

export function plainText(value: string): string {
  return value
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/[`*_>#\[\]()-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

export function firstMarkdownImage(value: string): { url: string, alt: string } | null {
  const match = value.match(/!\[([^\]]*)\]\((?:<([^>]+)>|([^\s)]+))(?:\s+["'][^)]*["'])?\)/)
  if (!match) return null
  const url = absoluteURL(match[2] || match[3] || '')
  return url ? { url, alt: match[1].trim() } : null
}
