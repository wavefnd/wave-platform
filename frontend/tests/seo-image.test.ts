import assert from 'node:assert/strict'
import test from 'node:test'
import { firstMarkdownImage } from '../src/services/seo.ts'

// Expected values match markdownImage in internal/web/seo.go, which resolves the
// same Markdown against the article canonical.
const blog = 'https://wave.example/blog/compiler-notes'
const release = 'https://wave.example/releases/v0.2.2'

test('relative, root-relative, and absolute HTTP(S) images resolve to the server URL', () => {
  const cases: Array<{ markdown: string, url: string, alt: string }> = [
    { markdown: '![Diagram](images/pipeline.png)', url: 'https://wave.example/blog/images/pipeline.png', alt: 'Diagram' },
    { markdown: '![](../media/diagram.webp)', url: 'https://wave.example/media/diagram.webp', alt: '' },
    { markdown: '![Hero](/img/hero.png)', url: 'https://wave.example/img/hero.png', alt: 'Hero' },
    { markdown: '![Absolute](https://cdn.example/a.png)', url: 'https://cdn.example/a.png', alt: 'Absolute' },
    { markdown: '![Insecure](http://cdn.example/a.png)', url: 'http://cdn.example/a.png', alt: 'Insecure' },
    { markdown: '![Protocol](//cdn.example/a.png)', url: 'https://cdn.example/a.png', alt: 'Protocol' },
    { markdown: '![Titled](images/pipeline.png "pipeline")', url: 'https://wave.example/blog/images/pipeline.png', alt: 'Titled' },
    { markdown: '![Spaced](<images/pipeline diagram.png>)', url: 'https://wave.example/blog/images/pipeline%20diagram.png', alt: 'Spaced' },
    // Unicode spaces are matched on both sides: the pattern only excludes the
    // ASCII whitespace Go's \s excludes.
    { markdown: '![Unicode](images/pipeline\u3000graph.png)', url: 'https://wave.example/blog/images/pipeline%E3%80%80graph.png', alt: 'Unicode' },
    { markdown: '![Unicode](<images/pipeline\u3000diagram.png>)', url: 'https://wave.example/blog/images/pipeline%E3%80%80diagram.png', alt: 'Unicode' },
    // net/url trims the captured value, so surrounding whitespace drops out.
    { markdown: '![Trimmed](images/a.png\u3000)', url: 'https://wave.example/blog/images/a.png', alt: 'Trimmed' },
    { markdown: '![Trimmed](<images/a.png\u3000>)', url: 'https://wave.example/blog/images/a.png', alt: 'Trimmed' },
  ]
  for (const item of cases) assert.deepEqual(firstMarkdownImage(item.markdown, blog), { url: item.url, alt: item.alt })
  assert.equal(firstMarkdownImage('# Article\n\nNo preview image.', blog), null)
})

test('a release article keeps its /releases canonical when reached through the legacy blog URL', () => {
  // The legacy /blog/<slug> route redirects to /releases/<slug> in
  // internal/web/router.go, and BlogPage.vue canonicalises release posts the
  // same way before resolving preview images.
  assert.deepEqual(firstMarkdownImage('![Graph](images/graph.png)', release), { url: 'https://wave.example/releases/images/graph.png', alt: 'Graph' })
  assert.notEqual(
    firstMarkdownImage('![Graph](images/graph.png)', release)?.url,
    firstMarkdownImage('![Graph](images/graph.png)', blog)?.url)
  assert.deepEqual(firstMarkdownImage('![Graph](../media/graph.png)', release), { url: 'https://wave.example/media/graph.png', alt: 'Graph' })
  assert.deepEqual(firstMarkdownImage('![Graph](/releases/v0.2.2/graph.png)', release), { url: 'https://wave.example/releases/v0.2.2/graph.png', alt: 'Graph' })
})

test('disallowed URL schemes stay rejected for every canonical', () => {
  const rejected = ['![x](javascript:alert(1))', '![x](data:image/png;base64,AAAA)', '![x](file:///etc/passwd)', '![x](mailto:team@wave.example)']
  for (const markdown of rejected) {
    assert.equal(firstMarkdownImage(markdown, blog), null)
    assert.equal(firstMarkdownImage(markdown, release), null)
  }
})

test('control characters are rejected like net/url rejects them', () => {
  assert.equal(firstMarkdownImage('![Vt](images/pipe\vline.png)', blog), null)
  assert.equal(firstMarkdownImage('![Vt](<images/pipe\vline.png>)', blog), null)
  assert.equal(firstMarkdownImage('![Del](<images/a.png\u007f>)', blog), null)
})

test('a rejected first image does not fall through to a later valid image', () => {
  // Both sides inspect only the first image in the document.
  assert.equal(firstMarkdownImage('![x](javascript:alert(1))\n\n![Ok](images/a.png)', blog), null)
})
