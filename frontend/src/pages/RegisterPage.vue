<script setup lang="ts">
import QRCode from 'qrcode'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import TurnstileWidget from '../components/auth/TurnstileWidget.vue'
import AuthenticatorEnrollmentDialog from '../components/auth/AuthenticatorEnrollmentDialog.vue'
import GmailDeliveryWarning from '../components/GmailDeliveryWarning.vue'
import { useI18n } from '../i18n'
import { containsGmailAddress } from '../services/email-address'
import { beginRegistration, getAuthConfig, getRegistrationAddress, type TOTPEnrollment } from '../services/http'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const displayName = ref('')
const username = ref('')
const suggestedUsername = ref('')
const addressChoiceRequired = ref(false)
const recoveryEmail = ref('')
const code = ref('')
const challenge = ref('')
const mailDomain = ref('wave-lang.dev')
const registrationOpen = ref(true)
const totpConfigured = ref(true)
const siteKey = ref('')
const enrollment = ref<TOTPEnrollment | null>(null)
const qrCode = ref('')
const error = ref('')

const addressPreview = computed(() => {
	const generated = displayName.value.trim().toLocaleLowerCase().replace(/[^\p{L}\p{N}]+/gu, '-').replace(/^-|-$/g, '')
	const local = addressChoiceRequired.value ? username.value.trim().toLocaleLowerCase() : (suggestedUsername.value || generated)
  return local ? `${local}@${mailDomain.value}` : ''
})
const gmailRecovery = computed(() => containsGmailAddress(recoveryEmail.value))

async function checkAddress() {
  if (!displayName.value.trim()) return
  try {
    const result = await getRegistrationAddress(displayName.value)
    suggestedUsername.value = result.localPart
    addressChoiceRequired.value = result.choiceRequired
    if (!result.choiceRequired) username.value = ''
  } catch { /* registration submit remains authoritative */ }
}

onMounted(async () => {
  try {
    const config = await getAuthConfig()
    mailDomain.value = config.mailDomain
    registrationOpen.value = config.registrationOpen
    totpConfigured.value = config.totpConfigured
    siteKey.value = config.turnstileSiteKey
  } catch { /* submit remains authoritative */ }
})

function safeRedirect(): string {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/mail'
  return redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : '/mail'
}

async function begin() {
  error.value = ''
  try {
    await checkAddress()
    if (addressChoiceRequired.value && !username.value.trim()) { error.value = t('auth.addressChoiceRequired'); return }
    const result = await beginRegistration(displayName.value, addressChoiceRequired.value ? username.value : suggestedUsername.value, recoveryEmail.value, challenge.value)
    qrCode.value = await QRCode.toDataURL(result.uri, { width: 208, margin: 1, errorCorrectionLevel: 'M' })
    enrollment.value = result
  } catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}

function closeEnrollment() {
  enrollment.value = null
  qrCode.value = ''
  code.value = ''
  error.value = ''
}

async function finish() {
  if (!enrollment.value) return
  error.value = ''
  try { await auth.completeSignUp(enrollment.value.token, code.value); await router.replace(safeRedirect()) }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}
</script>

<template>
  <main class="page shell auth-page">
    <section class="auth-panel auth-panel-wide">
      <header class="auth-heading"><img src="/img/wave-logo.ico" alt="" /><h1>{{ t('register.title') }}</h1></header>
      <form class="auth-form" @submit.prevent="begin">
        <label for="register-name">{{ t('auth.displayName') }}</label>
        <input id="register-name" v-model="displayName" autocomplete="name" required autofocus @blur="checkAddress" />
		<template v-if="addressChoiceRequired"><label for="register-address">{{ t('auth.waveAddress') }}</label>
		<div class="address-field"><input id="register-address" v-model="username" autocomplete="username" required maxlength="60" /><span>@{{ mailDomain }}</span></div>
		<p class="auth-help">{{ t('auth.duplicateNameAddressHelp') }}</p></template>
        <p v-if="addressPreview" class="address-preview">{{ addressPreview }}</p>
		<p v-if="!addressChoiceRequired" class="auth-help">{{ t('auth.waveAddressHelp') }}</p>
        <label for="register-recovery">{{ t('auth.recoveryEmail') }}</label>
        <input id="register-recovery" v-model="recoveryEmail" type="email" autocomplete="email" required />
        <GmailDeliveryWarning v-if="gmailRecovery" />
        <p class="auth-help">{{ t('auth.recoveryEmailHelp') }}</p>
        <TurnstileWidget :site-key="siteKey" action="register" @token="challenge = $event" />
        <button class="portal-button primary" type="submit" :disabled="!registrationOpen || !totpConfigured">{{ t('auth.setupAuthenticator') }}</button>
        <p v-if="!registrationOpen" class="form-notice" role="status">{{ t('auth.registrationClosed') }}</p>
        <p v-else-if="!totpConfigured" class="form-notice error" role="status">{{ t('auth.totpUnavailable') }}</p>
        <p v-if="error" class="form-notice error" role="alert">{{ error }}</p>
        <p class="auth-alternate"><RouterLink to="/login">{{ t('auth.signIn') }}</RouterLink></p>
      </form>
    </section>
    <AuthenticatorEnrollmentDialog
      v-if="enrollment"
      v-model:code="code"
      :title="t('auth.connectAuthenticator')"
      :help="t('auth.scanAuthenticator')"
      :qr-code="qrCode"
      :secret="enrollment.secret"
      :code-label="t('auth.authenticatorCode')"
      :manual-key-label="t('auth.manualKey')"
      :qr-alt="t('auth.qrCode')"
      :confirm-label="t('auth.finishAccount')"
      :cancel-label="t('common.cancel')"
      :error="error"
      :submitting="auth.loading"
      @confirm="finish"
      @close="closeEnrollment"
    />
  </main>
</template>
