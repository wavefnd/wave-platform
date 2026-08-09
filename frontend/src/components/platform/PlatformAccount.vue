<script setup lang="ts">
import { useI18n, type Locale } from '../../i18n'
import { useAuthStore } from '../../stores/auth'
import ThemeSelector from './ThemeSelector.vue'

const { locale, setLocale, t } = useI18n()
const auth = useAuthStore()

function changeLocale(event: Event) {
  setLocale((event.target as HTMLSelectElement).value as Locale)
}

async function signOut() {
	await auth.signOut()
}
</script>

<template>
  <div class="platform-account">
    <ThemeSelector />
    <select :value="locale" :aria-label="t('nav.language')" @change="changeLocale">
      <option value="en">English</option>
      <option value="ko">한국어</option>
    </select>
    <template v-if="auth.account">
      <span class="account-address" :title="auth.account.email">{{ auth.account.displayName }}</span>
	  <RouterLink :to="`/user/${encodeURIComponent(auth.account.username)}`">{{ t('user.profile') }}</RouterLink>
	  <RouterLink to="/account/security">{{ t('auth.settings') }}</RouterLink>
      <button class="account-signout" type="button" :disabled="auth.loading" @click="signOut">{{ t('auth.signOut') }}</button>
    </template>
    <template v-else>
      <RouterLink to="/login">{{ t('auth.signIn') }}</RouterLink>
      <RouterLink class="account-create" to="/register">{{ t('auth.signUp') }}</RouterLink>
    </template>
  </div>
</template>
