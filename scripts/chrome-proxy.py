#!/usr/bin/env python3
"""TCP proxy: forwards 0.0.0.0:9223 -> 127.0.0.1:9222.

Rewrites Host/Origin headers so Chrome accepts connections from
non-localhost origins (Docker containers).

Chrome ignores --remote-debugging-address=0.0.0.0 on newer versions.
This proxy bridges the gap.

Usage: python3 scripts/chrome-proxy.py

Start Chrome first:
  google-chrome --remote-debugging-port=9222 --user-data-dir=/tmp/chrome-debug
"""

import re
import socket
import threading
import sys

LOCAL_PORT = 9223
CHROME_HOST = "127.0.0.1"
CHROME_PORT = 9222


def rewrite_first_packet(data):
    """Rewrite Host and Origin headers in the HTTP/WebSocket handshake."""
    try:
        text = data.decode("latin-1")
        text = re.sub(r"Host: [^\r\n]+", f"Host: localhost:{CHROME_PORT}", text)
        text = re.sub(r"Origin: [^\r\n]+", f"Origin: http://localhost:{CHROME_PORT}", text)
        return text.encode("latin-1")
    except Exception:
        return data


def forward(src, dst, rewrite_first=False):
    first = True
    try:
        while True:
            data = src.recv(65536)
            if not data:
                break
            if rewrite_first and first:
                data = rewrite_first_packet(data)
                first = False
            dst.sendall(data)
    except (OSError, BrokenPipeError):
        pass
    finally:
        try:
            src.shutdown(socket.SHUT_RD)
        except OSError:
            pass
        try:
            dst.shutdown(socket.SHUT_WR)
        except OSError:
            pass


def handle_client(client_sock):
    try:
        remote = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        remote.settimeout(5)
        remote.connect((CHROME_HOST, CHROME_PORT))
        remote.settimeout(None)
    except OSError as e:
        print(f"Cannot connect to Chrome: {e}")
        client_sock.close()
        return

    t1 = threading.Thread(target=forward, args=(client_sock, remote, True), daemon=True)
    t2 = threading.Thread(target=forward, args=(remote, client_sock, False), daemon=True)
    t1.start()
    t2.start()


def main():
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind(("0.0.0.0", LOCAL_PORT))
    server.listen(32)
    print(f"Chrome proxy: 0.0.0.0:{LOCAL_PORT} -> {CHROME_HOST}:{CHROME_PORT}")

    try:
        while True:
            client, addr = server.accept()
            threading.Thread(target=handle_client, args=(client,), daemon=True).start()
    except KeyboardInterrupt:
        print("\nStopped")
        server.close()
        sys.exit(0)


if __name__ == "__main__":
    main()
