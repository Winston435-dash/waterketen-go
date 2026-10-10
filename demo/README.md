# Demo's

Met [VHS](https://github.com/charmbracelet/vhs) render je de TUI als GIF: de
tape-bestanden staan hier als bron, de GIF's horen ernaast te staan.

```bash
go build -o waterketen-tui ./cmd/waterketen-tui
vhs demo/waterketen-tui.tape          # 1100x800, breed scherm
vhs demo/waterketen-tui-mobiel.tape   # 460x620, telefoonformaat
```

Nodig: `vhs`, `ttyd`, `ffmpeg` en een Chromium. Op een headless machine met
snap-chromium faalt VHS (`browser exited unexpectedly`): snap kan de tijdelijke
`--user-data-dir` van VHS niet lezen. Zet dan een echte Chrome eerder in `PATH`
en draai met `VHS_NO_SANDBOX=1`.

Wat de demo's laten zien:

- `waterketen-tui.tape` — alle zeven panelen, `j`/`k`, de catalogus met `[`/`]`
  en `t`, een echte bronverversing (`v`) met voortgangsbalk, de kaart
  (zoomen, `g` gebied, `/` straat, `t` lagen, `o` solo/samen, `f`), en beide
  opdrachtenpaletten (`ctrl+p`, `?`).
- `waterketen-tui-mobiel.tape` — dezelfde app op 44 kolommen: de zijbalk wijkt
  voor de kaart, de paneeltoetsen staan in de header en Ontdek krijgt zijn eigen
  leesweergave.

VHS kan het venster **niet** halverwege een tape veranderen; daarom zijn het twee
tapes in plaats van één met een resize. Gemeten: de GIF bleef 1100x800 nadat
`Set Width`/`Set Height` halverwege wijzigde.

De demo's lezen de echte managed data (`~/.local/share/waterketen`) en doen bij
het starten echte netwerkcontroles. Een langzame of platte bron verandert dus
zichtbaar het beeld — dat is de ware toestand, geen gefingeerde.
