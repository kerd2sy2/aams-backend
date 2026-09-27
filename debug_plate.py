import cv2, pytesseract, re, glob, os, json

img = cv2.imread('/tmp/test_plates/WhatsApp Image 2026-09-26 at 10.01.31 PM.jpeg')
if img is not None:
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    clahe = cv2.createCLAHE(clipLimit=3.0, tileGridSize=(8,8))
    cl = clahe.apply(gray)
    
    txt_eng = pytesseract.image_to_string(cl, config='--psm 11')
    txt_digits = pytesseract.image_to_string(cl, config='--psm 6 -c tessedit_char_whitelist=0123456789')
    txt_ara = pytesseract.image_to_string(cl, lang='ara', config='--psm 11')
    
    print("ENG:", repr(txt_eng))
    print("DIGITS:", repr(txt_digits))
    print("ARA:", repr(txt_ara))
