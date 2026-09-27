import sys, os, cv2, glob
sys.path.insert(0, '/opt/aams-backend')
from smart_plate_detector import extract_plate_data

files = sorted(glob.glob('/tmp/new_plates/*.jpg'))
print(f"Total test plate files: {len(files)}")
success = 0
for f in files[:15]:
    img = cv2.imread(f)
    if img is not None:
        res = extract_plate_data(img)
        print(f"{os.path.basename(f)} -> DIGITS: '{res['digits']}', AR_LETTERS: '{res['arabic_letters']}', EN_LETTERS: '{res['english_letters']}'")
        if res['digits'] and res['arabic_letters']:
            success += 1

print(f"Successfully recognized: {success}/{min(len(files), 15)}")
