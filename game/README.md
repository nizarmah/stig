# game

Game is a Go package for the game client.

## Setup

### Browser

Before you can use the game client, you need to start a browser with remote debugging enabled.

```bash
# mac osx
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile
# linux
google-chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile
```

Copy the WebSocket URL. We'll use it in the next step.

```bash
> /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile

DevTools listening on ws://127.0.0.1:9222/devtools/browser/2b1ea713-517a-479e-a17b-3958601b23fb
```

### Environment

From the root directory of this repository, create the `.env` files.

```bash
make env
```

Replace the `BROWSER_WS_URL` with the WebSocket URL you copied earlier.

```diff
# ./game/env/.env
-BROWSER_WS_URL=
+BROWSER_WS_URL=ws://127.0.0.1:9222/devtools/browser/2b1ea713-517a-479e-a17b-3958601b23fb
```

## Run

The game client has different commands. Click on the command to learn how to run it.

- [`play`][cmd-play] - Play the game with AI.
- [`record`][cmd-record] - Record gameplay for training the AI.

[cmd-play]: ./cmd/play
[cmd-record]: ./cmd/record
