<script setup lang="ts">
import type { SourceLanguage } from './types'

defineProps<{
  languages: SourceLanguage[]
  title: string
  subtitle: string
  primaryLabel: string
}>()
</script>

<template>
  <div class="language-breakdown-content">
    <header>
      <strong>{{ title }}</strong>
      <small v-if="subtitle">{{ subtitle }}</small>
    </header>
    <div class="language-bar" :aria-label="title">
      <span
        v-for="language in languages"
        :key="language.name"
        :style="{ width: `${language.percent}%`, backgroundColor: language.color }"
      />
    </div>
    <ul class="language-legend">
      <li v-for="(language, index) in languages" :key="language.name">
        <i :style="{ backgroundColor: language.color }" />
        <strong>{{ language.name }}</strong>
        <span>{{ language.percent.toFixed(2) }}%</span>
        <small v-if="index === 0 && primaryLabel">{{ primaryLabel }}</small>
      </li>
    </ul>
  </div>
</template>
