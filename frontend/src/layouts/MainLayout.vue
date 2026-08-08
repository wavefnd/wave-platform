<script setup lang="ts">
import { computed, onMounted, ref, watch, watchEffect } from 'vue'
import { Menu, X } from '@lucide/vue'
import { RouterView, useRoute } from 'vue-router'

import PlatformAccount from '../components/platform/PlatformAccount.vue'
import PlatformIdentity from '../components/platform/PlatformIdentity.vue'
import ServiceSearch from '../components/platform/ServiceSearch.vue'
import ServiceSwitcher from '../components/platform/ServiceSwitcher.vue'
import { useI18n } from '../i18n'
import { useAuthStore } from '../stores/auth'
import { updateSEO } from '../services/seo'

const route = useRoute()
const { locale, t } = useI18n()
const menuOpen = ref(false)
const auth = useAuthStore()

onMounted(() => auth.initialize())

const service = computed(() => {
  const path = route.path
  if (path.startsWith('/mail')) return 'mail'
  if (path.startsWith('/community')) return 'community'
  if (path.startsWith('/lunastev')) return 'personal'
  if (path.startsWith('/docs')) return 'docs'
  if (path.startsWith('/questions')) return 'questions'
  if (path.startsWith('/source')) return 'source'
  if (path.startsWith('/admin')) return 'admin'
	if (path.startsWith('/account') || path.startsWith('/login') || path.startsWith('/register')) return 'account'
  if (path.startsWith('/releases')) return 'portal'
  if (path.startsWith('/search')) return 'portal'
  return 'portal'
})

const serviceLabel = computed(() => {
  if (service.value === 'portal') return t('brand.subtitle')
  if (service.value === 'account') return t('auth.account')
  return t(`nav.${service.value}`)
})
const searchLabel = computed(() => t('search.scope.portal'))
const showHeaderSearch = computed(() => !['source', 'account'].includes(service.value))
const showFooter = computed(() => !['mail', 'source'].includes(service.value))
const isPortal = computed(() => service.value === 'portal' && route.name === 'home')
const isAdmin = computed(() => service.value === 'admin')

watch(() => route.fullPath, () => { menuOpen.value = false })
watchEffect(() => updateSEO(route, locale.value, service.value as Parameters<typeof updateSEO>[2]))
</script>

<template>
  <div class="app-shell" :data-service="service">
    <a class="skip-link" href="#main-content">{{ t('common.skipContent') }}</a>
    <header v-if="!isAdmin" :class="['platform-header', { 'is-portal': isPortal }]">
      <div class="platform-header-row">
        <PlatformIdentity :service="serviceLabel" />
        <ServiceSearch v-if="showHeaderSearch" :label="searchLabel" :compact="!isPortal" />
        <PlatformAccount />
        <button
          class="platform-menu-button"
          type="button"
          :aria-expanded="menuOpen"
          :aria-label="t('nav.menu')"
          @click="menuOpen = !menuOpen"
        >
          <X v-if="menuOpen" :size="20" aria-hidden="true" />
          <Menu v-else :size="20" aria-hidden="true" />
        </button>
      </div>
      <div :class="['platform-switcher-row', { open: menuOpen }]">
        <ServiceSwitcher :compact="!isPortal" />
        <PlatformAccount class="mobile-platform-account" />
      </div>
    </header>

    <RouterView v-slot="{ Component }">
      <component :is="Component" id="main-content" />
    </RouterView>

    <footer v-if="showFooter && !isAdmin" class="site-footer">
      <div class="portal-width footer-row">
        <span>Wave Platform</span>
      </div>
    </footer>
  </div>
</template>
