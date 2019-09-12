APP_NAME = ddns
BUILD_DIR = build/

.PHONY: all
all: get build

.PHONY: get
get:
	go get

.PHONY: build
build:
	go build -o "$(BUILD_DIR)$(APP_NAME)"
	cp config.dist.yaml "$(BUILD_DIR)"
