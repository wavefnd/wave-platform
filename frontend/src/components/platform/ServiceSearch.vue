<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { Search } from '@lucide/vue'

defineProps<{ label: string; compact?: boolean }>()

const input = ref<HTMLInputElement | null>(null)

function focusSearch(event: KeyboardEvent) {
  if (event.key !== '/' || event.metaKey || event.ctrlKey || event.altKey) return
  const target = event.target as HTMLElement | null
  if (target?.matches('input, textarea, select, [contenteditable="true"]')) return
  event.preventDefault()
  input.value?.focus()
}

onMounted(() => window.addEventListener('keydown', focusSearch))
onBeforeUnmount(() => window.removeEventListener('keydown', focusSearch))
</script>

<template>
  <form :class="['service-search', { compact }]" action="/search" method="get" role="search">
    <Search :size="17" :stroke-width="1.8" aria-hidden="true" />
    <label class="sr-only" for="platform-service-search">{{ label }}</label>
    <input id="platform-service-search" ref="input" name="q" type="search" :placeholder="label" />
    <kbd>/</kbd>
  </form>
</template>
