"""
Package dataset provides functions to load the dataset.
"""

from pathlib import Path
from typing import Tuple

import re, cv2, numpy as np, tqdm, torch, torch.utils.data as data

from stig.internal.game import action
from stig.internal.dataset.image import process_from_path

FRAME_RE = re.compile(
    r"frame_(?P<frame_timestamp>\d+)_(?P<throttle>accelerate|brake|)_(?P<steering>left|right|)\.jpe?g$",
    re.IGNORECASE,
)

def build_dataset(
    model_name: str,
    datasets_dir: str,
    recordings_dir: str,
    size: Tuple[int, int],
) -> str:
    """
    Returns path to the .npz file (existing or freshly built).
    """
    # prepare directories
    datasets_root = Path(datasets_dir)
    recordings_root = Path(recordings_dir)

    # create the output path
    manifest = datasets_root / f"{model_name}.npz"

    # check if the dataset is up to date
    newest_src = _newest_mtime(recordings_root)
    if manifest.exists() and manifest.stat().st_mtime >= newest_src:
        return str(manifest)

    # count how many frames are in the recordings directory
    frames = _all_frames(recordings_root)

    # ensure we have some frames
    if not frames:
        raise RuntimeError(f"no frames found in {recordings_dir}")

    images, throttles, steerings = [], [], []

    for p in tqdm.tqdm(frames, desc="Building dataset"):
        re_match = FRAME_RE.match(p.name)
        if not re_match:
            raise RuntimeError(f"unexpected frame file: {p}")

        throttle = re_match.group("throttle")
        throttles.append(action.THROTTLE_LABELS_MAP[throttle])

        steering = re_match.group("steering")
        steerings.append(action.STEERING_LABELS_MAP[steering])

        img = process_from_path(str(p), size)
        images.append(img)

    with tqdm.tqdm(total=1, desc="Saving dataset") as pbar:
        np.savez_compressed(
            manifest,
            images=np.array(images),
            throttles=np.array(throttles),
            steerings=np.array(steerings),
        )

        pbar.update(1)

    return str(manifest)

def load_dataset(npz_path: str) -> data.TensorDataset:
    """
    Loads a dataset from a .npz file.
    """
    with tqdm.tqdm(total=1, desc="Loading dataset") as pbar:
        npz = np.load(npz_path)

        dataset = data.TensorDataset(
          torch.from_numpy(npz["images"]).float()[:, None], # (N,1,H,W),
          torch.from_numpy(npz["throttles"]),
          torch.from_numpy(npz["steerings"]),
        )

        pbar.update(1)

    return dataset

def _all_frames(root: Path) -> list[Path]:
    """Every file whose **name** matches FRAME_RE (any depth)."""
    return sorted(p for p in root.rglob("*") if FRAME_RE.match(p.name))

def _newest_mtime(root: Path) -> float:
    """Most recent mtime among all frame files."""
    return max(p.stat().st_mtime for p in _all_frames(root))
