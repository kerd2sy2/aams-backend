import cv2, pytesseract, re, glob, os, json

SAUDI_PLATES_FLEET = [
    "2381", "2383", "2387", "2458", "2459", "2565",
    "5097", "5098", "5099", "5442", "5443", "5445", "5446", "5447", "5448", "5449", "5450", "5451", "5452",
    "6032", "6238", "6239", "6240", "6241", "6242",
    "651", "652", "6530", "6531", "6533", "6534", "6535", "6536", "6537", "6538", "6540", "6541", "6542", "6543", "6544", "6545", "6546", "6547", "6548", "6549",
    "6568", "6576", "7030", "7036", "7037", "7038", "7039",
    "7568", "7569", "7570", "7571", "7572", "7573", "7574", "7577", "7578",
    "8020", "8021", "8022", "8023", "8034", "8035", "8036", "8037", "8039", "8040", "8041", "8042", "8043", "8044", "8045", "8046", "8047", "8048",
    "8872", "8873", "8874", "8875", "8876"
]

EN_TO_AR_PAIRS = {
    "AD": "د ا", "DA": "د ا",
    "BT": "ط ب", "TB": "ط ب",
    "BE": "ع ب", "EB": "ع ب",
    "AJ": "ح ا", "JA": "ح ا",
    "RE": "ر ع", "ER": "ر ع",
    "AH": "ا ح", "HA": "ا ح",
}

def scan_plate_fast(image_path):
    img = cv2.imread(image_path)
    if img is None:
        return {"digits": "", "letters": "", "full_plate": ""}
        
    h, w = img.shape[:2]
    scale = 640.0 / max(h, w)
    resized = cv2.resize(img, (int(w * scale), int(h * scale)))
    gray = cv2.cvtColor(resized, cv2.COLOR_BGR2GRAY)
    
    clahe = cv2.createCLAHE(clipLimit=2.5, tileGridSize=(8, 8))
    enhanced = clahe.apply(gray)
    
    # Bottom half (English row)
    eh, ew = enhanced.shape[:2]
    bot_half = enhanced[int(eh * 0.35):eh, 0:ew]
    _, bot_otsu = cv2.threshold(bot_half, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
    bot_inv = cv2.bitwise_not(bot_otsu)
    
    # Left side (digits)
    bot_left = bot_inv[:, 0:int(ew * 0.65)]
    # Right side (letters)
    bot_right = bot_inv[:, int(ew * 0.35):int(ew * 0.88)]
    
    digits_txt = pytesseract.image_to_string(bot_left, config='--oem 3 --psm 6 -c tessedit_char_whitelist=0123456789')
    letters_txt = pytesseract.image_to_string(bot_right, config='--oem 3 --psm 6 -c tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ').upper()
    
    # Whole bot_half backup
    all_digits = pytesseract.image_to_string(bot_inv, config='--oem 3 --psm 6 -c tessedit_char_whitelist=0123456789')
    
    found_digits = ""
    for bike in SAUDI_PLATES_FLEET:
        if bike in digits_txt or bike in all_digits:
            found_digits = bike
            break
            
    if not found_digits:
        cands = re.findall(r'\b\d{3,4}\b', f"{digits_txt} {all_digits}")
        if cands:
            found_digits = cands[0]
            
    found_letters = ""
    for en, ar in EN_TO_AR_PAIRS.items():
        if en in letters_txt:
            found_letters = ar
            break
            
    if not found_letters:
        if "AD" in letters_txt or "D" in letters_txt or "A" in letters_txt:
            found_letters = "د ا"
        elif "BE" in letters_txt or "E" in letters_txt or "B" in letters_txt:
            found_letters = "ع ب"
        elif "RE" in letters_txt or "R" in letters_txt:
            found_letters = "ر ع"
        elif "AJ" in letters_txt or "J" in letters_txt:
            found_letters = "ح ا"
        elif "BT" in letters_txt or "T" in letters_txt:
            found_letters = "ط ب"
            
    return {
        "digits": found_digits,
        "letters": found_letters,
        "full_plate": f"{found_digits} {found_letters}".strip() if found_digits else (found_letters or "")
    }

files = sorted(glob.glob('/tmp/new_plates/*.jpg'))[:8]
for f in files:
    res = scan_plate_fast(f)
    print(f"{os.path.basename(f)} -> {res}")
