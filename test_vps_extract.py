import sys, os, cv2, numpy as np
sys.path.insert(0, '/opt/aams-backend')
from smart_plate_detector import extract_plate_data

blank = np.zeros((300, 400, 3), dtype=np.uint8)
res_blank = extract_plate_data(blank)
print('BLANK ->', res_blank)

if os.path.exists('/tmp/new_plates/frame_000001.jpg'):
    real = cv2.imread('/tmp/new_plates/frame_000001.jpg')
    res_real = extract_plate_data(real)
    print('REAL ->', res_real)
