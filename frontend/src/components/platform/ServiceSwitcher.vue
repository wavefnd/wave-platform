<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { BookOpenText, ChevronDown, CircleHelp, GitFork, Grid2X2, Mail, MessagesSquare, Newspaper, NotebookPen, ScrollText, ShieldCheck } from '@lucide/vue'

import { useI18n } from '../../i18n'
import { useAuthStore } from '../../stores/auth'

defineProps<{ compact?: boolean }>()
const { t } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const menu = ref<HTMLElement | null>(null)
const moreOpen = ref(false)

const primaryServices = [
	{ to: '/', key: 'nav.home', icon: Grid2X2, exact: true },
	{ to: '/blog', key: 'nav.blog', icon: Newspaper, prefixes: ['/blog', '/releases'] },
	{ to: '/docs', key: 'nav.docs', icon: BookOpenText, prefixes: ['/docs'] },
	{ to: '/community', key: 'nav.community', icon: MessagesSquare, prefixes: ['/community'] },
	{ to: '/source', key: 'nav.source', icon: GitFork, prefixes: ['/source', '/patches'] },
]
const exploreGroups = computed(() => [
	{
		key: 'nav.discuss',
		items: [
			{ to: '/rfcs', key: 'nav.rfc', icon: ScrollText },
			{ to: '/questions', key: 'nav.questions', icon: CircleHelp },
		],
	},
	{
		key: 'nav.workspace',
		items: [
			{ to: '/mail', key: 'nav.mail', icon: Mail },
			{ to: '/lunastev', key: 'nav.personal', icon: NotebookPen },
		],
	},
	...(auth.account?.owner || auth.account?.administrator ? [{ key: 'nav.manage', items: [{ to: '/admin', key: 'nav.admin', icon: ShieldCheck }] }] : []),
])
const exploreActive = computed(() => exploreGroups.value.some((group) => group.items.some((item) => route.path === item.to || route.path.startsWith(`${item.to}/`))))

function primaryActive(item: typeof primaryServices[number]) {
	if (item.exact) return route.path === item.to
	return item.prefixes?.some((prefix) => route.path === prefix || route.path.startsWith(`${prefix}/`))
}

function closeOutside(event: PointerEvent) {
	if (moreOpen.value && !menu.value?.contains(event.target as Node)) moreOpen.value = false
}

function closeEscape(event: KeyboardEvent) {
	if (event.key === 'Escape') moreOpen.value = false
}

watch(() => route.fullPath, () => { moreOpen.value = false })
onMounted(() => {
	document.addEventListener('pointerdown', closeOutside)
	document.addEventListener('keydown', closeEscape)
})
onBeforeUnmount(() => {
	document.removeEventListener('pointerdown', closeOutside)
	document.removeEventListener('keydown', closeEscape)
})
</script>

<template>
  <nav :class="['service-switcher', { compact }]" :aria-label="t('nav.primary')">
    <div class="service-primary">
      <RouterLink v-for="service in primaryServices" :key="service.to" :to="service.to" :class="{ active: primaryActive(service) }">
        <component :is="service.icon" :size="compact ? 15 : 18" :stroke-width="1.8" aria-hidden="true" />
        <span>{{ t(service.key) }}</span>
      </RouterLink>
    </div>
    <div ref="menu" class="service-more" :class="{ open: moreOpen, active: exploreActive }">
      <button type="button" :aria-expanded="moreOpen" aria-haspopup="true" @click="moreOpen = !moreOpen">
        <span>{{ t('nav.explore') }}</span><ChevronDown :size="14" aria-hidden="true" />
      </button>
      <div class="service-more-panel">
        <section v-for="group in exploreGroups" :key="group.key">
          <strong>{{ t(group.key) }}</strong>
          <RouterLink v-for="item in group.items" :key="item.to" :to="item.to">
            <component :is="item.icon" :size="17" :stroke-width="1.8" aria-hidden="true" />
            <span>{{ t(item.key) }}</span>
          </RouterLink>
        </section>
      </div>
    </div>
  </nav>
</template>
