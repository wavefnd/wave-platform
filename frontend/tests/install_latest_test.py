"""Exercise the actual Bash installer with local release archives, never the network."""
import hashlib
import io
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import unittest

INSTALLER = Path(__file__).resolve().parents[1] / 'public/install.sh'

class InstallationTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix='wave install test ')
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bin = self.root / 'mock-bin'; self.bin.mkdir()
        self.home = self.root / 'home'; self.home.mkdir()
        self.install = self.home / 'Wave tools/bin'
        self.std = self.home / '.wave/lib/wave/std'
        self.config = self.root / 'http.json'
        self.urls = {}
        self.env = dict(os.environ, HOME=str(self.home), SHELL='/bin/bash',
                        PATH=str(self.bin)+os.pathsep+os.environ['PATH'],
                        WAVE_INSTALL_DIR=str(self.install), FIXTURE_CONFIG=str(self.config),
                        FIXTURE_OS='Linux', FIXTURE_ARCH='x86_64')
        for key in ['WAVE_VERSION', 'VEX_VERSION']: self.env.pop(key, None)
        self.script('curl', '''#!/usr/bin/env python3
import json,os,sys,shutil
from pathlib import Path
args=sys.argv[1:]; url=next(x for x in args if x.startswith('https://'))
entry=json.loads(Path(os.environ['FIXTURE_CONFIG']).read_text()).get(url)
if entry is None: sys.exit(22)
if '-o' in args: shutil.copyfile(entry['file'],args[args.index('-o')+1])
else: print(json.dumps(entry['json']))
''')
        self.script('uname', '#!/bin/sh\nif [ "$1" = -s ]; then echo "$FIXTURE_OS"; else echo "$FIXTURE_ARCH"; fi\n')
        self.script('sysctl', '#!/bin/sh\necho "${FIXTURE_ROSETTA:-0}"\n')
        self.script('mv', '''#!/bin/sh
case "$1" in
    */.wave-std.*/stage)
        if [ "${FAIL_STD_ACTIVATE:-0}" = 1 ]; then exit 1; fi;;
    */.wave-std.*/previous)
        if [ "${FAIL_STD_RESTORE:-0}" = 1 ]; then exit 1; fi;;
    */.wave-install.*/previous)
        if [ "${FAIL_COMPILER_RESTORE:-0}" = 1 ]; then exit 1; fi;;
esac
case "$1" in
    */stage) if [ "${FAIL_ACTIVATE:-0}" = 1 ]; then exit 1; fi;;
esac
exec '''+shutil.which('mv')+''' "$@"
''')
        self.script('cp', '''#!/bin/sh
case "${3:-}" in
    */.wave-std.*/stage/)
        if [ "${FAIL_STD_COPY:-0}" = 1 ]; then exit 1; fi;;
esac
exec '''+shutil.which('cp')+''' "$@"
''')

    def script(self, name, text):
        path=self.bin/name; path.write_text(text); path.chmod(0o755)

    def archive(self, repo, name, target, bad=False):
        path=self.root/name
        prefix=name.removesuffix('.tar.gz')
        is_wave=repo=='Wave'
        exe='wavec' if is_wave else 'vex'
        script='#!/bin/sh\n'
        if is_wave:
            script += f'''case "$*" in
--version) echo 'Wave 0.2.1-pre-beta';;
'print target-spec --format=json') echo '{{"triple":"{target}"}}';;
'print std-path') printf '%s\\n' "$HOME/.wave/lib/wave/std";;
run*)
    {'exit 7' if bad else ':'}
    std="$HOME/.wave/lib/wave/std"
    if [ "${{3:-}}" = --std-root ]; then std="$4"; fi
    [ -f "$std/manifest.json" ] && [ -f "$std/packaged.wave" ];;
*) exit 2;;
esac
'''
        else: script += "echo 'Vex 1.0.0'\n"
        files={exe:script, 'LICENSE':'notice'}
        if is_wave: files.update({'std/manifest.json':'{"name":"std"}', 'std/packaged.wave':'packaged std', 'llvm/bin/llvm-config':'fixture'})
        with tarfile.open(path,'w:gz') as archive:
            for file,text in files.items():
                data=text.encode(); entry=tarfile.TarInfo(prefix+'/'+file)
                entry.size=len(data); entry.mode=0o755 if file==exe else 0o644
                archive.addfile(entry,io.BytesIO(data))
        digest=hashlib.sha256(path.read_bytes()).hexdigest()
        version = 'v0.2.1-pre-beta' if is_wave else 'v1.0.0'
        self.urls[f'https://github.com/wavefnd/{repo}/releases/download/{version}/{name}']={'file':str(path)}
        return {'name':name,'state':'uploaded','digest':'sha256:'+digest}

    def releases(self, os_name='Linux', arch='x86_64', bad=False, vex=True):
        table={('Linux','x86_64'):('x86_64-unknown-linux-gnu','x86_64-linux-gnu','x86_64-unknown-linux-gnu'),
               ('Linux','aarch64'):('aarch64-unknown-linux-gnu','aarch64-linux-gnu','aarch64-unknown-linux-gnu'),
               ('Linux','riscv64'):('riscv64-unknown-linux-gnu','riscv64-linux-gnu','riscv64gc-unknown-linux-gnu'),
               ('Linux','loongarch64'):('loongarch64-unknown-linux-gnu','loongarch64-linux-gnu','loongarch64-unknown-linux-gnu'),
               ('Darwin','arm64'):('aarch64-apple-darwin',)*3,
               ('Darwin','x86_64'):('x86_64-apple-darwin',)*3,
               ('FreeBSD','amd64'):('x86_64-unknown-freebsd',)*3}
        target,wave_suffix,vex_suffix=table[os_name,arch]
        self.env.update(FIXTURE_OS=os_name,FIXTURE_ARCH=arch)
        assets={'Wave':[self.archive('Wave',f'wave-v0.2.1-pre-beta-{wave_suffix}.tar.gz',target,bad)],
                'Vex':[self.archive('Vex',f'vex-v1.0.0-{vex_suffix}.tar.gz',target)] if vex else []}
        self.urls['https://api.github.com/repos/wavefnd/Wave/releases/tags/v0.2.1-pre-beta'] = {'json': {'tag_name': 'v0.2.1-pre-beta', 'draft': False, 'assets': assets['Wave']}}
        for repo in assets:
            self.urls[f'https://api.github.com/repos/wavefnd/{repo}/releases?per_page=100&page=1']={'json':[
                {'id':2,'tag_name':'nightly','draft':False,'published_at':'2026-10-05T00:00:00Z','assets':[]},
                {'id':1,'tag_name':'v1.0.0','draft':False,'published_at':'2026-10-01T00:00:00Z','assets':assets[repo]}]}

    def run_install(self, *args, success=True):
        self.config.write_text(json.dumps(self.urls))
        result=subprocess.run(['bash',str(INSTALLER),*args],env=self.env,cwd=self.root,text=True,capture_output=True,timeout=20)
        if success: self.assertEqual(result.returncode,0,result.stdout+result.stderr)
        else: self.assertNotEqual(result.returncode,0,result.stdout+result.stderr)
        return result

    def seed(self):
        self.install.mkdir(parents=True, exist_ok=True)
        (self.install/'wavec').write_text('previous compiler')
        (self.install/'keep.txt').write_text('previous data')
        self.std.mkdir(parents=True, exist_ok=True)
        (self.std/'manifest.json').write_text('{"name":"std"}')
        (self.std/'keep.wave').write_text('previous std')

    def test_all_unix_targets(self):
        for os_name,arch in [('Linux','x86_64'),('Linux','aarch64'),('Linux','riscv64'),('Linux','loongarch64'),('Darwin','x86_64'),('Darwin','arm64'),('FreeBSD','amd64')]:
            with self.subTest(os=os_name,arch=arch):
                self.releases(os_name,arch,vex=False)
                result=self.run_install('--no-modify-path')
                self.assertIn('installing Wave only',result.stdout)
                self.assertTrue((self.install/'std/manifest.json').is_file())
                self.assertEqual((self.std/'packaged.wave').read_text(), 'packaged std')
                ordinary = subprocess.run(
                    [str(self.install/'wavec'), 'run', 'user.wave'], env=self.env,
                    text=True, capture_output=True,
                )
                self.assertEqual(ordinary.returncode, 0, ordinary.stdout+ordinary.stderr)
                selected = subprocess.run(
                    [str(self.install/'wavec'), 'print', 'std-path'], env=self.env,
                    text=True, capture_output=True,
                )
                self.assertEqual(selected.stdout.strip(), str(self.std))
                self.assertFalse((self.home/'.bashrc').exists())

    def test_vex_and_path_are_idempotent(self):
        self.releases()
        self.run_install('--with-vex')
        self.run_install('latest')
        self.assertTrue((self.install/'vex').is_file())
        self.assertEqual((self.home/'.bashrc').read_text().count('# Wave'),1)
        rc=subprocess.run(['bash','-c','source "$HOME/.bashrc"; command -v wavec'],env=self.env,capture_output=True,text=True)
        self.assertEqual(rc.stdout.strip(),str(self.install/'wavec'))

    def test_shell_path_is_quoted_literally(self):
        self.install=self.home / "Wave's $(touch should-not-exist) tools/bin"
        self.env['WAVE_INSTALL_DIR']=str(self.install)
        self.releases(vex=False)
        self.run_install()
        result=subprocess.run(['bash','-c','source "$HOME/.bashrc"; command -v wavec'],
                              env=self.env,cwd=self.home,capture_output=True,text=True)
        self.assertEqual(result.stdout.strip(),str(self.install/'wavec'))
        self.assertFalse((self.home/'should-not-exist').exists())

    def test_rosetta_selects_native_arm64(self):
        self.releases('Darwin','arm64',vex=False)
        self.env.update(FIXTURE_ARCH='x86_64',FIXTURE_ROSETTA='1')
        self.run_install('--no-modify-path')

    def test_failures_preserve_previous_installation(self):
        self.seed()
        for failure in ['digest','download','smoke','activate','std-copy','std-activate','vex-required']:
            with self.subTest(failure=failure):
                self.releases(bad=failure=='smoke',vex=failure!='vex-required')
                wave=self.urls['https://api.github.com/repos/wavefnd/Wave/releases/tags/v0.2.1-pre-beta']['json']['assets'][0]
                if failure=='digest': wave['digest']='sha256:'+'0'*64
                if failure=='download': del self.urls['https://github.com/wavefnd/Wave/releases/download/v0.2.1-pre-beta/'+wave['name']]
                self.env['FAIL_ACTIVATE']='1' if failure=='activate' else '0'
                self.env['FAIL_STD_ACTIVATE']='1' if failure=='std-activate' else '0'
                self.env['FAIL_STD_COPY']='1' if failure=='std-copy' else '0'
                self.run_install('--with-vex','--no-modify-path',success=False)
                self.assertEqual((self.install/'wavec').read_text(),'previous compiler')
                self.assertEqual((self.install/'keep.txt').read_text(),'previous data')
                self.assertEqual((self.std/'keep.wave').read_text(), 'previous std')
                self.assertFalse((self.std/'packaged.wave').exists())
                self.assertFalse(Path(str(self.install)+'.install-lock').exists())
                self.assertFalse(Path(str(self.std)+'.install-lock').exists())
                self.assertFalse(list(self.std.parent.glob('.wave-std.*')))

    def test_existing_std_is_replaced_by_the_matching_package(self):
        self.seed()
        self.releases(vex=False)
        self.run_install('--without-vex', '--no-modify-path')
        self.assertEqual((self.std/'packaged.wave').read_text(), 'packaged std')
        self.assertFalse((self.std/'keep.wave').exists())
        self.assertFalse(Path(str(self.std)+'.install-lock').exists())
        self.assertFalse(list(self.std.parent.glob('.wave-std.*')))

    def test_std_lock_is_shared_and_not_removed_by_another_install(self):
        self.seed()
        self.releases(vex=False)
        std_lock = Path(str(self.std)+'.install-lock')
        std_lock.mkdir()
        result = self.run_install('--without-vex', '--no-modify-path', success=False)
        self.assertIn('Another std install may be running', result.stderr)
        self.assertEqual((self.install/'wavec').read_text(), 'previous compiler')
        self.assertEqual((self.std/'keep.wave').read_text(), 'previous std')
        self.assertTrue(std_lock.is_dir())
        self.assertFalse(Path(str(self.install)+'.install-lock').exists())

    def test_failed_restoration_retains_each_backup_and_tries_both_trees(self):
        for failed_tree in ['std', 'compiler']:
            with self.subTest(tree=failed_tree):
                self.seed()
                self.releases(bad=True, vex=False)
                self.env['FAIL_STD_RESTORE'] = '1' if failed_tree == 'std' else '0'
                self.env['FAIL_COMPILER_RESTORE'] = '1' if failed_tree == 'compiler' else '0'
                result = self.run_install('--without-vex', '--no-modify-path', success=False)
                self.assertIn('Previous installation retained', result.stderr)
                if failed_tree == 'std':
                    backups = list(self.std.parent.glob('.wave-std.*/previous/keep.wave'))
                    self.assertTrue(backups)
                    self.assertEqual(backups[-1].read_text(), 'previous std')
                    self.assertEqual((self.install/'wavec').read_text(), 'previous compiler')
                else:
                    backups = list(self.install.parent.glob('.wave-install.*/previous/wavec'))
                    self.assertTrue(backups)
                    self.assertEqual(backups[-1].read_text(), 'previous compiler')
                    self.assertEqual((self.std/'keep.wave').read_text(), 'previous std')
                self.assertFalse(Path(str(self.install)+'.install-lock').exists())
                self.assertFalse(Path(str(self.std)+'.install-lock').exists())

    def test_overlapping_and_symlinked_std_destinations_are_rejected(self):
        self.seed()
        self.releases(vex=False)
        # Canonicalize a parent alias before comparing compiler/std destinations.
        alias = self.root/'std parent alias'
        alias.symlink_to(self.std.parent, target_is_directory=True)
        self.env['WAVE_INSTALL_DIR'] = str(alias/'std/toolchain')
        result = self.run_install('--without-vex', '--no-modify-path', success=False)
        self.assertIn('must not overlap', result.stderr)
        self.assertEqual((self.std/'keep.wave').read_text(), 'previous std')
        self.assertEqual((self.install/'wavec').read_text(), 'previous compiler')

        self.env['WAVE_INSTALL_DIR'] = str(self.home/'fresh/../.wave/lib/wave/std/compiler')
        result = self.run_install('--without-vex', '--no-modify-path', success=False)
        self.assertIn('must not overlap', result.stderr)
        self.assertEqual((self.std/'keep.wave').read_text(), 'previous std')

        self.env['WAVE_INSTALL_DIR'] = str(self.install)
        target = self.root/'linked std'
        self.std.rename(target)
        self.std.symlink_to(target, target_is_directory=True)
        result = self.run_install('--without-vex', '--no-modify-path', success=False)
        self.assertIn('must not be a symbolic link', result.stderr)
        self.assertEqual((target/'keep.wave').read_text(), 'previous std')
        self.assertTrue(self.std.is_symlink())

    def test_relative_or_missing_home_is_rejected_before_installation(self):
        self.releases(vex=False)
        for home_value in ['relative-home', None]:
            with self.subTest(home=home_value):
                if home_value is None:
                    self.env.pop('HOME', None)
                else:
                    self.env['HOME'] = home_value
                result = self.run_install('--without-vex', '--no-modify-path', success=False)
                self.assertIn('HOME must be an absolute path', result.stderr)
                self.assertFalse(self.install.exists())
                self.assertFalse((self.root/'relative-home').exists())

    def test_macos_rejects_case_variant_std_overlap(self):
        self.seed()
        self.releases('Darwin', 'arm64', vex=False)
        self.env['WAVE_INSTALL_DIR'] = str(self.std.parent/'STD/toolchain')
        result = self.run_install('--without-vex', '--no-modify-path', success=False)
        self.assertIn('must not overlap', result.stderr)
        self.assertEqual((self.std/'keep.wave').read_text(), 'previous std')
        self.assertEqual((self.install/'wavec').read_text(), 'previous compiler')

    def test_unmanaged_std_directory_is_preserved(self):
        self.std.mkdir(parents=True)
        (self.std/'keep.txt').write_text('unrelated data')
        self.releases(vex=False)
        result = self.run_install('--without-vex', '--no-modify-path', success=False)
        self.assertIn('not managed by Wave std', result.stderr)
        self.assertEqual((self.std/'keep.txt').read_text(), 'unrelated data')
        self.assertFalse(self.install.exists())

    def test_no_fallback_when_pinned_release_lacks_target(self):
        self.releases()
        self.urls['https://api.github.com/repos/wavefnd/Wave/releases/tags/v0.2.1-pre-beta']['json']['assets'] = []
        releases=self.urls['https://api.github.com/repos/wavefnd/Wave/releases?per_page=100&page=1']['json']
        releases.insert(0,{'id':4,'tag_name':'v2.0.0','draft':False,'published_at':'2026-10-04T00:00:00Z','assets':[]})
        result=self.run_install(success=False)
        self.assertIn('No older version',result.stderr)
        self.assertFalse(self.install.exists())

    def test_without_vex_does_not_query_vex(self):
        self.releases()
        del self.urls['https://api.github.com/repos/wavefnd/Vex/releases?per_page=100&page=1']
        self.run_install('--without-vex','--no-modify-path')
        self.assertFalse((self.install/'vex').exists())

    def test_invalid_arguments_and_lock_do_not_replace_files(self):
        self.seed(); self.releases()
        for args in [('--version','1.0.0'),('--vex-version=1.0.0',),('nightly',),('--with-vex','--without-vex')]:
            self.run_install(*args,success=False)
        Path(str(self.install)+'.install-lock').mkdir()
        self.run_install(success=False)
        self.assertEqual((self.install/'wavec').read_text(),'previous compiler')

if __name__=='__main__': unittest.main()
