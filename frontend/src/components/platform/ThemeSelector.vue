<script setup lang="ts">
import { Moon, Sun } from '@lucide/vue'

import { useI18n } from '../../i18n'
import { useTheme } from '../../theme'

const { t } = useI18n()
const { resolved, setTheme } = useTheme()

function toggleTheme() {
  setTheme(resolved.value === 'dark' ? 'light' : 'black')
}
</script>

<template>
  <button class="theme-toggle" type="button" :aria-label="resolved === 'dark' ? t('theme.switchLight') : t('theme.switchBlack')" :title="resolved === 'dark' ? t('theme.switchLight') : t('theme.switchBlack')" @click="toggleTheme">
    <Sun v-if="resolved === 'dark'" :size="17" aria-hidden="true" />
    <Moon v-else :size="17" aria-hidden="true" />
  </button>
</template>
