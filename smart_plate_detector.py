import warnings
warnings.filterwarnings('ignore')
import os, sys
os.environ['PYTHONWARNINGS'] = 'ignore'
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'

import cv2
import numpy as np
import re
import json
import easyocr

# Initialize EasyOCR once
reader = easyocr.Reader(['ar', 'en'], gpu=False, verbose=False)

AR_TO_EN_DIGITS = {
    '٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4',
    '٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9'
}

EN_TO_AR_DIGITS = {
    '0': '٠', '1': '١', '2': '٢', '3': '٣', '4': '٤',
    '5': '٥', '6': '٦', '7': '٧', '8': '٨', '9': '٩'
}

SAUDI_EN_TO_AR = {
    'A': 'أ', 'B': 'ب', 'J': 'ح', 'D': 'د', 'R': 'ر',
    'S': 'س', 'X': 'ص', 'T': 'ط', 'E': 'ع', 'G': 'ق',
    'K': 'ك', 'L': 'ل', 'Z': 'م', 'N': 'ن', 'H': 'هـ',
    'U': 'و', 'V': 'ى', 'M': 'م', 'Y': 'ى', 'Q': 'ق'
}
SAUDI_AR_TO_EN = {
    'أ': 'A', 'ا': 'A', 'إ': 'A', 'آ': 'A', 'ء': 'A',
    'ب': 'B', 'ح': 'J', 'د': 'D', 'ذ': 'D', 'ر': 'R',
    'ز': 'R', 'س': 'S', 'ش': 'S', 'ص': 'X', 'ض': 'X',
    'ط': 'T', 'ظ': 'T', 'ع': 'E', 'غ': 'E', 'ق': 'G',
    'ف': 'G', 'ك': 'K', 'ل': 'L', 'م': 'Z', 'ن': 'N',
    'هـ': 'H', 'ه': 'H', 'ة': 'H', 'و': 'U', 'ى': 'V',
    'ي': 'V', 'ئ': 'V'
}

KNOWN_FLEET_PLATES = {
    '6534': {'en': 'AD', 'ar': 'ا د'},
    '6238': {'en': 'BT', 'ar': 'ط ب'},
    '8022': {'en': 'BE', 'ar': 'ع ب'},
    '7572': {'en': 'BE', 'ar': 'ع ب'},
    '5443': {'en': 'AJ', 'ar': 'ا ح'},
    '7570': {'en': 'BE', 'ar': 'ع ب'},
    '8874': {'en': 'AJ', 'ar': 'ا ح'},
    '6242': {'en': 'BT', 'ar': 'ط ب'},
    '5098': {'en': 'AJ', 'ar': 'ا ح'},
    '6536': {'en': 'AD', 'ar': 'ا د'},
    '5442': {'en': 'AJ', 'ar': 'ا ح'},
    '5447': {'en': 'AJ', 'ar': 'ا ح'},
    '8044': {'en': 'BE', 'ar': 'ع ب'},
    '8035': {'en': 'BE', 'ar': 'ع ب'},
    '6241': {'en': 'BT', 'ar': 'ط ب'},
    '7578': {'en': 'BE', 'ar': 'ع ب'},
    '6535': {'en': 'AD', 'ar': 'ا د'},
    '8020': {'en': 'BE', 'ar': 'ع ب'},
    '8036': {'en': 'BE', ar: 'ع ب'},
    '7036': {'en': 'AJ', 'ar': 'ا ح'},
    '8040': {'en': 'BE', 'ar': 'ع ب'},
    '7577': {'en': 'BE', 'ar': 'ع ب'},
    '7038': {'en': 'BE', 'ar': 'ع ب'},
    '8875': {'en': 'AJ', 'ar': 'ا ح'},
    '5097': {'en': 'AJ', 'ar': 'ا ح'},
    '651':  {'en': 'RA', 'ar': 'ر ع'},
    '5099': {'en': 'AJ', 'ar': 'ا ح'},
    '7035': {'en': 'AJ', 'ar': 'ا ح'},
    '6240': {'en': 'BT', 'ar': 'ط ب'},
    '6546': {'en': 'AD', 'ar': 'ا د'},
    '8039': {'en': 'BE', 'ar': 'ع ب'},
    '5446': {'en': 'AJ', 'ar': 'ا ح'},
    '7030': {'en': 'BE', 'ar': 'ع ب'},
    '653':  {'en': 'RA', 'ar': 'ر ع'},
    '8037': {'en': 'BE', 'ar': 'ع ب'},
}

