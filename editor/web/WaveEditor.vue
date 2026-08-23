<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import type { WaveEditorCommand, WaveEditorTransform, WaveEditorTransformResult } from './types'

const props = withDefaults(defineProps<{
  modelValue: string
  label?: string
  placeholder?: string
  mode?: 'markdown' | 'plain'
  rows?: number
  minLength?: number
  maxLength?: number
  required?: boolean
  disabled?: boolean
  transform?: WaveEditorTransform
  toolbarLabel?: string
  statusLabel?: string
  commandLabels?: Partial<Record<WaveEditorCommand, string>>
  characterLabel?: string
  wordLabel?: string
  lineLabel?: string
  engineLabel?: string
  waveEngineLabel?: string
  compatibilityEngineLabel?: string
}>(), {
  label: '', placeholder: '', mode: 'markdown', rows: 8, minLength: undefined, maxLength: 200000,
  required: false, disabled: false, transform: undefined, toolbarLabel: 'Formatting', statusLabel: 'document status',
  commandLabels: () => ({}), characterLabel: 'chars', wordLabel: 'words', lineLabel: 'lines',
  engineLabel: 'WaveEditor', waveEngineLabel: 'Wave core', compatibilityEngineLabel: 'Wave compatible',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  save: []
}>()

const input = ref<HTMLTextAreaElement | null>(null)
const busy = ref(false)
const engine = ref(props.engineLabel)
const lineCount = computed(() => (props.modelValue.match(/\n/g)?.length ?? 0) + 1)
const wordCount = computed(() => props.modelValue.trim() ? props.modelValue.trim().split(/\s+/u).length : 0)
const characterCount = computed(() => Array.from(props.modelValue).length)

const commands: Array<{ command: WaveEditorCommand; text: string; title: string }> = [
  { command: 'bold', text: 'B', title: 'Bold' },
  { command: 'italic', text: 'I', title: 'Italic' },
  { command: 'inline-code', text: '</>', title: 'Inline code' },
  { command: 'heading', text: 'H2', title: 'Heading' },
  { command: 'quote', text: '“', title: 'Quote' },
  { command: 'unordered-list', text: '•', title: 'List' },
  { command: 'link', text: '↗', title: 'Link' },
]

function localTransform(content: string, start: number, end: number, command: WaveEditorCommand): WaveEditorTransformResult {
  const wrappers: Record<WaveEditorCommand, [string, string]> = {
    bold: ['**', '**'], italic: ['*', '*'], 'inline-code': ['`', '`'], heading: ['## ', ''],
    quote: ['> ', ''], 'unordered-list': ['- ', ''], link: ['[', '](https://)'],
  }
  const [prefix, suffix] = wrappers[command]
  return {
    content: content.slice(0, start) + prefix + content.slice(start, end) + suffix + content.slice(end),
    selectionStart: start + prefix.length, selectionEnd: end + prefix.length, engine: 'web',
  }
}

async function apply(command: WaveEditorCommand) {
  if (busy.value || props.disabled) return
  const element = input.value
  if (!element) return
  const start = element.selectionStart
  const end = element.selectionEnd
  busy.value = true
  try {
    const result = props.transform
      ? await props.transform(props.modelValue, start, end, command)
      : localTransform(props.modelValue, start, end, command)
    emit('update:modelValue', result.content)
    engine.value = result.engine === 'wave' ? props.waveEngineLabel : result.engine === 'go' ? props.compatibilityEngineLabel : props.engineLabel
    await nextTick()
    input.value?.focus()
    input.value?.setSelectionRange(result.selectionStart, result.selectionEnd)
  } finally {
    busy.value = false
  }
}

function update(event: Event) {
  emit('update:modelValue', (event.target as HTMLTextAreaElement).value)
}

function onKeydown(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
    event.preventDefault(); emit('save'); return
  }
  if (props.mode !== 'markdown' || !(event.metaKey || event.ctrlKey)) return
  if (event.key.toLowerCase() === 'b') { event.preventDefault(); void apply('bold') }
  if (event.key.toLowerCase() === 'i') { event.preventDefault(); void apply('italic') }
}

function focus() { input.value?.focus() }

function insert(prefix: string, suffix = '') {
  const element = input.value
  if (!element) return
  const start = element.selectionStart
  const end = element.selectionEnd
  emit('update:modelValue', props.modelValue.slice(0, start) + prefix + props.modelValue.slice(start, end) + suffix + props.modelValue.slice(end))
  void nextTick(() => {
    input.value?.focus()
    input.value?.setSelectionRange(start + prefix.length, end + prefix.length)
  })
}

defineExpose({ focus, insert })
</script>

<template>
  <section class="wave-editor" :class="{ 'is-disabled': disabled, 'is-plain': mode === 'plain' }">
    <header v-if="label || mode === 'markdown'" class="wave-editor__header">
      <label v-if="label" class="wave-editor__label" @click="focus">{{ label }}</label>
      <div v-if="mode === 'markdown'" class="wave-editor__toolbar" role="toolbar" :aria-label="toolbarLabel">
        <button v-for="item in commands" :key="item.command" type="button" :title="commandLabels[item.command] ?? item.title" :aria-label="commandLabels[item.command] ?? item.title"
          :disabled="disabled || busy" @click="apply(item.command)">{{ item.text }}</button>
      </div>
    </header>
    <textarea ref="input" class="wave-editor__input" :value="modelValue" :placeholder="placeholder" :rows="rows"
      :minlength="minLength" :maxlength="maxLength" :required="required" :disabled="disabled"
      @input="update" @keydown="onKeydown" />
    <footer class="wave-editor__status" :aria-label="statusLabel">
      <span>{{ engine }}</span>
      <span>{{ characterCount }} {{ characterLabel }} · {{ wordCount }} {{ wordLabel }} · {{ lineCount }} {{ lineLabel }}</span>
    </footer>
  </section>
</template>
