"""
BrainCNN — lightweight dual-head network that predicts throttle
and steering from a single grayscale frame.

Input  : (N, 1, 120, 160)  – pixel values 0-255
Output : tuple(
            throttle_logits (N, 3),   # 0 neutral, 1 accelerate, 2 brake
            steer_logits    (N, 3)    # 0 straight, 1 left, 2 right
         )
"""
import torch
import torch.nn as nn
import torch.nn.functional as F


class BrainCNN(nn.Module):
    def __init__(self) -> None:
        super().__init__()

        # ── feature extractor ──────────────────────────────────────────
        self.backbone = nn.Sequential(
            # (1, 120, 160) → (16, 58, 78)
            nn.Conv2d(1, 16, kernel_size=5, stride=2, padding=1),
            nn.ReLU(inplace=True),

            # (16, 58, 78) → (32, 29, 39)
            nn.Conv2d(16, 32, kernel_size=3, stride=2, padding=1),
            nn.ReLU(inplace=True),

            # (32, 29, 39) → (64, 15, 20)
            nn.Conv2d(32, 64, kernel_size=3, stride=2, padding=1),
            nn.ReLU(inplace=True),

            nn.Flatten(),  # → (*, flat_dim)
        )

        # ── compute flattened feature dimension programmatically ──────
        with torch.no_grad():
            dummy = torch.zeros(1, 1, 120, 160)        # (N=1, C=1, H=120, W=160)
            flat_dim = self.backbone(dummy).shape[1]    # e.g. 64*15*20 = 19 200

        # ── classification heads ──────────────────────────────────────
        self.fc_throttle = nn.Linear(flat_dim, 3)
        self.fc_steer    = nn.Linear(flat_dim, 3)

    # ──────────────────────────────────────────────────────────────────
    def forward(self, x: torch.Tensor):
        """
        x : (N, 1, 120, 160), dtype float32 or uint8
        returns (throttle_logits, steer_logits)
        """
        z = self.backbone(x.float() / 255.0)  # normalize to [0,1]
        return self.fc_throttle(z), self.fc_steer(z)
