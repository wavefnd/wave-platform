import Prism from 'prismjs'
import 'prismjs/components/prism-bash'
import 'prismjs/components/prism-c'
import 'prismjs/components/prism-cpp'
import 'prismjs/components/prism-docker'
import 'prismjs/components/prism-git'
import 'prismjs/components/prism-json'
import 'prismjs/components/prism-markdown'
import 'prismjs/components/prism-makefile'
import 'prismjs/components/prism-markup'
import 'prismjs/components/prism-rust'
import 'prismjs/components/prism-toml'
import 'prismjs/components/prism-typescript'
import 'prismjs/components/prism-yaml'

Prism.languages.wave = {
  comment: [/\/\*[\s\S]*?\*\//, /\/\/.*$/m],
  string: { pattern: /"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/, greedy: true },
  keyword: /\b(?:asm|as|break|class|clobber|const|continue|deref|else|enum|export|extern|for|fun|if|import|in|input|is|let|match|module|mut|null|out|print|println|proto|return|static|struct|true|false|type|var|while|xnand)\b/,
  builtin: /\b(?:array|bool|byte|char|f32|f64|i(?:8|16|32|64|128|256|512|1024)|isz|ptr|str|u(?:8|16|32|64|128|256|512|1024)|usz)\b/,
  number: /\b(?:0[xX][\da-fA-F]+|0[bB][01]+|\d+(?:\.\d+)?)\b/,
  operator: /->|=>|==|!=|<=|>=|&&|\|\||<<|>>|[-+*/%=&|^!<>~]/,
  punctuation: /[{}[\];(),.:]/,
}

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
  xml: 'markup', yaml: 'yaml', yml: 'yaml',
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
