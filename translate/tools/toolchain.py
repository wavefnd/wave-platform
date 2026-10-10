#!/usr/bin/env python3
"""Install the repository-pinned Wave release and verify its archive digest."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import tarfile
import tempfile
import urllib.request

DOWNLOAD = 'https://github.com/wavefnd/Wave/releases/download/v' + (Path(__file__).resolve().parents[2] / 'wave-version').read_text().strip() + '/'
DEFAULT_LOCK = Path(__file__).resolve().parents[1] / 'wave-toolchain.lock.json'


def open_url(url, accept='application/vnd.github+json'):
    headers = {'Accept': accept, 'User-Agent': 'Wave-Translation-bootstrap'}
    token = os.environ.get('GH_TOKEN')
    if token and url.startswith('https://api.github.com/'):
        headers['Authorization'] = 'Bearer ' + token
    return urllib.request.urlopen(urllib.request.Request(url, headers=headers), timeout=60)


def read_lock(path):
    lock = json.loads(path.read_text())
    version = (Path(__file__).resolve().parents[2] / 'wave-version').read_text().strip()
    if lock['schema_version'] != 2 or lock['compiler_version'] != version or lock['tag'] != 'v' + version:
        raise ValueError('Lock must match the repository Wave version')
    if lock['target'] not in ('x86_64-linux-gnu', 'aarch64-linux-gnu'):
        raise ValueError('Unsupported target')
    item = lock['archive']
    if item['name'] != 'wave-' + lock['tag'] + '-' + lock['target'] + '.tar.gz':
        raise ValueError('Invalid archive name')
    if not re.fullmatch('[0-9a-f]{64}', item['sha256']) or not isinstance(item['size'], int) or item['size'] <= 0:
        raise ValueError('Invalid archive digest or size')
    return lock


def verify_file(path, item):
    digest = hashlib.sha256()
    size = 0
    with path.open('rb') as source:
        while chunk := source.read(1024 * 1024):
            size += len(chunk)
            digest.update(chunk)
    if size != item['size'] or digest.hexdigest() != item['sha256']:
        raise ValueError('Artifact size or SHA-256 mismatch: ' + item['name'])


def retain(store, item, supplied=None, offline=False):
    destination = store / item['sha256'] / item['name']
    if destination.exists():
        verify_file(destination, item)
        return destination
    if offline and supplied is None:
        raise ValueError('Snapshot is not retained locally: ' + str(destination))
    destination.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(dir=destination.parent, delete=False) as temp:
        temporary = Path(temp.name)
        try:
            if supplied is not None:
                temp.write(supplied)
            else:
                with open_url(DOWNLOAD + item['name'], 'application/octet-stream') as response:
                    total = 0
                    while chunk := response.read(1024 * 1024):
                        total += len(chunk)
                        if total > item['size']:
                            raise ValueError('Download exceeds locked size')
                        temp.write(chunk)
            temp.flush()
            verify_file(temporary, item)
            temporary.replace(destination)
        finally:
            temporary.unlink(missing_ok=True)
    return destination


def extract(archive, destination, lock):
    if destination.exists():
        raise ValueError('Install destination already exists; use a new directory')
    destination.parent.mkdir(parents=True, exist_ok=True)
    staging = Path(tempfile.mkdtemp(prefix='.wave-stage-', dir=destination.parent))
    try:
        with tarfile.open(archive, 'r:gz') as tar:
            # data_filter rejects escaping paths and links and unsafe file types.
            if not hasattr(tarfile, 'data_filter'):
                raise ValueError('Python with tarfile.data_filter is required (3.12+ recommended)')
            tar.extractall(staging, filter='data')
        compilers = list(staging.glob('*/wavec'))
        if len(compilers) != 1 or not compilers[0].is_file():
            raise ValueError('Archive must contain one Wave package')
        package = compilers[0].parent
        if not (package / 'llvm').is_dir() or not (package / 'std').is_dir():
            raise ValueError('Wave package is missing its bundled LLVM or standard library')
        (package / 'wave-translation-install.json').write_text(json.dumps(lock, indent=2) + '\n')
        package.rename(destination)
    finally:
        shutil.rmtree(staging)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--lock', type=Path, default=DEFAULT_LOCK)
    parser.add_argument('--store', type=Path, required=True)
    parser.add_argument('--destination', type=Path, required=True)
    parser.add_argument('--offline', action='store_true')
    args = parser.parse_args()
    lock = read_lock(args.lock)
    archive = retain(args.store, lock['archive'], offline=args.offline)
    extract(archive, args.destination, lock)
    print(args.destination / 'wavec')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, KeyError, OSError, tarfile.TarError) as error:
        raise SystemExit(str(error))
