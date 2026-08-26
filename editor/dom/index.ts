import {
  analyzeDocument,
  createTransform,
  waveEditorCommands,
  type WaveEditorCommand,
  type WaveEditorTransform,
} from '../core'

export interface WaveEditorDOMOptions {
  value?: string
  label?: string
  placeholder?: string
  mode?: 'markdown' | 'plain'
  rows?: number
  minLength?: number
  maxLength?: number
  disabled?: boolean
  transform?: WaveEditorTransform
  commandLabels?: Partial<Record<WaveEditorCommand, string>>
  onChange?: (value: string) => void
  onSave?: (value: string) => void
}

export interface WaveEditorDOMController {
  readonly element: HTMLElement
  readonly input: HTMLTextAreaElement
  getValue(): string
  setValue(value: string): void
  setDisabled(disabled: boolean): void
  focus(): void
  destroy(): void
}

export function mountWaveEditor(host: HTMLElement, options: WaveEditorDOMOptions = {}): WaveEditorDOMController {
  if (!(host instanceof HTMLElement)) throw new TypeError('WaveEditor requires an HTMLElement host.')
  const mode = options.mode ?? 'markdown'
  const transform = options.transform ?? createTransform()
  const root = document.createElement('section')
  root.className = `wave-editor${mode === 'plain' ? ' is-plain' : ''}`
  const input = document.createElement('textarea')
  input.className = 'wave-editor__input'
  input.value = options.value ?? ''
  input.placeholder = options.placeholder ?? ''
  input.rows = options.rows ?? 8
  input.maxLength = options.maxLength ?? 200_000
  if (options.minLength !== undefined) input.minLength = options.minLength
  input.disabled = options.disabled ?? false

  let busy = false
  const status = document.createElement('footer')
  status.className = 'wave-editor__status'
  status.setAttribute('aria-label', 'document status')
  const engineStatus = document.createElement('span')
  engineStatus.textContent = 'WaveEditor'
  const metricsStatus = document.createElement('span')
  status.append(engineStatus, metricsStatus)

  const refreshMetrics = () => {
    const metrics = analyzeDocument(input.value)
    metricsStatus.textContent = `${metrics.characters} chars · ${metrics.words} words · ${metrics.lines} lines`
  }
  const apply = async (command: WaveEditorCommand) => {
    if (busy || input.disabled) return
    busy = true
    root.classList.add('is-busy')
    try {
      const result = await transform(input.value, input.selectionStart, input.selectionEnd, command)
      input.value = result.content
      engineStatus.textContent = result.engine === 'wave' ? 'Wave core' : 'WaveEditor'
      refreshMetrics()
      options.onChange?.(input.value)
      input.focus()
      input.setSelectionRange(result.selectionStart, result.selectionEnd)
    } finally {
      busy = false
      root.classList.remove('is-busy')
    }
  }

  if (options.label || mode === 'markdown') {
    const header = document.createElement('header')
    header.className = 'wave-editor__header'
    if (options.label) {
      const label = document.createElement('label')
      label.className = 'wave-editor__label'
      label.textContent = options.label
      label.addEventListener('click', () => input.focus())
      header.append(label)
    }
    if (mode === 'markdown') {
      const toolbar = document.createElement('div')
      toolbar.className = 'wave-editor__toolbar'
      toolbar.setAttribute('role', 'toolbar')
      toolbar.setAttribute('aria-label', 'Formatting')
      for (const command of waveEditorCommands) {
        const button = document.createElement('button')
        button.type = 'button'
        button.textContent = command.toolbarText
        button.title = options.commandLabels?.[command.id] ?? command.label
        button.setAttribute('aria-label', button.title)
        button.addEventListener('click', () => { void apply(command.id) })
        toolbar.append(button)
      }
      header.append(toolbar)
    }
    root.append(header)
  }
  root.append(input, status)
  host.replaceChildren(root)

  const onInput = () => { refreshMetrics(); options.onChange?.(input.value) }
  const onKeydown = (event: KeyboardEvent) => {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
      event.preventDefault(); options.onSave?.(input.value); return
    }
    if (mode !== 'markdown' || !(event.metaKey || event.ctrlKey)) return
    const command = waveEditorCommands.find((item) => item.shortcut === event.key.toLowerCase())
    if (command) { event.preventDefault(); void apply(command.id) }
  }
  input.addEventListener('input', onInput)
  input.addEventListener('keydown', onKeydown)
  refreshMetrics()

  return {
    element: root,
    input,
    getValue: () => input.value,
    setValue(value) { input.value = value; refreshMetrics() },
    setDisabled(disabled) { input.disabled = disabled; root.classList.toggle('is-disabled', disabled) },
    focus: () => input.focus(),
    destroy() { input.removeEventListener('input', onInput); input.removeEventListener('keydown', onKeydown); root.remove() },
  }
}
