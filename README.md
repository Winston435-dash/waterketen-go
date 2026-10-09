# waterketen (Go)

Betrouwbare TUI en CLI voor Rotterdamse waterketen-geodata — volledige
rewrite van `~/projects/waterketen-cli/` (Python/Textual) naar Go met de
[Bubble Tea](https://github.com/charmbracelet/bubbletea)-stack.

## Waarom Go

Het oorspronkelijke project bleek een TUI-design-systeem nodig te hebben
waarin een agent niet elke scherm vanaf nul opnieuw ontwerpt. In plaats van
opnieuw te pappen op een Textual-basis met 950 regels app-logica zonder
tokens, is dit een nieuwe codebase rond één regel: **schermen mogen zelf
nooit randen, kleuren of spacing definiëren**.

## Stack

- **Bubble Tea** — applicatie-architectuur (Elm-model).
- **Bubbles** — hergebruikte componenten: `help`, `key`, `table`, `viewport`.
- **Lip Gloss** — styling, via `internal/ui/theme.go` als token-bibliotheek.
- **Eigen `ui/` kit** — project-design-systeem:
  - `theme.go` — kleur/spacing/border-tokens.
  - `keymap.go` — alle toetsen.
  - `components/` — `Panel`, `CommandBar`, `StatusBadge`, `EmptyState`,
    `DataTable`, `Modal`.
  - `screens/` — de zes panelen (Bronnen, Lagen, Gebieden, Analyse, Kaart,
    Profiel).

## Regels voor uitbreidingen (de contracten)

1. **Nooit** `lipgloss.NewStyle()` met een hex of AdaptiveColor in
   `internal/ui/screens/` of `cmd/` schrijven. Gebruik
   `ui/components.Panel(...)`, `...StatusBadge(...)`, etc. Dit is
   **afgedwongen**: `internal/ui/screens/contract_test.go` faalt als een scherm
   `lipgloss` importeert.
2. **Gebruik alleen thema-tokens**: `ui.ColorPrimary`, `ui.SpacingSM`,
   `ui.BorderPanel`. Nieuwe kleuren of spacing horen in `internal/ui/theme.go`
   thuis als *named* token.
3. **Elke scherm volgt de vaste anatomie**: `Header → Content → CommandBar`.
   De `internal/app.Model` regelt dit; screens leveren alleen hun `View()`.
4. **Core is puur**: `internal/core/` mag NOOIT `lipgloss`, `bubbletea` of
   `bubbles` importeren. Strings voor het scherm komen uit
   `internal/ui/screens/` (of in de toekomst een `report.go`). Core bezit wél de
   **data-ink**: `core.CellColour` is de enige kleurtabel voor de kaart;
   `ui/components.RenderAtlas` zet die om naar lipgloss en voegt alleen
   bold/underline toe. De kaart rendert daardoor byte-identiek na de refactor.
5. **Sidebar is de legenda** voor paneel-keys. De `CommandBar` herhaalt
   alleen wat het actuele scherm zelf niet al zichtbaar maakt.

## Status

- ✅ Zes panelen in de TUI (Bronnen live, Kaart met atlas, Profiel, etc.).
- ✅ CLI: `sources`, `status`, `analyse`, `braille`.
- ✅ GeoPackage-lezing **zonder cgo** (pure-Go `modernc.org/sqlite` + eigen
  GeoPackageBinary/WKB-decoder).
- ✅ Live WFS/ArcGIS-probes (hits-only) via `net/http`.
- ✅ **Bronnenregister met licentie per bron.** Zes bedrade bronnen (GWSW, DAMO,
  Vewin, RIONED-OGC, Zuid-Holland drinkwater, RWS regiogebieden) met hun eigen
  `TypeName` en `License`, plus vijf gedetecteerde-niet-bedrade bronnen met
  reden. Licentie is een beslissing, geen detail: CC0, CC BY-SA 4.0 en publiek
  domein zijn verschillende rechten, en een dienst zonder licentieveld zegt
  "onbekend". `go test` faalt als een bron die velden mist.
- ✅ Renderer gebruikt dezelfde managed data als de Python-CLI
  (`~/.local/share/waterketen`).
- ✅ **Kaartbibliotheek in de TUI.** Paneel 5 leest de gepubliceerde snapshot
  rechtstreeks: net als in Python opent het op **heel Rotterdam** (`GM0599`,
  45,5 × 18,4 km) en `g` zoomt naar een wijk of buurt (115 CBS-gebieden).
  Viewports komen uit de R-tree-index van elke GeoPackage, dus een laag van
  87.146 rioolbuizen kost alleen zijn venster. `t` kiest de zichtbare lagen in
  DrawOrder. Wijkt bewust af van Python: **geen polygoon-clip en geen simplify**
  per viewport — de snapshot is bij de bouw al op Rotterdam geclipt, en de
  R-tree geeft de bbox. Waar dat tot sampling leidt (heel de stad op één scherm)
  zegt de kaart dat met "overzicht N/M kandidaten; zoom in voor alle objecten".
- ⏳ `fetch` (GWSW-paging + clip op CBS-grens) — nog in Python, zie
  `MIGRATION.md`.
- ⏳ `bibliotheek bouw` (osmium-import van PBF) — nog Python. **Lezen** van de
  snapshot is nu wel native.
- ⏳ `kaart` SVG/PNG-export — nog Python.

## Gebruik

```bash
cd ~/projects/waterketen-go
go build -o waterketen     ./cmd/waterketen
go build -o waterketen-tui ./cmd/waterketen-tui

./waterketen  sources --live
./waterketen  status
./waterketen  analyse --gebied delfshaven --laag riool
./waterketen-tui
```

Je kunt ook meteen installeren:

```bash
install -Dm755 waterketen     ~/.local/bin/waterketen-go
install -Dm755 waterketen-tui ~/.local/bin/waterketen-tui-go
```

De binaire namen bewust **anders** dan de Python-CLI (`waterketen` /
`waterketen-tui` blijven van de Python). Pas na feature-pariteit wisselen.

## Tests

```bash
go test ./...
```

De kern-tests lezen de echte managed data op
`~/.local/share/waterketen/processed/` en pinnen metingen (4.610 buizen,
4461 gemeten BOB's (beide uiteinden), 342 straten, 12 materialen, grenzen).
Dat maakt ze tot een echte regressievangnet: elke afwijking in WKB,
BOB-filtering of de render faalt direct.
