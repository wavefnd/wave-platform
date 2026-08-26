import { describe, expect, it } from 'vitest'

import {
  analyzeDocument,
  browserOffsetToUnicode,
  transformSelection,
  unicodeOffsetToBrowser,
  WaveEditorDocument,
} from '../core'

describe('WaveEditor core', () => {
  it('transforms a browser selection and returns complete metrics', () => {
    const result = transformSelection('Wave editor', 5, 11, 'bold')
    expect(result).toEqual({
      content: 'Wave **editor**', selectionStart: 7, selectionEnd: 13,
      engine: 'javascript', characters: 15, lines: 1, words: 2,
    })
  })

  it('counts Unicode characters independently from UTF-16 browser offsets', () => {
    const content = 'Wave 🌊 언어'
    expect(analyzeDocument(content)).toEqual({ characters: 9, lines: 1, words: 3 })
    expect(browserOffsetToUnicode(content, 7)).toBe(6)
    expect(unicodeOffsetToBrowser(content, 6)).toBe(7)
  })

  it('keeps bounded undo and redo history', async () => {
    const document = new WaveEditorDocument('Wave', 0, 4, 2)
    await document.apply('italic')
    expect(document.snapshot.content).toBe('*Wave*')
    expect(document.undo().content).toBe('Wave')
    expect(document.redo().content).toBe('*Wave*')
  })

  it('rejects invalid selections and oversized documents', () => {
    expect(() => transformSelection('Wave', 0, 5, 'bold')).toThrow(RangeError)
    expect(() => analyzeDocument('x'.repeat(200_001))).toThrow(RangeError)
  })
})
