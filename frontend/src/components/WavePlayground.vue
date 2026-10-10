<script setup lang="ts">
import { Play, RotateCcw, SlidersHorizontal, Square, Terminal } from '@lucide/vue'
import { computed, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import { tokenizeWave } from '../services/waveSyntax'
import { useI18n } from '../i18n'
import { compileWave, WAVE_VERSION } from '../services/playground/client'

const props = defineProps<{ source: string; compact?: boolean; embedded?: boolean; stdin?: string }>()
const { t } = useI18n()
const id = useId()
const code = ref(props.source)
const highlighted = computed(() => tokenizeWave(code.value + '\n'))
const sourceInput = ref<HTMLTextAreaElement | null>(null)
const sourceHighlight = ref<HTMLPreElement | null>(null)
function syncScroll() {
  if (!sourceInput.value || !sourceHighlight.value) return
  sourceHighlight.value.style.width = `${sourceInput.value.clientWidth + 2}px`
  sourceHighlight.value.style.height = `${sourceInput.value.clientHeight + 2}px`
  sourceHighlight.value.scrollTop = sourceInput.value.scrollTop
  sourceHighlight.value.scrollLeft = sourceInput.value.scrollLeft
}
watch(code, syncScroll, { flush: 'post' })
const stdin = ref(props.stdin ?? '')
const optionsOpen = ref(!!props.stdin)
const output = ref('')
const failure = ref('')
const phase = ref<'idle' | 'compile' | 'run' | 'done' | 'stopped'>('idle')
const exitCode = ref<number | null>(null)
const sourceRows = computed(() => props.compact || props.embedded ? Math.min(15, Math.max(3, code.value.split('\n').length)) : 15)
const busy = computed(() => phase.value === 'compile' || phase.value === 'run')
let resizeObserver: ResizeObserver | null = null
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
function reset() {
  stop(); code.value = props.source; stdin.value = props.stdin ?? ''
  output.value = ''; failure.value = ''; exitCode.value = null; phase.value = 'idle'
}
watch(() => [props.source, props.stdin], () => { reset(); optionsOpen.value = !!props.stdin })
onMounted(() => {
  resizeObserver = new ResizeObserver(syncScroll)
  if (sourceInput.value) resizeObserver.observe(sourceInput.value)
  syncScroll()
})
onBeforeUnmount(() => { resizeObserver?.disconnect(); generation++; release() })
</script>

<template>
  <section class="wave-playground" :class="{ compact, embedded }" :aria-label="t('playground.title')">
    <div class="playground-file">
      <label :for="`${id}-source`">main.wave</label>
      <span v-if="!embedded && !compact" class="playground-version">Wave {{ WAVE_VERSION }}</span>
      <div class="playground-actions">
        <button type="button" :aria-label="t('playground.options')" :title="t('playground.options')" :aria-expanded="optionsOpen" :aria-controls="`${id}-options`" @click="optionsOpen = !optionsOpen"><SlidersHorizontal :size="15" aria-hidden="true" /></button>
        <button type="button" :aria-label="t('playground.reset')" :title="t('playground.reset')" @click="reset"><RotateCcw :size="15" aria-hidden="true" /></button>
        <button v-if="busy" type="button" @click="stop"><Square :size="13" aria-hidden="true" />{{ t('playground.stop') }}</button>
        <button v-else type="button" class="playground-run" @click="run"><Play :size="14" aria-hidden="true" />{{ t('playground.run') }}</button>
      </div>
    </div>
    <div class="playground-editor">
      <pre ref="sourceHighlight" class="playground-highlight" aria-hidden="true"><code><span v-for="(segment, index) in highlighted" :key="index" :class="segment.kind ? ['token', segment.kind] : undefined">{{ segment.text }}</span></code></pre>
      <textarea :id="`${id}-source`" ref="sourceInput" v-model="code" class="playground-source" :rows="sourceRows" :disabled="busy" wrap="off" spellcheck="false" autocapitalize="off" autocomplete="off" @scroll="syncScroll" @input="syncScroll" @keydown.ctrl.enter.prevent="run" @keydown.meta.enter.prevent="run" />
    </div>
    <div v-show="optionsOpen" :id="`${id}-options`" class="playground-input">
      <label :for="`${id}-stdin`">{{ t('playground.stdin') }}</label>
      <textarea :id="`${id}-stdin`" v-model="stdin" :disabled="busy" :aria-describedby="`${id}-stdin-help`" rows="3" maxlength="8192" spellcheck="false" />
      <p :id="`${id}-stdin-help`">{{ t('playground.stdinHelp') }}</p>
      <p>{{ t('playground.limits') }}</p>
    </div>
    <div class="status-hidden" role="status">
      {{ t(`playground.${phase === 'run' ? 'running' : phase}`) }}<template v-if="exitCode !== null"> · {{ t('playground.exit') }} {{ exitCode }}</template>
    </div>
    <section v-if="phase !== 'idle'" class="playground-result" :aria-labelledby="`${id}-terminal`">
      <div class="terminal-header">
        <span :id="`${id}-terminal`" class="terminal-title"><Terminal :size="14" aria-hidden="true" />{{ t('playground.terminal') }}</span>
        <span class="terminal-status" :class="{ 'terminal-failed': failure || (exitCode !== null && exitCode !== 0) }">
          <template v-if="failure">{{ t('playground.error') }}</template>
          <template v-else-if="exitCode !== null">{{ t('playground.exit') }} {{ exitCode }}</template>
          <template v-else>{{ t(`playground.${phase === 'run' ? 'running' : phase}`) }}</template>
        </span>
      </div>
      <div class="terminal-output" tabindex="0" :aria-label="t('playground.output')">
        <pre v-if="output">{{ output }}</pre>
        <pre v-if="failure" class="playground-error" role="alert">{{ failure }}</pre>
        <p v-if="!output && !failure && phase === 'done'">{{ t('playground.noOutput') }}</p>
      </div>
    </section>
    <RouterLink v-if="compact" class="playground-open" to="/playground">{{ t('playground.open') }} →</RouterLink>
  </section>
</template>

<style scoped>
.wave-playground { text-align: left; min-width: 0; color: var(--wave-text); border: 1px solid var(--wave-border); border-radius: var(--wave-radius-sm); }
.playground-file { display: flex; align-items: center; gap: 12px; padding: 4px 8px 4px 14px; border-bottom: 1px solid var(--wave-border); font: 12px var(--wave-font-code); color: var(--wave-text-muted); }
.playground-version { font-size: 11px; }
.wave-playground textarea { display: block; box-sizing: border-box; width: 100%; border: 1px solid var(--wave-border); border-radius: var(--wave-radius-sm); padding: 12px; resize: vertical; font: 13px/1.6 var(--wave-font-code); tab-size: 4; }
.playground-editor { position: relative; background: #1e1e1e; }
.wave-playground .playground-source, .wave-playground .playground-highlight { margin: 0; padding: 12px; border: 1px solid var(--wave-border); font: 13px/1.6 var(--wave-font-code); tab-size: 4; white-space: pre; overflow-wrap: normal; word-break: normal; letter-spacing: normal; }
.wave-playground .playground-highlight { position: absolute; inset: 0; box-sizing: border-box; overflow: hidden; pointer-events: none; color: #e8e8e8; background: transparent; border-radius: var(--wave-radius-sm); }
.wave-playground .playground-highlight code { font: inherit; color: inherit; background: transparent; padding: 0; font-variant-ligatures: normal; }
.wave-playground .playground-source { position: relative; background: transparent; color: transparent; caret-color: #e8e8e8; min-height: 86px; }
.playground-source::selection { background: #315b83; color: #fff; }
.wave-playground .playground-highlight .token.comment { color: var(--wave-code-comment); }
.wave-playground .playground-highlight .token.keyword { color: var(--wave-code-keyword); }
.wave-playground .playground-highlight .token.builtin, .wave-playground .playground-highlight .token.number { color: var(--wave-code-type); }
.wave-playground .playground-highlight .token.string { color: var(--wave-code-string); }
.wave-playground .playground-highlight .token.operator { color: var(--wave-code-operator); }
@media (forced-colors: active) {
  .wave-playground .playground-source { color: CanvasText; caret-color: CanvasText; }
  .playground-highlight { visibility: hidden; }
}
.wave-playground textarea:focus-visible, .wave-playground button:focus-visible { outline: 2px solid var(--wave-accent); outline-offset: 2px; }
.playground-input { padding: 12px 14px; border-top: 1px solid var(--wave-border); font-size: 12px; }
.playground-input label { display: block; margin-bottom: 8px; color: var(--wave-text-secondary); }
.wave-playground .playground-input p { margin: 8px 0 0; color: var(--wave-text-muted); font-size: 11px; line-height: 1.5; }
.playground-input textarea { background: var(--wave-surface); color: var(--wave-text); }
.playground-actions { display: flex; align-items: center; gap: 2px; margin-left: auto; }
.playground-actions button { display: inline-flex; align-items: center; justify-content: center; gap: 5px; min-width: 32px; min-height: 32px; padding: 6px 8px; border: 0; border-radius: var(--wave-radius-sm); background: transparent; color: var(--wave-text-muted); cursor: pointer; font-size: 12px; font-family: inherit; white-space: nowrap; }
.playground-actions button:hover, .playground-actions button[aria-expanded="true"] { background: var(--wave-surface); color: var(--wave-text); }
.playground-actions .playground-run { color: var(--wave-accent); }
.status-hidden { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; border: 0; }
.wave-playground .playground-open { display: block; padding: 8px 14px; border-top: 1px solid var(--wave-border); font-size: 12px; color: var(--wave-accent); }
.wave-playground .playground-source, .wave-playground .playground-highlight { border-color: transparent; border-radius: 0; }
.wave-playground .playground-source:focus-visible { outline-offset: -2px; }
.playground-result { border-top: 1px solid var(--wave-border); background: #101214; color: #d8dee5; border-radius: 0 0 var(--wave-radius-sm) var(--wave-radius-sm); }
.terminal-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 8px 14px; border-bottom: 1px solid #252a30; color: #a5adb8; font: 11px/1.5 var(--wave-font-code); }
.terminal-title { display: inline-flex; align-items: center; gap: 7px; }
.terminal-status { text-align: right; }
.terminal-output { max-height: 240px; min-height: 48px; overflow: auto; padding: 12px 14px; }
.terminal-output:focus-visible { outline: 2px solid var(--wave-accent); outline-offset: -2px; }
.wave-playground .terminal-output pre { margin: 0; padding: 0; border: 0; border-radius: 0; background: transparent; color: #d8dee5; white-space: pre-wrap; overflow-wrap: anywhere; font: 13px/1.7 var(--wave-font-code); }
.wave-playground .terminal-output .playground-error, .terminal-failed { color: #f28b82; }
.wave-playground .terminal-output p { margin: 0; color: #a5adb8; font: 12px/1.7 var(--wave-font-code); }
@media (forced-colors: active) {
  .playground-result, .wave-playground .terminal-output pre, .wave-playground .terminal-output p, .terminal-header, .terminal-failed { color: CanvasText; background: Canvas; }
  .terminal-header { border-color: CanvasText; }
}
@media (max-width: 420px) { .playground-version { display: none; } }
</style>
