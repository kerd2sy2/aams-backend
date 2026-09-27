import warnings
warnings.filterwarnings('ignore')
import os, sys
os.environ['PYTHONWARNINGS'] = 'ignore'

import cv2
import easyocr
import numpy as np
import glob
import re

reader = easyocr.Reader(['ar', 'en'], gpu=False, verbose=False)

def crop_plate_candidates(img):
    crops = [img]
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    
    # 1. CLAHE enhanced
    clahe = cv2.createCLAHE(clipLimit=3.0, tileGridSize=(8, 8))
    enhanced = clahe.apply(gray)
    crops.append(cv2.cvtColor(enhanced, cv2.COLOR_GRAY2BGR))
    
    # 2. Find rectangular plate contours
    blurred = cv2.GaussianBlur(gray, (5, 5), 0)
    thresh = cv2.adaptiveThreshold(blurred, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY_INV, 19, 9)
    contours, _ = cv2.findContours(thresh, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
    
    h, w = img.shape[:2]
    img_area = h * w
    for cnt in sorted(contours, key=cv2.contourArea, reverse=True)[:10]:
        area = cv2.contourArea(cnt)
        if area < img_area * 0.04 or area > img_area * 0.95:
            continue
        x, y, cw, ch = cv2.boundingRect(cnt)
        aspect = cw / float(ch)
        if 0.8 <= aspect <= 3.8:
            px = max(0, x - 15)
            py = max(0, y - 15)
            pw = min(w - px, cw + 30)
            ph = min(h - py, ch + 30)
            crops.append(img[py:py+ph, px:px+pw])
            break
            
    return crops

samples = sorted(glob.glob('/tmp/new_plates/*.jpg'))[-8:]
for p in samples:
    img = cv2.imread(p)
    if img is None: continue
    h, w = img.shape[:2]
    small = cv2.resize(img, (800, int(h * 800 / w)))
    
    crops = crop_plate_candidates(small)
    all_texts = []
    for c in crops:
        res = reader.readtext(c)
        for r in res:
            if r[2] > 0.10:
                all_texts.append(r[1])
                
    raw = " ".join(set(all_texts))
    print(f"{os.path.basename(p)} -> Found: {raw}")
