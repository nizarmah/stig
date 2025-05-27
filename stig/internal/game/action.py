"""
Package game provides entities for the game.
"""

THROTTLE_LABELS_MAP = {
  "": 0,
  "accelerate": 1,
  "brake": 2,
}

THROTTLE_VALUES_MAP = {v: k for k, v in THROTTLE_LABELS_MAP.items()}

STEERING_LABELS_MAP = {
  "": 0,
  "left": 1,
  "right": 2,
}

STEERING_VALUES_MAP = {v: k for k, v in STEERING_LABELS_MAP.items()}
