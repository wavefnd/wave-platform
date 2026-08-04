<script setup lang="ts">
withDefaults(defineProps<{ rows?: number }>(), { rows: 4 })
</script>

<template>
  <div class="ui-skeleton-rows" aria-hidden="true">
    <div v-for="row in rows" :key="row">
      <span />
      <i />
      <small />
    </div>
  </div>
</template>