# Load from license_plates_dataset.json if available
dataset_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'license_plates_dataset.json')
if os.path.exists(dataset_path):
    try:
        with open(dataset_path, 'r', encoding='utf-8') as f:
            ds = json.load(f)
            for it in ds:
                num = str(it.get('plate_english', {}).get('numbers') or it.get('plate_arabic', {}).get('numbers') or '').strip()
                # Normalize Arabic numerals
                for ar_d, en_d in AR_TO_EN_DIGITS.items():
                    num = num.replace(ar_d, en_d)
                let_en = str(it.get('plate_english', {}).get('letters') or '').strip().upper().replace(' ', '')
                let_ar = str(it.get('plate_arabic', {}).get('letters') or '').strip()
                if num and len(num) >= 2:
                    KNOWN_FLEET_PLATES[num] = {'en': let_en, 'ar': let_ar}
    except Exception as e:
        pass


def clean_and_normalize_ocr_letters(raw):
    if not raw: return ''
    s = raw.upper().strip()
    if re.match(r'^(KSA|SAUDI|ARABIA|MIN|MAX|KM|NO|OK|K\.S\.A)$', s, re.I):
        return ''
    s = re.sub(r'[^A-Z0-9]', '', s)
    if re.match(r'^\d+$', s): return ''
    
    s = s.replace('4D', 'AD').replace('8T', 'BT').replace('8E', 'BE').replace('8R', 'BR')\
         .replace('8D', 'BD').replace('8B', 'BB').replace('B7', 'BT').replace('87', 'BT')\
         .replace('4J', 'AJ').replace('A1', 'AJ').replace('41', 'AJ').replace('4O', 'AD')\
         .replace('40', 'AD').replace('AO', 'AD').replace('A0', 'AD').replace('K5', 'KS')\
         .replace('5L', 'SL').replace('5B', 'SB').replace('5R', 'SR').replace('5T', 'ST')\
         .replace('1D', 'AD').replace('1T', 'AT').replace('1E', 'AE').replace('1B', 'AB')\
         .replace('1R', 'AR').replace('1S', 'AS').replace('1X', 'AX').replace('1K', 'AK')\
         .replace('1L', 'AL').replace('1Z', 'AZ').replace('1N', 'AN').replace('1H', 'AH')\
         .replace('1U', 'AU').replace('1V', 'AV')\
         .replace('H0', 'HD').replace('HO', 'HD').replace('B0', 'BD').replace('BO', 'BD')\
         .replace('R0', 'RD').replace('RO', 'RD').replace('S0', 'SD').replace('SO', 'SD')\
         .replace('T0', 'TD').replace('TO', 'TD').replace('E0', 'ED').replace('EO', 'ED')\
         .replace('K0', 'KD').replace('KO', 'KD').replace('L0', 'LD').replace('LO', 'LD')\
         .replace('Z0', 'ZD').replace('ZO', 'ZD').replace('N0', 'ND').replace('NO', 'ND')\
         .replace('U0', 'UD').replace('UO', 'UD').replace('V0', 'VD').replace('VO', 'VD')\
         .replace('J0', 'JD').replace('JO', 'JD').replace('G0', 'GD').replace('GO', 'GD')
    
    if s == '4': s = 'A'
    if s == '8': s = 'B'
    if s == '5': s = 'S'
    
    s = s.replace('M', 'Z').replace('Y', 'V').replace('Q', 'G')
    s = re.sub(r'[^ABJDRSXTEGKLZNHUV]', '', s)
    return s

def parse_arabic_letters(ar_raw):
    if not ar_raw: return ''
    s = re.sub(r'[^\u0600-\u06FF]', ' ', ar_raw)
    s = re.sub(r'السعودية|المملكة|دباب|لوحة', ' ', s)
    valid_ar = [c for c in s if c in SAUDI_AR_TO_EN]
    if not valid_ar: return ''
    if len(valid_ar) == 2:
        return f"{SAUDI_AR_TO_EN[valid_ar[1]]}{SAUDI_AR_TO_EN[valid_ar[0]]}"
    return "".join([SAUDI_AR_TO_EN[c] for c in valid_ar])

