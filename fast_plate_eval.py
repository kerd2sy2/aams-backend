import cv2
import easyocr
import time
import os
import glob
import re

print("Loading EasyOCR Engine...")
reader = easyocr.Reader(['en', 'ar'], gpu=False, verbose=False)
print("EasyOCR Engine is ready!")

def clean_arabic(txt):
    return re.sub(r'[^\u0600-\u06FF\s]', '', txt).strip()

def clean_digits(txt):
    nums = re.findall(r'\d+', txt)
    return ''.join(nums)

def clean_latin(txt):
    return re.sub(r'[^A-Za-z\s]', '', txt).strip().upper()

samples = sorted(glob.glob('/tmp/new_plates/*.jpg'))[:10]

for path in samples:
    t0 = time.time()
    img = cv2.imread(path)
    if img is None:
        continue
    h, w = img.shape[:2]
    # Resize for fast neural OCR (width 800)
    target_w = 800
    scale = target_w / float(w)
    target_h = int(h * scale)
    small = cv2.resize(img, (target_w, target_h), interpolation=cv2.INTER_AREA)
    
    # Read text with EasyOCR directly
    results = reader.readtext(small)
    
    all_texts = [r[1] for r in results if r[2] > 0.15]
    raw_str = ' '.join(all_texts)
    
    digits = clean_digits(raw_str)
    latin = clean_latin(raw_str)
    arabic = clean_arabic(raw_str)
    
    dt = time.time() - t0
    print(f"[{os.path.basename(path)}] ({dt:.2f}s) -> Raw: \"{raw_str}\" | Digits: \"{digits}\" | Latin: \"{latin}\" | Arabic: \"{arabic}\"")
