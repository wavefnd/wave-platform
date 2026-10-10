import { createHost, GuestExit } from './host'

self.onmessage = async (event: MessageEvent<{ wasm: ArrayBuffer; stdin: string }>) => {
  let output = ''
  try {
    let instance: WebAssembly.Instance
    const host = createHost(() => instance.exports.memory as WebAssembly.Memory, event.data.stdin, text => { output += text })
    const module = await WebAssembly.compile(event.data.wasm)
    const imports = host.imports as Record<string, Record<string, unknown>>
    const missing = WebAssembly.Module.imports(module).filter(item =>
      !Object.hasOwn(imports, item.module) || !Object.hasOwn(imports[item.module], item.name))
    if (missing.length) throw new Error(`This browser playground does not provide: ${missing.map(item => `${item.module}.${item.name}`).join(', ')}`)
    instance = await WebAssembly.instantiate(module, host.imports)
    let status = 0
    try { (instance.exports._start as () => void)() }
    catch (error) { if (error instanceof GuestExit) status = error.status; else throw error }
    host.finish()
    self.postMessage({ output, status })
  } catch (error) {
    self.postMessage({ output, error: error instanceof Error ? error.message : String(error) })
  }
}
