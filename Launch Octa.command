#!/bin/zsh
set -e
cd -- "$(dirname -- "$0")"
exec ./bin/run "$@"
