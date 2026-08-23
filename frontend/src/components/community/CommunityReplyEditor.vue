<script setup lang="ts">
import { nextTick, ref } from 'vue'

import MarkdownContent from '../MarkdownContent.vue'
import PlatformWaveEditor from '../editor/PlatformWaveEditor.vue'
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

const input = ref<{ focus: () => void } | null>(null)
const { t } = useI18n()

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
        <button type="button" class="preview-toggle" @click="emit('update:preview', !preview)">{{ preview ? t('community.edit') : t('community.preview') }}</button>
      </div>
      <MarkdownContent v-if="preview" class="community-editor-preview" :source="modelValue" />
      <PlatformWaveEditor v-else ref="input" :model-value="modelValue" required :max-length="10000" :rows="6"
        @update:model-value="emit('update:modelValue', $event)" />
      <p v-if="error" class="community-action-error" role="alert">{{ error }}</p>
      <footer><button class="ui-button primary" type="submit" :disabled="submitting">{{ t('community.postComment') }}</button></footer>
    </template>
    <RouterLink v-else class="community-signin-action" :to="{ name: 'login', query: { redirect: loginTarget } }">{{ t('community.signInToComment') }}</RouterLink>
  </form>
</template>
