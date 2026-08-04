<script setup lang="ts">
import { Code2, Quote } from '@lucide/vue'
import { nextTick, ref } from 'vue'

import MarkdownContent from '../MarkdownContent.vue'
import { useI18n } from '../../i18n'

const props = defineProps<{
  modelValue: string
  preview: boolean
  authenticated: boolean
  submitting: boolean
  error: string
  replying: boolean
  loginTarget: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:preview': [value: boolean]
  submit: []
  cancel: []
}>()

const input = ref<HTMLTextAreaElement | null>(null)
const { t } = useI18n()

async function insertMarkup(prefix: string, suffix = '') {
  const element = input.value
  if (!element) return
  const start = element.selectionStart
  const end = element.selectionEnd
  const selected = props.modelValue.slice(start, end)
  emit('update:modelValue', props.modelValue.slice(0, start) + prefix + selected + suffix + props.modelValue.slice(end))
  await nextTick()
  element.focus()
  element.setSelectionRange(start + prefix.length, end + prefix.length)
}

function focus() {
  nextTick(() => input.value?.focus())
}

defineExpose({ focus })
</script>

<template>
  <form class="community-reply-editor" :class="{ 'is-inline': replying }" @submit.prevent="emit('submit')">
    <header>
      <strong>{{ replying ? t('community.replying') : t('community.addComment') }}</strong>
      <button v-if="replying" type="button" @click="emit('cancel')">{{ t('common.cancel') }}</button>
    </header>
    <template v-if="authenticated">
      <div class="community-editor-toolbar">
        <button type="button" :title="t('community.code')" @click="insertMarkup('\n```wave\n', '\n```\n')"><Code2 :size="15" /></button>
        <button type="button" :title="t('community.quote')" @click="insertMarkup('> ')"><Quote :size="15" /></button>
        <button type="button" class="preview-toggle" @click="emit('update:preview', !preview)">{{ preview ? t('community.edit') : t('community.preview') }}</button>
      </div>
      <MarkdownContent v-if="preview" class="community-editor-preview" :source="modelValue" />
      <textarea v-else ref="input" :value="modelValue" required maxlength="10000" rows="6"
        @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)" />
      <p v-if="error" class="community-action-error" role="alert">{{ error }}</p>
      <footer><button class="ui-button primary" type="submit" :disabled="submitting">{{ t('community.postComment') }}</button></footer>
    </template>
    <RouterLink v-else class="community-signin-action" :to="{ name: 'login', query: { redirect: loginTarget } }">{{ t('community.signInToComment') }}</RouterLink>
  </form>
</template>
