"""Integration checks against a running Wave service (no mock compiler)."""
import http.client
import json
import os
import unittest
from urllib.parse import urlparse

URL = urlparse(os.environ.get('PLAYGROUND_URL', 'http://localhost:8080'))


class Playground(unittest.TestCase):
    def request(self, source, path='/api/playground/compile', headers=None, method='POST'):
        connection = http.client.HTTPConnection(URL.hostname, URL.port or 80, timeout=20)
        try:
            connection.request(method, path, body=source, headers=headers or {'Content-Type': 'text/plain'})
            response = connection.getresponse()
            return response.status, dict(response.getheaders()), response.read()
        finally:
            connection.close()

    def test_compiles_real_wasm_and_clears_previous_jobs(self):
        for source in ['fun main() { println("Hello, Wave!"); }', 'fun main() -> i32 { return 7; }']:
            status, headers, body = self.request(source)
            self.assertEqual(status, 200, body)
            self.assertEqual(headers['Content-Type'], 'application/wasm')
            self.assertEqual(headers['X-Wave-Version'], '0.2.1-pre-beta')
            self.assertEqual(body[:8], b'\x00asm\x01\x00\x00\x00')
        status, _, body = self.request('fun main() { var n: i32 = "bad"; }')
        self.assertEqual(status, 422)
        error = json.loads(body)['error']
        self.assertGreater(error['line'], 0)
        self.assertGreater(error['column'], 0)
        self.assertNotEqual(body[:4], b'\x00asm')

    def test_wasi_input_and_memory_programs_compile(self):
        source = '''import("std::io::fd")::{io_read, io_write};
fun main() -> i32 { var b: array<u8, 64>; var n: i64 = io_read(0, &b[0], 64);
if (n < 0) { return 1; } io_write(1, &b[0], n); return 0; }'''
        self.assertEqual(self.request(source)[0], 200)

    def test_empty_oversized_and_wrong_content_type(self):
        self.assertEqual(self.request('')[0], 400)
        self.assertEqual(self.request('a' * 32769)[0], 413)
        self.assertEqual(self.request('{}', headers={'Content-Type': 'application/json'})[0], 400)

    def test_bad_source_does_not_poison_next_compilation(self):
        self.assertEqual(self.request(b'fun main() {\x00')[0], 422)
        self.assertEqual(self.request('fun main() {}')[0], 200)

    def test_no_shell_interpolation(self):
        status, _, _ = self.request('fun main() { println("$(touch /work/injected)"); }')
        self.assertEqual(status, 200)


if __name__ == '__main__':
    unittest.main()
