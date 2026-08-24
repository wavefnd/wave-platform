import type { DocumentLocale } from './http'

export const documentLocales: Array<{ id: DocumentLocale; label: string }> = [
  { id: 'en', label: 'English' },
  { id: 'ko', label: '한국어' },
  { id: 'ja', label: '日本語' },
  { id: 'zh', label: '简体中文' },
  { id: 'es', label: 'Español' },
  { id: 'de', label: 'Deutsch' },
  { id: 'ru', label: 'Русский' },
  { id: 'id', label: 'Bahasa Indonesia · Melayu' },
  { id: 'vi', label: 'Tiếng Việt' },
]

const supported = new Set<DocumentLocale>(documentLocales.map((item) => item.id))

export function isDocumentLocale(value: unknown): value is DocumentLocale {
  return typeof value === 'string' && supported.has(value as DocumentLocale)
}

export function initialDocumentLocale(): DocumentLocale {
  const saved = localStorage.getItem('wave-doc-locale')
  if (isDocumentLocale(saved)) return saved
  const browser = navigator.language.toLowerCase().split('-')[0]
  if (browser === 'ms') return 'id'
  return isDocumentLocale(browser) ? browser : 'en'
}

export function saveDocumentLocale(locale: DocumentLocale) {
  localStorage.setItem('wave-doc-locale', locale)
}
