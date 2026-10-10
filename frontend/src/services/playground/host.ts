// SPDX-License-Identifier: MPL-2.0
// fixedDouble is adapted from wavefnd/Wave v0.2.1-pre-beta src/runtime/wasm_host.mjs.
export function fixedDouble(value: number) {
  const bits = new DataView(new ArrayBuffer(8));
  bits.setFloat64(0, value, true);
  const raw = bits.getBigUint64(0, true);
  const sign = raw >> 63n ? "-" : "";
  const exponent = Number((raw >> 52n) & 2047n);
  const fraction = raw & ((1n << 52n) - 1n);
  if (exponent === 2047) return sign + (fraction ? "nan" : "inf");
  let scaled = (exponent ? fraction | (1n << 52n) : fraction) * 1000000n;
  const shift = exponent ? exponent - 1023 - 52 : -1074;
  if (shift >= 0) {
    scaled <<= BigInt(shift);
  } else {
    const divisor = 1n << BigInt(-shift);
    const quotient = scaled / divisor;
    const twiceRemainder = (scaled % divisor) * 2n;
    scaled = quotient + BigInt(twiceRemainder > divisor ||
      (twiceRemainder === divisor && (quotient & 1n) !== 0n));
  }
  return sign + (scaled / 1000000n) + "." +
    (scaled % 1000000n).toString().padStart(6, "0");
}

export class GuestExit {
  status: number
  constructor(status: number) { this.status = status }
}

// Only explicit byte I/O and clocks are exposed. No browser/network/file APIs
// or platform credentials are passed to guest code.
export function createHost(getMemory: () => WebAssembly.Memory, input: string, emit: (text: string) => void) {
  const decoder = new TextDecoder()
  const encoder = new TextEncoder()
  const stdin = encoder.encode(input)
  let inputOffset = 0
  let outputSize = 0
  const memory = () => new Uint8Array(getMemory().buffer)
  const view = () => new DataView(getMemory().buffer)
  function bytes(at: number, size: number) {
    at >>>= 0; size >>>= 0
    if (at + size > memory().length) throw new Error('Memory access is out of bounds.')
    return memory().subarray(at, at + size)
  }
  function write(value: Uint8Array) {
    outputSize += value.length
    if (outputSize > 65536) throw new Error('Output limit exceeded (64 KiB).')
    emit(decoder.decode(value, { stream: true }))
  }
  function cstring(at: number) {
    at >>>= 0
    const data = memory()
    if (at >= data.length) throw new Error('String is outside memory.')
    const end = data.subarray(at, Math.min(data.length, at + 65537)).indexOf(0)
    if (end < 0) throw new Error('String is unterminated or exceeds the output limit.')
    return data.subarray(at, at + end)
  }
  const output = {
    puts(at: number) { const value = cstring(at); write(value); write(encoder.encode('\n')); return value.length + 1 },
    printf(format: number, argv: number) {
      const fmt = cstring(format)
      let cursor = argv >>> 0
      let total = 0
      const chunk = (data: Uint8Array) => { write(data); total += data.length }
      const string = (value: string) => chunk(encoder.encode(value))
      function arg(size: number, floating = false): number | bigint {
        cursor = Math.ceil(cursor / size) * size
        bytes(cursor, size)
        const value = floating ? view().getFloat64(cursor, true) : size === 8 ? view().getBigUint64(cursor, true) : view().getUint32(cursor, true)
        cursor += size
        return value
      }
      let start = 0
      for (let i = 0; i < fmt.length; i++) {
        if (fmt[i] !== 37) continue
        chunk(fmt.subarray(start, i))
        switch (fmt[++i]) {
          case 37: string('%'); break
          case 115: chunk(cstring(Number(arg(4)))); break
          case 99: chunk(new Uint8Array([Number(arg(4)) & 255])); break
          case 102: string(fixedDouble(Number(arg(8, true)))); break
          case 112: { const pointer = Number(arg(4)); string(pointer === 0 ? '(nil)' : `0x${pointer.toString(16)}`); break }
          default: throw new Error('Unsupported output format.')
        }
        start = i + 1
      }
      chunk(fmt.subarray(start))
      return total
    },
  }
  const wasi = {
    ...output,
    args_sizes_get(argc: number, size: number) { view().setUint32(argc, 0, true); view().setUint32(size, 0, true); return 0 },
    args_get() { return 0 },
    environ_sizes_get(count: number, size: number) { view().setUint32(count, 0, true); view().setUint32(size, 0, true); return 0 },
    environ_get() { return 0 },
    fd_write(fd: number, vectors: number, count: number, written: number) {
      if (fd !== 1 && fd !== 2) return 8
      if (count > 1024) return 28
      bytes(vectors, count * 8)
      let total = 0
      for (let i = 0; i < count; i++) {
        const at = view().getUint32(vectors + i * 8, true)
        const size = view().getUint32(vectors + i * 8 + 4, true)
        write(bytes(at, size)); total += size
      }
      view().setUint32(written, total, true)
      return 0
    },
    fd_read(fd: number, vectors: number, count: number, read: number) {
      if (fd !== 0) return 8
      if (count > 1024) return 28
      bytes(vectors, count * 8)
      let total = 0
      for (let i = 0; i < count; i++) {
        const at = view().getUint32(vectors + i * 8, true)
        const size = view().getUint32(vectors + i * 8 + 4, true)
        const destination = bytes(at, size)
        const consumed = Math.min(size, stdin.length - inputOffset)
        destination.set(stdin.subarray(inputOffset, inputOffset + consumed))
        inputOffset += consumed; total += consumed
      }
      view().setUint32(read, total, true)
      return 0
    },
    fd_close(fd: number) { return fd <= 2 ? 0 : 8 },
    fd_seek() { return 70 },
    fd_fdstat_get(fd: number, at: number) {
      if (fd > 2) return 8
      bytes(at, 24).fill(0); view().setUint8(at, 2)
      view().setBigUint64(at + 8, fd === 0 ? 2n : 64n, true)
      return 0
    },
    fd_prestat_get() { return 8 },
    fd_prestat_dir_name() { return 8 },
    clock_time_get(id: number, _precision: bigint, at: number) {
      if (id !== 0 && id !== 1) return 28
      const ns = id === 0 ? BigInt(Date.now()) * 1000000n : BigInt(Math.floor(performance.now() * 1e6))
      view().setBigUint64(at, ns, true); return 0
    },
    random_get(at: number, size: number) {
      const target = bytes(at, size)
      for (let i = 0; i < size; i += 65536) crypto.getRandomValues(target.subarray(i, i + 65536))
      return 0
    },
    proc_exit(status: number) { throw new GuestExit(status) },
  }
  return { imports: { env: output, wasi_snapshot_preview1: wasi }, finish: () => emit(decoder.decode()) }
}
