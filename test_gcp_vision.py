import os, glob, re, json
from google.cloud import vision

os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/opt/aams-backend/gcp-vision-credentials.json"

client = vision.ImageAnnotatorClient()

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
    "AD": "د ا", "DA": "د ا", "A D": "د ا", "D A": "د ا",
    "BT": "ط ب", "TB": "ط ب", "B T": "ط ب", "T B": "ط ب",
    "BE": "ع ب", "EB": "ع ب", "B E": "ع ب", "E B": "ع ب",
    "AJ": "ح ا", "JA": "ح ا", "A J": "ح ا", "J A": "ح ا",
    "RE": "ر ع", "ER": "ر ع", "R E": "ر ع", "E R": "ر ع",
    "AH": "ا ح", "HA": "ا ح", "A H": "ا ح", "H A": "ا ح",
    "د ا": "د ا", "ط ب": "ط ب", "ع ب": "ع ب", "ح ا": "ح ا", "ر ع": "ر ع"
}

def vision_ocr_plate(image_path):
    with open(image_path, "rb") as image_file:
        content = image_file.read()

    image = vision.Image(content=content)
    response = client.text_detection(image=image)
    texts = response.text_annotations

    if not texts:
        return {"digits": "", "letters": "", "full_plate": ""}

    full_text = texts[0].description
    norm_text = full_text
    for ar_d, en_d in [('٠','0'), ('١','1'), ('٢','2'), ('٣','3'), ('٤','4'), ('٥','5'), ('٦','6'), ('٧','7'), ('٨','8'), ('٩','9')]:
        norm_text = norm_text.replace(ar_d, en_d)

    found_digits = ""
    for bike in SAUDI_PLATES_FLEET:
        if re.search(r'(?<!\d)' + re.escape(bike) + r'(?!\d)', norm_text):
            found_digits = bike
            break

    if not found_digits:
        cands = re.findall(r'\b\d{3,4}\b', norm_text)
        if cands:
            found_digits = cands[0]

    found_letters = ""
    upper_text = full_text.upper()
    for key, ar_val in EN_TO_AR_PAIRS.items():
        if key in upper_text:
            found_letters = ar_val
            break

    full_plate = f"{found_digits} {found_letters}".strip() if found_digits else (found_letters or "")
    return {
        "digits": found_digits,
        "letters": found_letters,
        "full_plate": full_plate,
        "raw_text": repr(full_text.replace('\n', ' '))[:80]
    }

files = sorted(glob.glob('/tmp/new_plates/*.jpg'))[:10]
for f in files:
    try:
        res = vision_ocr_plate(f)
        print(f"{os.path.basename(f)} -> {res['full_plate']} | {res['digits']} | {res['letters']} (Raw: {res['raw_text']})")
    except Exception as e:
        print(f"Error on {f}: {e}")
