export const WAVE_VERSION = '0.2.1-pre-beta'
export const MAX_SOURCE_BYTES = 32768

export function diagnosticText(text: string): string {
  return text.split('\n').filter(Boolean).map(line => {
    try {
      const parsed = JSON.parse(line)
      const error = parsed.error ?? parsed
      return `${error.line ? `${error.line}:${error.column ?? 1}: ` : ''}${error.message ?? line}`
    } catch { return line }
  }).join('\n').replaceAll('/work/job/main.wave', 'main.wave')
}

export async function compileWave(source: string, signal: AbortSignal): Promise<ArrayBuffer> {
  if (!source.trim()) throw new Error('Enter a Wave program first.')
  if (new TextEncoder().encode(source).length > MAX_SOURCE_BYTES) throw new Error('Source limit exceeded (32 KiB).')
  const response = await fetch('/api/playground/compile', {
    method: 'POST', headers: { 'Content-Type': 'text/plain' }, body: source, signal,
    credentials: 'omit',
  })
  if (!response.ok) throw new Error(diagnosticText(await response.text()) || `Compilation failed (${response.status}).`)
  if (response.headers.get('X-Wave-Version') !== WAVE_VERSION || !response.headers.get('Content-Type')?.startsWith('application/wasm')) {
    throw new Error('The playground compiler is unavailable or has an incompatible version.')
  }
  const wasm = await response.arrayBuffer()
  if (wasm.byteLength > 4194304) throw new Error('Compiled program exceeds the size limit.')
  return wasm
}
