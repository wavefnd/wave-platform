<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import TurnstileWidget from '../components/auth/TurnstileWidget.vue'
import { useI18n } from '../i18n'
import { getAuthConfig, requestRecovery } from '../services/http'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const identifier = ref('')
const code = ref('')
const challenge = ref('')
const siteKey = ref('')
const recoveryMode = ref(false)
const recoverySent = ref(false)
const error = ref('')

onMounted(async () => { try { siteKey.value = (await getAuthConfig()).turnstileSiteKey } catch { /* submit remains authoritative */ } })

function safeRedirect(): string {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/mail'
  return redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : '/mail'
}

async function submit() {
  error.value = ''
  try {
    await auth.signIn(identifier.value, code.value, challenge.value)
    await router.replace(safeRedirect())
  } catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}

async function recover() {
  error.value = ''
  try { await requestRecovery(identifier.value, challenge.value); recoverySent.value = true }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}
</script>

<template>
  <main class="page shell auth-page">
    <section class="auth-panel">
      <header class="auth-heading"><img src="/img/wave-logo.ico" alt="" /><h1>{{ recoveryMode ? t('auth.recoveryTitle') : t('login.title') }}</h1></header>
      <form v-if="!recoveryMode" class="auth-form" @submit.prevent="submit">
        <label for="login-username">{{ t('auth.username') }}</label>
        <input id="login-username" v-model="identifier" name="username" autocomplete="username" required autofocus />
        <label for="login-code">{{ t('auth.authenticatorCode') }}</label>
        <input id="login-code" v-model="code" name="one-time-code" inputmode="numeric" autocomplete="one-time-code" pattern="[0-9]{6}" maxlength="6" required />
        <TurnstileWidget :site-key="siteKey" action="login" @token="challenge = $event" />
        <button class="portal-button primary" type="submit" :disabled="auth.loading">{{ t('auth.submitSignIn') }}</button>
        <button class="auth-text-action" type="button" @click="recoveryMode = true; challenge = ''">{{ t('auth.lostAuthenticator') }}</button>
        <p v-if="error" class="form-notice error" role="alert">{{ error }}</p>
        <p class="auth-alternate"><RouterLink to="/register">{{ t('auth.signUp') }}</RouterLink></p>
      </form>
      <form v-else class="auth-form" @submit.prevent="recover">
        <p class="auth-help">{{ t('auth.recoveryHelp') }}</p>
        <label for="recovery-identifier">{{ t('auth.username') }}</label>
        <input id="recovery-identifier" v-model="identifier" autocomplete="username" required autofocus />
        <TurnstileWidget :site-key="siteKey" action="recovery" @token="challenge = $event" />
        <button class="portal-button primary" type="submit" :disabled="recoverySent">{{ t('auth.sendRecovery') }}</button>
        <p v-if="recoverySent" class="form-notice" role="status">{{ t('auth.recoverySent') }}</p>
        <p v-if="error" class="form-notice error" role="alert">{{ error }}</p>
        <button class="auth-text-action" type="button" @click="recoveryMode = false; challenge = ''">{{ t('auth.backToSignIn') }}</button>
      </form>
    </section>
  </main>
</template>
