<script setup lang="ts">
defineProps<{ message: string; action?: string }>()
defineEmits<{ action: [] }>()
</script>

<template>
  <div class="ui-inline-state" role="status">
    <span>{{ message }}</span>
    <button v-if="action" type="button" @click="$emit('action')">{{ action }}</button>
  </div>
</template>
