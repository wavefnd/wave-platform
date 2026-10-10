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
        self.script('mv', '#!/bin/sh\ncase "$1" in */stage) if [ "${FAIL_ACTIVATE:-0}" = 1 ]; then exit 1; fi;; esac\nexec '+shutil.which('mv')+' "$@"\n')

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
--version) echo 'Wave 1.0.0';;
'print target-spec --format=json') echo '{{"triple":"{target}"}}';;
run*) {'exit 7' if bad else '[ "$3" = --std-root ] && [ -f "$4/manifest.json" ]'};;
*) exit 2;;
esac
'''
        else: script += "echo 'Vex 1.0.0'\n"
        files={exe:script, 'LICENSE':'notice'}
        if is_wave: files.update({'std/manifest.json':'{}','llvm/bin/llvm-config':'fixture'})
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
        result=subprocess.run(['bash',str(INSTALLER),*args],env=self.env,text=True,capture_output=True,timeout=20)
        if success: self.assertEqual(result.returncode,0,result.stdout+result.stderr)
        else: self.assertNotEqual(result.returncode,0,result.stdout+result.stderr)
        return result

    def seed(self):
        self.install.mkdir(parents=True)
        (self.install/'wavec').write_text('previous compiler')
        (self.install/'keep.txt').write_text('previous data')

    def test_all_unix_targets(self):
        for os_name,arch in [('Linux','x86_64'),('Linux','aarch64'),('Linux','riscv64'),('Linux','loongarch64'),('Darwin','x86_64'),('Darwin','arm64'),('FreeBSD','amd64')]:
            with self.subTest(os=os_name,arch=arch):
                self.releases(os_name,arch,vex=False)
                result=self.run_install('--no-modify-path')
                self.assertIn('installing Wave only',result.stdout)
                self.assertTrue((self.install/'std/manifest.json').is_file())
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
        for failure in ['digest','download','smoke','activate','vex-required']:
            with self.subTest(failure=failure):
                self.releases(bad=failure=='smoke',vex=failure!='vex-required')
                wave=self.urls['https://api.github.com/repos/wavefnd/Wave/releases/tags/v0.2.1-pre-beta']['json']['assets'][0]
                if failure=='digest': wave['digest']='sha256:'+'0'*64
                if failure=='download': del self.urls['https://github.com/wavefnd/Wave/releases/download/v0.2.1-pre-beta/'+wave['name']]
                self.env['FAIL_ACTIVATE']='1' if failure=='activate' else '0'
                self.run_install('--with-vex','--no-modify-path',success=False)
                self.assertEqual((self.install/'wavec').read_text(),'previous compiler')
                self.assertEqual((self.install/'keep.txt').read_text(),'previous data')
                self.assertFalse(Path(str(self.install)+'.install-lock').exists())

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
