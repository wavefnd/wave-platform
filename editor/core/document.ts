import { WaveEditorEngine } from './engine'
import { normalizeSelection } from './metrics'
import type { WaveEditorCommand, WaveEditorSnapshot } from './types'

export class WaveEditorDocument {
  readonly historyLimit: number
  private current: WaveEditorSnapshot
  private readonly undoStack: WaveEditorSnapshot[] = []
  private readonly redoStack: WaveEditorSnapshot[] = []

  constructor(content = '', selectionStart = 0, selectionEnd = selectionStart, historyLimit = 100) {
    if (!Number.isSafeInteger(historyLimit) || historyLimit < 1) throw new RangeError('History limit must be a positive integer.')
    const selection = normalizeSelection(content, selectionStart, selectionEnd)
    this.current = { content, ...selection }
    this.historyLimit = historyLimit
  }

  get snapshot(): Readonly<WaveEditorSnapshot> { return { ...this.current } }
  get canUndo() { return this.undoStack.length > 0 }
  get canRedo() { return this.redoStack.length > 0 }

  replace(content: string, selectionStart = content.length, selectionEnd = selectionStart) {
    const selection = normalizeSelection(content, selectionStart, selectionEnd)
    this.remember(this.current)
    this.current = { content, ...selection }
    this.redoStack.length = 0
    return this.snapshot
  }

  async apply(command: WaveEditorCommand, engine = new WaveEditorEngine()) {
    const before = this.snapshot
    const result = await engine.transform({
      content: before.content, selectionStart: before.start, selectionEnd: before.end, command,
    })
    this.remember(before)
    this.current = { content: result.content, start: result.selectionStart, end: result.selectionEnd }
    this.redoStack.length = 0
    return result
  }

  undo() { return this.restore(this.undoStack, this.redoStack) }
  redo() { return this.restore(this.redoStack, this.undoStack) }

  private remember(snapshot: WaveEditorSnapshot) {
    this.undoStack.push({ ...snapshot })
    if (this.undoStack.length > this.historyLimit) this.undoStack.shift()
  }

  private restore(source: WaveEditorSnapshot[], destination: WaveEditorSnapshot[]) {
    const snapshot = source.pop()
    if (!snapshot) return this.snapshot
    destination.push({ ...this.current })
    this.current = snapshot
    return this.snapshot
  }
}
