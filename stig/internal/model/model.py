"""
Package model provides the neural network model.
"""

from typing import Tuple

import torch, torch.nn as nn

class StigNet(nn.Module):
    """
    A *tiny* convolutional network that turns one grayscale frame to:
        throttle_logits (0: neutral, 1: accelerate, 2: brake)
        steering_logits (0: straight, 1: left, 2: right)

    Parameters:
        size: [Height, Width] of the dataset frames.
    """

    def __init__(
        self,
        size: Tuple[int, int], # (height, width)
    ):
        super().__init__()

        self.cnn = nn.Sequential(
            nn.Conv2d(1, 16, 5, stride=2, padding=1), nn.ReLU(),
            nn.Conv2d(16, 32, 3, stride=2, padding=1), nn.ReLU(),
            nn.Conv2d(32, 64, 3, stride=2, padding=1), nn.ReLU(),
            nn.Flatten()
        )

        with torch.no_grad():
            flat = self.cnn(
                torch.zeros(1,1,size[0],size[1])
            ).shape[1] # auto-derive

        self.fc_t = nn.Linear(flat, 3)
        self.fc_s = nn.Linear(flat, 3)

    def forward(
        self,
        x: torch.Tensor
    ) -> Tuple[torch.Tensor, torch.Tensor]:
        x = x.to(dtype=torch.float32, copy=False) / 255.0
        z = self.cnn(x)
        return self.fc_t(z), self.fc_s(z)
