import warnings
warnings.filterwarnings('ignore')
import os, sys
os.environ['PYTHONWARNINGS'] = 'ignore'
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'

import cv2
import numpy as np
import re
import json
import base64
import easyocr
import pytesseract
from http.server import HTTPServer, BaseHTTPRequestHandler

print("Pre-loading EasyOCR models into memory (Warm Cache)...")
reader = easyocr.Reader(['ar', 'en'], gpu=False, verbose=False)
print("EasyOCR Warm Cache Ready!")

AR_DIGIT_TO_EN = {
    '٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4',
    '٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9'
}

EN_DIGIT_TO_AR = {
    '0': '٠', '1': '١', '2': '٢', '3': '٣', '4': '٤',
    '5': '٥', '6': '٦', '7': '٧', '8': '٨', '9': '٩'
}

LETTER_PAIRS_MAP = [
    ('ر ع', 'RE', 'ER'),
    ('د ا', 'AD', 'DA'),
    ('ط ب', 'BT', 'TB'),
    ('ع ب', 'BE', 'EB'),
    ('ح ا', 'AJ', 'JA'),
    ('ا ح', 'AH', 'HA'),
    ('ح ل', 'HL', 'LH'),
    ('س ل', 'SL', 'LS'),
    ('ق م', 'QM', 'MQ'),
    ('ك س', 'KS', 'SK'),
    ('ب ر', 'BR', 'RB'),
    ('ن ر', 'NR', 'RN'),
    ('د ب', 'DB', 'BD'),
    ('ص ب', 'SB', 'BS'),
    ('ع ل', 'EL', 'LE'),
    ('ا ب', 'AB', 'BA'),
    ('ط ر', 'TR', 'RT'),
    ('د ر', 'DR', 'RD'),
    ('م ر', 'MR', 'RM'),
    ('ح ب', 'HB', 'BH'),
    ('س ب', 'SB', 'BS'),
    ('ن ب', 'NB', 'BN'),
    ('ل ب', 'LB', 'BL'),
    ('ك ب', 'KB', 'BK'),
    ('ق ب', 'QB', 'BQ'),
    ('ص ر', 'SR', 'RS'),
    ('س ر', 'SR', 'RS'),
    ('ك ر', 'KR', 'RK'),
    ('ل ر', 'LR', 'RL'),
    ('ح ر', 'HR', 'RH'),
]

def crop_plate(img):
    h, w = img.shape[:2]
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    blur = cv2.bilateralFilter(gray, 7, 50, 50)
    thresh = cv2.adaptiveThreshold(blur, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY_INV, 19, 9)
    contours, _ = cv2.findContours(thresh, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
    img_area = h * w
    
    best_crop = None
    best_area = 0
    for cnt in sorted(contours, key=cv2.contourArea, reverse=True)[:10]:
        area = cv2.contourArea(cnt)
        if area < img_area * 0.05 or area > img_area * 0.95:
            continue
        x, y, cw, ch = cv2.boundingRect(cnt)
        aspect = cw / float(ch)
        if 1.0 <= aspect <= 3.8 and area > best_area:
            best_area = area
            px = max(0, x - 10)
            py = max(0, y - 10)
            pw = min(w - px, cw + 20)
            ph = min(h - py, ch + 20)
            best_crop = img[py:py+ph, px:px+pw]
            
    if best_crop is not None:
        return best_crop
    return img[int(h*0.1):int(h*0.9), int(w*0.05):int(w*0.95)]

def process_image(img):
    h, w = img.shape[:2]
    scale = 650.0 / max(h, w)
    resized = cv2.resize(img, (int(w * scale), int(h * scale)))
    
    plate = crop_plate(resized)
    norm_plate = cv2.resize(plate, (500, 300))
    
    # Run warm in-memory EasyOCR
    easy_res = reader.readtext(norm_plate)
    easy_texts = [r[1] for r in easy_res if r[2] > 0.12]
    easy_full = " ".join(easy_texts)
    
    # 1. Extract Digits
    all_digit_candidates = []
    for t in easy_texts:
        en_nums = re.findall(r'\b\d{2,4}\b', t)
        all_digit_candidates.extend(en_nums)
        ar_nums = re.findall(r'[٠-٩]{2,4}', t)
        for an in ar_nums:
            conv = "".join([AR_DIGIT_TO_EN.get(c, c) for c in an])
            all_digit_candidates.append(conv)
            
    for ar_c, en_c in AR_DIGIT_TO_EN.items():
        easy_full = easy_full.replace(ar_c, en_c)
    all_digit_candidates.extend(re.findall(r'\b\d{2,4}\b', easy_full))
    
    valid_cands = [n for n in all_digit_candidates if n not in ['2024', '2025', '2026', '1000', '100']]
    final_digits = ""
    if valid_cands:
        from collections import Counter
        counts = Counter(valid_cands)
        sorted_cands = sorted(counts.keys(), key=lambda k: (len(k) in [3, 4], counts[k], len(k)), reverse=True)
        final_digits = sorted_cands[0]
        
    found_en_digits = final_digits
    found_ar_digits = "".join([EN_DIGIT_TO_AR.get(d, d) for d in final_digits]) if final_digits else ""
    
    # 2. Extract Letters
    found_ar_letters = ""
    found_en_letters = ""
    clean_text = easy_full.replace("السعودية", "").replace("KSA", "").replace("ksa", "").upper()
    
    for ar_pair, en1, en2 in LETTER_PAIRS_MAP:
        ar_comp = ar_pair.replace(" ", "")
        if ar_pair in clean_text or ar_comp in clean_text.replace(" ", "") or en1 in clean_text or en2 in clean_text:
            found_ar_letters = ar_pair
            found_en_letters = f"{en1[0]} {en1[1]}"
            break
            
    if not found_ar_letters:
        ar_matches = re.findall(r'[\u0600-\u06FF]', clean_text)
        clean_chars = [c for c in ar_matches if c not in 'السعودية٠١٢٣٤٥٦٧٨٩ \t\n']
        if len(clean_chars) >= 2:
            found_ar_letters = f"{clean_chars[0]} {clean_chars[1]}"
        elif len(clean_chars) == 1:
            found_ar_letters = clean_chars[0]
            
    full_plate = f"{found_en_digits} {found_ar_letters}".strip()
    
    return {
        "digits": found_en_digits,
        "letters": found_ar_letters,
        "full_plate": full_plate,
        "arabic_digits": found_ar_digits,
        "arabic_letters": found_ar_letters,
        "english_letters": found_en_letters
    }

class OCRServerHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        try:
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            req = json.loads(post_data.decode('utf-8'))
            
            raw_b64 = req.get('image', '')
            if ',' in raw_b64:
                raw_b64 = raw_b64.split(',')[1]
                
            img_bytes = base64.b64decode(raw_b64)
            nparr = np.frombuffer(img_bytes, np.uint8)
            img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
            
            if img is None:
                res = {"digits": "", "letters": "", "full_plate": ""}
            else:
                res = process_image(img)
                
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(res, ensure_ascii=False).encode('utf-8'))
        except Exception as e:
            self.send_response(500)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"error": str(e)}).encode('utf-8'))
            
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    port = 5005
    server = HTTPServer(('127.0.0.1', port), OCRServerHandler)
    print(f"Fast Warm Plate OCR Daemon listening on 127.0.0.1:{port}...")
    server.serve_forever()
