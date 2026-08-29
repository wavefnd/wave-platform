<script setup lang="ts">
import { ChevronDown } from '@lucide/vue'
import { onBeforeUnmount, onMounted, ref } from 'vue'

import { useI18n, type Locale } from '../../i18n'
import { useAuthStore } from '../../stores/auth'
import DiscordMark from '../icons/DiscordMark.vue'
import GitHubMark from '../icons/GitHubMark.vue'
import ThemeSelector from './ThemeSelector.vue'

const { locale, setLocale, t } = useI18n()
const auth = useAuthStore()
const accountMenu = ref<HTMLElement | null>(null)
const menuOpen = ref(false)

function changeLocale(event: Event) {
  setLocale((event.target as HTMLSelectElement).value as Locale)
}

async function signOut() {
	menuOpen.value = false
	await auth.signOut()
}

function closeFromOutside(event: PointerEvent) {
	if (menuOpen.value && !accountMenu.value?.contains(event.target as Node)) menuOpen.value = false
}

function closeFromEscape(event: KeyboardEvent) {
	if (event.key === 'Escape') menuOpen.value = false
}

onMounted(() => {
	document.addEventListener('pointerdown', closeFromOutside)
	document.addEventListener('keydown', closeFromEscape)
})
onBeforeUnmount(() => {
	document.removeEventListener('pointerdown', closeFromOutside)
	document.removeEventListener('keydown', closeFromEscape)
})
</script>

<template>
  <div class="platform-account">
    <ThemeSelector />
    <a class="external-icon-link" href="https://github.com/wavefnd/Wave" target="_blank" rel="noopener noreferrer" :aria-label="t('nav.github')" :title="t('nav.github')">
      <GitHubMark :size="17" />
    </a>
    <a class="external-icon-link" href="https://discord.gg/3nev5nHqq9" target="_blank" rel="noopener noreferrer" :aria-label="t('nav.discord')" :title="t('nav.discord')">
      <DiscordMark :size="17" />
    </a>
    <select :value="locale" :aria-label="t('nav.language')" @change="changeLocale">
      <option value="en">English</option>
      <option value="ko">한국어</option>
    </select>
    <template v-if="auth.account">
	  <div ref="accountMenu" class="account-menu">
		<button class="account-menu-trigger" type="button" :aria-expanded="menuOpen" aria-haspopup="menu" @click="menuOpen = !menuOpen">
		  <span :title="auth.account.email">{{ auth.account.displayName }}</span><ChevronDown :size="14" aria-hidden="true" />
		</button>
		<div v-if="menuOpen" class="account-menu-panel" role="menu">
		  <RouterLink role="menuitem" :to="`/user/${encodeURIComponent(auth.account.username)}`" @click="menuOpen = false">{{ t('user.profile') }}</RouterLink>
		  <RouterLink role="menuitem" to="/account/security" @click="menuOpen = false">{{ t('auth.settings') }}</RouterLink>
		  <button class="account-signout" role="menuitem" type="button" :disabled="auth.loading" @click="signOut">{{ t('auth.signOut') }}</button>
		</div>
	  </div>
    </template>
    <template v-else>
      <RouterLink to="/login">{{ t('auth.signIn') }}</RouterLink>
      <RouterLink class="account-create" to="/register">{{ t('auth.signUp') }}</RouterLink>
    </template>
  </div>
</template>
