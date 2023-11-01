#!/bin/bash

set -euo pipefail

DIR="$1"
if command -v golangci-lint >/dev/null 2>&1; then
  echo "Running golangci-lint on $DIR/*.go"
  if [ -n "${CI:-}" ]; then
    golangci-lint run -c "$DIR/.golangci.yml"
  else
    golangci-lint run -c "$DIR/.golangci.yml" --fix
  fi
else
  echo "golangci-lint is not installed. Skipping shell script formatting."
  echo "Follow instructions here to install: https://golangci-lint.run/usage/install/"
  echo "or use 'devbox shell' (https://www.jetpack.io/devbox/docs/quickstart/)"
  if [ -n "$CI" ]; then
    exit 1
  fi
fi
