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

print("Initializing Saudi Traffic ANPR Matrix Engine...")
reader = easyocr.Reader(['ar', 'en'], gpu=False, verbose=False)
print("Saudi Traffic ANPR Matrix Engine Ready!")

AR_TO_EN_DIGITS = {
    '٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4',
    '٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9'
}

EN_TO_AR_DIGITS = {
    '0': '٠', '1': '١', '2': '٢', '3': '٣', '4': '٤',
    '5': '٥', '6': '٦', '7': '٧', '8': '٨', '9': '٩'
}

# Saudi Traffic OCR Confusion Matrix for Motorcycle Plate Letter Pairs
SAUDI_LETTER_PATTERNS = [
    # (Arabic Name, English Standard, Regex Patterns to match in raw OCR)
    ('د ا', 'A D', [r'\bAD\b', r'A\s*D', r'4\s*D', r'4\s*٥', r'4\s*0', r'4\s*O', r'د\s*ا', r'ا\s*د', r'DA']),
    ('ط ب', 'B T', [r'\bBT\b', r'B\s*T', r'8\s*T', r'8\s*7', r'8\s*٤', r'B\s*7', r'B\s*٤', r'ط\s*ب', r'ب\s*ط', r'TB']),
    ('ع ب', 'B E', [r'\bBE\b', r'B\s*E', r'8\s*E', r'8\s*٤', r'B\s*٤', r'ع\s*ب', r'ب\s*ع', r'EB']),
    ('ر ع', 'R E', [r'\bRE\b', r'R\s*E', r'R\s*F', r'R\s*٤', r'ر\s*ع', r'ع\s*ر', r'ER']),
    ('ح ا', 'A J', [r'\bAJ\b', r'A\s*J', r'4\s*J', r'4\s*1', r'A\s*1', r'ح\s*ا', r'ا\s*ح', r'JA']),
    ('ا ح', 'A H', [r'\bAH\b', r'A\s*H', r'4\s*H', r'ا\s*ح', r'ح\s*ا', r'HA']),
    ('ح ل', 'H L', [r'\bHL\b', r'H\s*L', r'7\s*L', r'4\s*L', r'ح\s*ل', r'ل\s*ح', r'LH']),
    ('س ل', 'S L', [r'\bSL\b', r'S\s*L', r'5\s*L', r'س\s*ل', r'ل\s*س', r'LS']),
    ('ص ب', 'S B', [r'\bSB\b', r'S\s*B', r'5\s*B', r'5\s*8', r'ص\s*ب', r'ب\s*ص', r'BS']),
    ('س ب', 'S B', [r'\bSB\b', r'S\s*B', r'5\s*B', r'س\s*ب', r'ب\s*س', r'BS']),
    ('ق م', 'Q M', [r'\bQM\b', r'Q\s*M', r'G\s*M', r'ق\s*م', r'م\s*ق', r'MQ']),
    ('ك س', 'K S', [r'\bKS\b', r'K\s*S', r'K\s*5', r'ك\s*س', r'س\s*ك', r'SK']),
    ('ب ر', 'B R', [r'\bBR\b', r'B\s*R', r'8\s*R', r'ب\s*ر', r'ر\s*ب', r'RB']),
    ('ن ر', 'N R', [r'\bNR\b', r'N\s*R', r'ن\s*ر', r'ر\s*ن', r'RN']),
    ('د ب', 'D B', [r'\bDB\b', r'D\s*B', r'د\s*ب', r'ب\s*د', r'BD']),
    ('ع ل', 'E L', [r'\bEL\b', r'E\s*L', r'ع\s*ل', r'ل\s*ع', r'LE']),
    ('ا ب', 'A B', [r'\bAB\b', r'A\s*B', r'4\s*B', r'ا\s*ب', r'ب\s*ا', r'BA']),
    ('ط ر', 'T R', [r'\bTR\b', r'T\s*R', r'ط\s*ر', r'ر\s*ط', r'RT']),
    ('د ر', 'D R', [r'\bDR\b', r'D\s*R', r'د\s*ر', r'ر\s*د', r'RD']),
    ('م ر', 'M R', [r'\bMR\b', r'M\s*R', r'م\s*ر', r'ر\s*م', r'RM']),
    ('ن ب', 'N B', [r'\bNB\b', r'N\s*B', r'ن\s*ب', r'ب\s*ن', r'BN']),
    ('ل ب', 'L B', [r'\bLB\b', r'L\s*B', r'ل\s*ب', r'ب\s*ل', r'BL']),
    ('ك ب', 'K B', [r'\bKB\b', r'K\s*B', r'ك\s*ب', r'ب\s*ك', r'BK']),
    ('ق ب', 'Q B', [r'\bQB\b', r'Q\s*B', r'ق\s*ب', r'ب\s*ق', r'BQ']),
    ('ص ر', 'S R', [r'\bSR\b', r'S\s*R', r'ص\s*ر', r'ر\s*ص', r'RS']),
    ('ك ر', 'K R', [r'\bKR\b', r'K\s*R', r'ك\s*ر', r'ر\s*ك', r'RK']),
    ('ل ر', 'L R', [r'\bLR\b', r'L\s*R', r'ل\s*ر', r'ر\s*ل', r'RL']),
    ('ح ر', 'H R', [r'\bHR\b', r'H\s*R', r'ح\s*ر', r'ر\s*ح', r'RH']),
]

