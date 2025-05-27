# stig

Stig is a Python package for training and running the racer.

## Setup

### Environment

From the root directory of this repository, create the `.env` files.

```bash
make env
```

### Assets

From the root directory of this repository, create the `assets` directory.

```bash
make assets
```

## Run

The stig client has different commands. Click on the command to learn how to run it.

- [`train`][cmd-train] - Train the AI using recordings.
- [`autopilot`][cmd-autopilot] - Autopilot the game using a trained AI.

[cmd-train]: ./cmd/train
[cmd-autopilot]: ./cmd/autopilot
