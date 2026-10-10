"""Direct service framing checks; set PLAYGROUND_URL to the internal service."""
import http.client
import os
import socket
import time
import unittest
from urllib.parse import urlparse

URL = urlparse(os.environ['PLAYGROUND_URL'])


class Framing(unittest.TestCase):
    def exchange(self, parts):
        with socket.create_connection((URL.hostname, URL.port), timeout=15) as stream:
            for part in parts:
                stream.sendall(part)
            response = http.client.HTTPResponse(stream)
            response.begin()
            return response.status, response.read()

    def test_rejects_ambiguous_framing(self):
        for extra, expected in [
            (b'Host: second\r\n', 400),
            (b'Content-Length: 1\r\nContent-Length: 1\r\n', 400),
            (b'Transfer-Encoding: chunked\r\nContent-Length: 1\r\n', 400),
            (b'Transfer-Encoding: chunked\r\n', 501),
            (b'Content-Length: 99999999999999999999999999\r\n', 413),
            (b'Bad Header: x\r\n', 400),
        ]:
            with self.subTest(extra=extra):
                request = b'POST /api/playground/compile HTTP/1.1\r\nHost: local\r\n' + extra + b'\r\n'
                self.assertEqual(self.exchange([request])[0], expected)

    def test_fragmented_body_and_head_health(self):
        source = b'fun main() { println("fragmented"); }'
        head = b'POST /api/playground/compile HTTP/1.1\r\nHost: local\r\nContent-Type: text/plain\r\nContent-Length: ' + str(len(source)).encode() + b'\r\n\r\n'
        status, body = self.exchange([head[:24], head[24:], source[:6], source[6:]])
        self.assertEqual(status, 200)
        self.assertEqual(body[:4], b'\x00asm')
        with socket.create_connection((URL.hostname, URL.port), timeout=5) as stream:
            stream.sendall(b'HEAD /healthz HTTP/1.1\r\nHost: local\r\n\r\n')
            response = http.client.HTTPResponse(stream, method='HEAD')
            response.begin()
            self.assertEqual(response.status, 200)
            self.assertEqual(response.read(), b'')

    def test_incomplete_body_has_total_deadline(self):
        start = time.monotonic()
        status, _ = self.exchange([b'POST /api/playground/compile HTTP/1.1\r\nHost: local\r\nContent-Type: text/plain\r\nContent-Length: 10\r\n\r\nx'])
        self.assertEqual(status, 408)
        self.assertLess(time.monotonic() - start, 6)
        self.assertEqual(self.exchange([b'GET /healthz HTTP/1.1\r\nHost: local\r\n\r\n'])[0], 200)


if __name__ == '__main__':
    unittest.main()