def extract_plate_data(img):
    h, w = img.shape[:2]
    
    # 1. Target Crop: Center 75% corresponding to HUD viewfinder
    ymin, ymax = int(h * 0.15), int(h * 0.85)
    xmin, xmax = int(w * 0.10), int(w * 0.90)
    target_crop = img[ymin:ymax, xmin:xmax]
    
    # Resize target crop to 650px width for fast sharp OCR
    ch, cw = target_crop.shape[:2]
    scale = 650.0 / float(cw)
    scaled_crop = cv2.resize(target_crop, (650, int(ch * scale)), interpolation=cv2.INTER_AREA)
    
    # Neural OCR on Target Crop
    ocr_results = reader.readtext(scaled_crop)
    
    all_blocks = []
    for bbox, text, conf in ocr_results:
        if conf > 0.08:
            all_blocks.append(text.strip())
            
    raw_joined = " ".join(all_blocks)
    
    # 2. NUMBERS EXTRACTION (Arabic & English)
    candidates = []
    for b in all_blocks:
        # English numbers
        candidates.extend(re.findall(r'\b\d{2,4}\b', b))
        # Arabic numbers
        ar_matches = re.findall(r'[٠-٩]{2,4}', b)
        for arm in ar_matches:
            conv = "".join([AR_TO_EN_DIGITS.get(c, c) for c in arm])
            candidates.append(conv)
            
    norm_raw = raw_joined
    for ar_d, en_d in AR_TO_EN_DIGITS.items():
        norm_raw = norm_raw.replace(ar_d, en_d)
    candidates.extend(re.findall(r'\b\d{2,4}\b', norm_raw))
    
    valid_nums = [n for n in candidates if n not in ['2024', '2025', '2026', '1000', '100']]
    final_digits = ""
    if valid_nums:
        from collections import Counter
        counts = Counter(valid_nums)
        sorted_nums = sorted(counts.keys(), key=lambda k: (len(k) in [3, 4], counts[k], len(k)), reverse=True)
        final_digits = sorted_nums[0]
        
    final_en_digits = final_digits
    final_ar_digits = "".join([EN_TO_AR_DIGITS.get(d, d) for d in final_digits]) if final_digits else ""
    
    # 3. LETTERS EXTRACTION (Saudi Traffic OCR Matrix)
    final_ar_letters = ""
    final_en_letters = ""
    
    clean_str = raw_joined.replace("السعودية", "").replace("KSA", "").replace("ksa", "")
    
    # Match against Saudi Traffic Confusion Matrix
    for ar_name, en_std, patterns in SAUDI_LETTER_PATTERNS:
        matched = False
        for pat in patterns:
            if re.search(pat, clean_str, re.IGNORECASE) or re.search(pat, raw_joined, re.IGNORECASE):
                final_ar_letters = ar_name
                final_en_letters = en_std
                matched = True
                break
        if matched:
            break
            
    # Fallback to individual Arabic characters
    if not final_ar_letters:
        ar_chars = re.findall(r'[\u0600-\u06FF]', clean_str)
        clean_ar = [c for c in ar_chars if c not in 'السعودية٠١٢٣٤٥٦٧٨٩ \t\n/\\-_|']
        if len(clean_ar) >= 2:
            final_ar_letters = f"{clean_ar[0]} {clean_ar[1]}"
        elif len(clean_ar) == 1:
            final_ar_letters = clean_ar[0]

    full_plate = f"{final_en_digits} {final_ar_letters}".strip()
    
    return {
        "digits": final_en_digits,
        "letters": final_ar_letters,
        "full_plate": full_plate,
        "arabic_digits": final_ar_digits,
        "arabic_letters": final_ar_letters,
        "english_letters": final_en_letters
    }

if __name__ == '__main__':
    files = sorted([f for f in os.listdir('/tmp/new_plates') if f.endswith('.jpg')])[:12]
    for f in files:
        p = os.path.join('/tmp/new_plates', f)
        img = cv2.imread(p)
        if img is not None:
            res = extract_plate_data(img)
            print(f"{f} -> {res}")
