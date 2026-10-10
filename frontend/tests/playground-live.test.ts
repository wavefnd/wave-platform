// Explicit live integration suite; the unit suite never requires Docker.
import assert from 'node:assert/strict'
import { test } from 'node:test'
import { createHost, GuestExit } from '../src/services/playground/host.ts'
import { WAVE_VERSION } from '../src/services/playground/client.ts'

const base = process.env.PLAYGROUND_URL || 'http://localhost:8080'
async function run(source: string, input = '') {
  const response = await fetch(`${base}/api/playground/compile`, {
    method: 'POST', headers: { 'Content-Type': 'text/plain' }, body: source,
    signal: AbortSignal.timeout(20000),
  })
  assert.equal(response.status, 200, response.status === 200 ? '' : await response.text())
  assert.equal(response.headers.get('X-Wave-Version'), WAVE_VERSION)
  let instance: WebAssembly.Instance
  let output = ''
  let status = 0
  const host = createHost(() => instance.exports.memory as WebAssembly.Memory, input, text => { output += text })
  const module = await WebAssembly.compile(await response.arrayBuffer())
  instance = await WebAssembly.instantiate(module, host.imports)
  try { (instance.exports._start as () => void)() }
  catch (error) { if (error instanceof GuestExit) status = error.status; else throw error }
  host.finish()
  assert.throws(() => (instance.exports.memory as WebAssembly.Memory).grow(1024), RangeError)
  return { output, status }
}

test('pinned compiler output runs with the browser host and bounded memory', async () => {
  assert.deepEqual(await run('fun main() { println("Hello, Wave!"); }'), { output: 'Hello, Wave!\n', status: 0 })
  assert.deepEqual(await run('fun main() -> i32 { return 7; }'), { output: '', status: 7 })
})

test('real WASI stdin preserves Korean UTF-8 and reaches EOF', async () => {
  const source = `import("std::io::fd")::{io_read, io_write};
fun main() -> i32 {
    var buffer: array<u8, 4>;
    while (true) {
        var size: i64 = io_read(0, &buffer[0], 4);
        if (size == 0) { break; }
        if (size < 0) { return 1; }
        io_write(1, &buffer[0], size);
    }
    return 0;
}`
  assert.deepEqual(await run(source, '안녕하세요 Wave!\n'), { output: '안녕하세요 Wave!\n', status: 0 })
})

// Every embedded official example is tested with the real Wasm compiler/host,
// including alternate stdin cases. Merely passing a native build is insufficient.
test('all embedded documentation examples run with their manifest input and output', async t => {
  const { readFileSync, readdirSync } = await import('node:fs')
  const root = new URL('../../wavedoc/', import.meta.url)
  const cases = JSON.parse(readFileSync(new URL('examples.json', root), 'utf8'))
  const sources = new Map<string, string>()
  for (const path of readdirSync(new URL('ko/', root), { recursive: true })) {
    if (!String(path).endsWith('.md')) continue
    const text = readFileSync(new URL('ko/' + path, root), 'utf8')
    for (const match of text.matchAll(/<!-- wave-example: ([a-z0-9-]+) -->\s*```wave[ \t]+playground[ \t]*\n([\s\S]*?)\n```/gi)) sources.set(match[1], match[2])
  }
  assert.deepEqual([...sources.keys()].sort(), cases.filter(item => item.playground).map(item => item.id).sort())
  for (const example of cases.filter(item => item.playground)) {
    await t.test(example.id, async () => {
      for (const input of [example, ...(example.runs ?? [])]) {
        const result = await run(sources.get(example.id)!, input.stdin ?? '')
        assert.equal(result.status, input.exit ?? 0, example.id)
        if (input.stdout != null) assert.equal(result.output, input.stdout, example.id)
      }
    })
  }
})
