"""Run the standalone library example, independently of platform services."""
import http.client
import os
import socket
import subprocess
import time
import unittest


class HTTPExample(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        with socket.socket() as available:
            available.bind(('127.0.0.1', 0))
            cls.port = available.getsockname()[1]
        cls.process = subprocess.Popen(
            [os.environ.get('WAVE_HTTP_EXAMPLE', '/tmp/wave-http-example')],
            env={**os.environ, 'WAVE_HTTP_PORT': str(cls.port)},
            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        cls.addClassCleanup(cls.stop)
        for _ in range(100):
            if cls.process.poll() is not None:
                raise RuntimeError('HTTP example exited before listening')
            try:
                with socket.create_connection(('127.0.0.1', cls.port), timeout=.1):
                    return
            except OSError:
                time.sleep(.02)
        raise RuntimeError('HTTP example did not listen')

    @classmethod
    def stop(cls):
        if cls.process.poll() is None:
            cls.process.terminate()
        cls.process.wait(timeout=5)

    def request(self, method, path, body=None):
        connection = http.client.HTTPConnection('127.0.0.1', self.port, timeout=5)
        try:
            connection.request(method, path, body)
            response = connection.getresponse()
            return response.status, dict(response.getheaders()), response.read()
        finally:
            connection.close()

    def test_standalone_service_and_head(self):
        status, headers, body = self.request('GET', '/')
        self.assertEqual((status, body), (200, b'Hello, HTTP!\n'))
        self.assertNotIn('X-Wave-Version', headers)
        self.assertEqual(int(headers['Content-Length']), len(body))
        head = self.request('HEAD', '/')
        self.assertEqual(head, (200, headers, b''))
        self.assertEqual(self.request('HEAD', '/missing')[::2], (404, b''))

    def test_binary_body_and_service_selected_limit(self):
        body = bytes(range(256)) * 16
        status, headers, response = self.request('POST', '/echo', body)
        self.assertEqual((status, response), (201, body))
        self.assertEqual(headers['Content-Type'], 'application/octet-stream')
        self.assertEqual(self.request('POST', '/echo', body + b'x')[0], 413)

    def test_no_content_response_has_no_body_or_content_length(self):
        status, headers, body = self.request('GET', '/empty')
        self.assertEqual((status, body), (204, b''))
        self.assertNotIn('Content-Length', headers)


if __name__ == '__main__':
    unittest.main()
