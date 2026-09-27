import warnings
warnings.filterwarnings('ignore')
import os, sys
os.environ['PYTHONWARNINGS'] = 'ignore'

import cv2
import easyocr
import re
import glob

reader = easyocr.Reader(['ar', 'en'], gpu=False, verbose=False)

AR_TO_EN_DIGITS = {
    '٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4',
    '٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9'
}

KNOWN_AR_PAIRS = [
    ('ر ع', 'RE'), ('د ا', 'AD'), ('ط ب', 'BT'), ('ع ب', 'BE'),
    ('ح ا', 'AJ'), ('ا ح', 'AH'), ('ح ل', 'HL'), ('س ل', 'SL'),
    ('ق م', 'QM'), ('ك س', 'KS'), ('ب ر', 'BR'), ('ن ر', 'RN'),
    ('د ب', 'DB'), ('ص ب', 'SB'), ('ع ل', 'EL'), ('ا ب', 'AB'),
    ('ط ر', 'TR'), ('د ر', 'DR'), ('م ر', 'MR')
]

for p in sorted(glob.glob('/tmp/new_plates/*.jpg')):
    img = cv2.imread(p)
    if img is None: continue
    h, w = img.shape[:2]
    small = cv2.resize(img, (800, int(h * 800 / w)))
    
    # Read with EasyOCR
    res = reader.readtext(small)
    texts = [r[1] for r in res]
    full_txt = " ".join(texts)
    
    # 1. Convert all Arabic digits to English
    norm_txt = full_txt
    for ar, en in AR_TO_EN_DIGITS.items():
        norm_txt = norm_txt.replace(ar, en)
        
    found_nums = re.findall(r'\d{3,4}', norm_txt)
    num_val = found_nums[0] if found_nums else ""
    
    # 2. Extract Letters
    found_letters = ""
    for ar_pair, en_pair in KNOWN_AR_PAIRS:
        ar_compact = ar_pair.replace(" ", "")
        if ar_pair in full_txt or ar_compact in full_txt.replace(" ", "") or en_pair in norm_txt.upper():
            found_letters = ar_pair
            break
            
    if not found_letters:
        ar_letters = re.findall(r'[\u0600-\u06FF]', full_txt)
        clean_ar = [c for c in ar_letters if c not in 'السعودية']
        if len(clean_ar) >= 2:
            found_letters = f"{clean_ar[0]} {clean_ar[1]}"
        elif len(clean_ar) == 1:
            found_letters = clean_ar[0]

    print(f"{os.path.basename(p)} -> Found: Num={num_val} | Letters='{found_letters}' | Raw={texts}")
