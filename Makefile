APP_NAME = ddns
BIN_DIR = bin/

.PHONY build:
	go build -o "$(BIN_DIR)$(APP_NAME)"
