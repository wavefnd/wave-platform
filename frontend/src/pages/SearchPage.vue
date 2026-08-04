<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useI18n } from '../i18n'

const route = useRoute()
const { t } = useI18n()
const query = computed(() => typeof route.query.q === 'string' ? route.query.q.trim() : '')
</script>

<template>
  <main class="page shell utility-page search-page">
    <div class="portal-heading">
      <h1>{{ t('search.title') }}</h1>
      <p>{{ t('search.lead') }}</p>
    </div>

    <section class="utility-surface search-results">
      <p v-if="query" class="search-query"><span>“</span>{{ query }}<span>”</span></p>
      <p class="empty-state">{{ t('search.empty') }}</p>
    </section>
  </main>
</template>
