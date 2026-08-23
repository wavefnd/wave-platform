<script setup lang="ts">
import { Inbox, Mail, ShieldCheck, Users } from '@lucide/vue'
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import MailingListWorkspace from '../components/mail/MailingListWorkspace.vue'
import MailboxWorkspace from '../components/mail/MailboxWorkspace.vue'
import { useI18n } from '../i18n'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const mode = computed<'personal' | 'lists' | 'team'>(() => {
	const current = String(route.params.mode ?? '')
	if (current === 'lists' || route.path === '/mail/lists' || route.path.startsWith('/mail/lists/')) return 'lists'
	if (current === 'team' || route.path === '/mail/team') return 'team'
	return 'personal'
})
const canUseTeamMail = computed(() => Boolean(auth.account?.owner || auth.account?.administrator))
const modes = computed(() => [
	{ id: 'personal', label: t('mail.personal'), detail: t('mail.personalDetail'), icon: Inbox, to: '/mail/personal' },
	{ id: 'lists', label: t('mail.lists.title'), detail: t('mail.lists.modeDetail'), icon: Users, to: '/mail/lists' },
	...(canUseTeamMail.value ? [{ id: 'team', label: t('mail.team'), detail: t('mail.teamDetail'), icon: ShieldCheck, to: '/mail/team' }] : []),
])

function changeMobileMode(event: Event) {
	const target = event.target as HTMLSelectElement
	void router.push(target.value)
}

async function enforceTeamAccess() {
	if (auth.initialized && mode.value === 'team' && !canUseTeamMail.value) await router.replace('/mail/personal')
}

onMounted(async () => {
	await auth.initialize()
	await enforceTeamAccess()
})
watch([mode, () => auth.initialized], enforceTeamAccess)
</script>

<template>
	<main class="mail-service">
		<section v-if="auth.initialized && !auth.account" class="mail-signin" aria-labelledby="mail-title">
			<Mail :size="34" :stroke-width="1.5" aria-hidden="true" />
			<h1 id="mail-title">Wave Mail</h1>
			<p>{{ t('mail.signInRequired') }}</p>
			<RouterLink class="ui-button primary" :to="{ path: '/login', query: { redirect: route.fullPath } }">{{ t('mail.signIn') }}</RouterLink>
			<RouterLink class="mail-create-account" to="/register">{{ t('auth.signUp') }}</RouterLink>
		</section>

		<template v-else-if="auth.account">
			<header class="mail-hub-header">
				<div class="mail-hub-title"><Mail :size="20" aria-hidden="true" /><div><strong>Wave Mail</strong><span>{{ modes.find((item) => item.id === mode)?.detail }}</span></div></div>
				<nav class="mail-mode-tabs" :aria-label="t('mail.hubModes')">
					<RouterLink v-for="item in modes" :key="item.id" :to="item.to" :class="{ active: mode === item.id }">
						<component :is="item.icon" :size="16" aria-hidden="true" />{{ item.label }}
					</RouterLink>
				</nav>
				<label class="mail-mode-mobile"><span class="sr-only">{{ t('mail.hubModes') }}</span><select :value="modes.find((item) => item.id === mode)?.to" @change="changeMobileMode"><option v-for="item in modes" :key="item.id" :value="item.to">{{ item.label }}</option></select></label>
			</header>

			<MailboxWorkspace v-if="mode === 'personal'" mode="personal" />
			<MailingListWorkspace v-else-if="mode === 'lists'" />
			<MailboxWorkspace v-else-if="canUseTeamMail" mode="team" />
		</template>
	</main>
</template>
