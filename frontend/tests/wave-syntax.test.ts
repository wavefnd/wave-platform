import assert from 'node:assert/strict'
import { test } from 'node:test'
import { languageForPath, normalizeFenceLanguage, Prism } from '../src/components/source/syntax.ts'
import { tokenizeWave, highlightWave } from '../src/services/waveSyntax.ts'

function kindOf(source: string, word: string) {
  const offset = source.indexOf(word)
  assert.ok(offset >= 0)
  let cursor = 0
  for (const segment of tokenizeWave(source)) {
    cursor += segment.text.length
    if (cursor > offset) return segment.kind
  }
}

test('new keywords, explicit memory types and larger integers', () => {
  const source = '#[target(os = "linux")]\npub async fun work(value: ptr<array<u1024, 4>>) -> ! { await next(); }\nvariant Result { Value(isz), Error(usz) }'
  for (const word of ['pub', 'async', 'fun', 'await', 'variant']) assert.equal(kindOf(source, word), 'keyword')
  for (const word of ['ptr', 'array', 'u1024', 'isz', 'usz']) assert.equal(kindOf(source, word), 'builtin')
  assert.equal(kindOf(source, '#'), 'punctuation')
  assert.equal(kindOf(source, '!'), 'operator')
  assert.equal(kindOf(source, '"linux"'), 'string')
  for (const type of ['byte', 'char', 'bool', 'void', 'str', 'f32', 'f64', 'i128', 'i256', 'i512']) assert.equal(kindOf(type, type), 'builtin')
})

test('complete identifiers never acquire keyword or numeric fragments', () => {
  const source = 'async_task awaitable pub123 var한글 한글var item42 i1024_suffix'
  assert.deepEqual(tokenizeWave(source), [{ kind: '', text: source }])
})

test('nested comments and escaped strings preserve boundaries', () => {
  const source = '/* outer /* inner */ var still_comment */ var name: str = "\\\" // <tag>"; // comment\rawait work();'
  assert.equal(kindOf(source, 'still_comment'), 'comment')
  assert.equal(kindOf(source, 'var name'), 'keyword')
  assert.equal(kindOf(source, '<tag>'), 'string')
  assert.equal(kindOf(source, 'await'), 'keyword')
  assert.deepEqual(tokenizeWave('/* unfinished /* nested */'), [{ kind: 'comment', text: '/* unfinished /* nested */' }])
  for (const quote of ['"', "'"]) {
    assert.equal(kindOf(`${quote}unfinished\\\nvar next: i32;`, 'var'), 'keyword')
  }
})

test('radix literals, separators and exponent signs do not swallow arithmetic', () => {
  const source = '0xFE-1 + 0o7_55 + 0B1010 + 1_024 + 1.25e-3 + 2E+10'
  assert.deepEqual(tokenizeWave(source).filter(segment => segment.kind === 'number').map(segment => segment.text), ['0xFE', '1', '0o7_55', '0B1010', '1_024', '1.25e-3', '2E+10'])
  assert.equal(kindOf('0xFE-1', '-'), 'operator')
})

test('display preserves text, including incomplete input, and escapes HTML', () => {
  for (const source of ['', 'fun main() {\n', '\t"한글 & <script>"\r\n', '/* '.repeat(5000), "'\\", 'value123\n']) {
    assert.equal(tokenizeWave(source).map(segment => segment.text).join(''), source)
  }
  const html = highlightWave('println("</span><img src=x onerror=alert(1)>&");')
  assert.ok(!html.includes('<img'))
  assert.ok(html.includes('&lt;/span&gt;&lt;img'))
  assert.ok(html.includes('&amp;'))
})

test('Wave paths and fences resolve without automatic DOM rewriting', () => {
  assert.equal(languageForPath('src/main.wave'), 'wave')
  assert.equal(languageForPath('src/EXAMPLE.WAVE'), 'wave')
  assert.equal(normalizeFenceLanguage('Wave example'), 'wave')
  assert.equal(Prism.manual, true)
  assert.ok(Prism.highlight('const value = 42', Prism.languages.javascript, 'javascript').includes('token keyword'))
})
