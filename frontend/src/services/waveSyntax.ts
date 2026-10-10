// Keep in sync with wave/source-analyzer/main.wave and the pinned compiler's
// front/lexer/src/{ident,scan,trivia,literals}.rs (v0.2.1-pre-beta).
// This is a tolerant display lexer; compilation remains the syntax validator.
const keywords = new Set('async await fun extern export pub type enum variant static var deref let mut const if else proto struct while for module class in out clobber is as asm xnand import return continue print input println match break true false null'.split(' '))
const types = new Set('void array bool byte char f32 f64 isz usz ptr str i8 i16 i32 i64 i128 i256 i512 i1024 u8 u16 u32 u64 u128 u256 u512 u1024'.split(' '))
const identifier = /[a-zA-Z0-9_\u0080-\uffff]/
const numericByte = /[a-zA-Z0-9_]/

export type WaveSegment = { kind: '' | 'keyword' | 'builtin' | 'string' | 'comment' | 'number' | 'operator' | 'punctuation'; text: string }

export function tokenizeWave(source: string): WaveSegment[] {
  const segments: WaveSegment[] = []
  let index = 0
  while (index < source.length) {
    const start = index
    const current = source[index]
    let kind: WaveSegment['kind'] = ''
    if (source.startsWith('//', index)) {
      kind = 'comment'; index += 2
      while (index < source.length && !/[\r\n]/.test(source[index])) index++
    } else if (source.startsWith('/*', index)) {
      kind = 'comment'; index += 2
      let depth = 1
      while (index < source.length && depth) {
        if (source.startsWith('/*', index)) { depth++; index += 2 }
        else if (source.startsWith('*/', index)) { depth--; index += 2 }
        else index++
      }
    } else if (current === '"' || current === "'") {
      kind = 'string'; index++
      while (index < source.length && !/[\r\n]/.test(source[index])) {
        if (source[index] === current) { index++; break }
        if (source[index] === '\\' && index + 1 < source.length && !/[\r\n]/.test(source[index + 1])) index++
        index++
      }
    } else if (/[0-9]/.test(current)) {
      const radix = current === '0' && /[xXoObB]/.test(source[index + 1] ?? '')
      kind = 'number'; index++
      // Consume the compiler's numeric candidate, including exponent signs,
      // but leave range punctuation separate. Invalid candidates are not diagnostics.
      while (index < source.length) {
        const next = source[index]
        if (numericByte.test(next)) index++
        else if (!radix && next === '.' && source[index + 1] !== '.' && source[index - 1] !== '.') index++
        else if (!radix && (next === '+' || next === '-') && /[eE]/.test(source[index - 1])) index++
        else break
      }
    } else if (identifier.test(current)) {
      index++
      // Consume the entire identifier so digits or keywords inside names stay plain.
      while (index < source.length && identifier.test(source[index])) index++
      const word = source.slice(start, index)
      if (keywords.has(word)) kind = 'keyword'
      else if (types.has(word)) kind = 'builtin'
    } else {
      index++
      if (/[-+*/%=&|^!<>~?]/.test(current)) kind = 'operator'
      else if (/[{}[\];(),.:#]/.test(current)) kind = 'punctuation'
    }
    const text = source.slice(start, index)
    const previous = segments.at(-1)
    if (previous?.kind === kind) previous.text += text
    else segments.push({ kind, text })
  }
  return segments
}

export function highlightWave(source: string): string {
  return tokenizeWave(source).map(({ kind, text }) => {
    const escaped = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    return kind ? `<span class="token ${kind}">${escaped}</span>` : escaped
  }).join('')
}
