ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

lint:
	"${ROOT_DIR}lint/shellcheck.sh" "${ROOT_DIR}"
	"${ROOT_DIR}lint/actionlint.sh" "${ROOT_DIR}"
	"${ROOT_DIR}lint/golangci.sh" "${ROOT_DIR}"

test:
	go test -v ./...

build:
	go build

.PHONY: lint test
