import requests, base64, json, cv2, numpy as np, os

# 1. Login with actual credentials
res_login = requests.post('http://127.0.0.1:8081/api/v1/login', json={'login': '0500626432', 'password': 'password123'})
login_data = res_login.json()
token = login_data.get('access_token')
print('LOGIN STATUS:', res_login.status_code)
print('TOKEN RECEIVED:', bool(token))

headers = {'Authorization': f'Bearer {token}', 'Content-Type': 'application/json'}

# 2. Test blank black image
blank = np.zeros((300, 400, 3), dtype=np.uint8)
_, buf = cv2.imencode('.jpg', blank)
b64_blank = base64.b64encode(buf).decode('utf-8')

res_blank = requests.post('http://127.0.0.1:8081/api/v1/work/scan-plate', headers=headers, json={'image': f'data:image/jpeg;base64,{b64_blank}'})
print('BLANK SCAN STATUS:', res_blank.status_code)
print('BLANK SCAN BODY:', res_blank.text)

# 3. Test Real Plate Image (6534 AD)
if os.path.exists('/tmp/new_plates/frame_000001.jpg'):
    with open('/tmp/new_plates/frame_000001.jpg', 'rb') as f:
        real_b64 = base64.b64encode(f.read()).decode('utf-8')
    res_real = requests.post('http://127.0.0.1:8081/api/v1/work/scan-plate', headers=headers, json={'image': f'data:image/jpeg;base64,{real_b64}'})
    print('REAL PLATE STATUS:', res_real.status_code)
    print('REAL PLATE BODY:', res_real.text)
