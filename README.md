# stig

> The ultimate racer.

A mono-repo for the AI that plays the [Shopify Horizon Drive game][shopify-drive].

## Setup

### Docker

Make sure you have [Docker](https://docs.docker.com/get-docker/) installed.

### Environment

From the root directory of this repository, create the `.env` files.

```bash
make env
```

## Quick Start

We still don't have a proper guide, yet.

For now, you can check the README of each component to learn how to run it.

- [`game/`][repo-game] - The game client.
    * Used to record gameplay for supervised learning.
    * Used to play the game with the trained AI.
- [`stig/`][repo-stig] - The stig client.
    * Used to train an AI model using supervised learning.
    * Used to autopilot the driving using an AI model.

[shopify-drive]: https://www.shopify.com/ca/editions/summer2025/drive

[repo-game]: ./game
[repo-track]: ./stig
