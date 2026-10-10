<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import { useI18n } from '../i18n'
import { compileWave, WAVE_VERSION } from '../services/playground/client'

const props = defineProps<{ source: string; compact?: boolean; autorun?: boolean }>()
const { t } = useI18n()
const id = useId()
const code = ref(props.source)
const stdin = ref('')
const output = ref('')
const failure = ref('')
const phase = ref<'idle' | 'compile' | 'run' | 'done' | 'stopped'>('idle')
const exitCode = ref<number | null>(null)
const busy = computed(() => phase.value === 'compile' || phase.value === 'run')
let generation = 0
let controller: AbortController | null = null
let worker: Worker | null = null
let timer: ReturnType<typeof setTimeout> | null = null

function release() {
  controller?.abort(); controller = null
  worker?.terminate(); worker = null
  if (timer !== null) clearTimeout(timer)
  timer = null
}
function stop() { generation++; release(); phase.value = 'stopped' }
async function run() {
  if (busy.value) return
  release()
  const current = ++generation
  output.value = ''; failure.value = ''; exitCode.value = null; phase.value = 'compile'
  controller = new AbortController()
  timer = setTimeout(() => {
    if (current !== generation) return
    generation++; release(); phase.value = 'done'; failure.value = t('playground.compileTimeout')
  }, 20000)
  try {
    const wasm = await compileWave(code.value, controller.signal)
    if (current !== generation) return
    if (timer !== null) clearTimeout(timer)
    phase.value = 'run'
    worker = new Worker(new URL('../services/playground/runner.worker.ts', import.meta.url), { type: 'module' })
    timer = setTimeout(() => {
      if (current !== generation) return
      generation++; release(); phase.value = 'done'; failure.value = t('playground.runTimeout')
    }, 3000)
    worker.onmessage = event => {
      if (current !== generation) return
      output.value = event.data.output ?? ''
      failure.value = event.data.error ?? ''
      exitCode.value = event.data.status ?? null
      phase.value = 'done'; release()
    }
    worker.onerror = event => {
      if (current !== generation) return
      failure.value = event.message || t('playground.failed'); phase.value = 'done'; release()
    }
    worker.postMessage({ wasm, stdin: stdin.value }, [wasm])
  } catch (error) {
    if (current !== generation) return
    failure.value = error instanceof Error ? error.message : t('playground.failed')
    phase.value = 'done'; release()
  }
}
watch(() => props.source, value => { stop(); code.value = value; output.value = ''; failure.value = ''; exitCode.value = null; phase.value = 'idle' })
onMounted(() => { if (props.autorun) void run() })
onBeforeUnmount(() => { generation++; release() })
</script>

<template>
  <section class="wave-playground" :class="{ compact }" :aria-label="t('playground.title')">
    <div class="playground-file"><label :for="`${id}-source`">main.wave</label><span>Wave {{ WAVE_VERSION }}</span></div>
    <textarea :id="`${id}-source`" v-model="code" class="playground-source" :rows="compact ? 5 : 15" :disabled="busy" spellcheck="false" autocapitalize="off" autocomplete="off" @keydown.ctrl.enter.prevent="run" @keydown.meta.enter.prevent="run" />
    <details class="playground-input"><summary>{{ t('playground.stdin') }}</summary><label :for="`${id}-stdin`">{{ t('playground.stdinHelp') }}</label><textarea :id="`${id}-stdin`" v-model="stdin" :disabled="busy" rows="3" maxlength="8192" spellcheck="false" /></details>
    <div class="playground-actions">
      <button v-if="busy" type="button" @click="stop">{{ t('playground.stop') }}</button>
      <button v-else type="button" class="playground-run" @click="run">{{ t('playground.run') }}</button>
      <span role="status">{{ t(`playground.${phase === 'run' ? 'running' : phase}`) }}<template v-if="exitCode !== null"> · {{ t('playground.exit') }} {{ exitCode }}</template></span>
      <RouterLink v-if="compact" to="/playground">{{ t('playground.open') }}</RouterLink>
    </div>
    <div v-if="output || failure || phase === 'done'" class="playground-result" :aria-label="t('playground.output')">
      <pre v-if="output">{{ output }}</pre>
      <pre v-if="failure" class="playground-error" role="alert">{{ failure }}</pre>
      <p v-if="!output && !failure">{{ t('playground.noOutput') }}</p>
    </div>
    <p class="playground-limit">{{ t('playground.limits') }}</p>
  </section>
</template>

<style scoped>
.wave-playground { min-width: 0; color: var(--wave-text); }
.playground-file, .playground-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 10px; }
.playground-file { justify-content: space-between; padding-bottom: 8px; font: 12px var(--wave-font-code); }
.playground-file span, .playground-actions span { color: var(--wave-text-muted); font-size: 12px; }
.wave-playground textarea { display: block; box-sizing: border-box; width: 100%; border: 1px solid var(--wave-border); border-radius: var(--wave-radius-sm); padding: 12px; resize: vertical; font: 13px/1.6 var(--wave-font-code); tab-size: 4; }
.playground-source { background: #1e1e1e; color: #e8e8e8; min-height: 140px; }
.wave-playground textarea:focus-visible, .wave-playground button:focus-visible { outline: 2px solid var(--wave-accent); outline-offset: 2px; }
.playground-input { margin: 10px 0; font-size: 12px; }
.playground-input summary { cursor: pointer; }
.playground-input label { display: block; margin: 8px 0; color: var(--wave-text-secondary); }
.playground-input textarea { background: var(--wave-surface); color: var(--wave-text); }
.playground-actions { margin: 12px 0; }
.playground-actions button { padding: 6px 16px; border: 1px solid var(--wave-border-strong); border-radius: var(--wave-radius-sm); background: var(--wave-surface); color: var(--wave-text); cursor: pointer; font-weight: 600; }
.playground-actions .playground-run { color: var(--wave-accent); }
.playground-actions a { margin-left: auto; font-size: 12px; color: var(--wave-accent); }
.playground-result { max-height: 280px; overflow: auto; border-top: 1px solid var(--wave-border); padding-top: 10px; }
.wave-playground .playground-result pre { margin: 0; padding: 8px 0; background: transparent; color: var(--wave-text); white-space: pre-wrap; overflow-wrap: anywhere; font: 13px/1.6 var(--wave-font-code); }
.wave-playground .playground-result .playground-error { color: var(--wave-danger, #c93838); }
.wave-playground .playground-limit { margin: 8px 0 0; color: var(--wave-text-muted); font-size: 11px; line-height: 1.5; }
</style>
