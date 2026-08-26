import { MAX_DOCUMENT_CHARACTERS, type WaveEditorMetrics, type WaveEditorSelection } from './types'

export function analyzeDocument(content: string): WaveEditorMetrics {
  const characters = Array.from(content).length
  if (characters > MAX_DOCUMENT_CHARACTERS) {
    throw new RangeError(`WaveEditor documents may contain at most ${MAX_DOCUMENT_CHARACTERS} characters.`)
  }
  return {
    characters,
    lines: (content.match(/\n/g)?.length ?? 0) + 1,
    words: content.trim() ? content.trim().split(/\s+/u).length : 0,
  }
}

export function normalizeSelection(content: string, start: number, end: number): WaveEditorSelection {
  if (!Number.isSafeInteger(start) || !Number.isSafeInteger(end) || start < 0 || end < start || end > content.length) {
    throw new RangeError('WaveEditor selection is outside the document.')
  }
  return { start, end }
}

export function browserOffsetToUnicode(content: string, browserOffset: number): number {
  normalizeSelection(content, browserOffset, browserOffset)
  return Array.from(content.slice(0, browserOffset)).length
}

export function unicodeOffsetToBrowser(content: string, unicodeOffset: number): number {
  if (!Number.isSafeInteger(unicodeOffset) || unicodeOffset < 0) throw new RangeError('Unicode offset is invalid.')
  const characters = Array.from(content)
  if (unicodeOffset > characters.length) throw new RangeError('Unicode offset is outside the document.')
  return characters.slice(0, unicodeOffset).join('').length
}
