#!/usr/bin/env python3
"""Check documentation links and canonical Korean Wave examples.

Run with a compiler and its matching std. Builds and program-created files stay
in temporary directories; no installed compiler or std is selected implicitly.
"""
import argparse
import json
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
MARKER = re.compile(r'<!-- wave-example: ([a-z0-9-]+) -->\s*```wave\n(.*?)\n```', re.S)

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--compiler', type=Path)
    parser.add_argument('--std-root', type=Path)
    parser.add_argument('--links-only', action='store_true')
    parser.add_argument('--case', action='append', default=[])
    args = parser.parse_args()
    examples = {}
    paths = set()
    locales = {item['id'] for item in json.loads((ROOT / 'wavedoc/locales.json').read_text())}
    docs = list((ROOT / 'wavedoc').glob('*/*/*.md'))
    translations = []
    for file in docs:
        source = file.read_text()
        path = re.search(r'^path: (.+)$', source, re.M).group(1)
        paths.add((file.relative_to(ROOT / 'wavedoc').parts[0], path))
        locale = file.relative_to(ROOT / 'wavedoc').parts[0]
        for name, code in MARKER.findall(source):
            if locale != 'ko':
                translations.append((file, name, code))
                continue
            if name in examples:
                raise ValueError(f'duplicate example {name}')
            examples[name] = (file, code)
    for file, name, code in translations:
        if name not in examples or examples[name][1] != code:
            raise ValueError(f'{file}: translated example {name} differs from Korean source')
    for file in docs:
        for locale, path in re.findall(r'\]\(/docs/([\w-]+)(?:/([^\s)#]*))?(?:#[^\s)]*)?\)', file.read_text()):
            path = path or ''
            if locale not in locales or (path not in ('', 'stdlib', 'whale') and (locale, path) not in paths and ('en', path) not in paths):
                raise ValueError(f'{file}: broken link /docs/{locale}/{path}')
    print('Documentation links and translated examples: PASS', flush=True)
    cases = json.loads((ROOT / 'wavedoc/examples.json').read_text())
    if len({c['id'] for c in cases}) != len(cases) or set(examples) != {c['id'] for c in cases}:
        raise ValueError('example markers and manifest must correspond exactly')
    if args.links_only:
        return
    if not args.compiler or not args.std_root:
        parser.error('--compiler and --std-root are required for examples')
    unknown = set(args.case) - set(examples)
    if unknown:
        parser.error(f'unknown examples: {sorted(unknown)}')
    failures = []
    compiler, std = str(args.compiler.resolve()), str(args.std_root.resolve())
    subprocess.run(['python3', str(ROOT / 'tools/check-wave-version.py'), compiler], check=True)
    for case in cases:
        name = case['id']
        if args.case and name not in args.case:
            continue
        file, code = examples[name]
        try:
            with tempfile.TemporaryDirectory(prefix='wave-doc-') as directory:
                work = Path(directory)
                source = work / 'main.wave'
                source.write_text(code)
                for filename, content in case.get('files', {}).items():
                    target = work / filename
                    if not target.resolve().is_relative_to(work):
                        raise ValueError('fixture path escapes temporary directory')
                    target.parent.mkdir(parents=True, exist_ok=True)
                    target.write_text(content)
                mode = case.get('mode', 'run')
                command = [compiler, '--std-root', std]
                command += ['check', str(source)] if mode in ('check', 'reject') else ['build', str(source), '-o', str(work / 'program')]
                build = subprocess.run(command, cwd=work, capture_output=True, text=True, timeout=90)
                if mode == 'reject':
                    assert build.returncode != 0, 'invalid example was accepted'
                    assert case['diagnostic'] in build.stdout + build.stderr, build.stdout + build.stderr
                else:
                    assert build.returncode == 0, build.stdout + build.stderr
                    if mode == 'run':
                        for run in [case, *case.get('runs', [])]:
                            result = subprocess.run([str(work / 'program')], cwd=work, input=run.get('stdin', ''), capture_output=True, text=True, timeout=15)
                            assert result.returncode == run.get('exit', 0), f'exit={result.returncode}\n{result.stdout}{result.stderr}'
                            if run.get('stdout') is not None:
                                assert result.stdout == run['stdout'], f'expected {run["stdout"]!r}, got {result.stdout!r}'
            print(f'PASS {name}', flush=True)
        except (AssertionError, subprocess.TimeoutExpired) as error:
            failures.append(name)
            print(f'FAIL {name} ({file.relative_to(ROOT)}): {error}', flush=True)
    if failures:
        raise SystemExit(f'{len(failures)} example(s) failed: {", ".join(failures)}')

if __name__ == '__main__':
    main()