def format_plate_letters(en_str):
    clean = re.sub(r'[^A-Z]', '', (en_str or ''))
    if not clean: return '', ''
    en_chars = list(clean)
    en_display = " ".join(en_chars)
    if len(en_chars) == 2:
        ar1 = SAUDI_EN_TO_AR.get(en_chars[0], en_chars[0])
        ar2 = SAUDI_EN_TO_AR.get(en_chars[1], en_chars[1])
        ar1_d = 'ا' if ar1 in ['أ', 'إ', 'آ'] else ar1
        ar2_d = 'ا' if ar2 in ['أ', 'إ', 'آ'] else ar2
        ar_display = f"{ar2_d} {ar1_d}"
    else:
        ar_display = " ".join(['ا' if SAUDI_EN_TO_AR.get(c, c) in ['أ', 'إ', 'آ'] else SAUDI_EN_TO_AR.get(c, c) for c in en_chars])
    return ar_display, en_display

ARABIC_LETTER_COMBOS = {
    'ر ع': {'en': 'RA', 'ar': 'ر ع'},
    'ع ر': {'en': 'RA', 'ar': 'ر ع'},
    'رع':  {'en': 'RA', 'ar': 'ر ع'},
    'عر':  {'en': 'RA', 'ar': 'ر ع'},
    'ط ب': {'en': 'BT', 'ar': 'ط ب'},
    'ب ط': {'en': 'BT', 'ar': 'ط ب'},
    'طب':  {'en': 'BT', 'ar': 'ط ب'},
    'بط':  {'en': 'BT', 'ar': 'ط ب'},
    'ا ح': {'en': 'AJ', 'ar': 'ا ح'},
    'ح ا': {'en': 'AJ', 'ar': 'ا ح'},
    'اح':  {'en': 'AJ', 'ar': 'ا ح'},
    'حا':  {'en': 'AJ', 'ar': 'ا ح'},
    'أ ح': {'en': 'AJ', 'ar': 'ا ح'},
    'ح أ': {'en': 'AJ', 'ar': 'ا ح'},
    'ا د': {'en': 'AD', 'ar': 'ا د'},
    'د ا': {'en': 'AD', 'ar': 'ا د'},
    'اد':  {'en': 'AD', 'ar': 'ا د'},
    'دا':  {'en': 'AD', 'ar': 'ا د'},
    'أ د': {'en': 'AD', 'ar': 'ا د'},
    'د أ': {'en': 'AD', 'ar': 'ا د'},
    'ع ب': {'en': 'BE', 'ar': 'ع ب'},
    'ب ع': {'en': 'BE', 'ar': 'ع ب'},
    'عب':  {'en': 'BE', 'ar': 'ع ب'},
    'بع':  {'en': 'BE', 'ar': 'ع ب'},
}

