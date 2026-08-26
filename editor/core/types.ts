export const WAVE_EDITOR_VERSION = '0.2.0'
export const MAX_DOCUMENT_CHARACTERS = 200_000

export type WaveEditorCommand = 'bold' | 'italic' | 'inline-code' | 'heading' | 'quote' | 'unordered-list' | 'link'

export interface WaveEditorSelection {
  start: number
  end: number
}

export interface WaveEditorMetrics {
  characters: number
  lines: number
  words: number
}

export interface WaveEditorTransformRequest {
  content: string
  selectionStart: number
  selectionEnd: number
  command: WaveEditorCommand
}

export interface WaveEditorTransformResult extends WaveEditorMetrics {
  content: string
  selectionStart: number
  selectionEnd: number
  engine: string
}

export type WaveEditorTransform = (
  content: string,
  selectionStart: number,
  selectionEnd: number,
  command: WaveEditorCommand,
) => Promise<WaveEditorTransformResult>

export interface WaveEditorAdapter {
  readonly name: string
  transform(request: WaveEditorTransformRequest): Promise<WaveEditorTransformResult>
  analyze?(content: string): Promise<WaveEditorMetrics>
}

export interface WaveEditorSnapshot extends WaveEditorSelection {
  content: string
}
