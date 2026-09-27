import os, glob, re, sys
import pytesseract
from PIL import Image, ImageEnhance, ImageFilter, ImageOps

# Saudi plate valid letters
SAUDI_AR_LETTERS = set('ابحدرسطعقكلمنهوي')
SAUDI_EN_LETTERS = set('ABJDRSXTEGKZMNHUV')

AR_TO_EN = {
    'ا': 'A', 'أ': 'A', 'إ': 'A', 'آ': 'A',
    'ب': 'B', 'ح': 'J', 'د': 'D', 'ر': 'R',
    'س': 'S', 'ص': 'X', 'ط': 'T', 'ع': 'E',
    'ق': 'G', 'ك': 'K', 'ل': 'L', 'م': 'Z',
    'ن': 'N', 'ه': 'H', 'و': 'U', 'ي': 'V', 'ى': 'V'
}

EN_TO_AR = {v: k for k, v in AR_TO_EN.items()}

def parse_plate_from_image(img_path):
    img = Image.open(img_path)
    w, h = img.size
    
    # Try multiple pre-processing filters for best character extraction
    # 1. Grayscale + high contrast
    gray = ImageOps.grayscale(img)
    enhanced = ImageEnhance.Contrast(gray).enhance(2.5)
    
    # Also crop bottom 70% or center if too much background
    ocr_text = pytesseract.image_to_string(enhanced, lang='ara+eng', config='--psm 11')
    
    # Also run with PSM 6 (uniform block of text)
    ocr_text_psm6 = pytesseract.image_to_string(enhanced, lang='ara+eng', config='--psm 6')
    
    full_text = ocr_text + " " + ocr_text_psm6
    
    # Normalize Arabic digits ٠-٩ to 0-9
    ar_digits = {'٠':'0','١':'1','٢':'2','٣':'3','٤':'4','٥':'5','٦':'6','٧':'7','٨':'8','٩':'9'}
    normalized = full_text
    for ar, en in ar_digits.items():
        normalized = normalized.replace(ar, en)
        
    # Extract 3-4 digit numbers
    digits_matches = re.findall(r'\b\d{3,4}\b', normalized)
    
    # Extract 2-3 letter combinations
    found_ar_letters = []
    found_en_letters = []
    
    for word in full_text.split():
        clean_ar = [c for c in word if c in SAUDI_AR_LETTERS]
        if 2 <= len(clean_ar) <= 3:
            found_ar_letters.append(" ".join(clean_ar))
            
        clean_en = [c for c in word.upper() if c in SAUDI_EN_LETTERS]
        if 2 <= len(clean_en) <= 3:
            found_en_letters.append(" ".join(clean_en))
            
    return {
        "raw": repr(full_text.strip()[:100]),
        "digits_candidates": digits_matches,
        "ar_letters_candidates": found_ar_letters,
        "en_letters_candidates": found_en_letters,
    }

if __name__ == '__main__':
    files = glob.glob('/tmp/test_plates/*.jpeg')
    print(f"Testing OCR on {len(files)} plate images:")
    for f in files:
        res = parse_plate_from_image(f)
        print(f"[{os.path.basename(f)}]")
        print("  Digits:", res["digits_candidates"])
        print("  Ar Letters:", res["ar_letters_candidates"])
        print("  En Letters:", res["en_letters_candidates"])
        print("  Raw Sample:", res["raw"])
        print()
