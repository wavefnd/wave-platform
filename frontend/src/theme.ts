import { computed, ref } from 'vue'
import markdownDarkUrl from 'github-markdown-css/github-markdown-dark.css?url'
import markdownLightUrl from 'github-markdown-css/github-markdown-light.css?url'

export type ThemePreference = 'system' | 'light' | 'black'
export type ResolvedTheme = 'light' | 'dark'

const storageKey = 'wave-theme'
const media = window.matchMedia('(prefers-color-scheme: dark)')
const stored = localStorage.getItem(storageKey)
const preference = ref<ThemePreference>(stored === 'light' || stored === 'black' ? stored : 'system')
const systemDark = ref(media.matches)
const resolved = computed<ResolvedTheme>(() => preference.value === 'system'
  ? (systemDark.value ? 'dark' : 'light')
  : preference.value === 'black' ? 'dark' : 'light')

function applyTheme() {
  document.documentElement.dataset.theme = resolved.value
  document.documentElement.dataset.coreuiTheme = resolved.value
  document.documentElement.style.colorScheme = resolved.value
  document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
    ?.setAttribute('content', resolved.value === 'dark' ? '#0b0c0f' : '#6654F1')
  let markdown = document.querySelector<HTMLLinkElement>('link[data-wave-markdown-theme]')
  if (!markdown) {
    markdown = document.createElement('link')
    markdown.rel = 'stylesheet'
    markdown.dataset.waveMarkdownTheme = 'true'
    document.head.append(markdown)
  }
  markdown.href = resolved.value === 'dark' ? markdownDarkUrl : markdownLightUrl
}

function setTheme(value: ThemePreference) {
  preference.value = value
  localStorage.setItem(storageKey, value)
  applyTheme()
}

media.addEventListener('change', (event) => {
  systemDark.value = event.matches
  if (preference.value === 'system') applyTheme()
})
applyTheme()

export function useTheme() {
  return { preference, resolved, setTheme }
}
