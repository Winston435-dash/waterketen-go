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
Set PlaybackSpeed 1.5

Type "cd ~/projects/waterketen-go && ./waterketen-tui"
Enter
Sleep 7s

# --- Ontdek (7): de vraaggestuurde catalogus, het startscherm ---------------
Sleep 1500ms
Down
Sleep 800ms
Down
Sleep 800ms
Enter
Sleep 2500ms
PageDown
Sleep 1500ms
Escape
Sleep 800ms
Type "]"
Sleep 2s
Type "]"
Sleep 2s
Type "t"
Sleep 2500ms
Type "t"
Sleep 800ms
Type "["
Sleep 800ms

# --- Bronnen (1): live broncontrole met echte voortgangsbalk -----------------
Type "1"
Sleep 2s
Type "v"
Sleep 5s

# --- Lagen (2) en Gebieden (3) ----------------------------------------------
Type "2"
Sleep 2s
Type "3"
Sleep 2s

# --- Analyse (4): echte metingen, scrollbaar ---------------------------------
Type "4"
Sleep 1500ms
Down
Sleep 400ms
Down
Sleep 400ms
PageDown
Sleep 1500ms

# --- Kaart (5): braille-atlas van heel Rotterdam -----------------------------
Type "5"
Sleep 6s
Type "+"
Sleep 1500ms
Type "+"
Sleep 1500ms

# gebied kiezen: Delfshaven
Type "g"
Sleep 1500ms
Type "delfshaven"
Sleep 1s
Enter
Sleep 5s

# straat zoeken
Type "/"
Sleep 1s
Type "westzeedijk"
Sleep 1s
Enter
Sleep 4s

# straatnamen aan/uit
Type "l"
Sleep 2s
Type "l"
Sleep 800ms

# lagen aan/uit: putten en gemalen erbij (spatie schakelt, enter past toe)
Type "t"
Sleep 1500ms
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
Sleep 5s

# solo-laag (o): riool → drinkwaterleidingen → straten, dan weer samen
Type "o"
Sleep 3s
Type "]"
Sleep 3s
Type "["
Sleep 1s
Type "o"
Sleep 2s
Type "f"
Sleep 2500ms

# --- Profiel (6): dwarsdoorsnede, laag per laag ------------------------------
Type "6"
Sleep 2s
Type "]"
Sleep 2s
Type "]"
Sleep 2s

# --- Opdrachten: ctrl+p (globaal) en ? (paneel) ------------------------------
Ctrl+P
Sleep 2500ms
Escape
Sleep 800ms
Type "?"
Sleep 2500ms
Escape
Sleep 800ms

# --- j/k wisselt van paneel --------------------------------------------------
Type "k"
Sleep 1500ms
Type "k"
Sleep 1500ms

Type "q"
Sleep 1500ms
