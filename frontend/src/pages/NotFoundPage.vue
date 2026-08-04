<script setup lang="ts">
import { useI18n } from '../i18n'

const { t } = useI18n()
</script>

<template>
  <main class="page shell utility-page not-found-page">
    <section class="utility-surface">
      <span class="not-found-code">404</span>
      <h1 class="page-title">{{ t('notFound.title') }}</h1>
      <p class="page-lead">{{ t('notFound.lead') }}</p>
      <div class="hero-actions">
        <RouterLink class="button" to="/">{{ t('notFound.home') }}</RouterLink>
      </div>
    </section>
  </main>
</template>
