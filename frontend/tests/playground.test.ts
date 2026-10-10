import assert from 'node:assert/strict'
import { test } from 'node:test'
import { createHost, fixedDouble, GuestExit } from '../src/services/playground/host.ts'
import { diagnosticText, compileWave } from '../src/services/playground/client.ts'

test('WASI uses bounded UTF-8 input/output and exposes no environment or files', () => {
  const memory = new WebAssembly.Memory({ initial: 1, maximum: 1 })
  const view = new DataView(memory.buffer)
  let output = ''
  const host = createHost(() => memory, '안녕하세요\n', text => { output += text })
  const wasi = host.imports.wasi_snapshot_preview1
  view.setUint32(0, 128, true); view.setUint32(4, 100, true)
  assert.equal(wasi.fd_read(0, 0, 1, 16), 0)
  const size = view.getUint32(16, true)
  view.setUint32(4, size, true)
  assert.equal(wasi.fd_write(1, 0, 1, 20), 0)
  host.finish()
  assert.equal(output, '안녕하세요\n')
  assert.equal(wasi.fd_read(0, 0, 1, 16), 0)
  assert.equal(view.getUint32(16, true), 0)
  assert.equal(wasi.fd_read(3, 0, 1, 16), 8)
  assert.equal(wasi.fd_write(3, 0, 1, 16), 8)
  assert.equal(wasi.fd_prestat_get(), 8)
  assert.equal(wasi.environ_sizes_get(24, 28), 0)
  assert.equal(view.getUint32(24, true), 0)
  assert.throws(() => wasi.proc_exit(7), error => error instanceof GuestExit && error.status === 7)
  assert.throws(() => wasi.fd_write(1, 65535, 1, 20), /bounds/)
})

test('stdout limits and unterminated guest strings cannot allocate unbounded output', () => {
  const memory = new WebAssembly.Memory({ initial: 2, maximum: 2 })
  const host = createHost(() => memory, '', () => {})
  const view = new DataView(memory.buffer)
  view.setUint32(0, 1024, true); view.setUint32(4, 65536, true)
  host.imports.wasi_snapshot_preview1.fd_write(1, 0, 1, 16)
  assert.throws(() => host.imports.wasi_snapshot_preview1.fd_write(1, 0, 1, 16), /Output limit/)
  new Uint8Array(memory.buffer).fill(65, 1024)
  assert.throws(() => host.imports.env.puts(1024), /unterminated/)
})

test('typed printf preserves float rounding, strings and Unicode byte boundaries', () => {
  assert.equal(fixedDouble(-0), '-0.000000')
  assert.equal(fixedDouble(1e21), '1000000000000000000000.000000')
  const memory = new WebAssembly.Memory({ initial: 1 })
  const bytes = new Uint8Array(memory.buffer)
  bytes.set(new TextEncoder().encode('%s %f\0'), 0)
  bytes.set(new TextEncoder().encode('Wave\0'), 64)
  const view = new DataView(memory.buffer)
  view.setUint32(128, 64, true); view.setFloat64(136, 1.25, true)
  let output = ''
  const host = createHost(() => memory, '', text => { output += text })
  host.imports.env.printf(0, 128)
  assert.equal(output, 'Wave 1.250000')
})

test('diagnostics retain line and column, without exposing the work directory', () => {
  assert.equal(diagnosticText('{"error":{"line":2,"column":8,"message":"bad type"}}'), '2:8: bad type')
  assert.equal(diagnosticText('/work/job/main.wave: error'), 'main.wave: error')
})

test('client rejects oversized sources before sending and checks compiler identity', async () => {
  await assert.rejects(compileWave('a'.repeat(32769), new AbortController().signal), /Source limit/)
  const previous = globalThis.fetch
  try {
    globalThis.fetch = async () => new Response('not wasm', { headers: { 'Content-Type': 'text/html' } })
    await assert.rejects(compileWave('fun main() {}', new AbortController().signal), /incompatible version/)
  } finally { globalThis.fetch = previous }
})
