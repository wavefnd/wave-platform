#!/usr/bin/env python3
"""Regression checks for explicit, registered embedded documentation examples."""
import importlib.util
from pathlib import Path
import sys
import unittest

sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location('examples', Path(__file__).with_name('check-doc-examples.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)

class PlaygroundTests(unittest.TestCase):
    source = '<!-- wave-example: hello -->\n```wave playground\nfun main() {}\n```'

    def test_extracts_both_fence_forms(self):
        for source in (self.source, self.source.replace('wave playground', 'wave'), self.source.replace('wave playground', 'Wave  Playground ')):
            self.assertEqual(module.MARKER.findall(source), [('hello', 'fun main() {}')])

    def test_requires_registered_complete_program(self):
        self.assertEqual(module.validate_playgrounds(self.source, {'hello': {'playground': True}}, 'doc'), {'hello'})
        for cases in ({}, {'hello': {}}, {'hello': {'playground': True, 'files': {'file': 'data'}}}, {'hello': {'playground': True, 'mode': 'reject'}}):
            with self.assertRaises(ValueError):
                module.validate_playgrounds(self.source, cases, 'doc')
        with self.assertRaises(ValueError):
            module.validate_playgrounds('```wave playground\nfun main() {}\n```', {}, 'doc')

if __name__ == '__main__':
    unittest.main()
