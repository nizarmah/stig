.PHONY: env game-play game-record track-train

# setup env files
env:
	@echo "Setting up environment files..."
	@if [ ! -f game/env/.env ]; then \
		cp game/env/example.env game/env/.env; \
		echo "Created game/env/.env"; \
	else \
		echo "game/env/.env already exists"; \
	fi
	@if [ ! -f game/env/play/.env ]; then \
		cp game/env/play/example.env game/env/play/.env; \
		echo "Created game/env/play/.env"; \
	else \
		echo "game/env/play/.env already exists"; \
	fi
	@if [ ! -f game/env/record/.env ]; then \
		cp game/env/record/example.env game/env/record/.env; \
		echo "Created game/env/record/.env"; \
	else \
		echo "game/env/record/.env already exists"; \
	fi
	@echo "Done."

# play the game
game-play:
	@docker compose run --rm --build \
			game-play

# record the game
game-record:
	@docker compose run --rm --build \
			game-record

# train the model
track-train:
	@docker compose run --rm --build \
			track-train --epochs 10
