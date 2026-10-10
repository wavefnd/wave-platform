<script setup lang="ts">
import WavePlayground from '../components/WavePlayground.vue'
import { firstWaveProgram } from '../services/homeOnboarding'
import { useI18n } from '../i18n'
const { t } = useI18n()
</script>
<template>
  <main class="portal-width playground-page">
    <h1>{{ t('playground.title') }}</h1>
    <p>{{ t('playground.intro') }}</p>
    <WavePlayground :source="firstWaveProgram" />
    <p><RouterLink to="/docs/language/program-structure">{{ t('docs.firstProgram') }}</RouterLink></p>
  </main>
</template>
<style scoped>
.playground-page { max-width: 960px; padding-top: 32px; padding-bottom: 48px; }
.playground-page > p { color: var(--wave-text-secondary); }
</style>
