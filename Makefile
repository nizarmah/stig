.PHONY: env game-play game-record stig-train stig-drive

# copy a file if it doesn't exist
define copy-file
	@if [ ! -f $(1) ]; then \
		cp $(2) $(1); \
		echo "✅ $(1)"; \
	fi
endef

# create a directory if it doesn't exist
define create-dir
	@if [ ! -d $(1) ]; then \
		mkdir -p $(1); \
		echo "✅ $(1)"; \
	fi
endef

# setup env files
env:
	$(call copy-file,game/env/.env,game/env/example.env)
	$(call copy-file,game/env/play/.env,game/env/play/example.env)
	$(call copy-file,game/env/record/.env,game/env/record/.env)
	$(call copy-file,stig/env/.env,stig/env/example.env)
	$(call copy-file,stig/env/autopilot/.env,stig/env/autopilot/example.env)
	$(call copy-file,stig/env/train/.env,stig/env/train/example.env)

# setup assets directories
assets:
	$(call create-dir,assets/datasets)
	$(call create-dir,assets/models)
	$(call create-dir,assets/recordings)

# play the game
game-play:
	@docker compose run --rm --build game-play

# record the game
game-record:
	@docker compose run --rm --build game-record

# run the autopilot
stig-autopilot:
	@docker compose run --rm --build stig-autopilot

# train the model
stig-train:
	@docker compose run --rm --build stig-train
