import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import * as api from '../services/http'

export const useNotificationStore = defineStore('notifications', () => {
	const items = ref<api.NotificationItem[]>([])
	const unreadCount = ref(0)
	const loading = ref(false)
	const loaded = ref(false)
	const error = ref('')
	const accountId = ref('')

	const hasUnread = computed(() => unreadCount.value > 0)

	async function load(currentAccountId: string, force = false) {
		if (!currentAccountId) return
		if (accountId.value !== currentAccountId) {
			reset()
			accountId.value = currentAccountId
		}
		if (loading.value || (loaded.value && !force)) return
		loading.value = true
		error.value = ''
		try {
			const result = await api.getNotifications()
			items.value = result.items
			unreadCount.value = result.unreadCount
			loaded.value = true
		} catch (cause) {
			error.value = cause instanceof Error ? cause.message : String(cause)
		} finally {
			loading.value = false
		}
	}

	async function markRead(id: string) {
		const current = items.value.find((item) => item.id === id)
		if (!current || current.readAt) return
		const updated = await api.markNotificationRead(id)
		items.value = items.value.map((item) => item.id === id ? updated : item)
		unreadCount.value = Math.max(0, unreadCount.value - 1)
	}

	async function markAllRead() {
		if (!hasUnread.value) return
		await api.markAllNotificationsRead()
		const readAt = new Date().toISOString()
		items.value = items.value.map((item) => item.readAt ? item : { ...item, readAt })
		unreadCount.value = 0
	}

	function reset() {
		items.value = []
		unreadCount.value = 0
		loading.value = false
		loaded.value = false
		error.value = ''
		accountId.value = ''
	}

	return { items, unreadCount, loading, loaded, error, hasUnread, load, markRead, markAllRead, reset }
})
