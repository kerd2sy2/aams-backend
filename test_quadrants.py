import cv2, pytesseract, re, numpy as np, os, glob

def test_crop_quadrants(image_path):
    img = cv2.imread(image_path)
    if img is None:
        return
        
    h, w = img.shape[:2]
    # Resize max dimension to 1000
    scale = 1000.0 / max(h, w)
    resized = cv2.resize(img, (int(w * scale), int(h * scale)))
    gray = cv2.cvtColor(resized, cv2.COLOR_BGR2GRAY)
    
    # Try finding plate contour
    blur = cv2.GaussianBlur(gray, (5, 5), 0)
    thresh = cv2.adaptiveThreshold(blur, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY_INV, 19, 9)
    
    contours, _ = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    
    best_plate = None
    best_area = 0
    
    for c in contours:
        x, y, cw, ch = cv2.boundingRect(c)
        area = cw * ch
        aspect = float(cw) / max(1, ch)
        # Saudi bike plates have aspect ratio between 1.2 and 2.4 and occupy at least 15% of image
        if 1.2 <= aspect <= 2.8 and area > (resized.shape[0] * resized.shape[1] * 0.10):
            if area > best_area:
                best_area = area
                best_plate = (x, y, cw, ch)
                
    if best_plate:
        x, y, cw, ch = best_plate
        plate_crop = resized[y:y+ch, x:x+cw]
    else:
        # If no contour found, crop the central 70%
        margin_y = int(resized.shape[0] * 0.15)
        margin_x = int(resized.shape[1] * 0.10)
        plate_crop = resized[margin_y:resized.shape[0]-margin_y, margin_x:resized.shape[1]-margin_x]
        
    ph, pw = plate_crop.shape[:2]
    plate_gray = cv2.cvtColor(plate_crop, cv2.COLOR_BGR2GRAY)
    
    # 4 Quadrants:
    # Top-Left: Arabic digits
    # Top-Right: Arabic letters
    # Bottom-Left: English digits (y: ph//2 to ph, x: 0 to int(pw * 0.55))
    # Bottom-Right: English letters (y: ph//2 to ph, x: int(pw * 0.50) to int(pw * 0.85))
    
    bot_left = plate_gray[int(ph*0.45):ph, 0:int(pw*0.58)]
    bot_right = plate_gray[int(ph*0.45):ph, int(pw*0.48):int(pw*0.85)]
    
    # Preprocess bottom-left (English digits)
    _, bl_thresh = cv2.threshold(bot_left, 0, 255, cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU)
    # Preprocess bottom-right (English letters)
    _, br_thresh = cv2.threshold(bot_right, 0, 255, cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU)
    
    digits_raw = pytesseract.image_to_string(bl_thresh, config='--oem 3 --psm 7 -c tessedit_char_whitelist=0123456789')
    digits_raw2 = pytesseract.image_to_string(bot_left, config='--oem 3 --psm 7 -c tessedit_char_whitelist=0123456789')
    
    letters_raw = pytesseract.image_to_string(br_thresh, config='--oem 3 --psm 7 -c tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ')
    letters_raw2 = pytesseract.image_to_string(bot_right, config='--oem 3 --psm 7 -c tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ')
    
    print(f"File: {os.path.basename(image_path)}")
    print(f"  Bot-Left Digits: raw1={repr(digits_raw.strip())} raw2={repr(digits_raw2.strip())}")
    print(f"  Bot-Right Letters: raw1={repr(letters_raw.strip())} raw2={repr(letters_raw2.strip())}")
    print()

for f in sorted(glob.glob('/tmp/test_plates/*.jpeg')):
    test_crop_quadrants(f)
