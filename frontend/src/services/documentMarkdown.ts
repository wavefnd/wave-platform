import { Marked, Renderer, type Token, type Tokens } from 'marked'
import { gfmHeadingId } from 'marked-gfm-heading-id'
import manifest from '../../../wavedoc/examples.json' with { type: 'json' }

export type DocumentExample = { source: string; stdin?: string; expectedOutput?: string; expectedExit?: number }
type ExampleMetadata = { id: string; stdin?: string; stdout?: string | null; exit?: number }
const examples = new Map<string, ExampleMetadata>(manifest.map(item => [item.id, item]))
const marker = /^<!--\s*wave-example: ([a-z0-9-]+)\s*-->\s*$/

export function renderDocumentMarkdown(source: string, options: {
  playgrounds: boolean
  slotPrefix: string
  renderCode: (text: string, language?: string) => string
}) {
  const slots: DocumentExample[] = []
  const ids = new WeakMap<Tokens.Code, string>()
  // Only an immediately preceding marker supplies metadata. Visit nested lists
  // and quotes as well, while keeping the full document in a single parse so
  // heading IDs and reference links work across embedded examples.
  function annotate(tokens: Token[]) {
    let pending: string | undefined
    for (const token of tokens) {
      if (token.type === 'space') continue
      if (token.type === 'html') pending = marker.exec(token.text)?.[1]
      else {
        if (token.type === 'code' && pending) ids.set(token as Tokens.Code, pending)
        pending = undefined
        if (token.type === 'blockquote') annotate((token as Tokens.Blockquote).tokens)
        if (token.type === 'list') for (const item of (token as Tokens.List).items) annotate(item.tokens)
      }
    }
  }
  const renderer = new Renderer()
  renderer.code = token => {
    if (options.playgrounds && /^wave\s+playground$/i.test(token.lang?.trim() ?? '')) {
      const example = examples.get(ids.get(token) ?? '')
      const index = slots.push({
        source: token.text,
        stdin: example?.stdin,
        expectedOutput: typeof example?.stdout === 'string' ? example.stdout : undefined,
        expectedExit: example ? example.exit ?? 0 : undefined,
      }) - 1
      return `<div class="document-playground" data-wave-playground="${options.slotPrefix}-${index}"></div>`
    }
    return options.renderCode(token.text, token.lang)
  }
  const parser = new Marked(gfmHeadingId(), {
    async: false, gfm: true, breaks: false, renderer,
    hooks: { processAllTokens(tokens) { annotate(tokens); return tokens } },
  })
  return { html: parser.parse(source) as string, slots }
}
