<script setup lang="ts">
import { X } from '@lucide/vue'
import { onMounted, ref } from 'vue'

defineProps<{
  title: string
  help: string
  qrCode: string
  secret: string
  code: string
  codeLabel: string
  manualKeyLabel: string
  qrAlt: string
  confirmLabel: string
  cancelLabel: string
  error?: string
  submitting?: boolean
}>()

const emit = defineEmits<{
  (event: 'update:code', value: string): void
  (event: 'confirm'): void
  (event: 'close'): void
}>()

const codeInput = ref<HTMLInputElement | null>(null)
onMounted(() => codeInput.value?.focus())
</script>

<template>
  <Teleport to="body">
    <div class="auth-layer" role="presentation" @click.self="emit('close')">
      <section class="auth-layer-dialog" role="dialog" aria-modal="true" :aria-labelledby="'authenticator-dialog-title'" @keydown.esc="emit('close')">
        <header class="auth-layer-header">
          <h2 id="authenticator-dialog-title">{{ title }}</h2>
          <button class="icon-button" type="button" :aria-label="cancelLabel" @click="emit('close')"><X :size="18" /></button>
        </header>
        <form class="auth-form enrollment-form" @submit.prevent="emit('confirm')">
          <p class="auth-help">{{ help }}</p>
          <img class="totp-qr" :src="qrCode" :alt="qrAlt" />
          <details><summary>{{ manualKeyLabel }}</summary><code class="totp-secret">{{ secret }}</code></details>
          <label for="authenticator-enrollment-code">{{ codeLabel }}</label>
          <input
            id="authenticator-enrollment-code"
            ref="codeInput"
            :value="code"
            inputmode="numeric"
            autocomplete="one-time-code"
            pattern="[0-9]{6}"
            maxlength="6"
            required
            @input="emit('update:code', ($event.target as HTMLInputElement).value)"
          />
          <div class="auth-layer-actions">
            <button class="portal-button" type="button" @click="emit('close')">{{ cancelLabel }}</button>
            <button class="portal-button primary" type="submit" :disabled="submitting">{{ confirmLabel }}</button>
          </div>
          <p v-if="error" class="form-notice error" role="alert">{{ error }}</p>
        </form>
      </section>
    </div>
  </Teleport>
</template>
