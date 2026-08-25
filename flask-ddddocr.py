from flask import Flask, request, jsonify
import ddddocr
from PIL import Image
import io
import base64
import hashlib
import time
import random
import numpy as np
import requests
import cv2
import os
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# 全局初始化 DdddOcr
ocr_detector = ddddocr.DdddOcr(det=True)

# 从环境变量获取 OCR 服务 URL
OCR_URL = os.getenv("UMI_OCR_URL", "http://127.0.0.1:1224/api/ocr")

def generate_unique_sha256():
    unique_str = f"{time.time()}-{random.random()}"
    sha = hashlib.sha256(unique_str.encode('utf-8')).hexdigest()
    return sha

def send_ocr_request(image_array, angle=0):
    img = image_array.copy()
    if angle != 0:
        (h, w) = img.shape[:2]
        center = (w // 2, h // 2)
        M = cv2.getRotationMatrix2D(center, angle, 1.0)
        img = cv2.warpAffine(img, M, (w, h))

    _, buffer = cv2.imencode(".png", img)
    img_base64 = base64.b64encode(buffer).decode("utf-8")

    payload = {
        "base64": img_base64,
        "options": {
            "ocr.language": "models/config_chinese.txt",
            "ocr.cls": True,
            "ocr.limit_side_len": 4320,
            "tbpu.parser": "multi_none",
            "data.format": "text"
        }
    }
    try:
        response = requests.post(OCR_URL, json=payload, timeout=15)
        if response.status_code == 200:
            return response.json()
        else:
            return {"error": response.status_code, "msg": response.text}
    except Exception as e:
        return {"error": str(e)}

def get_word_det(image_array):
    _, buffer = cv2.imencode(".png", image_array)
    bboxes = ocr_detector.detection(buffer.tobytes())
    return bboxes

def crop_square_from_2point(image_array, coord1, coord2):
    x1, y1 = coord1
    x2, y2 = coord2
    h, w = image_array.shape[:2]

    left = max(min(x1, x2) - 10, 0)
    top = max(min(y1, y2) - 10, 0)
    right = min(max(x1, x2) + 10, w)
    bottom = min(max(y1, y2) + 10, h)

    cropped = image_array[top:bottom, left:right]
    return cropped

@app.route("/ocr_match", methods=["POST"])
def ocr_match():
    data = request.json
    base64_str = data.get("base64_str")
    word_list = data.get("word_list", [])

    if not base64_str or not word_list:
        return jsonify({"error": "Missing base64_str or word_list"}), 400

    # 解码 Base64 图片
    image_data = base64.b64decode(base64_str)
    image_array = np.array(Image.open(io.BytesIO(image_data)).convert("RGB"))
    image_array = cv2.cvtColor(image_array, cv2.COLOR_RGB2BGR)

    det_list = get_word_det(image_array)
    results = {}

    for det in det_list:
        x1, y1, x2, y2 = det
        if abs(x1 - x2) <= 30 or abs(y1 - y2) <= 30:
            continue

        cropped_img = crop_square_from_2point(image_array, (x1, y1), (x2, y2))
        results[generate_unique_sha256()] = {"det": det, "words": []}

        for i in range(-45, 45, 2):
            ocr_result = send_ocr_request(cropped_img, i)
            if "data" in ocr_result and ocr_result["data"]:
                w = ocr_result["data"][0]
                if w != 'N':
                    results[list(results.keys())[-1]]["words"].append(w)
                    if w in word_list:
                        break

    ans = []
    for word in word_list:
        for key in results:
            if word in results[key]["words"]:
                x1, y1, x2, y2 = det = results[key]["det"]
                ans.append({"x": (x1 + x2) / 2, "y": (y1 + y2) / 2})
                break

    return jsonify({"result": ans})

if __name__ == "__main__":
    print(f"Initializing DdddOcr... OCR URL: {OCR_URL}")
    app.run(host='0.0.0.0', port=5000, debug=True)
