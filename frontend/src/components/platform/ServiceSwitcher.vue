<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { BookOpenText, CircleHelp, GitFork, Grid2X2, Mail, MessagesSquare, Newspaper, NotebookPen, ShieldCheck } from '@lucide/vue'

import { useI18n } from '../../i18n'
import { useAuthStore } from '../../stores/auth'

defineProps<{ compact?: boolean }>()
const { t } = useI18n()
const auth = useAuthStore()
const route = useRoute()

const publicServices = [
  { to: '/', key: 'nav.home', icon: Grid2X2 },
	{ to: '/blog', key: 'nav.blog', icon: Newspaper },
  { to: '/mail', key: 'nav.mail', icon: Mail },
  { to: '/community', key: 'nav.community', icon: MessagesSquare },
  { to: '/lunastev', key: 'nav.personal', icon: NotebookPen },
  { to: '/docs', key: 'nav.docs', icon: BookOpenText },
  { to: '/questions', key: 'nav.questions', icon: CircleHelp },
  { to: '/source', key: 'nav.source', icon: GitFork },
]
const services = computed(() => auth.account?.owner || auth.account?.administrator
  ? [...publicServices, { to: '/admin', key: 'nav.admin', icon: ShieldCheck }]
  : publicServices)
</script>

<template>
  <nav :class="['service-switcher', { compact }]" :aria-label="t('nav.primary')">
    <RouterLink v-for="service in services" :key="service.to" :to="service.to" :class="{ 'router-link-active': service.key === 'nav.blog' && route.path.startsWith('/releases') }">
      <component :is="service.icon" :size="compact ? 15 : 18" :stroke-width="1.8" aria-hidden="true" />
      <span>{{ t(service.key) }}</span>
    </RouterLink>
  </nav>
</template>
