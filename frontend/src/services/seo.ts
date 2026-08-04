import type { RouteLocationNormalizedLoaded } from 'vue-router'

import type { Locale } from '../i18n'

const descriptions = {
  en: {
    portal: 'The official home for the Wave programming language, documentation, releases, community, questions, source, and mail.',
    docs: 'Official Wave programming language guides and reference documentation.',
    community: 'Wave programming language community posts and technical discussions.',
    questions: 'Technical questions and answers about the Wave programming language.',
    source: 'Read-only source browser for official Wave Git mirrors.',
    mail: 'Wave Mail.', account: 'Wave account security and authentication.', admin: 'Wave Platform administration.',
  },
  ko: {
    portal: 'Wave 프로그래밍 언어의 공식 문서, 릴리즈, 커뮤니티, 질문, 소스와 메일을 제공하는 플랫폼입니다.',
    docs: 'Wave 프로그래밍 언어의 공식 안내서와 레퍼런스 문서입니다.',
    community: 'Wave 프로그래밍 언어 커뮤니티 글과 기술 토론입니다.',
    questions: 'Wave 프로그래밍 언어에 관한 기술 질문과 답변입니다.',
    source: 'Wave 공식 Git 미러를 탐색하는 읽기 전용 소스 브라우저입니다.',
    mail: 'Wave Mail입니다.', account: 'Wave 계정 보안과 인증입니다.', admin: 'Wave Platform 관리입니다.',
  },
} as const

const labels = { portal: 'Wave', docs: 'Documentation', community: 'Community', questions: 'Questions', source: 'Source', mail: 'Mail', account: 'Account', admin: 'Administration', personal: 'LunaStev' } as const
type Service = keyof typeof labels

const nonIndexableRoutes = new Set([
  'login', 'register', 'account-recover', 'verify-recovery', 'account-security', 'admin',
  'search', 'not-found', 'community-new', 'personal-space-new', 'question-new',
])

interface PageSEO {
  title: string
  description: string
  locale: Locale
  path?: string
  schema?: Record<string, unknown>
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

function schemaType(service: Service, detail: boolean): string {
  if (service === 'docs') return detail ? 'TechArticle' : 'CollectionPage'
  if (service === 'source') return detail ? 'SoftwareSourceCode' : 'CollectionPage'
  if (['community', 'questions', 'personal'].includes(service)) return detail ? 'WebPage' : 'CollectionPage'
  return 'WebPage'
}

function canonicalURL(pathname: string): string {
  const value = new URL(pathname || '/', window.location.origin)
  value.search = ''
  value.hash = ''
  return value.toString()
}

export function applyPageSEO(options: PageSEO) {
  const canonical = canonicalURL(options.path ?? window.location.pathname)
  const robots = options.noIndex ? 'noindex, nofollow, noarchive' : 'index, follow, max-image-preview:large'
  const schemaName = String(options.schema?.['@type'] ?? '')
  const openGraphType = ['Article', 'TechArticle', 'BlogPosting', 'DiscussionForumPosting'].includes(schemaName) ? 'article' : 'website'
  document.title = options.title
  document.documentElement.lang = options.locale
  upsertMeta('meta[name="description"]', { name: 'description', content: options.description })
  upsertMeta('meta[name="robots"]', { name: 'robots', content: robots })
  upsertMeta('meta[property="og:title"]', { property: 'og:title', content: options.title })
  upsertMeta('meta[property="og:description"]', { property: 'og:description', content: options.description })
  upsertMeta('meta[property="og:type"]', { property: 'og:type', content: openGraphType })
  upsertMeta('meta[property="og:url"]', { property: 'og:url', content: canonical })
  upsertMeta('meta[property="og:site_name"]', { property: 'og:site_name', content: 'Wave' })
  upsertMeta('meta[property="og:locale"]', { property: 'og:locale', content: options.locale === 'ko' ? 'ko_KR' : 'en_US' })
  upsertMeta('meta[name="twitter:card"]', { name: 'twitter:card', content: 'summary' })
  upsertMeta('meta[name="twitter:title"]', { name: 'twitter:title', content: options.title })
  upsertMeta('meta[name="twitter:description"]', { name: 'twitter:description', content: options.description })
  upsertLink('link[rel="canonical"]', { rel: 'canonical', href: canonical })

  let script = document.head.querySelector<HTMLScriptElement>('script[data-wave-schema]')
  if (options.noIndex || !options.schema) {
    script?.remove()
    return
  }
  if (!script) { script = document.createElement('script'); script.type = 'application/ld+json'; script.dataset.waveSchema = 'true'; document.head.append(script) }
  const organization = {
    '@type': 'Organization', '@id': `${window.location.origin}/#organization`, name: 'Wave Foundation', url: window.location.origin,
    logo: { '@type': 'ImageObject', url: new URL('/img/wave-logo.ico', window.location.origin).toString(), width: 256, height: 256 },
  }
  const website = {
    '@type': 'WebSite', '@id': `${window.location.origin}/#website`, name: 'Wave Programming Language', url: window.location.origin,
    publisher: { '@id': organization['@id'] }, inLanguage: ['en', 'ko'],
    potentialAction: { '@type': 'SearchAction', target: `${window.location.origin}/search?q={search_term_string}`, 'query-input': 'required name=search_term_string' },
  }
  const page = {
    ...options.schema, '@id': `${canonical}#page`, url: canonical, name: options.title, description: options.description,
    isPartOf: { '@id': website['@id'] }, inLanguage: options.locale, publisher: { '@id': organization['@id'] },
  }
  script.textContent = JSON.stringify({ '@context': 'https://schema.org', '@graph': [organization, website, page] })
}

export function updateSEO(route: RouteLocationNormalizedLoaded, locale: Locale, service: Service) {
  const localDescriptions = descriptions[locale]
  const description = localDescriptions[service === 'personal' ? 'community' : service] ?? localDescriptions.portal
  const suffix = labels[service]
  const title = service === 'portal' && route.name === 'home' ? 'Wave Programming Language' : `${suffix} · Wave`
  const privateService = ['mail', 'account', 'admin'].includes(service)
  const noIndex = privateService || nonIndexableRoutes.has(String(route.name ?? ''))
  const detail = Boolean(route.params.pathMatch || route.params.thread || route.params.question || route.params.repository || route.params.slug)
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
