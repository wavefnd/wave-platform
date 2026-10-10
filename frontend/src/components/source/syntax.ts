import Prism from 'prismjs'
import 'prismjs/components/prism-bash.js'
import 'prismjs/components/prism-c.js'
import 'prismjs/components/prism-cpp.js'
import 'prismjs/components/prism-docker.js'
import 'prismjs/components/prism-git.js'
import 'prismjs/components/prism-json.js'
import 'prismjs/components/prism-markdown.js'
import 'prismjs/components/prism-makefile.js'
import 'prismjs/components/prism-markup.js'
import 'prismjs/components/prism-rust.js'
import 'prismjs/components/prism-toml.js'
import 'prismjs/components/prism-typescript.js'
import 'prismjs/components/prism-yaml.js'

// Vue owns the rendered DOM; Prism auto-highlighting would erase Wave tokens
// on initial page loads and overwrite the source viewer's line markup.
Prism.manual = true

const filenameLanguages: Record<string, string> = {
  dockerfile: 'docker',
  makefile: 'makefile',
  'cargo.lock': 'toml',
}

const extensionLanguages: Record<string, string> = {
  bash: 'bash', c: 'c', cc: 'cpp', cpp: 'cpp', cxx: 'cpp',
  gitignore: 'git', h: 'c', hpp: 'cpp', hxx: 'cpp',
  html: 'markup', htm: 'markup', js: 'javascript', json: 'json',
  md: 'markdown', rs: 'rust', sh: 'bash', toml: 'toml', ts: 'typescript',
  wave: 'wave', xml: 'markup', yaml: 'yaml', yml: 'yaml',
}

export function languageForPath(path: string) {
  const name = path.split('/').at(-1)?.toLowerCase() ?? ''
  if (filenameLanguages[name]) return filenameLanguages[name]
  const extension = name.includes('.') ? name.split('.').at(-1) ?? '' : ''
  return extensionLanguages[extension] ?? ''
}

export function normalizeFenceLanguage(value = '') {
  const normalized = value.trim().toLowerCase().split(/\s+/, 1)[0]
  const aliases: Record<string, string> = {
    shell: 'bash', sh: 'bash', zsh: 'bash',
    html: 'markup', xml: 'markup',
    js: 'javascript', ts: 'typescript',
    md: 'markdown', yml: 'yaml',
  }
  const language = aliases[normalized] ?? normalized
  return /^[a-z\d-]+$/.test(language) ? language : ''
}

export function grammarForLanguage(language: string) {
  return Prism.languages[language]
}

export { Prism }
