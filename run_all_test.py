import glob, sys, json
from smart_plate_detector import scan_image

files = sorted(glob.glob('/tmp/test_plates/*.jpeg'))
for f in files:
    res = scan_image(f)
    print(f"{f.split('/')[-1]}: {json.dumps(res, ensure_ascii=False)}")
