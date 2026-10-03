#!/bin/zsh
set -e
cd -- "$(dirname -- "$0")"
if command -v godot >/dev/null 2>&1; then
  exec godot --path "$PWD"
elif [[ -x /Applications/Godot.app/Contents/MacOS/Godot ]]; then
  exec /Applications/Godot.app/Contents/MacOS/Godot --path "$PWD"
else
  echo "Install Godot 4.4 or newer, then open project.godot and press F6 or F5."
  read -r "?Press Return to close."
fi
