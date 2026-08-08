<script setup lang="ts">
import { useI18n } from '../../i18n'
import { useTheme, type ThemePreference } from '../../theme'

const { t } = useI18n()
const { preference, setTheme } = useTheme()

function changeTheme(event: Event) {
  setTheme((event.target as HTMLSelectElement).value as ThemePreference)
}
</script>

<template>
  <select class="theme-selector" :value="preference" :aria-label="t('theme.label')" :title="t('theme.label')" @change="changeTheme">
    <option value="system">{{ t('theme.system') }}</option>
    <option value="light">{{ t('theme.light') }}</option>
    <option value="black">{{ t('theme.black') }}</option>
  </select>
</template>
