<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ patch: string }>()

const lines = computed(() => props.patch.replaceAll('\r\n', '\n').split('\n').map((text) => {
  let kind = 'context'
  if (text.startsWith('diff --git ')) kind = 'file'
  else if (text.startsWith('@@')) kind = 'hunk'
  else if (text.startsWith('+') && !text.startsWith('+++')) kind = 'addition'
  else if (text.startsWith('-') && !text.startsWith('---')) kind = 'deletion'
  else if (text.startsWith('index ') || text.startsWith('--- ') || text.startsWith('+++ ')) kind = 'meta'
  return { text, kind }
}))
</script>

<template>
  <pre class="source-diff"><code><span
    v-for="(line, index) in lines"
    :key="index"
    :class="`source-diff-${line.kind}`"
  >{{ line.text }}
</span></code></pre>
</template>
