"""
Command train trains a model on a given dataset.
"""

from pathlib import Path
from typing import Tuple

import tqdm, torch, torch.nn as nn, torch.utils.data as data

from stig.internal.env import env
from stig.internal.dataset.dataset import build_dataset, load_dataset
from stig.internal.model.model import StigNet

def train(
    model_name: str,
    datasets_dir: str,
    models_dir: str,
    recordings_dir: str,
    frame_size: Tuple[int, int],
    epochs: int,
    batch_size: int,
    learning_rate: float,
    device: str,
):
    # Create the dataset.
    dataset_path = build_dataset(model_name, datasets_dir, recordings_dir, frame_size)

    # Load the dataset.
    dataset = load_dataset(dataset_path)
    dataloader = data.DataLoader(dataset, batch_size=batch_size, shuffle=True)

    # Create the model.
    net = StigNet(frame_size).to(device)
    opt = torch.optim.Adam(net.parameters(), learning_rate)
    ce = nn.CrossEntropyLoss()

    # Train the model.
    for ep in range(epochs):
        tot = tc = sc = 0

        for x, t, s in tqdm.tqdm(dataloader, desc=f"epoch {ep:02d}"):
            ot, os = net(x)
            loss = ce(ot, t) + ce(os, s)

            opt.zero_grad()
            loss.backward()
            opt.step()

            tot += x.size(0)
            tc += (ot.argmax(1) == t).sum().item()
            sc += (os.argmax(1) == s).sum().item()

        print(
          f"epoch {ep:02d}\t"
          f"throttle_acc={tc/tot:.3f}\t"
          f"steering_acc={sc/tot:.3f}\t"
        )

    # Save the model.
    model_path = Path(models_dir) / f"{model_name}.pt"
    torch.jit.script(net).save(str(model_path))
    print(f"✅ saved {model_path}")

if __name__ == "__main__":
    model_name = env.lookup("MODEL_NAME")

    datasets_dir = env.lookup("DATASETS_DIR")
    models_dir = env.lookup("MODELS_DIR")
    recordings_dir = env.lookup("RECORDINGS_DIR")

    frame_height = env.lookupInt("FRAME_HEIGHT")
    frame_width = env.lookupInt("FRAME_WIDTH")
    frame_size = (frame_height, frame_width)

    epochs = env.lookupInt("EPOCHS")
    batch_size = env.lookupInt("BATCH_SIZE")
    learning_rate = env.lookupFloat("LEARNING_RATE")
    device = "cuda" if torch.cuda.is_available() else "cpu"

    train(
      model_name,
      datasets_dir,
      models_dir,
      recordings_dir,
      frame_size,
      epochs,
      batch_size,
      learning_rate,
      device,
    )
