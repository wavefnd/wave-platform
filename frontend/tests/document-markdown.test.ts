import assert from 'node:assert/strict'
import { test } from 'node:test'
import { renderDocumentMarkdown } from '../src/services/documentMarkdown.ts'
const options = { playgrounds: true, slotPrefix: 'example-test', renderCode: (text: string) => `<pre>${text.replace(/</g, '&lt;')}</pre>` }
const code = 'fun main() { println("Hello"); }'
const fence = '```wave playground\n' + code + '\n```'

test('only an explicit playground fence becomes an embedded editor', () => {
  for (const info of ['wave', 'wave snippet', 'wave playground invalid', 'rust playground']) {
    const result = renderDocumentMarkdown('```' + info + '\n' + code + '\n```', options)
    assert.equal(result.slots.length, 0)
    assert.match(result.html, /<pre>/)
  }
  assert.deepEqual(renderDocumentMarkdown(fence, options).slots.map(item => item.source), [code])
  assert.equal(renderDocumentMarkdown(fence, { ...options, playgrounds: false }).slots.length, 0)
  assert.equal(renderDocumentMarkdown('    ```wave playground\n    example\n    ```', options).slots.length, 0)
})

test('metadata is supplied by the adjacent registered example only', () => {
  const result = renderDocumentMarkdown('<!-- wave-example: input-calculator -->\n' + fence, options)
  assert.equal(result.slots[0].stdin, '3 1200\n')
  assert.equal(result.slots[0].expectedOutput, 'total=3600\n')
  assert.equal(result.slots[0].expectedExit, 0)
  for (const prefix of ['<!-- wave-example: unknown -->\n', '<!-- wave-example: input-calculator -->\n\nA paragraph.\n\n']) {
    assert.equal(renderDocumentMarkdown(prefix + fence, options).slots[0].stdin, undefined)
  }
})

test('a single parse preserves headings, references and surrounding lists', () => {
  const result = renderDocumentMarkdown('## Title\n\n' + fence + '\n\n## Title\n\n[Guide][guide]\n\n- before\n- after\n\n[guide]: /docs/en\n', options)
  assert.match(result.html, /id="title"/)
  assert.match(result.html, /id="title-1"/)
  assert.match(result.html, /href="\/docs\/en"/)
  assert.match(result.html, /<ul>/)
  assert.equal(result.slots.length, 1)
  assert.match(renderDocumentMarkdown('## Title', options).html, /id="title"/)
})

test('nested examples retain container markup and metadata', () => {
  const nested = ('<!-- wave-example: book-hello -->\n' + fence).split('\n').map(line => '> ' + line).join('\n')
  const result = renderDocumentMarkdown(nested, options)
  assert.match(result.html, /^<blockquote>/)
  assert.match(result.html, /<\/blockquote>/)
  assert.equal(result.slots[0].expectedOutput, 'Hello, Wave!\n')
})

test('example text stays data, including HTML and closing-tag strings', () => {
  const source = 'fun main() { println("</div><img src=x onerror=alert(1)>"); }'
  const result = renderDocumentMarkdown('```wave playground\n' + source + '\n```', options)
  assert.equal(result.slots[0].source, source)
  assert.ok(!result.html.includes('<img'))
  assert.equal(result.slots.length, 1)
})
