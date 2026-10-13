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
  - `screens/` — acht panelen: Ontdek, Bronnen, Lagen, Gebieden, Analyse,
    Kaart, Profiel en Query. Bestaande paneeltoetsen 1–6 blijven gelijk;
    Ontdek is 7 en Query is 8.

De TUI schildert een expliciete **RGB-zwarte achtergrond (`#000000`)** over
header, panelen, lege cellen en commandobar. De achtergrond is niet afhankelijk
van de terminalkleur of diens configureerbare ANSI-zwart; kaartkleuren blijven
ongewijzigd. Voor tekst gebruikt de TUI altijd de donkere themavariant.

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

- ✅ Acht panelen in de TUI, met Ontdek als vraaggestuurde startpagina.
- ✅ **Offline gebiedsquery.** Paneel 8 telt de gekozen snapshotlagen tegen de
  volledige CBS-polygon, inclusief gaten en randintersecties. Gebiedskiezer,
  laagvinkjes en objectdetails; paginering begrenst nooit de telling.
- ✅ CLI: `ontdek`, `sources`, `status`, `analyse`, `braille`.
- ✅ GeoPackage-lezing **zonder cgo** (pure-Go `modernc.org/sqlite` + eigen
  GeoPackageBinary/WKB-decoder).
- ✅ Live WFS/ArcGIS-probes (hits-only) via `net/http`.
- ✅ Expliciete IMWA-import: duikers, gemalen, sluizen en stuwen via PDOK,
  geclipt op de Rotterdamse gemeentegrens en gepubliceerd als nieuwe snapshot.
- ✅ **Bronnenregister met licentie per bron.** Zeventien bedrade bronlagen (GWSW,
  DAMO, Vewin, RIONED-OGC, Zuid-Holland, IMWA en RWS) met hun eigen
  `TypeName` en `License`, plus zes gedetecteerde-niet-bedrade bronnen met
  reden. Licentie is een beslissing, geen detail: CC0, CC BY-SA 4.0 en publiek
  domein zijn verschillende rechten, en een dienst zonder licentieveld zegt
  "onbekend". `go test` faalt als een bron die velden mist.
  Paneel 1 scrollt met `↑`/`↓`, `PgUp`/`PgDown`, `Home` en `End`, zodat alle
  bronnen en licenties ook op een kleine terminal bereikbaar blijven.
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
  snapshot en aanvullen met IMWA-kunstwerken is nu wel native.
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

## Ontdek: eerst je vraag, daarna de techniek

De TUI opent met **Ontdek** (toets `7`). Kies een vraag met `↑`/`↓` en druk
`Enter` om te lezen. Op een brede terminal staat het antwoord ook naast de lijst;
op een telefoon krijgt het antwoord een eigen leesweergave. `Esc` brengt je terug.
Met `[`/`]` wissel je tussen **Vragen**, **Bronnen**, **Lagen** en **Begrippen**.
Tijdens lezen scrollen `↑`/`↓`, `PgUp`/`PgDown`, `Home` en `End`; `t` schakelt de
technische verdieping in of uit. Geen animaties of downloads in deze catalogus.
De bestaande app-startup blijft bronnen controleren; dat is geen catalogusactie.

```bash
./waterketen ontdek
./waterketen ontdek --vraag drinkwater
./waterketen ontdek --laag riool --technisch
./waterketen ontdek --overzicht bronnen
./waterketen ontdek --bron damo_leiding --technisch
./waterketen ontdek --begrip bob
./waterketen ontdek --json
./waterketen ontdek --markdown --technisch
```

Officiële namen blijven naast de uitleg staan. Iedere laag vertelt wat het is,
wat de kaartvorm betekent, welke vraag je ermee beantwoordt en wat je er **niet**
uit kunt afleiden. Herkomst, licentie, lokale bestandstatus en actualiteit staan
apart: metadata-aantallen zijn geen bewijs van een compleet fysiek netwerk.
BOB is niet automatisch straatdiepte; een verzorgingsgebied is geen leidingnet;
de gemeentelijke Evides-levering is gedeeltelijk. Onbekende rechten of
brondatums worden niet ingevuld met aannames. Tijdstempels worden in
Amsterdam-tijd getoond; JSON bewaart de oorspronkelijke waarden.

API-adressen en de volledige voorbeeldaanvraag staan achter `t` / `--technisch`,
met uitleg per parameter. De catalogus voert deze aanvragen niet uit. Zij
beschrijft het bestaande bronnenregister en bekende lagen, **niet alle lagen van
elke externe dienst**. Een bron die hier uitgelegd wordt, is niet automatisch
bedraad, gedownload of op de kaart zichtbaar.

De inhoud woont uitsluitend in `internal/core/catalog*.go`. TUI, CLI, JSON en
Markdown gebruiken dezelfde catalogus en dezelfde uitlegfunctie. De
[gegenereerde catalogus](docs/catalogus.md) bevat ook API-voorbeelden. Regenereer
zonder lokale snapshotinformatie met een lege datastore:

