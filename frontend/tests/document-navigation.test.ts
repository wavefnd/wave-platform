import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
import { homeDocumentationLinks, firstWaveProgram } from '../src/services/homeOnboarding.ts'
import { documentLocales, isDocumentLocale, initialDocumentLocale } from '../src/services/documentLocale.ts'
import { canonicalDocumentPath, documentationCatalog, documentationPath, documentationProject, projectDocuments } from '../src/services/documentNavigation.ts'

test('URL selects the project and separates catalogue from document paths', () => {
  for (const path of ['', 'language/types', 'toolchain/whale-overview', 'whales/overview']) {
    assert.equal(documentationProject(path), 'wave')
    assert.equal(documentationPath(path), path)
  }
  assert.equal(documentationProject('whale'), 'whale')
  assert.equal(documentationPath('whale'), '')
  assert.equal(documentationProject('whale/overview'), 'whale')
  assert.equal(documentationPath('whale/overview'), 'whale/overview')
  for (const locale of documentLocales.map(({ id }) => id)) {
    assert.equal(documentationCatalog(locale, 'wave'), `/docs/${locale}`)
    assert.equal(documentationCatalog(locale, 'whale'), `/docs/${locale}/whale`)
  }
})

test('navigation merges translations before sorting within the selected project', () => {
  const english = [
    { path: 'language/types', title: 'Wave', groupOrder: 1, order: 1 },
    { path: 'toolchain/whale-overview', title: 'Legacy Wave path', groupOrder: 4, order: 1 },
    { path: 'whale/symbols', title: 'Symbols', groupOrder: 5, order: 4 },
    { path: 'whale/overview', title: 'Overview', groupOrder: 5, order: 1 },
    { path: 'whale/alignment', title: 'Alignment', groupOrder: 5, order: 4 },
    { path: 'whale/other', title: 'Other group', groupOrder: 6, order: 1 },
  ]
  const translated = [{ ...english[2], title: '심볼' }]
  const whale = projectDocuments(translated, english, 'whale')
  assert.deepEqual(whale.map(item => item.path), ['whale/overview', 'whale/alignment', 'whale/symbols', 'whale/other'])
  assert.equal(whale[2].title, '심볼')
  assert.deepEqual(projectDocuments(translated, english, 'wave').map(item => item.path), ['language/types', 'toolchain/whale-overview'])
  // Search and previous/next consume this same scoped list.
  assert.equal(whale.filter(item => item.title.includes('Wave')).length, 0)
  assert.equal(whale[whale.findIndex(item => item.path === 'whale/symbols') - 1].path, 'whale/alignment')
  assert.deepEqual(projectDocuments([], [], 'whale'), [])
  assert.equal(english[2].title, 'Symbols')
})

test('standard library has an isolated catalogue, preserves old URLs and translation fallback', () => {
  const legacy = ['reference/standard-library', 'reference/string-and-bytes', 'reference/memory-and-buffer', 'reference/system-io-network-process']
  for (const path of [...legacy, 'stdlib/buffer']) {
    assert.equal(documentationProject(path), 'stdlib')
    assert.equal(documentationPath(path), path)
  }
  assert.equal(documentationProject('stdlib'), 'stdlib')
  assert.equal(documentationPath('stdlib'), '')
  assert.equal(documentationProject('stdlib-extra/page'), 'wave')
  for (const locale of ['ko', 'en', 'ja']) assert.equal(documentationCatalog(locale, 'stdlib'), `/docs/${locale}/stdlib`)
  const english = [
    { path: legacy[0], title: 'Library', groupOrder: 1, order: 1 },
    { path: 'stdlib/buffer', title: 'Buffer', groupOrder: 1, order: 2 },
    { path: 'language/types', title: 'Types', groupOrder: 3, order: 1 },
    { path: 'whale/overview', title: 'Whale', groupOrder: 5, order: 1 },
  ]
  const translated = [{ ...english[0], title: '표준 라이브러리' }]
  assert.deepEqual(projectDocuments(translated, english, 'stdlib').map(item => item.title), ['표준 라이브러리', 'Buffer'])
  assert.deepEqual(projectDocuments(translated, english, 'wave').map(item => item.path), ['language/types'])
  assert.deepEqual(projectDocuments([], [], 'stdlib'), [])
})


test('retired lesson and toolchain URLs point at the single canonical collection', () => {
  assert.equal(canonicalDocumentPath('learn/functions'), 'language/functions-and-generics')
  assert.equal(canonicalDocumentPath('learn/strings'), 'language/strings')
  assert.equal(canonicalDocumentPath('toolchain/build-link-targets'), 'whale/build-link-targets')
  assert.equal(canonicalDocumentPath('getting-started/design-goals'), 'getting-started/overview')
  assert.equal(canonicalDocumentPath('language/control-flow'), 'language/control-flow')
  assert.equal(documentationProject(canonicalDocumentPath('toolchain/whale-cli')), 'whale')
})


test('new documentation locales accept language codes and normalize browser regions', (t) => {
  const originalStorage = Object.getOwnPropertyDescriptor(globalThis, 'localStorage')
  const originalNavigator = Object.getOwnPropertyDescriptor(globalThis, 'navigator')
  t.after(() => {
    for (const [key, descriptor] of [['localStorage', originalStorage], ['navigator', originalNavigator]] as const) {
      if (descriptor) Object.defineProperty(globalThis, key, descriptor)
      else Reflect.deleteProperty(globalThis, key)
    }
  })
  Object.defineProperty(globalThis, 'localStorage', { configurable: true, value: { getItem: () => null } })
  for (const [code, device] of [['pt', 'pt-BR'], ['fr', 'fr-CA'], ['pl', 'pl-PL'], ['nl', 'nl-BE'], ['tr', 'tr-TR'], ['it', 'it-IT']]) {
    Object.defineProperty(globalThis, 'navigator', { configurable: true, value: { language: device } })
    assert.equal(isDocumentLocale(code), true)
    assert.equal(isDocumentLocale(device), false)
    assert.equal(initialDocumentLocale(), code)
    assert.equal(documentationCatalog(code, 'stdlib'), `/docs/${code}/stdlib`)
  }
  Object.defineProperty(globalThis, 'localStorage', { configurable: true, value: { getItem: () => 'pt' } })
  assert.equal(initialDocumentLocale(), 'pt')
})


test('home onboarding links select installation, the first lesson, and the course index', () => {
  assert.deepEqual(homeDocumentationLinks.map(({ label, path }) => [label, path]), [
    ['docs.installation', 'getting-started/install'],
    ['docs.firstProgram', 'language/program-structure'],
    ['docs.language', 'getting-started/overview'],
  ])
  for (const locale of ['en', 'ko', 'ja', 'zh']) {
    for (const { path } of homeDocumentationLinks) {
      assert.equal(canonicalDocumentPath(path), path)
      const document = readFileSync(new URL(`../../wavedoc/${locale}/${path}.md`, import.meta.url), 'utf8')
      assert.ok(document.includes(`path: ${path}\n`), `${locale}/${path} must be a published document`)
    }
  }
})

test('home example stays aligned with the complete first-lesson program', () => {
  for (const locale of ['en', 'ko']) {
    const lesson = readFileSync(new URL(`../../wavedoc/${locale}/language/program-structure.md`, import.meta.url), 'utf8')
    const example = lesson.match(/<!-- wave-example: book-hello -->\s*```wave\n([\s\S]*?)\n```/)
    assert.equal(example?.[1], firstWaveProgram)
  }
})
