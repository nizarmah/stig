# stig

The ultimate racer.

## Setup

### Environment variables

1. Copy `example.env` to `.env`.

    ```bash
    cp example.env .env
    ```

1. Keep `BROWSER_WS_URL` empty. We'll add it later.

## Run

1. Start Chrome with remote debugging enabled.

    ```bash
    # linux
    google-chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile
    # mac osx
    /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile
    ```

1. Copy the browser's WebSocket URL into `.env` as `BROWSER_WS_URL`.

    ```bash
    > /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222 --user-data-dir=/tmp/stig-profile

    DevTools listening on ws://127.0.0.1:9222/devtools/browser/b2bee96f-9556-4af4-94aa-e6843af576b9
    ```

    ```diff
    diff --git a/.env b/.env
    index 90c66f5..1d96ca6 100644
    --- a/.env
    +++ b/.env
    @@ -1,2 +1,2 @@
    -BROWSER_WS_URL=
    +BROWSER_WS_URL=ws://127.0.0.1:9222/devtools/browser/b2bee96f-9556-4af4-94aa-e6843af576b9
    ```

1. Run stig.

    ```bash
    make run
    ```
