.PHONY: game-play game-record

# Play the game.
game-play:
	include game/env/play/.env
	export
	@go run game/cmd/play/main.go

# Record gameplay.
game-record:
	include game/env/record/.env
	export
	@go run game/cmd/record/main.go
