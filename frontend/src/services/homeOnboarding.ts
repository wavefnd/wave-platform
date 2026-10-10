// Legacy /docs URLs preserve the reader's saved documentation language.
export const homeDocumentationLinks = [
  { path: 'getting-started/install', label: 'docs.installation', detail: 'docs.installation.detail' },
  { path: 'language/program-structure', label: 'docs.firstProgram', detail: 'docs.firstProgram.detail' },
  { path: 'getting-started/overview', label: 'docs.language', detail: 'docs.language.detail' },
] as const

// Keep this complete program aligned with the first language lesson.
export const firstWaveProgram = `fun main() {
    println("Hello, Wave!");
}`
