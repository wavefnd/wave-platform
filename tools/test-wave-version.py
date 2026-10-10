"""The installed and packaged compilers format their version output differently."""
import contextlib
import io
from pathlib import Path
import runpy
import subprocess
import unittest
from unittest.mock import patch

SCRIPT = Path(__file__).with_name('check-wave-version.py')
VERSION = (SCRIPT.parent.parent / 'wave-version').read_text().strip()


class VersionOutput(unittest.TestCase):
    def check(self, output):
        result = subprocess.CompletedProcess([], 0, stdout=output)
        with patch('subprocess.run', return_value=result), patch('sys.argv', [str(SCRIPT)]), contextlib.redirect_stdout(io.StringIO()):
            runpy.run_path(str(SCRIPT), run_name='__main__')

    def test_plain_and_colored_output(self):
        self.check(f'wavec {VERSION} (Linux)\n  backend: LLVM 21.1.8\n')
        self.check(f'\x1b[38;2;2;161;47mwavec\x1b[0m \x1b[38;2;2;161;47m{VERSION}\x1b[0m (Linux)\n')

    def test_other_versions_and_unrecognized_output_still_fail(self):
        for output in ['wavec nightly', '\x1b[32mwavec 0.2.0-pre-beta\x1b[0m', 'different compiler']:
            with self.subTest(output=output), self.assertRaises(SystemExit):
                self.check(output)


if __name__ == '__main__':
    unittest.main()
