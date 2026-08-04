import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import * as api from '../services/http'

export const useAuthStore = defineStore('auth', () => {
	const account = ref<api.AccountSession | null>(null)
	const initialized = ref(false)
	const loading = ref(false)

	const authenticated = computed(() => account.value !== null)

	async function initialize(force = false) {
		if ((initialized.value && !force) || loading.value) return
		loading.value = true
		try { account.value = await api.getCurrentAccount() }
		catch { account.value = null }
		finally { initialized.value = true; loading.value = false }
	}

	async function signIn(identifier: string, code: string, challenge = '') {
		loading.value = true
		try { account.value = await api.login(identifier, code, challenge); initialized.value = true }
		finally { loading.value = false }
	}

	async function completeSignUp(token: string, code: string) {
		loading.value = true
		try { account.value = await api.finishRegistration(token, code); initialized.value = true }
		finally { loading.value = false }
	}

	async function signOut() {
		loading.value = true
		try { await api.logout(); account.value = null; initialized.value = true }
		finally { loading.value = false }
	}

	return { account, initialized, loading, authenticated, initialize, signIn, completeSignUp, signOut }
})
