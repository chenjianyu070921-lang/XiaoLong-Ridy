# -*- coding: utf-8 -*-
"""
本地假高德服务器(mock_amap.py)
用于压测时不消耗真实高德配额:把 locationsvc 配置里的 MapService.BaseUrl
指向 http://127.0.0.1:18080/v3,所有高德调用都会打到本脚本。

响应格式与高德 v3 API 完全一致,对照 internal/geo/amap.go 的解析逻辑构造。
启动方式: python mock_amap.py
"""
import json
import random
import urllib.parse
from http.server import BaseHTTPRequestHandler, HTTPServer

PORT = 18080


def build_regeo_response():
    """模拟 逆地理编码 /geocode/regeo 的响应"""
    return {
        "status": "1",
        "info": "OK",
        "regeocode": {
            "formatted_address": "北京市朝阳区望京街道阜通东大街6号",
            "addressComponent": {
                "province": "北京市",
                "city": [],  # 直辖市高德返回空数组
                "district": "朝阳区",
                "township": "望京街道",
            },
            "pois": [
                {"name": "望京SOHO", "type": "商务住宅;楼宇"},
                {"name": "阜通东大街", "type": "道路附属设施;路口名"},
            ],
        },
    }


def build_poi_response():
    """模拟 周边搜索 /place/around 的响应"""
    names = ["肯德基(望京店)", "麦当劳(阜通店)", "瑞幸咖啡(望京SOHO店)", "海底捞(望京店)", "星巴克(阜通大街店)"]
    pois = []
    for i in range(random.randint(3, 5)):
        pois.append({
            "name": names[i % len(names)],
            "address": "北京市朝阳区望京街道%d号" % (i + 1),
            "location": "%.6f,%.6f" % (116.4 + random.uniform(0, 0.02), 39.9 + random.uniform(0, 0.01)),
            "type": "餐饮服务;中餐厅",
            "distance": str(random.randint(50, 900)),
        })
    return {
        "status": "1",
        "info": "OK",
        "count": str(len(pois)),
        "pois": pois,
    }


def build_route_response():
    """模拟 驾车路径规划 /direction/driving 的响应"""
    return {
        "status": "1",
        "info": "OK",
        "route": {
            "origin": "116.407400,39.904200",
            "destination": "116.318600,39.984700",
            "paths": [
                {
                    "distance": "12580",
                    "duration": "1680",
                    "strategy": "速度优先",
                    "steps": [
                        {
                            "instruction": "沿阜通东大街向北行驶1.2公里",
                            "polyline": "116.407400,39.904200;116.407500,39.910000;116.407900,39.916000",
                        },
                        {
                            "instruction": "左转进入北四环东路",
                            "polyline": "116.407900,39.916000;116.410000,39.920000;116.413000,39.924000",
                        },
                    ],
                }
            ],
        },
    }


class MockAmapHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        path = parsed.path

        if path.endswith("/geocode/regeo"):
            body = build_regeo_response()
        elif path.endswith("/place/around"):
            body = build_poi_response()
        elif path.endswith("/direction/driving"):
            body = build_route_response()
        else:
            body = {"status": "0", "info": "MOCK:未知路径 %s" % path}

        data = json.dumps(body, ensure_ascii=False).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, fmt, *args):  # 静默日志,避免压测时刷屏
        pass


if __name__ == "__main__":
    server = HTTPServer(("127.0.0.1", PORT), MockAmapHandler)
    print("Mock Amap server listening on http://127.0.0.1:%d/v3" % PORT)
    server.serve_forever()
