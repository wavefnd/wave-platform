<script setup lang="ts">
import QRCode from 'qrcode'
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { useI18n } from '../i18n'
import { finishRecovery, getRecoveryEnrollment, type TOTPEnrollment } from '../services/http'

const { t } = useI18n()
const route = useRoute()
const enrollment = ref<TOTPEnrollment | null>(null)
const qrCode = ref('')
const code = ref('')
const error = ref('')
const complete = ref(false)
const token = typeof route.query.token === 'string' ? route.query.token : ''

onMounted(async () => {
  try {
    enrollment.value = await getRecoveryEnrollment(token)
    qrCode.value = await QRCode.toDataURL(enrollment.value.uri, { width: 208, margin: 1, errorCorrectionLevel: 'M' })
  } catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.recoveryExpired') }
})

async function submit() {
  error.value = ''
  try { await finishRecovery(token, code.value); complete.value = true }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}
</script>

<template>
  <main class="page shell auth-page">
    <section class="auth-panel auth-panel-wide">
      <header class="auth-heading"><img src="/img/wave-logo.ico" alt="" /><h1>{{ t('auth.resetAuthenticator') }}</h1></header>
      <div v-if="complete" class="auth-form"><p class="form-notice" role="status">{{ t('auth.resetComplete') }}</p><RouterLink class="portal-button primary centered" to="/login">{{ t('auth.signIn') }}</RouterLink></div>
      <form v-else-if="enrollment" class="auth-form enrollment-form" @submit.prevent="submit">
        <p class="auth-help">{{ t('auth.scanNewAuthenticator') }}</p>
        <img class="totp-qr" :src="qrCode" :alt="t('auth.qrCode')" />
        <details><summary>{{ t('auth.manualKey') }}</summary><code class="totp-secret">{{ enrollment.secret }}</code></details>
        <label for="recovery-code">{{ t('auth.authenticatorCode') }}</label>
        <input id="recovery-code" v-model="code" inputmode="numeric" autocomplete="one-time-code" pattern="[0-9]{6}" maxlength="6" required />
        <button class="portal-button primary" type="submit">{{ t('auth.resetAuthenticator') }}</button>
        <p v-if="error" class="form-notice error" role="alert">{{ error }}</p>
      </form>
      <div v-else class="auth-form"><p v-if="error" class="form-notice error" role="alert">{{ error }}</p></div>
    </section>
  </main>
</template>
