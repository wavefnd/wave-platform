<script setup lang="ts">
import { Bell, CheckCheck } from '@lucide/vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import { useI18n } from '../../i18n'
import type { NotificationItem } from '../../services/http'
import { useAuthStore } from '../../stores/auth'
import { useNotificationStore } from '../../stores/notifications'

const { locale, t } = useI18n()
const notifications = useNotificationStore()
const auth = useAuthStore()
const root = ref<HTMLElement | null>(null)
const open = ref(false)

const badge = computed(() => notifications.unreadCount > 99 ? '99+' : String(notifications.unreadCount))
const triggerLabel = computed(() => notifications.hasUnread
	? `${t('notifications.title')}: ${notifications.unreadCount} ${t('notifications.unread')}`
	: t('notifications.title'))

function kindLabel(item: NotificationItem) {
	return t(`notifications.kind.${item.kind}`)
}

function relativeTime(value: string) {
	const instant = new Date(value)
	if (Number.isNaN(instant.getTime())) return ''
	const seconds = Math.round((instant.getTime() - Date.now()) / 1000)
	const formatter = new Intl.RelativeTimeFormat(locale.value === 'ko' ? 'ko' : 'en', { numeric: 'auto' })
	if (Math.abs(seconds) < 60) return formatter.format(seconds, 'second')
	const minutes = Math.round(seconds / 60)
	if (Math.abs(minutes) < 60) return formatter.format(minutes, 'minute')
	const hours = Math.round(minutes / 60)
	if (Math.abs(hours) < 24) return formatter.format(hours, 'hour')
	return formatter.format(Math.round(hours / 24), 'day')
}

async function toggle() {
	open.value = !open.value
	if (open.value && auth.account) await notifications.load(auth.account.id, true)
}

async function select(item: NotificationItem) {
	open.value = false
	try { await notifications.markRead(item.id) } catch { /* Navigation remains available if the read update fails. */ }
}

async function markAll() {
	try { await notifications.markAllRead() } catch { /* Keep unread state if the update fails. */ }
}

function closeFromOutside(event: PointerEvent) {
	if (open.value && !root.value?.contains(event.target as Node)) open.value = false
}

function closeFromEscape(event: KeyboardEvent) {
	if (event.key === 'Escape') open.value = false
}

onMounted(() => {
	if (auth.account) void notifications.load(auth.account.id)
	document.addEventListener('pointerdown', closeFromOutside)
	document.addEventListener('keydown', closeFromEscape)
})

onBeforeUnmount(() => {
	document.removeEventListener('pointerdown', closeFromOutside)
	document.removeEventListener('keydown', closeFromEscape)
})
</script>

<template>
	<div ref="root" class="notification-menu">
		<button class="notification-trigger" type="button" :aria-label="triggerLabel" :title="t('notifications.title')" aria-haspopup="dialog" :aria-expanded="open" @click="toggle">
			<Bell :size="17" aria-hidden="true" />
			<span v-if="notifications.hasUnread" class="notification-badge" aria-hidden="true">{{ badge }}</span>
		</button>
		<section v-if="open" class="notification-panel" role="dialog" :aria-label="t('notifications.title')">
			<header>
				<strong>{{ t('notifications.title') }}</strong>
				<button v-if="notifications.hasUnread" type="button" :disabled="notifications.loading" @click="markAll">
					<CheckCheck :size="15" aria-hidden="true" />{{ t('notifications.markAllRead') }}
				</button>
			</header>
			<p v-if="notifications.loading && !notifications.loaded" class="notification-state">{{ t('notifications.loading') }}</p>
			<p v-else-if="notifications.error && !notifications.items.length" class="notification-state">{{ t('notifications.failed') }}</p>
			<p v-else-if="!notifications.items.length" class="notification-state">{{ t('notifications.empty') }}</p>
			<ul v-else class="notification-list">
				<li v-for="item in notifications.items" :key="item.id" :class="{ unread: !item.readAt }">
					<RouterLink :to="item.url" @click="select(item)">
						<span class="notification-copy">
							<strong>{{ kindLabel(item) }}</strong>
							<span>{{ item.subject }}<template v-if="item.detail"> · {{ item.detail }}</template></span>
							<small><template v-if="item.actorName">{{ item.actorName }} · </template>{{ relativeTime(item.createdAt) }}</small>
						</span>
						<span v-if="!item.readAt" class="notification-unread-dot" :aria-label="t('notifications.unread')"></span>
					</RouterLink>
				</li>
			</ul>
		</section>
	</div>
</template>
