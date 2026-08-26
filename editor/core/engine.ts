import { transformSelection } from './commands'
import { analyzeDocument } from './metrics'
import type {
  WaveEditorAdapter,
  WaveEditorMetrics,
  WaveEditorTransformRequest,
  WaveEditorTransformResult,
} from './types'

export class WaveEditorEngine {
  readonly adapter?: WaveEditorAdapter

  constructor(adapter?: WaveEditorAdapter) {
    this.adapter = adapter
  }

  async transform(request: WaveEditorTransformRequest): Promise<WaveEditorTransformResult> {
    if (this.adapter) return this.adapter.transform(request)
    return transformSelection(request.content, request.selectionStart, request.selectionEnd, request.command)
  }

  async analyze(content: string): Promise<WaveEditorMetrics> {
    if (this.adapter?.analyze) return this.adapter.analyze(content)
    return analyzeDocument(content)
  }
}

export function createTransform(engine = new WaveEditorEngine()) {
  return (content: string, selectionStart: number, selectionEnd: number, command: WaveEditorTransformRequest['command']) =>
    engine.transform({ content, selectionStart, selectionEnd, command })
}
