#!/usr/bin/env python3
"""Offline regression tests for pinned release installation and isolated builds."""
import copy
import hashlib
import io
import json
from pathlib import Path
import sys
import tarfile
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / 'tools'))
import build
import toolchain


def artifact(name, data):
    return {'name': name, 'size': len(data), 'sha256': hashlib.sha256(data).hexdigest()}


def snapshot():
    return {
        'schema_version': 2, 'compiler_version': '0.2.1-pre-beta', 'tag': 'v0.2.1-pre-beta',
        'target': 'x86_64-linux-gnu',
        'archive': artifact('wave-v0.2.1-pre-beta-x86_64-linux-gnu.tar.gz', b'archive'),
    }


class ToolingTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.lock = snapshot()
        self.lock_path = self.root / 'lock.json'
        self.lock_path.write_text(json.dumps(self.lock))

    def test_version_is_pinned(self):
        self.assertEqual(toolchain.read_lock(self.lock_path), self.lock)
        for version in ['nightly', '0.2.0-pre-beta', '0.2.2']:
            wrong = {**self.lock, 'compiler_version': version}
            self.lock_path.write_text(json.dumps(wrong))
            with self.assertRaises(ValueError):
                toolchain.read_lock(self.lock_path)

    def test_lock_rejects_paths_and_cross_generation_names(self):
        for field, value in [('name', '../escape'), ('name', 'different.tar.gz'),
                             ('sha256', 'bad'), ('size', 0)]:
            changed = copy.deepcopy(self.lock)
            changed['archive'][field] = value
            self.lock_path.write_text(json.dumps(changed))
            with self.subTest(field=field, value=value), self.assertRaises(ValueError):
                toolchain.read_lock(self.lock_path)

    def test_retained_snapshot_works_after_upstream_disappears(self):
        store = self.root / 'store'
        toolchain.retain(store, self.lock['archive'], b'archive')
        with patch.object(toolchain, 'open_url', side_effect=AssertionError('No network allowed')):
            self.assertEqual(toolchain.retain(store, self.lock['archive'], offline=True).read_bytes(), b'archive')

    def test_missing_or_corrupted_store_never_falls_forward(self):
        item = self.lock['archive']
        with patch.object(toolchain, 'open_url', side_effect=AssertionError('No network allowed')):
            with self.assertRaisesRegex(ValueError, 'not retained'):
                toolchain.retain(self.root, item, offline=True)
            path = toolchain.retain(self.root, item, b'archive')
            path.write_bytes(b'corrupt')
            with self.assertRaisesRegex(ValueError, 'mismatch'):
                toolchain.retain(self.root, item)

    def test_failed_download_leaves_no_published_or_temporary_file(self):
        for data in [b'wrong!!', b'oversized response']:
            with patch.object(toolchain, 'open_url', return_value=io.BytesIO(data)), self.assertRaises(ValueError):
                toolchain.retain(self.root, self.lock['archive'])
            self.assertEqual(list((self.root / self.lock['archive']['sha256']).iterdir()), [])

    def test_streamed_download_is_verified_and_retained(self):
        with patch.object(toolchain, 'open_url', return_value=io.BytesIO(b'archive')):
            self.assertEqual(toolchain.retain(self.root, self.lock['archive']).read_bytes(), b'archive')

    def archive(self, include_std=True, unsafe=False):
        path = self.root / 'package.tar.gz'
        with tarfile.open(path, 'w:gz') as tar:
            names = ['package/wavec', 'package/llvm/lib/placeholder']
            if include_std:
                names.append('package/std/placeholder')
            if unsafe:
                names.append('../escaped')
            for name in names:
                member = tarfile.TarInfo(name)
                member.size = 1
                tar.addfile(member, io.BytesIO(b'x'))
        return path

    def test_extract_requires_complete_package_and_preserves_existing_install(self):
        destination = self.root / 'installed'
        with self.assertRaisesRegex(ValueError, 'standard library'):
            toolchain.extract(self.archive(include_std=False), destination, self.lock)
        self.assertFalse(destination.exists())
        toolchain.extract(self.archive(), destination, self.lock)
        self.assertEqual(json.loads((destination / 'wave-translation-install.json').read_text()), self.lock)
        with self.assertRaisesRegex(ValueError, 'already exists'):
            toolchain.extract(self.archive(), destination, self.lock)

    def test_escaping_archive_path_is_rejected(self):
        with self.assertRaises(tarfile.FilterError):
            toolchain.extract(self.archive(unsafe=True), self.root / 'installed', self.lock)
        self.assertFalse((self.root / 'escaped').exists())
        self.assertFalse((self.root / 'installed').exists())

    @patch.object(build.platform, 'system', return_value='Linux')
    @patch.object(build.platform, 'machine', return_value='x86_64')
    def test_build_uses_locked_std_and_rejects_other_install(self, *_):
        sdk = self.root / 'toolchain'
        toolchain.extract(self.archive(), sdk, self.lock)
        with patch.object(build.subprocess, 'run') as run:
            build.build(sdk, self.lock_path, self.root / 'server', self.root / 'build')
            command = run.call_args.args[0]
            self.assertEqual(command[:3], [str(sdk / 'wavec'), '--std-root', str(sdk / 'std')])
            self.assertEqual(command[3:5], ['--dep', 'wave_http=' + str(build.ROOT.parent / 'http')])
            receipt = sdk / 'wave-translation-install.json'
            receipt.write_text('{}')
            with self.assertRaisesRegex(ValueError, 'differs'):
                build.build(sdk, self.lock_path, self.root / 'server', self.root / 'build')
            self.assertEqual(run.call_count, 1)


if __name__ == '__main__':
    unittest.main()
