import pathlib, re, cv2, numpy as np, tqdm, os

PATTERN = re.compile(r"frame_\d+_(accelerate|brake|)_(left|right|)\.png$")
LABELS_T = {"":0,"accelerate":1,"brake":2}
LABELS_S = {"":0,"left":1,"right":2}

def build_if_needed(data_root, out="dataset/dataset.npz", size=(120,160)):
    if pathlib.Path(out).exists():
        return
    imgs, tl, sl = [], [], []
    for p in tqdm.tqdm(list(pathlib.Path(data_root).rglob("frame_*.png")),
                       desc="Building dataset"):
        m = PATTERN.search(p.name)
        if not m: continue
        throttle, steer = m.groups()
        img = cv2.imread(str(p), cv2.IMREAD_GRAYSCALE)
        img = cv2.resize(img, size, interpolation=cv2.INTER_AREA)
        imgs.append(img)
        tl.append(LABELS_T[throttle]); sl.append(LABELS_S[steer])
    os.makedirs(pathlib.Path(out).parent, exist_ok=True)
    np.savez_compressed(out, images=np.stack(imgs),
                        throttle=np.array(tl), steer=np.array(sl))