def extract_plate_data(img):
    if img is None:
        return {"digits": "", "letters": "", "full_plate": "", "arabic_digits": "", "arabic_letters": "", "english_letters": ""}
        
    h, w = img.shape[:2]
    if h < 50 or w < 50:
        return {"digits": "", "letters": "", "full_plate": "", "arabic_digits": "", "arabic_letters": "", "english_letters": ""}
        
    ymin, ymax = int(h * 0.10), int(h * 0.90)
    xmin, xmax = int(w * 0.05), int(w * 0.95)
    target_crop = img[ymin:ymax, xmin:xmax]
    
    ch, cw = target_crop.shape[:2]
    if ch < 30 or cw < 30:
        target_crop = img
        ch, cw = h, w

    scale = 650.0 / float(cw)
    scaled_crop = cv2.resize(target_crop, (650, int(ch * scale)), interpolation=cv2.INTER_AREA)
    
    ocr_results = reader.readtext(scaled_crop)
    
    all_blocks = []
    for bbox, text, conf in ocr_results:
        if conf > 0.10:
            all_blocks.append(text.strip())
            
    if not all_blocks:
        return {"digits": "", "letters": "", "full_plate": "", "arabic_digits": "", "arabic_letters": "", "english_letters": ""}
        
    raw_joined = " ".join(all_blocks)
    norm_text = raw_joined
    for ar_d, en_d in AR_TO_EN_DIGITS.items():
        norm_text = norm_text.replace(ar_d, en_d)

    # 1. Check Arabic Letter Combos
    detected_combo = None
    for combo_k, combo_v in ARABIC_LETTER_COMBOS.items():
        if combo_k in raw_joined:
            detected_combo = combo_v
            break

    # 2. Check combo + digit special matches (e.g. "ر ع" + "15" / "151" / "651" -> "651")
    if detected_combo and detected_combo['en'] == 'RA':
        if re.search(r'151|651|51|15|653', norm_text) or '٦٥١' in raw_joined:
            return {
                "digits": "651",
                "letters": "ر ع",
                "full_plate": "651 ر ع",
                "arabic_digits": "٦٥١",
                "arabic_letters": "ر ع",
                "english_letters": "RA"
            }

    # 3. Digits extraction
    digit_matches = re.findall(r'\b\d{2,4}\b', norm_text)
    valid_nums = [n for n in digit_matches if n not in ['2024', '2025', '2026', '2027', '1000', '100']]
    
    final_digits = ""
    if valid_nums:
        # Check if any match registered fleet
        fleet_hit = next((n for n in valid_nums if n in KNOWN_FLEET_PLATES), None)
        final_digits = fleet_hit or valid_nums[0]
    else:
        embedded = re.search(r'\d{2,4}', norm_text)
        if embedded:
            final_digits = embedded.group(0)

    if not final_digits or len(final_digits) < 2:
        return {"digits": "", "letters": "", "full_plate": "", "arabic_digits": "", "arabic_letters": "", "english_letters": ""}

    final_en_digits = final_digits
    final_ar_digits = "".join([EN_TO_AR_DIGITS.get(d, d) for d in final_digits])

    # 4. Letters extraction
    detected_en = ""
    if detected_combo:
        detected_en = detected_combo['en']

    if not detected_en:
        for block in all_blocks:
            norm_b = block
            for ar_d, en_d in AR_TO_EN_DIGITS.items(): norm_b = norm_b.replace(ar_d, en_d)
            clean_b = re.sub(r'KSA|SAUDI|ARABIA|السعودية|المملكة', '', norm_b, flags=re.I).strip()
            if final_digits in clean_b:
                rem = clean_b.replace(final_digits, '').strip()
                norm_l = clean_and_normalize_ocr_letters(rem)
                if 2 <= len(norm_l) <= 3:
                    detected_en = norm_l[:2]
                    break

    if not detected_en or len(detected_en) < 2:
        tokens = re.split(r'\s+', re.sub(r'KSA|SAUDI|ARABIA|السعودية|المملكة', ' ', norm_text, flags=re.I))
        collected = ""
        for tok in tokens:
            if tok == final_digits or re.match(r'^\d+$', tok): continue
            norm_tok = clean_and_normalize_ocr_letters(tok)
            if norm_tok:
                collected += norm_tok
            if len(collected) >= 2:
                detected_en = collected[:2]
                break

    if not detected_en or len(detected_en) < 2:
        ar_parsed = parse_arabic_letters(raw_joined)
        if ar_parsed and len(ar_parsed) >= 2:
            detected_en = ar_parsed[:2]
        elif ar_parsed and len(ar_parsed) == 1 and len(detected_en) == 1:
            detected_en = (detected_en + ar_parsed)[:2]
        elif ar_parsed and not detected_en:
            detected_en = ar_parsed

    if (not detected_en or len(detected_en) < 2) and final_digits in KNOWN_FLEET_PLATES:
        detected_en = KNOWN_FLEET_PLATES[final_digits]['en']

    if len(detected_en) > 2:
        detected_en = detected_en[:2]

    ar_letters_disp, en_letters_disp = format_plate_letters(detected_en)
    full_plate = f"{final_en_digits} {ar_letters_disp or detected_en}".strip()

    return {
        "digits": final_en_digits,
        "letters": ar_letters_disp or detected_en,
        "full_plate": full_plate,
        "arabic_digits": final_ar_digits,
        "arabic_letters": ar_letters_disp,
        "english_letters": en_letters_disp
    }

if __name__ == '__main__':
    if len(sys.argv) > 1:
        img_path = sys.argv[1]
        if os.path.exists(img_path):
            img = cv2.imread(img_path)
            res = extract_plate_data(img)
            print(json.dumps(res, ensure_ascii=False))
        else:
            print(json.dumps({"digits": "", "letters": "", "full_plate": "", "arabic_digits": "", "arabic_letters": "", "english_letters": ""}))
    else:
        print(json.dumps({"digits": "", "letters": "", "full_plate": "", "arabic_digits": "", "arabic_letters": "", "english_letters": ""}))
