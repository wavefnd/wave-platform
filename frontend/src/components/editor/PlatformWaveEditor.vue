<script setup lang="ts">
import { WaveEditor } from '@wavefnd/editor'
import type { WaveEditorCommand } from '@wavefnd/editor'
import { computed, ref } from 'vue'

import { useI18n } from '../../i18n'
import { transformEditorDocument } from '../../services/http'

withDefaults(defineProps<{
  modelValue: string
  label?: string
  placeholder?: string
  mode?: 'markdown' | 'plain'
  rows?: number
  minLength?: number
  maxLength?: number
  required?: boolean
  disabled?: boolean
}>(), {
  label: '', placeholder: '', mode: 'markdown', rows: 8, minLength: undefined, maxLength: 200000,
  required: false, disabled: false,
})

const emit = defineEmits<{ 'update:modelValue': [value: string]; save: [] }>()
const { t } = useI18n()
const editor = ref<{ focus: () => void; insert: (prefix: string, suffix?: string) => void } | null>(null)
const commandLabels = computed<Record<WaveEditorCommand, string>>(() => ({
  bold: t('editor.bold'), italic: t('editor.italic'), 'inline-code': t('editor.inlineCode'), heading: t('editor.heading'),
  quote: t('editor.quote'), 'unordered-list': t('editor.list'), link: t('editor.link'),
}))

async function transform(content: string, start: number, end: number, command: WaveEditorCommand) {
  return transformEditorDocument(content, start, end, command)
}

function focus() { editor.value?.focus() }
function insert(prefix: string, suffix = '') { editor.value?.insert(prefix, suffix) }
defineExpose({ focus, insert })
</script>

<template>
  <WaveEditor ref="editor" :model-value="modelValue" :label="label" :placeholder="placeholder" :mode="mode"
    :rows="rows" :min-length="minLength" :max-length="maxLength" :required="required" :disabled="disabled"
    :toolbar-label="t('editor.toolbar')" :status-label="t('editor.status')" :command-labels="commandLabels"
    :character-label="t('editor.characters')" :word-label="t('editor.words')" :line-label="t('editor.lines')"
    :wave-engine-label="t('editor.waveCore')" :compatibility-engine-label="t('editor.compatibility')" :transform="transform"
    @update:model-value="emit('update:modelValue', $event)" @save="emit('save')" />
</template>
