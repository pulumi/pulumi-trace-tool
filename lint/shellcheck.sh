#!/bin/bash

set -euo pipefail

DIR="$1"
if command -v shellcheck >/dev/null 2>&1; then
  if ! find "$DIR" -type f -name '*.sh' -exec false {} + >/dev/null 2>&1; then
    echo "Running shellcheck on $DIR/**/*.sh"
    find "$DIR" -type f -name '*.sh' -exec shellcheck -s bash {} +
  fi
  if [ -f "$DIR"/Makefile ]; then
    echo "Running shellcheck on $DIR/Makefile"
    # We unset Make related variables so that we can check the Makefile's dry-run output without
    # Make thinking it's running recursively.
    (
      unset MAKELEVEL MAKEFLAGS MFLAGS
      make -f "$DIR"/Makefile -n | shellcheck -s bash -
    )
  fi
else
  echo "shellcheck is not installed. Skipping shell script linting."
  echo "Follow instructions here to install: https://github.com/koalaman/shellcheck#installing"
  echo "or use 'devbox shell' (https://www.jetpack.io/devbox/docs/quickstart/)"
  if [ -n "$CI" ]; then
    exit 1
  fi
fi
