# stig

> The ultimate racer.

A mono-repo for the AI that plays the [Shopify Horizon Drive game][shopify-drive].

## Prerequisites

- [Docker](https://www.docker.com/get-started) - Required for running the game client and autopilot.
- [Google Chrome](https://www.google.com/chrome/) - Required for browser automation (other browsers may work but are not documented).

## Quick Start

Follow these steps to get Stig up and running:

### 1. Initial Setup

First, set up the environment configuration files and create necessary directories:

```bash
# Create environment files from templates
make env

# Create asset directories (datasets, models, recordings)
make assets
```

### 2. Download Pre-trained Model

Download the pre-trained novice model:

```bash
make stig-novice
```

### 3. Start Chrome with Remote Debugging

Open a new terminal and start Chrome with remote debugging enabled:

**macOS:**
```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile
```

**Linux:**
```bash
google-chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile
```

### 4. Configure WebSocket URL

After starting Chrome, you'll see output like this:
```
DevTools listening on ws://127.0.0.1:9222/devtools/browser/2b1ea713-517a-479e-a17b-3958601b23fb
```

Copy the WebSocket URL (the `ws://...` part) and update your environment files:

1. Edit `game/env/.env`
2. Set `BROWSER_WS_URL` to the WebSocket URL you copied

Example:
```properties
BROWSER_WS_URL=ws://127.0.0.1:9222/devtools/browser/2b1ea713-517a-479e-a17b-3958601b23fb
```

### 5. Run Stig

Open two new terminals and run these commands:

**Terminal 1 - Start the autopilot:**
```bash
make stig-autopilot
```

**Terminal 2 - Start the game client:**
```bash
make game-play
```

### Additional Commands

- **Record gameplay:** `make game-record`
- **Train a new model:** `make stig-train`

[shopify-drive]: https://www.shopify.com/ca/editions/summer2025/drive