```bash
empty=$(mktemp -d "${TMPDIR:-$HOME/.hermes/cache/scratch}/waterketen-docs.XXXXXX")
./waterketen ontdek --markdown --technisch --data-root "$empty" > docs/catalogus.md
rmdir "$empty"
```

## Lagen en gebieden: één lokale inventaris

- **2 Lagen** gebruikt dezelfde catalogus als **7 Ontdek**: alle beschreven lagen,
  alle managed leveringen en alle snapshotlagen, ook nieuwe laagnamen zonder
  aparte uitleg. Groen betekent **bestand aanwezig**, niet inhoudelijk gevalideerd
  of volledig. Ontbrekende bestanden en alleen beschreven bronlagen blijven apart
  zichtbaar. `l` wisselt tussen alle lagen en alleen lagen met een lokaal bestand.
- **3 Gebieden** leest daadwerkelijk `areas.gpkg`: gemeente, wijken en buurten
  met CBS-code. Het aantal in de titel is het gelezen aantal; een afwijking van het
  manifest wordt gemeld. Zonder CBS-register zijn alleen managed gebiedsnamen
  zichtbaar, niet verzonnen officiële grenzen.
- Beide overzichten scrollen met `↑`/`↓`, `PgUp`/`PgDown`, `Home` en `End`.
  `/` opent het naam-/codefilter; `Enter` past toe, `Esc` annuleert tijdens typen
  en wist daarbuiten het filter.
- **4 Analyse** scheidt berekeningen over managed gebiedsbestanden van
  **snapshotmetadata**. Snapshot-aantallen zijn geen geometrieanalyse; voor een
  exacte gebiedstelling gebruik je **8 Query**. **6 Profiel** toont alleen managed
  BOB-lagen en benoemt buishoogte in NAP, niet diepte onder straatmaaiveld.
- **5 Kaart** en **8 Query** herkennen nieuwe gepubliceerde vectorlagen zonder
  uitbreiding van een namen-allowlist. Bestaande kaartkleuren/tekenvolgorde blijven
  gelijk; nieuwe lagen krijgen neutrale data-ink en een geometriepassende legenda.

Een broncontrole downloadt geen dataset. Nieuwe bronlagen moeten expliciet door
de producer worden opgehaald en in een snapshot opgenomen voordat Kaart en Query
ze lokaal kunnen gebruiken. Alternatieve leveringen worden niet opgeteld als
extra fysiek netwerk.

### IMWA-kunstwerken op de kaart

Met een bestaande kaartbibliotheek in RD (`EPSG:28992`) en `ogr2ogr` (GDAL):

```bash
go run ./cmd/waterketen bibliotheek imwa
# Of met de geïnstalleerde Go-CLI:
waterketen-go bibliotheek imwa
```

Dit haalt **duikers, gemalen, sluizen en stuwen** van Het Waterschapshuis via
PDOK op binnen de Rotterdamse bbox, met gecontroleerde WFS-paginering en behoud
van bronattributen en echte punt-/lijn-/vlakgeometrie. GDAL knipt op de volledige
CBS-gemeentegrens (`GM0599`), niet alleen de bbox. Brondatum blijft onbekend;
ophaaldatum, CC0-licentie en herkomst worden afzonderlijk bewaard.

Alle vier de lagen moeten geldig zijn voordat de nieuwe snapshot actief wordt.
De vorige snapshot blijft intact; bestaande vectorbestanden worden via hard links
hergebruikt, niet nogmaals gekopieerd. Bestaande manifestvelden blijven behouden.
De import ondersteunt ook `--data-root` en `--json`. Geen automatische download
bij TUI-start, broncontrole of openen van de catalogus.

**Herstart de TUI**, open **5 Kaart → t** en vink **IMWA — duikers**, **IMWA —
gemalen**, **IMWA — sluizen** en **IMWA — stuwen** aan met `Spatie`; bevestig met
`Enter`. Ze blijven standaard uit om de bestaande kaart niet te veranderen.
Ze staan ook lokaal beschikbaar in **2 Lagen** en **8 Query**. De IMWA-gemalen
zijn een alternatieve registratie, niet extra gemeentelijke/GWSW-gemalen.

## Query: gebied kiezen, lagen tellen

Open paneel **8 Query**. `g` kiest een gemeente, wijk of buurt: zoek op naam
of CBS-code en bevestig met `Enter`. Wijk Delfshaven (`WK059903`) en buurt
Delfshaven (`BU05990320`) zijn verschillende gebieden.

Met `l` kies je lagen; `Spatie` schakelt een vinkje en `Enter` past de keuze
toe. `Esc` annuleert de selectie. Druk daarna `Enter` om offline te tellen.
Selecteer een resultaat met `↑`/`↓`; `Enter` opent de objecten en nogmaals
`Enter` de bronattributen. `PgUp`/`PgDown`, `Home` en `End` navigeren over alle
objecten, ook voorbij de eerste pagina van 50. `Esc` gaat terug of breekt
lopend querywerk af.

De query telt bronobjecten die de volledige gebiedsgrens raken, niet het
rechthoekige kaartvak en niet een kaartsample. **Rioolgemalen** filtert de
gemengde GWSW-pomplaag op het expliciete type `Rioolgemaal`; de ongefilterde
laag en oppervlaktewatergemalen blijven aparte keuzes. Tel alternatieve
leveringen niet bij elkaar op. Een nul is een succesvolle lege selectie;
`—` met een melding betekent niet beschikbaar of een leesfout.

