#!/bin/zsh
# Screenshot structures without disturbing the desktop: Godot is launched hidden
# (open -j) and in the background (open -g), so it never takes focus. Godot's real
# --headless mode has no GPU renderer, so it cannot produce images.
#   tools/capture.sh [--shapes=cube,castle] [--views=hero,front,side,top]
#                    [--modes=solid,xray] [--at=0.3,0.6] [--gap=0] [--out=captures]
# Output: captures/<name>-<view>.png; log in captures/capture.log.
cd -- "$(dirname -- "$0")/.."
APP=${GODOT_APP:-/Applications/Godot.app}
mkdir -p captures
: > captures/capture.log
open -g -j -n -W -a "$APP" --stdout "$PWD/captures/capture.log" --stderr "$PWD/captures/capture.log" \
  --args --path "$PWD" --resolution 1900x1000 --position 4000,4000 --script tools/capture.gd -- "$@"
grep -E "Showing|Captured|Bench|ERROR|WARNING" captures/capture.log
! grep -q "ERROR" captures/capture.log
