<script setup lang="ts">
import QRCode from 'qrcode'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import AuthenticatorEnrollmentDialog from '../components/auth/AuthenticatorEnrollmentDialog.vue'
import GmailDeliveryWarning from '../components/GmailDeliveryWarning.vue'
import { useI18n } from '../i18n'
import { containsGmailAddress } from '../services/email-address'
import { beginTOTPRotation, changeRecoveryEmail, finishTOTPRotation, getAccountSecurity, type AccountSecurity, type TOTPEnrollment } from '../services/http'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const security = ref<AccountSecurity | null>(null)
const currentCode = ref('')
const newCode = ref('')
const recoveryEmail = ref('')
const enrollment = ref<TOTPEnrollment | null>(null)
const qrCode = ref('')
const notice = ref('')
const error = ref('')
const gmailRecovery = computed(() => containsGmailAddress(recoveryEmail.value || security.value?.recoveryEmail || ''))

async function load() { security.value = await getAccountSecurity() }
onMounted(async () => { try { await load() } catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') } })

async function beginRotation() {
  error.value = ''; notice.value = ''
  try {
    const result = await beginTOTPRotation(currentCode.value)
    qrCode.value = await QRCode.toDataURL(result.uri, { width: 208, margin: 1 })
    enrollment.value = result
  }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}

function closeRotation() { enrollment.value = null; qrCode.value = ''; newCode.value = ''; error.value = '' }

async function finishRotation() {
  if (!enrollment.value) return
  error.value = ''
  try { await finishTOTPRotation(enrollment.value.token, newCode.value); await auth.initialize(true); await router.replace('/login') }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}

async function updateRecovery() {
  error.value = ''; notice.value = ''
  try { await changeRecoveryEmail(recoveryEmail.value, currentCode.value); notice.value = t('auth.verificationSent'); recoveryEmail.value = ''; currentCode.value = '' }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
}

</script>

<template>
  <main class="page shell settings-page">
    <header class="utility-heading"><h1 class="page-title">{{ t('auth.securitySettings') }}</h1></header>
    <section class="settings-section">
      <h2>{{ t('auth.authenticator') }}</h2>
      <p>{{ t('auth.authenticatorStatus') }} <strong>{{ security?.totpEnabled ? t('common.enabled') : t('common.disabled') }}</strong></p>
      <form class="settings-form" @submit.prevent="beginRotation">
        <label for="rotate-current">{{ t('auth.currentCode') }}</label><input id="rotate-current" v-model="currentCode" inputmode="numeric" autocomplete="one-time-code" pattern="[0-9]{6}" maxlength="6" required />
        <button class="portal-button" type="submit">{{ t('auth.replaceAuthenticator') }}</button>
      </form>
    </section>
    <section class="settings-section">
      <h2>{{ t('auth.recoveryEmail') }}</h2>
      <p v-if="security">{{ security.recoveryEmail }} · {{ security.recoveryVerified ? t('auth.verified') : t('auth.unverified') }}</p>
      <form class="settings-form" @submit.prevent="updateRecovery">
        <label for="recovery-new">{{ t('auth.newRecoveryEmail') }}</label><input id="recovery-new" v-model="recoveryEmail" type="email" autocomplete="email" required />
        <GmailDeliveryWarning v-if="gmailRecovery" />
        <label for="recovery-code">{{ t('auth.currentCode') }}</label><input id="recovery-code" v-model="currentCode" inputmode="numeric" autocomplete="one-time-code" pattern="[0-9]{6}" maxlength="6" required />
        <button class="portal-button" type="submit">{{ t('auth.sendVerification') }}</button>
      </form>
    </section>
    <p v-if="notice" class="form-notice" role="status">{{ notice }}</p>
    <p v-if="notice" class="mail-delivery-hint">{{ t('auth.checkSpam') }}</p>
    <p v-if="error" class="form-notice error" role="alert">{{ error }}</p>
    <AuthenticatorEnrollmentDialog
      v-if="enrollment"
      v-model:code="newCode"
      :title="t('auth.replaceAuthenticator')"
      :help="t('auth.scanAuthenticator')"
      :qr-code="qrCode"
      :secret="enrollment.secret"
      :code-label="t('auth.newCode')"
      :manual-key-label="t('auth.manualKey')"
      :qr-alt="t('auth.qrCode')"
      :confirm-label="t('auth.confirmReplacement')"
      :cancel-label="t('common.cancel')"
      :error="error"
      @confirm="finishRotation"
      @close="closeRotation"
    />
  </main>
</template>
