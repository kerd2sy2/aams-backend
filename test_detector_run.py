import cv2, os, easyocr
from smart_plate_detector import extract_plate_data

reader = easyocr.Reader(['ar', 'en'], gpu=False, verbose=False)
for f in sorted(os.listdir('/tmp/new_plates/'))[:10]:
    p = os.path.join('/tmp/new_plates', f)
    img = cv2.imread(p)
    if img is not None:
        res = extract_plate_data(img)
        print(f, '->', res)
