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

def get_or_create_dataset(
    model_name: str,
    datasets_dir: str,
    recordings_dir: str,
    size: Tuple[int, int],
) -> data.TensorDataset:
    """
    Returns a dataset (existing or freshly built).
    """
    # Prepare directories.
    datasets_root = Path(datasets_dir)
    recordings_root = Path(recordings_dir)

    # Create the dataset directory.
    dataset_dir = datasets_root / model_name
    dataset_dir.mkdir(parents=True, exist_ok=True)

    # Create the output path.
    imgs_path = dataset_dir / "images.npy"
    meta_path = dataset_dir / "meta.npz"

    # Check if the dataset is up to date.
    newest_src = _newest_mtime(recordings_root)
    if imgs_path.exists() and imgs_path.stat().st_mtime >= newest_src:
        return load_dataset(str(imgs_path), str(meta_path))

    # Count how many frames are in the recordings directory.
    frames = _all_frames(recordings_root)

    # Ensure we have some frames.
    if not frames:
        raise RuntimeError(f"no frames found in {recordings_dir}")

    # Create as numpy arrays so we don't need to convert later.
    images = np.empty((len(frames), size[0], size[1]), np.uint8)
    throttles = np.empty(len(frames), np.int64)
    steerings = np.empty(len(frames), np.int64)

    # Build the dataset.
    for i, p in enumerate(tqdm.tqdm(frames, desc="Building dataset")):
        re_match = FRAME_RE.match(p.name)
        if not re_match:
            raise RuntimeError(f"unexpected frame file: {p}")

        throttles[i] = action.THROTTLE_LABELS_MAP[re_match.group("throttle")]
        steerings[i] = action.STEERING_LABELS_MAP[re_match.group("steering")]

        images[i] = process_from_path(str(p), size)

    # Save the dataset.
    with tqdm.tqdm(total=2, desc="Saving dataset") as pbar:
        np.save(imgs_path, images)
        pbar.update(1)

        np.savez_compressed(meta_path, throttles=throttles, steerings=steerings)
        pbar.update(1)

    # Load the dataset.
    return load_dataset(str(imgs_path), str(meta_path))

def load_dataset(imgs_path: str, meta_path: str) -> data.TensorDataset:
    """
    Loads a dataset from a .npy and .npz file.
    """
    with tqdm.tqdm(total=1, desc="Loading dataset") as pbar:
        # Separate loading to use memory mapping.
        imgs = np.load(imgs_path, mmap_mode="r+")
        meta = np.load(meta_path)

        # Create the dataset.
        dataset = data.TensorDataset(
          torch.from_numpy(imgs)[:, None], # uint8 → (N,1,H,W)
          torch.from_numpy(meta["throttles"]),
          torch.from_numpy(meta["steerings"]),
        )
        pbar.update(1)

    return dataset

def _all_frames(root: Path) -> list[Path]:
    """Every file whose **name** matches FRAME_RE (any depth)."""
    return sorted(p for p in root.rglob("*") if FRAME_RE.match(p.name))

def _newest_mtime(root: Path) -> float:
    """Most recent mtime among all frame files."""
    return max(p.stat().st_mtime for p in _all_frames(root))
