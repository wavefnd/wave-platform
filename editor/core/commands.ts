import type { WaveEditorCommand, WaveEditorTransformResult } from './types'
import { analyzeDocument, normalizeSelection } from './metrics'

export interface WaveEditorCommandDefinition {
  id: WaveEditorCommand
  label: string
  toolbarText: string
  prefix: string
  suffix: string
  shortcut?: string
}

export const waveEditorCommands: readonly WaveEditorCommandDefinition[] = Object.freeze([
  { id: 'bold', label: 'Bold', toolbarText: 'B', prefix: '**', suffix: '**', shortcut: 'b' },
  { id: 'italic', label: 'Italic', toolbarText: 'I', prefix: '*', suffix: '*', shortcut: 'i' },
  { id: 'inline-code', label: 'Inline code', toolbarText: '</>', prefix: '`', suffix: '`' },
  { id: 'heading', label: 'Heading', toolbarText: 'H2', prefix: '## ', suffix: '' },
  { id: 'quote', label: 'Quote', toolbarText: '“', prefix: '> ', suffix: '' },
  { id: 'unordered-list', label: 'List', toolbarText: '•', prefix: '- ', suffix: '' },
  { id: 'link', label: 'Link', toolbarText: '↗', prefix: '[', suffix: '](https://)' },
])

const commandsByID = new Map(waveEditorCommands.map((command) => [command.id, command]))

export function commandDefinition(command: WaveEditorCommand): WaveEditorCommandDefinition {
  const definition = commandsByID.get(command)
  if (!definition) throw new RangeError(`Unsupported WaveEditor command: ${String(command)}`)
  return definition
}

export function transformSelection(
  content: string,
  selectionStart: number,
  selectionEnd: number,
  command: WaveEditorCommand,
): WaveEditorTransformResult {
  const selection = normalizeSelection(content, selectionStart, selectionEnd)
  const definition = commandDefinition(command)
  const nextContent = content.slice(0, selection.start)
    + definition.prefix
    + content.slice(selection.start, selection.end)
    + definition.suffix
    + content.slice(selection.end)
  return {
    content: nextContent,
    selectionStart: selection.start + definition.prefix.length,
    selectionEnd: selection.end + definition.prefix.length,
    engine: 'javascript',
    ...analyzeDocument(nextContent),
  }
}
