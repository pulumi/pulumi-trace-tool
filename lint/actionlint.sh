#!/bin/bash

set -euo pipefail

DIR="$1/.github/workflows"
if command -v actionlint >/dev/null 2>&1; then
  echo "Running actionlint on $DIR/*.(yaml|yml)"
  find "$DIR" -type f -name '*.yml' -exec actionlint {} +
  find "$DIR" -type f -name '*.yaml' -exec actionlint {} +
else
  echo "actionlint is not installed. Skipping shell script formatting."
  echo "Follow instructions here to install: https://github.com/rhysd/actionlint/blob/main/docs/install.md"
  echo "or use 'devbox shell' (https://www.jetpack.io/devbox/docs/quickstart/)"
  if [ -n "$CI" ]; then
    exit 1
  fi
fi
