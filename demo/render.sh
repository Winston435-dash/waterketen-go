#!/bin/bash
# Render de demo's: VHS maakt de mp4's, daarna worden de GIF's eruit gebouwd.
#
#   cd demo && ./render.sh
#
# Nodig: vhs, ttyd, ffmpeg en een Chromium op de PATH.
#
# Waarom de GIF niet rechtstreeks uit VHS komt: VHS's .gif-pad gebruikt
# split -> palettegen -> paletteuse, en dat buffert de HELE rgba-stream in RAM
# (een minuut op 1000x700 is al ~2 GB). De tweetraps palettegen hieronder schrijft
# het palet naar een PNG en stroomt daarna, dus het geheugengebruik blijft klein.
set -euo pipefail
cd "$(dirname "$0")"

vhs waterketen-tui.tape
vhs waterketen-tui-mobiel.tape

mk_gif () { # in.mp4 out.gif breedte
  local pal
  pal=$(mktemp --suffix=.png)
  ffmpeg -v error -y -i "$1" -vf "fps=20,scale=$3:-2:flags=lanczos,palettegen=stats_mode=diff" "$pal"
  ffmpeg -v error -y -i "$1" -i "$pal" \
    -lavfi "fps=20,scale=$3:-2:flags=lanczos[x];[x][1:v]paletteuse=dither=sierra2_4a" \
    -loop 0 "$2"
  rm -f "$pal"
}

mk_gif waterketen-tui.mp4        waterketen-tui.gif        900
mk_gif waterketen-tui-mobiel.mp4 waterketen-tui-mobiel.gif 420

ls -la ./*.mp4 ./*.gif
