<script setup lang="ts">
defineProps<{ service: string }>()
</script>

<template>
  <RouterLink class="platform-identity" to="/" aria-label="Wave Platform">
    <img src="/img/wave-logo.ico" alt="" />
    <span>Wave</span>
    <small v-if="service">{{ service }}</small>
  </RouterLink>
</template>
