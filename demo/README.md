# Demo's

Met [VHS](https://github.com/charmbracelet/vhs) zijn de TUI-demo's code: de
tape-bestanden staan hier als bron, de renders horen ernaast te staan.

```bash
go build -o waterketen-tui ./cmd/waterketen-tui
cd demo && ./render.sh
```

`render.sh` draait VHS voor beide tapes en bouwt daarna de GIF's uit de mp4's.
Los draaien kan ook, maar **vanuit deze map** (het pad in `Output` is relatief):

```bash
cd demo
vhs waterketen-tui.tape           # 1000x700 -> 97x36 cellen, breed scherm
vhs waterketen-tui-mobiel.tape    # 460x620  -> 44x31 cellen, telefoonformaat
```

Nodig: `vhs`, `ttyd`, `ffmpeg` en een Chromium.

Wat de demo's laten zien:

- `waterketen-tui.tape` — alle acht panelen, `j`/`k`, de catalogus met `[`/`]`
  en `t`, een echte bronverversing (`v`) met voortgangsbalk, de kaart (zoomen,
  `g` gebied, `/` straat zoeken, `t` lagen, `o` solo/samen met `[`/`]`, `f`
  passend), een offline telling op wijk Delfshaven in Query (`8`, `g`, `Enter`)
  en beide opdrachtenpaletten (`ctrl+p` en `?`).
- `waterketen-tui-mobiel.tape` — dezelfde app op 44 kolommen: de zijbalk wijkt
  voor de kaart, de paneeltoetsen staan in de header en Ontdek krijgt zijn eigen
  leesweergave.

## Waarom het is zoals het is (alle drie gemeten)

1. **Het zijn twee tapes, geen resize.** VHS kan het venster niet halverwege een
   tape veranderen: na een `Set Width`/`Set Height` halverwege bleef de uitvoer
   even groot (1000x700), en beide `stty size`-probes gaven 97x36.
2. **Geen `Set WindowBar`, geen `Set BorderRadius`.** Elk daarvan voegt een
   `loop=-1`-input plus een overlay/alphamerge-stap aan de ffmpeg-filtergraph
   toe. Met die twee erin deed dezelfde tape er 20+ minuten over en groeide
   ffmpeg naar 2,2 GB (OOM-killer); zonder die twee is het een kwestie van
   seconden en blijft de piek onder 1,2 GB. Een strakke raamrand is dat niet
   waard.
3. **De GIF komt uit de mp4, niet rechtstreeks uit VHS.** VHS's `.gif`-pad
   gebruikt `split -> palettegen -> paletteuse` en dat buffert de HELE
   rgba-stream in RAM (een minuut op 1000x700 is al ~2 GB). De tweetraps
   `palettegen` naar een PNG in `render.sh` stroomt en blijft klein.

## Op een kleine machine

Rendert VHS zwaar of word je zelf door de OOM-killer gepakt, begrens het in een
eigen cgroup in plaats van in die van je shell/desktop — zo blijft de klap
binnen de render:

```bash
systemd-run --user --unit=wkdemo --collect \
  -p MemoryMax=1500M -p MemorySwapMax=0 \
  -p Environment=PATH="$HOME/.local/bin:$PATH" \
  -p Environment=VHS_NO_SANDBOX=1 \
  bash -c 'cd demo && ./render.sh'
```

Twee valkuilen daarbij: een transient unit erft je login-`PATH` **niet** (zet
`vhs`/`ttyd`/`ffmpeg`/Chromium expliciet op de `PATH`), en op een machine met
snap-chromium faalt VHS met `browser exited unexpectedly` — snap kan de
tijdelijke `--user-data-dir` van VHS niet lezen. Zet dan een echte Chrome eerder
op de `PATH` en draai met `VHS_NO_SANDBOX=1`.

## Wat de demo eerlijk laat zien

De tapes lezen de echte managed data (`~/.local/share/waterketen`) en doen bij
het starten echte netwerkcontroles. Een langzame of platte bron verandert dus
zichtbaar het beeld — dat is de ware toestand, geen gefingeerde. De kaart toont
echte objecten uit de snapshot van Rotterdam; de aantallen in Analyse zijn de
gemeten aantallen.
