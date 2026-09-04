"""本地前端预览代理：静态提供构建产物，并把 /admin 请求转发到 admin 网关。"""

import http.client
import http.server
import os


class PreviewHandler(http.server.SimpleHTTPRequestHandler):
    """处理前端静态文件和后台 API 反向代理请求。"""

    def do_GET(self):
        self._handle_request()

    def do_POST(self):
        self._handle_request()

    def do_PUT(self):
        self._handle_request()

    def do_DELETE(self):
        self._handle_request()

    def _handle_request(self):
        if not self.path.startswith("/admin"):
            return super().do_GET()

        body = None
        content_length = int(self.headers.get("Content-Length", "0"))
        if content_length:
            body = self.rfile.read(content_length)

        headers = {
            key: value
            for key, value in self.headers.items()
            if key.lower() not in {"host", "content-length"}
        }
        if body is not None:
            headers["Content-Length"] = str(len(body))

        upstream = http.client.HTTPConnection("127.0.0.1", 8717, timeout=30)
        upstream.request(self.command, self.path, body=body, headers=headers)
        response = upstream.getresponse()
        data = response.read()

        self.send_response(response.status)
        for key, value in response.getheaders():
            if key.lower() not in {"connection", "transfer-encoding"}:
                self.send_header(key, value)
        self.end_headers()
        self.wfile.write(data)
        upstream.close()


if __name__ == "__main__":
    # 优先使用独立预览产物，避免清理正在运行的 dist 目录。
    output_dir = os.environ.get(
        "ADMIN_PREVIEW_DIR",
        os.path.join(os.path.dirname(__file__), "dist-admin-preview"),
    )
    os.chdir(output_dir)
    # 允许通过环境变量切换端口，避免占用已有预览服务。
    port = int(os.environ.get("ADMIN_PREVIEW_PORT", "4173"))
    server = http.server.ThreadingHTTPServer(("127.0.0.1", port), PreviewHandler)
    server.serve_forever()
