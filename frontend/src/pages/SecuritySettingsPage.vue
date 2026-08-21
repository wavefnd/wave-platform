<script setup lang="ts">
import QRCode from 'qrcode'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import AuthenticatorEnrollmentDialog from '../components/auth/AuthenticatorEnrollmentDialog.vue'
import GmailDeliveryWarning from '../components/GmailDeliveryWarning.vue'
import { useI18n } from '../i18n'
import { containsGmailAddress } from '../services/email-address'
import {
  beginTOTPRotation, changeRecoveryEmail, deleteAccountWebhook, finishTOTPRotation, getAccountSecurity, getAccountWebhooks,
  saveAccountWebhook, testAccountWebhook, type AccountSecurity, type TOTPEnrollment, type WebhookAdminView, type WebhookInput,
} from '../services/http'
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
const webhooks = ref<WebhookAdminView>({ supportedEvents: [], endpoints: [], deliveries: [] })
const webhookForm = ref<WebhookInput>({ id: '', name: '', kind: 'generic', url: '', events: [], enabled: true, rotateSecret: false })
const webhookSecret = ref('')
const webhookBusy = ref(false)

async function load() { [security.value, webhooks.value] = await Promise.all([getAccountSecurity(), getAccountWebhooks()]) }
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

function newWebhook() { webhookSecret.value = ''; webhookForm.value = { id: '', name: '', kind: 'generic', url: '', events: [], enabled: true, rotateSecret: false } }
function editWebhook(id: string) {
  const item = webhooks.value.endpoints.find((endpoint) => endpoint.id === id); if (!item) return
  webhookSecret.value = ''; webhookForm.value = { id, name: item.name, kind: item.kind, url: '', events: [...item.events], enabled: item.enabled, rotateSecret: false }
}
async function saveWebhook() {
  webhookBusy.value = true; error.value = ''; notice.value = ''
  try { const saved = await saveAccountWebhook(webhookForm.value); webhooks.value = await getAccountWebhooks(); editWebhook(saved.id); webhookSecret.value = saved.signingSecret; notice.value = t('auth.webhookSaved') }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
  finally { webhookBusy.value = false }
}
async function sendWebhookTest(id: string) {
  webhookBusy.value = true; error.value = ''; notice.value = ''
  try { await testAccountWebhook(id); webhooks.value = await getAccountWebhooks(); notice.value = t('auth.webhookTestSent') }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
  finally { webhookBusy.value = false }
}
async function removeWebhook(id: string) {
  if (!window.confirm(t('admin.confirmDeleteWebhook'))) return
  webhookBusy.value = true; error.value = ''
  try { await deleteAccountWebhook(id); newWebhook(); webhooks.value = await getAccountWebhooks() }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('auth.failed') }
  finally { webhookBusy.value = false }
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
    <section class="settings-section account-webhooks">
      <header><div><h2>{{ t('admin.webhooks') }}</h2><p>{{ t('auth.webhookHelp') }}</p></div><button class="text-button" type="button" @click="newWebhook">{{ t('admin.newWebhook') }}</button></header>
      <form class="settings-form webhook-settings-form" @submit.prevent="saveWebhook">
        <label for="webhook-name">{{ t('admin.webhookName') }}</label><input id="webhook-name" v-model="webhookForm.name" required maxlength="80" />
        <label for="webhook-kind">{{ t('admin.webhookKind') }}</label><select id="webhook-kind" v-model="webhookForm.kind"><option value="generic">Generic JSON</option><option value="discord">Discord</option></select>
        <label for="webhook-url">{{ t('admin.webhookUrl') }}</label><input id="webhook-url" v-model="webhookForm.url" type="url" :required="!webhookForm.id" placeholder="https://…" /><small v-if="webhookForm.id">{{ t('admin.webhookUrlRetained') }}</small>
        <fieldset><legend>{{ t('admin.webhookEvents') }}</legend><label v-for="event in webhooks.supportedEvents" :key="event"><input v-model="webhookForm.events" type="checkbox" :value="event" /> <code>{{ event }}</code></label></fieldset>
        <label class="checkbox-label"><input v-model="webhookForm.enabled" type="checkbox" /> {{ t('admin.enabled') }}</label>
        <label v-if="webhookForm.id" class="checkbox-label"><input v-model="webhookForm.rotateSecret" type="checkbox" /> {{ t('admin.rotateWebhookSecret') }}</label>
        <button class="portal-button" type="submit" :disabled="webhookBusy || webhookForm.events.length === 0">{{ t('common.save') }}</button>
      </form>
      <div v-if="webhookSecret" class="form-notice"><strong>{{ t('admin.webhookSecretOnce') }}</strong><code class="webhook-secret">{{ webhookSecret }}</code></div>
      <div class="account-webhook-list">
        <article v-for="endpoint in webhooks.endpoints" :key="endpoint.id">
          <div><strong>{{ endpoint.name }}</strong><small>{{ endpoint.kind }} · {{ endpoint.destination }}</small><code v-for="event in endpoint.events" :key="event">{{ event }}</code></div>
          <div><button type="button" @click="editWebhook(endpoint.id)">{{ t('common.edit') }}</button><button type="button" :disabled="webhookBusy" @click="sendWebhookTest(endpoint.id)">{{ t('admin.testWebhook') }}</button><button type="button" :disabled="webhookBusy" @click="removeWebhook(endpoint.id)">{{ t('common.delete') }}</button></div>
        </article>
        <p v-if="webhooks.endpoints.length === 0">{{ t('admin.noWebhooks') }}</p>
      </div>
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