Snapshotdatum, brondatum en licentie blijven afzonderlijk zichtbaar.
**Registratie ≠ volledig fysiek netwerk.** De query zelf downloadt niets;
de bestaande broncontroles bij het starten van de app blijven ongewijzigd.

## Mobiel (Termux / smalle terminals)

De TUI is bruikbaar op een telefoon. Beneden ~76 kolommen of ~24 rijen wijkt de
zijbalk voor de kaart — die is de inhoud, de zijbalk is comfort — en wat de
zijbalk normaal vertelt, verhuist naar de twee vrije regels:

- **header**: `waterketen — 4 Analyse ‹ 5 Kaart › 6 Profiel` (de paneeltoetsen,
  want de zijbalk wás de legenda)
- **paneltitel**: `Kaart · GM0599 · samen` (wat `Info()` op een breed scherm in de
  zijbalk zet)
- **commandobar**: alleen de toetsen, geen beschrijvingen (`← · → · ↑ · ↓ · + · - · f · [ · ] · o · l · /`)

Drie regels die dit betaalbaar maakten, alle drie uit een echte capture:

1. `components.SingleLine` is de ENIGE manier om een vaste één-regel-slot te
   vullen (header, commandobar). `lipgloss.Width()` vult aan maar kapt niet af, en
   `Padding` komt er bovenop — dus `Width(n).Padding(0,1)` rendert n+2 cellen en de
   terminal wikkelt het naar een tweede rij. Gemeten op 30x12: header=2, bar=2,
   totaal 14 rijen op een scherm van 12. Afkappen eindigt op `…`, zodat een
   afgekapte CBS-code niet als typefout leest.
2. `Model.layout()` is de ENIGE plek die de schermgeometrie bepaalt. `panelBox()`
   en `renderBody()` lezen allebei die struct, zodat de box die een scherm
   *krijgt* niet kan afwijken van de box waarin het *getekend* wordt. Dat verschil
   was de mobiele bug: `panelBox` trok de zijbalk af vanaf 44 kolommen, terwijl
   `renderBody` hem pas onder 80 liet vallen — dus op 72 kolommen tekende de kaart
   48 breed in een paneel van 72 en bleef een derde van het scherm leeg.
3. Overlays klemmen hun eigen dialoogbreedte. Een dialoog breder dan het paneel
   laat lipgloss de PANEELRAND wikkelen in plaats van de dialoog, en een te hoge
   dialoog duwt de hint van het scherm. Beide gemeten op 48 kolommen.

De regressietests staan in `internal/app/layout_test.go` (geometrie op acht
formaten) en `internal/ui/screens/overlay_mobile_test.go` (overlays op zes
formaten).

Je kunt ook meteen installeren:

```bash
install -Dm755 waterketen     ~/.local/bin/waterketen-go
install -Dm755 waterketen-tui ~/.local/bin/waterketen-tui-go
```

De binaire namen bewust **anders** dan de Python-CLI (`waterketen` /
`waterketen-tui` blijven van de Python). Pas na feature-pariteit wisselen.

## Tests

```bash
gofmt -l . && go build ./... && go vet ./... && go test -count=1 ./...
go test -race -count=1 ./...
```

De standaardtests gebruiken kleine, deterministische SQLite-/R-tree-fixtures en
lokale HTTP-testservers. Ze vereisen geen managed datastore en benaderen geen
externe diensten. CI controleert formatting, build, vet, tests, races en WKB-fuzzing.

De echte managed-data-regressies zijn expliciet opt-in. Ze lezen `core.DataRoot()`
(`$WATERKETEN_DATA_DIR`, anders `~/.local/share/waterketen`) en pinnen metingen:
4.610 buizen, 4.461 gemeten BOB's (minstens één uiteinde), 342 straten,
12 materialen en de 115 CBS-gebieden. Een ontbrekende datastore is dan een fout,
niet een stil overgeslagen controle.

```bash
WATERKETEN_INTEGRATION=1 go test -count=1 ./internal/core
# Apart: echte externe diensten controleren (kan falen door storing/netwerk).
WATERKETEN_LIVE_TESTS=1 go test -run '^TestLiveSources$' -count=1 ./internal/core
# Parser-fuzzing: tellingen worden vóór allocatie gevalideerd; nesting is begrensd.
go test ./internal/core -run '^$' -fuzz '^FuzzWKB$' -fuzztime=30s -parallel=2
go test ./internal/core -run '^$' -fuzz '^FuzzGeoPackageBinary$' -fuzztime=30s -parallel=2
```

CLI en TUI gebruiken dezelfde laagloader en samenvatting in
`internal/core/analysis.go`. Analyse cachet zijn rapport totdat de data verandert;
leesfouten blijven naast de leesbare lagen zichtbaar. Een lege kaartselectie of
uitsnede blijft een geldige kaart, zonder impliciet opnieuw netwerken aan te zetten.
