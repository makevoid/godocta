#!/bin/zsh
# Record the one-minute UI walkthrough without disturbing the desktop: Godot runs hidden
# (open -j) in the background (open -g) and streams frames to ffmpeg.
#   tools/walkthrough.sh [--out=captures/walkthrough.mp4] [--fps=30] [--fast=2] [--stills=DIR]
# Log: captures/walkthrough.log. Needs ffmpeg (Homebrew).
cd -- "$(dirname -- "$0")/.."
APP=${GODOT_APP:-/Applications/Godot.app}
LOG=${WALKTHROUGH_LOG:-$PWD/captures/walkthrough.log}
FPS=30
for arg in "$@"; do [[ $arg = --fps=* ]] && FPS=${arg#--fps=}; done
mkdir -p captures "${LOG:h}"
: > "$LOG"
open -g -j -n -W -a "$APP" --stdout "$LOG" --stderr "$LOG" \
  --args --path "$PWD" --resolution 640x360 --position 4000,4000 --fixed-fps "$FPS" \
  --script tools/walkthrough.gd -- "$@" || exit $?
grep -E "Walkthrough|ERROR|WARNING" "$LOG"
! grep -q "ERROR" "$LOG"
