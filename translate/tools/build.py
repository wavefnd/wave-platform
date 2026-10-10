#!/usr/bin/env python3
"""Build the Wave service with the compiler and standard library from its lock."""
import argparse
import json
from pathlib import Path
import platform
import subprocess

from toolchain import DEFAULT_LOCK, read_lock

ROOT = Path(__file__).resolve().parents[1]


def build(toolchain, lock_path, output, build_dir):
    lock = read_lock(lock_path)
    targets = {'x86_64': 'x86_64-linux-gnu', 'aarch64': 'aarch64-linux-gnu'}
    if platform.system() != 'Linux' or targets.get(platform.machine()) != lock['target']:
        raise ValueError('Locked toolchain does not match this host')
    toolchain = toolchain.resolve()
    receipt = toolchain / 'wave-translation-install.json'
    if json.loads(receipt.read_text()) != lock:
        raise ValueError('Installed snapshot differs from the lock; install the locked snapshot first')
    if not (toolchain / 'std').is_dir():
        raise ValueError('Installed snapshot is missing its standard library')
    output = output.resolve()
    build_dir = build_dir.resolve()
    output.parent.mkdir(parents=True, exist_ok=True)
    build_dir.mkdir(parents=True, exist_ok=True)
    # An explicit std root avoids silently using an older ~/.wave installation.
    subprocess.run([
        str(toolchain / 'wavec'), '--std-root', str(toolchain / 'std'),
        'build', str(ROOT / 'src/main.wave'), '--target-dir', str(build_dir),
        '-o', str(output),
    ], check=True, cwd=ROOT)
    print(output)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--toolchain', type=Path, required=True)
    parser.add_argument('--lock', type=Path, default=DEFAULT_LOCK)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--build-dir', type=Path, required=True)
    args = parser.parse_args()
    build(args.toolchain, args.lock, args.output, args.build_dir)


if __name__ == '__main__':
    try:
        main()
    except (ValueError, KeyError, OSError, subprocess.CalledProcessError) as error:
        raise SystemExit(str(error))
