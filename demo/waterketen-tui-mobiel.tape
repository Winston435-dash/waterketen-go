# waterketen-tui — demo op telefoonformaat (VHS)
#
#   go build -o waterketen-tui ./cmd/waterketen-tui
#   vhs demo/waterketen-tui-mobiel.tape
#
# Zelfde app, smal venster (44x31 op 460x620): de zijbalk wijkt voor de kaart,
# de paneeltoetsen verhuizen naar de header en Ontdek krijgt een eigen
# leesweergave. Dit staat in een aparte tape omdat VHS het venster niet
# halverwege een tape kan veranderen (gemeten: de GIF bleef 1100x800).

Output waterketen-tui-mobiel.mp4

Set Shell "bash"
Set FontSize 16
Set Width 460
Set Height 620
Set Padding 12
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms
Set PlaybackSpeed 1.25

Type "cd ~/projects/waterketen-go && ./waterketen-tui"
Enter
Sleep 8s

# --- Ontdek (7): op een smal scherm krijgt het antwoord een leesweergave -----
Sleep 2s
Down
Sleep 1s
Enter
Sleep 3s
PageDown
Sleep 3s
Escape
Sleep 2s
Type "]"
Sleep 3s
Type "["
Sleep 1s

# --- Analyse (4) ------------------------------------------------------------
Type "4"
Sleep 2s
Down
Sleep 1s
PageDown
Sleep 2s

# --- Kaart (5): inzoomen op een wijk ---------------------------------------
Type "5"
Sleep 7s
Type "g"
Sleep 2s
Type "delfshaven"
Sleep 1s
Enter
Sleep 6s
Type "+"
Sleep 3s

# lagen erbij: putten en gemalen
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
Sleep 5s

# straat zoeken
Type "/"
Sleep 1s
Type "westzeedijk"
Sleep 1s
Enter
Sleep 5s

Type "q"
Sleep 2s
