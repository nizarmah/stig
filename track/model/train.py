# track/model/train.py
import argparse, os, pathlib, numpy as np, torch, torch.nn as nn
import torch.optim as optim, torch.utils.data as d
from .brain import BrainCNN           # CNN you defined earlier
from dataset.build import build_if_needed

DATA_ROOT  = os.getenv("DATA_ROOT",  "../assets/recordings")
MODEL_DIR  = pathlib.Path(os.getenv("MODEL_DIR", "../models"))
MODEL_DIR.mkdir(parents=True, exist_ok=True)

class Frames(d.Dataset):
    def __init__(self, npz):
        d = np.load(npz)
        self.x = torch.tensor(d["images"]).unsqueeze(1).float()
        self.t = torch.tensor(d["throttle"]).long()
        self.s = torch.tensor(d["steer"]).long()
    def __len__(self): return len(self.t)
    def __getitem__(self, i): return self.x[i], self.t[i], self.s[i]

def train(args):
    build_if_needed(DATA_ROOT)                     # ensures dataset.npz exists
    ds  = Frames("dataset/dataset.npz")
    dl  = d.DataLoader(ds, batch_size=args.bs, shuffle=True, num_workers=4)
    net = BrainCNN().to(args.dev)
    optim_ = optim.Adam(net.parameters(), lr=args.lr)
    ce = nn.CrossEntropyLoss()

    for epoch in range(args.epochs):
        n, correct_t, correct_s = 0, 0, 0
        for x, t, s in dl:
            x, t, s = x.to(args.dev), t.to(args.dev), s.to(args.dev)
            out_t, out_s = net(x)
            loss = (ce(out_t, t) + ce(out_s, s)) / 2
            optim_.zero_grad(); loss.backward(); optim_.step()

            n += x.size(0)
            correct_t += (out_t.argmax(1) == t).sum().item()
            correct_s += (out_s.argmax(1) == s).sum().item()

        print(f"epoch {epoch:02d} "
              f"thr_acc={correct_t/n:.3f} "
              f"str_acc={correct_s/n:.3f}")

    torch.jit.script(net).save(MODEL_DIR / "brain-v0.pt")
    print("✅  Saved", MODEL_DIR / "brain-v0.pt")

if __name__ == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument("--epochs", type=int, default=5)
    p.add_argument("--bs",     type=int, default=128)
    p.add_argument("--lr",     type=float, default=1e-3)
    p.add_argument("--dev",    default="cuda" if torch.cuda.is_available() else "cpu")
    train(p.parse_args())
