<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps<{ siteKey: string; action: string }>()
const emit = defineEmits<{ token: [value: string] }>()
const host = ref<HTMLElement | null>(null)
let widgetId: string | undefined

declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: Record<string, unknown>) => string
      remove: (id: string) => void
    }
  }
}

function renderWidget() {
  if (!props.siteKey || !host.value || !window.turnstile || widgetId) return
  widgetId = window.turnstile.render(host.value, {
    sitekey: props.siteKey,
    action: props.action,
    appearance: 'interaction-only',
    callback: (value: string) => emit('token', value),
    'expired-callback': () => emit('token', ''),
    'error-callback': () => emit('token', ''),
  })
}

function load() {
  if (!props.siteKey) return
  if (window.turnstile) { renderWidget(); return }
  const existing = document.querySelector<HTMLScriptElement>('script[data-wave-turnstile]')
  if (existing) { existing.addEventListener('load', renderWidget, { once: true }); return }
  const script = document.createElement('script')
  script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
  script.async = true
  script.defer = true
  script.dataset.waveTurnstile = 'true'
  script.addEventListener('load', renderWidget, { once: true })
  document.head.append(script)
}

onMounted(load)
watch(() => props.siteKey, load)
onBeforeUnmount(() => { if (widgetId && window.turnstile) window.turnstile.remove(widgetId) })
</script>

<template><div v-if="siteKey" ref="host" class="turnstile-slot" /></template>
