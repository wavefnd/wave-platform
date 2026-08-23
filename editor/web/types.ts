export type WaveEditorCommand = 'bold' | 'italic' | 'inline-code' | 'heading' | 'quote' | 'unordered-list' | 'link'

export interface WaveEditorTransformResult {
  content: string
  selectionStart: number
  selectionEnd: number
  engine?: string
  lines?: number
  words?: number
}

export type WaveEditorTransform = (
  content: string,
  selectionStart: number,
  selectionEnd: number,
  command: WaveEditorCommand,
) => Promise<WaveEditorTransformResult>
