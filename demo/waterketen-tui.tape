# waterketen-tui — demo van alle functionaliteit (VHS)
#
#   go build -o waterketen-tui ./cmd/waterketen-tui   # eerst de binary
#   vhs demo/waterketen-tui.tape                      # vanuit de repo-root
#
# VHS heeft ttyd + ffmpeg + Chromium nodig. De demo leest de echte managed data
# op ~/.local/share/waterketen en doet bij het starten echte broncontroles.
#
# Op 1100x800 met FontSize 16 is het venster 107x41 — breed genoeg voor de
# zijbalk. De smalle (telefoon)layout staat in demo/waterketen-tui-mobiel.tape:
# VHS kan het venster NIET halverwege een tape veranderen (gemeten: de GIF
# bleef 1100x800 nadat Set Width/Height halverwege wijzigde).
#
# Afdrukbare toetsen gaan door Type "..." — een kale `]` is geen VHS-commando.

Output waterketen-tui.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 800
Set Padding 16
Set BorderRadius 10
Set WindowBar Colorful
Set WindowBarSize 28
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms
Set PlaybackSpeed 1.25

Type "cd ~/projects/waterketen-go && ./waterketen-tui"
Enter
Sleep 8s

# --- Ontdek (7): de vraaggestuurde catalogus, het startscherm ---------------
Sleep 2s
Down
Sleep 1s
Down
Sleep 1s
Enter
Sleep 3s
PageDown
Sleep 2s
Escape
Sleep 1s
Type "]"
Sleep 3s
Type "]"
Sleep 3s
Type "t"
Sleep 3s
Type "t"
Sleep 1s
Type "["
Sleep 1s

# --- Bronnen (1): live broncontrole met echte voortgangsbalk -----------------
Type "1"
Sleep 3s
Type "v"
Sleep 6s

# --- Lagen (2) en Gebieden (3) ----------------------------------------------
Type "2"
Sleep 3s
Type "3"
Sleep 3s

# --- Analyse (4): echte metingen, scrollbaar ---------------------------------
Type "4"
Sleep 2s
Down
Sleep 500ms
Down
Sleep 500ms
PageDown
Sleep 2s

# --- Kaart (5): braille-atlas van heel Rotterdam -----------------------------
Type "5"
Sleep 7s
Type "+"
Sleep 2s
Type "+"
Sleep 2s
Type "g"
Sleep 2s
Type "delfshaven"
Sleep 1s
Enter
Sleep 6s
Type "/"
Sleep 1s
Type "westzeedijk"
Sleep 1s
Enter
Sleep 5s
Type "l"
Sleep 3s
Type "l"
Sleep 1s

# lagen aan/uit: putten en gemalen erbij (spatie schakelt, enter past toe)
Type "t"
Sleep 2s
Down
Down
Down
Down
Down
Down
Down
Space
Down
Space
Enter
Sleep 6s

# solo-laag, laag wisselen, weer samen, dan passend maken
Type "o"
Sleep 4s
Type "]"
Sleep 4s
Type "["
Sleep 1s
Type "o"
Sleep 3s
Type "f"
Sleep 3s

# --- Profiel (6): dwarsdoorsnede, laag per laag ------------------------------
Type "6"
Sleep 3s
Type "]"
Sleep 3s
Type "]"
Sleep 3s

# --- Opdrachten: ctrl+p (globaal) en ? (paneel) ------------------------------
Ctrl+P
Sleep 3s
Escape
Sleep 1s
Type "?"
Sleep 3s
Escape
Sleep 1s

# --- j/k wisselt van paneel --------------------------------------------------
Type "k"
Sleep 2s
Type "k"
Sleep 2s

Type "q"
Sleep 2s
