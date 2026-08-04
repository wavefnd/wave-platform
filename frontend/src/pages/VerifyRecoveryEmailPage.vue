<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { useI18n } from '../i18n'
import { verifyRecoveryEmail } from '../services/http'

const { t } = useI18n()
const route = useRoute()
const state = ref<'loading' | 'complete' | 'error'>('loading')

onMounted(async () => {
  const token = typeof route.query.token === 'string' ? route.query.token : ''
  try { await verifyRecoveryEmail(token); state.value = 'complete' } catch { state.value = 'error' }
})
</script>

<template>
  <main class="page shell auth-page"><section class="auth-panel"><header class="auth-heading"><img src="/img/wave-logo.ico" alt="" /><h1>{{ t('auth.verifyRecovery') }}</h1></header><div class="auth-form"><p v-if="state === 'loading'">{{ t('common.loading') }}</p><p v-else-if="state === 'complete'" class="form-notice">{{ t('auth.recoveryVerified') }}</p><p v-else class="form-notice error">{{ t('auth.verificationExpired') }}</p><RouterLink class="portal-button centered" to="/account/security">{{ t('auth.securitySettings') }}</RouterLink></div></section></main>
</template>
