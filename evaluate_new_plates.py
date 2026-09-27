import cv2
import numpy as np
import pytesseract
import glob, os, re, json

files = sorted(glob.glob('/tmp/new_plates/*.jpg'))
print(f"Total new plate photos: {len(files)}")

for f in files:
    img = cv2.imread(f)
    if img is None:
        continue
    h, w = img.shape[:2]
    # Resize
    scale = 1000.0 / max(h, w)
    resized = cv2.resize(img, (int(w * scale), int(h * scale)))
    gray = cv2.cvtColor(resized, cv2.COLOR_BGR2GRAY)
    
    # Try multiple enhancements:
    clahe = cv2.createCLAHE(clipLimit=3.0, tileGridSize=(8,8))
    cl = clahe.apply(gray)
    _, otsu = cv2.threshold(cl, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
    otsu_inv = cv2.bitwise_not(otsu)
    
    # 1. OCR digits on whole and lower half
    txt_digits = pytesseract.image_to_string(otsu, config='--psm 6 -c tessedit_char_whitelist=0123456789')
    txt_digits_inv = pytesseract.image_to_string(otsu_inv, config='--psm 6 -c tessedit_char_whitelist=0123456789')
    txt_digits_cl = pytesseract.image_to_string(cl, config='--psm 6 -c tessedit_char_whitelist=0123456789')
    
    txt_all = f"{txt_digits} {txt_digits_inv} {txt_digits_cl}"
    
    # 2. OCR letters
    txt_letters = pytesseract.image_to_string(cl, config='--psm 11 -c tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ')
    txt_ara = pytesseract.image_to_string(cl, lang='ara', config='--psm 11')
    
    nums = re.findall(r'\b\d{3,4}\b', txt_all)
    
    print(f"[{os.path.basename(f)}]")
    print(f"  Numbers detected: {nums}")
    print(f"  EN letters: {repr(txt_letters.strip()[:40])}")
    print(f"  ARA letters: {repr(txt_ara.strip()[:40])}")
    print("-" * 50)
