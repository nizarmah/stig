"""
Package image provides image processing utilities.
"""

from typing import Tuple

import base64, cv2, numpy as np, torch

def process_from_bytes(
    img_bytes: bytes,
    frame_size: Tuple[int, int],
    device: str
) -> torch.Tensor:
    """
    Process an image for model inference.
    """
    img = cv2.imdecode(np.frombuffer(img_bytes, np.uint8), cv2.IMREAD_GRAYSCALE)
    if img is None:
        raise ValueError(f"Failed to decode image")

    img = resize_image(
      img,
      height=frame_size[0],
      width=frame_size[1],
    )

    return img

def process_from_path(
    image_path: str,
    frame_size: Tuple[int, int]
) -> np.ndarray:
    """
    Process an image from file path for dataset building.
    """
    img = cv2.imread(str(image_path), cv2.IMREAD_GRAYSCALE)
    if img is None:
        raise ValueError(f"Failed to load image from {image_path}")

    img = resize_image(
      img,
      height=frame_size[0],
      width=frame_size[1],
    )

    return img

def resize_image(
    img: np.ndarray,
    height: int,
    width: int,
) -> np.ndarray:
    """
    Resize an image to the specified size.
    """
    return cv2.resize(img, (width, height), cv2.INTER_AREA)

def image_to_tensor(
    img: np.ndarray,
    device: str,
) -> torch.Tensor:
    """
    Convert a numpy image array to a PyTorch tensor.
    """
    tensor = torch.from_numpy(img).float()[None, None] / 255.0
    return tensor.to(device)
