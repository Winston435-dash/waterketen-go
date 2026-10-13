# TUI-review en verbeterplan

Status: uitvoering van de zes geaccordeerde prioriteiten gestart.
Review: 13 oktober 2026. **Gewenste stack: Charm v2**, op verzoek van Winston.

## Uitvoeringslog

- [x] 1. Charm v2-stack: Bubble Tea 2.1.0, Bubbles 2.2.1, Lip Gloss 2.0.6,
  Glamour 2.0.1. Invoer-/paste-/muismessages, views, componentafmetingen,
  architectuurguards en tests gemigreerd. Build en tests groen.
  Echte kaartcaptures op vier formaten: alle zichtbare cellen, kleuren,
  nadruk, zwarte achtergrond en geometrie identiek. Op 30×12 ook byte-identiek;
  op de andere formaten codeert de v2-renderer de onzichtbare voorgrondkleur
  van spaties anders. De exacte diffs en `step1-verification.json` staan bij
  de scratch-artefacten. Dit is een expliciete afwijking van de strikte
  byte-identiteitsregel, geen stilzwijgend geslaagde bytecheck.
- [x] 2. Vindbare toetsen en correcte overlay-hulp. Kaartacties zoeken/gebied/lagen
  hebben voorrang, `? hulp` blijft zichtbaar, overlays tonen geen ontoegankelijke
  globale toetsen en laagkiezers tonen Spatie/Enter/Esc. ANSI-/graphemebewuste
  afkapping en titel-ellipsis getest. Vier echte `step2-*`-captureformaten gecontroleerd.
- [x] 3. Kaartgeometrie en compacte legenda. De componentlaag meet de werkelijke
  legenda in plaats van zes vaste reserveregels. Op korte schermen één regel;
  `L` opent een scrollbare volledige legenda. Op 30×12 zeven kaartregels in
  plaats van twee. Vier echte `step3-*`-captureformaten en geometrie-/toetstests.
- [x] 4. Query-resultaten eerst. Compact formulier op korte schermen, na uitvoering
  resultaten bovenaan, `ctrl+r` als aparte herhaalactie. Selectie blijft zichtbaar
  na terugkeer en resize; pagina-/begin-/eindtoetsen navigeren ook resultaten.
  Vier echte `step4-*`-captureformaten: alle drie Delfshaven-resultaten passen
  nu zelfs op 30×12. Objectdetails/paginering en annuleren blijven getest.
- [ ] 5. Scrollbaar BOB-overzicht.
- [ ] 6. Compacte inventarissen en taakgericht opdrachtenpalet.

## Conclusie

De basis is sterk: vraaggestuurde uitleg, expliciete herkomst/licenties,
een eigen componentlaag en onderscheid tussen lokale bestanden, metadata en
echte gebiedstellingen. De grootste winst zit in **vindbare bediening,
zichtbare resultaten en beter gebruik van kleine schermen**, niet in meer kleur
of animatie. Migreer eerst naar v2 zonder tegelijk opnieuw te ontwerpen.

## Onderzoek en bewijs

- DeepWiki-MCP gebruikt via de in `~/.hermes/config.yaml` ingestelde
  `https://mcp.deepwiki.com/mcp`: `read_wiki_structure` en `read_wiki_contents`
  voor Bubble Tea, Bubbles en Lip Gloss. Geen externe AI-vraag gesteld.
- De lokale skills `waterketen-cli`, `bubbletea-tui-development` en
  `terminal-ui-preview` uit `~/.hermes/skills/` gelezen.
- Een nieuwe scratch-build van de huidige broncode daadwerkelijk onder tmux
  bediend, niet de mogelijk verouderde repo-binary.
- **58 echte ANSI-/tekstcaptures**, inclusief alle acht panelen, lezen,
  gebied-/laagkiezers, beide paletten en Query-resultaten/objectdetails.
  Formaten: **120×36, 80×24, 48×20 en 30×12**. Niet iedere detailtoestand
  is op ieder formaat getest. Geen volledige toets-, muis- of prestatietest.
- Echte lokale snapshot gelezen; de bestaande automatische bronprobes uitgevoerd.
  Geen datasets gedownload of opgeslagen. Bronlatenties zijn momentopnamen.
- Artefacten buiten de repo:
  `~/.hermes/cache/scratch/waterketen-ui-review-20261013/`.
  `manifest.json` inventariseert de captures; `captures.html` toont een selectie
  rechtstreeks uit de ANSI-output, zonder mock of ontwerpvarianten.
- Gemeten SGR: accent `38;2;255;123;245` (`#ff7bf5`), gedeelde zwarte
  achtergrond `48;2;0;0;0` (`#000000`).
- Baseline groen: `gofmt -l .`, `go build ./...`, `go vet ./...` en
  `go test -count=1 ./...`.

## Bevindingen

| Prioriteit | Waargenomen probleem | Voorstel |
|---|---|---|
| P0 | `go.mod` gebruikt Bubble Tea v1.3.10, Bubbles v1.0.0 en Lip Gloss v1, terwijl v2 gewenst is. | Migreer de hele UI-stack als afzonderlijke, gedragbehoudende wijziging. |
| P1 | De footer kapt ook op 120 kolommen af. Op 48×20 verdwijnen bij Kaart onder andere zoeken, gebied en laagkeuze; op 30×12 eindigt hij al bij `+`. Twee palet-hints nemen steeds de eerste plaatsen in. | Een korte, per modus geprioriteerde helpregel; volledige hulp/opdrachten blijven bereikbaar. |
| P1 | Op 30×12 heeft de kaart twee tekenregels, maar de legenda vijf regels. Op 120×36 blijven na de legenda vijf lege regels over. | Reserveer hoogte op basis van de werkelijke legenda; compacte, uitklapbare legenda op korte schermen. |
| P1 | Op 48×20 staat het Query-formulier boven de resultaten. Na terugkeer uit objectdetails is slechts het eerste resultaat onderaan zichtbaar. Op 30×12 verdwijnt zelfs de uitvoeractie uit beeld. | Compacte samenvatting van gebied/lagen, resultaten eerst na uitvoering en geselecteerde rij altijd in beeld. |
| P1 | Query-laagkiezer adverteert `ctrl+p` en `?`, maar bezit de volledige toetsstroom. `ctrl+p` werkelijk ingedrukt: geen palet geopend. De eigen hint kapt af; Spatie ontbreekt in de footer. | Modus bepaalt zowel toetsen als hulp. Toon Spatie/Enter/Esc volledig; adverteer geen ontoegankelijke globale acties. |
| P1 | Profiel toont op 120×36 Tjalklaan met 54 buizen en “30 meer”, maar heeft geen scroll- of objectnavigatie. De stippen zijn gesorteerde BOB-waarden, geen ruimtelijke doorsnede. | Alle gegevens bereikbaar maken; voorlopig eerlijk benoemen als BOB-overzicht. Een echt profiel apart ontwerpen. |
| P2 | Lagen toont slechts ongeveer vijf registraties per desktopscherm door alle leveringsmetadata inline. Bronnen heeft variabele kolomuitlijning en geen vaste kolomkoppen. | Compacte selecteerbare lijst/tabel met details op aanvraag, en smalle kaartjes op mobiel. |
| P2 | Het globale palet bevat veel opdrachten zoals “Bronnen · omhoog” en “Lagen · omlaag”; echte taken raken ertussen verstopt. | Globaal palet voor taken; scrolltoetsen in paneelhulp, niet als tientallen globale opdrachten. |
| P2 | Tijdens de broncontrole worden al ontvangen resultaten verborgen achter een vrijwel leeg scherm met spinner/balk, bijvoorbeeld bij 16/17 antwoorden. | Deelresultaten zichtbaar houden en voortgang erboven tonen. |
| P2 | Ontdek legt uit wat je kunt doen, maar biedt nog geen directe route naar Kaart/Query met die context. Lange uitleg verdringt op mobiel de kern. | Korte kernuitleg, verdieping eronder en expliciete acties naar de juiste lokale gegevens. |

Bewijscaptures: onder andere `30x12-kaart`, `48x20-kaart`,
`48x20-query-delfshaven-resultaat`, `48x20-query-laagkiezer`,
`120x36-profiel`, `120x36-lagen`, `120x36-opdrachten` en
`120x36-bronnen`. De afgebroken Querytitel op 30 kolommen heeft bovendien
geen ellipsis: neem ook titelafkapping mee.

## Fase 1 — Charm v2, zonder redesign

**Doel:** v2 als basis, met dezelfde inhoud, kaartdata-ink en bestaande bediening.

Beschikbare versies gecontroleerd met `go list -m -json ...@latest` tijdens
de review; voorgestelde pins:

| Module | Versie |
|---|---|
| `charm.land/bubbletea/v2` | `v2.1.0` |
| `charm.land/bubbles/v2` | `v2.2.1` |
| `charm.land/lipgloss/v2` | `v2.0.6` |
| `charm.land/glamour/v2` | `v2.0.1` |

Deze modules vragen geen nieuwere Go-versie dan de huidige Go 1.26.0.
Controleer de onderlinge afhankelijkheden en de definitieve modulegraph bij
implementatie; doe geen brede upgrade van ongerelateerde dependencies.

1. Versterk eerst de architectuurtests. De schermtest herkent momenteel alleen
   `charmbracelet/lipgloss`; hij moet ook `charm.land/lipgloss/v2` afwijzen.
   Controleer ook de pure-core-importguards en leg een v2-stackguard vast.
2. Migreer imports, inclusief tests en Glamour. Inventariseer eventuele transitieve
   v1-restanten; geen v1-componenten in de actieve v2-UI.
3. Laat alleen het applicatiemodel `View() tea.View` retourneren.
   De eigen `Screen`-interface mag tekst blijven leveren aan de app-shell.
   Zet alt-screen/muismodus declaratief op de view, ook tijdens initialiseren.
4. Vervang toetsconstructie en -routing door `tea.KeyPressMsg`, `Code`, `Text`
   en modifiers. Behandel `tea.PasteMsg` afzonderlijk: een hernoeming van
   `KeyMsg` alleen behoudt plakken niet. Migreer alle testhelpers en palet-replay.
5. Migreer klikken/wiel/motion naar de v2-muistypen. Behoud kaartcoördinaten,
   zoomankers en overlay-eigenaarschap; releases mogen geen dubbele actie geven.
6. Gebruik Bubbles-getters/setters voor afmetingen en nieuwe progresskleur-API's.
   Behoud één app-spinner, echte tellers en stoppende animaties wanneer idle.
7. Vervang `AdaptiveColor` en v1-kleurprofieltests. Behoud voorlopig de expliciete
   zwarte app-oppervlakte en dezelfde donkere tokens uit de huidige README.
   Oudere skills adviseren transparantie; dat is geen reden om tijdens deze
   migratie het huidige repo-contract stilzwijgend terug te draaien.
8. Werk documentatie bij: v2-imports/API's en de huidige tabnavigatie, niet de
   inmiddels verdwenen sidebar als uitlegmodel.

**Betrokken:** `go.mod`, `go.sum`, `internal/app/`, `internal/ui/` en UI-tests.
Coreberekeningen en managed data blijven ongemoeid.

**Gereed wanneer:** alle controles groen zijn en identieke kaarttoestanden op
dezelfde geometrie/snapshot worden vergeleken met de v1-baseline. Streef naar
byte-identieke `tmux capture-pane -e`-kaartoutput volgens het repo-contract.
Bij verschillen door de nieuwe renderer: niet automatisch accepteren; toon
de exacte diff en controleer cellen, kleuren, nadruk en geometrie afzonderlijk.

## Fase 2 — Bedienbaarheid en mobiel

Kleine, afzonderlijk te controleren wijzigingen, in deze volgorde:

1. **Contextuele hulp en focus.** Maak korte versus volledige hulp expliciet,
   zoals Bubbles' `ShortHelp`/`FullHelp`. Gebruik actieve acties, geen generieke
   “omhoog/omlaag” waar “verplaatsen”, “kiezen” of “lezen” bedoeld wordt.
   Geef op mobiel voorrang aan uitvoeren/zoeken/gebied/lagen/terug;
   behoud een zichtbare route naar de overige opdrachten. Toon focus met een
   cursor/marker en tekst, niet alleen kleur. Bestaande toetsen 1–8 blijven gelijk.
2. **Kaartgeometrie.** Eén berekening voor kaart, legenda en context.
   Verkort/klap de legenda in op korte schermen zonder haar betekenis te verliezen;
   schaal en noorden blijven zichtbaar. Legenda open/gesloten heeft bijpassende
   hulp. Wijzig geen laagkleuren, tekenvolgorde of samplingsemantiek.
3. **Query-layout.** Vóór uitvoeren: formulier. Na uitvoeren: resultaten met een
   compacte gebied-/laagcontext en een duidelijke aparte herhaalactie.
   Enter blijft contextueel, maar zijn huidige effect is altijd leesbaar.
   Houd selectie zichtbaar bij uitvoeren, terugkeren, scrollen en resizen.
4. **Profielbereikbaarheid.** Scrollbare/selecteerbare BOB-lijst en straatkeuze.
   Geen “meer” zonder toegang tot die gegevens. Toon NAP-eenheden en de beperking
   “geen diepte onder maaiveld”; presenteer de huidige stippen niet als doorsnede.
5. **Afkapping/Unicode.** Eén ANSI-/graphemebewuste afkapfunctie op basis van
   `x/ansi`, ook voor paneltitels. Test accenten, samengestelde tekens, lange
   CBS-labels en lange invoer. Laat iedere afkapping herkenbaar eindigen op `…`.

**Acceptatie:** primaire acties en annuleren zijn op 48×20 en 30×12 vindbaar;
alle resultaten blijven bereikbaar; footer/header blijven één regel; geen
geadverteerde toets zonder aantoonbaar effect in de betreffende modus.

## Fase 3 — Informatiehiërarchie en taken

- Bouw een herbruikbare inventorycomponent: Bubbles `list`/`table` met
  compacte regels en selectie; details in een leesviewport. Op smalle terminals
  geen geplette desktopkolommen. Licentie, beschikbaarheid en onzekerheid blijven
  zichtbaar in samenvatting of direct bereikbare details.
- Gebruik Bubbles `textinput` voor consistente zoekvelden: zichtbare cursor,
  bewerken in het midden, Unicode, plakken en horizontale scroll. Behoud
  domeinmatching op naam/CBS-code en voorkom globale shortcuts tijdens typen.
- Maak het globale palet taakgericht: “Kies gebied”, “Zoek straat”,
  “Kies kaartlagen”, “Voer gebiedstelling uit”, “Ververs bronnen”. Gebruik
  bestaande paneelhulp voor navigatietoetsen. Toon niet-uitvoerbare acties niet
  als uitvoerbaar; geen verborgen automatische downloads.
- Houd bij bronprobes resultaten zichtbaar met voortgang erboven. Gebruik een
  vaste, begrijpelijke ordening; de huidige core-volgorde niet stilzwijgend wijzigen.
- Voeg vanuit Ontdek/Lagen/Gebieden expliciete acties naar Kaart/Query toe.
  Definieer gebied-/laagoverdracht eerst: managed Delfshaven en snapshotwijk
  Delfshaven zijn niet automatisch dezelfde selectie of telling.
- Geef Ontdek een korte kern en leesbare vervolgstappen; gebruik technische
  verdieping voor volledige API-details, maar verstop licenties/beperkingen niet.

## Fase 4 — Echt profiel, apart domeinontwerp

Een ruimtelijk profiel vraagt een afstandsas, echte buissegmenten en begin-/eind-
BOB's, niet gesorteerde stippen. Definieer straatprojectie en richting eerst.
Alleen met expliciete maaiveldgegevens mag straatdiepte worden afgeleid; geen
verzonnen maaiveldlijn of impliciete AHN-download. Dit is meer dan UI-styling en
verdient een eigen voorstel en validatie tegen de brongegevens.

## Verificatie per wijziging

```bash
gofmt -l . && go build ./... && go vet ./... && go test -count=1 ./...
go test -race -count=1 ./...
```

- Test toetsgedrag, niet alleen aanwezigheid in `KeyMap()`.
- Test overlays: Spatie, Enter, Esc, Unicode/plakken, annuleren zonder wijziging,
  resize en geen lekkende globale acties.
- Controleer Query: nul ≠ ontbrekend ≠ leesfout, complete telling onafhankelijk
  van paginering, terugkeer naar geselecteerde rij en annuleren tijdens werk.
- Capture alle acht panelen en relevante toestanden opnieuw op de vier formaten,
  plus bestaande geometrie-/overlayregressies. Voeg idle CPU/repaintmeting toe aan
  de v2- en animatiewijzigingen. Geen vaste prestatieclaim zonder meting.
- Gebruik gecontroleerde lokale fixtures voor deterministische vergelijking;
  gebruik daarnaast echte snapshotdata. Geen verzonnen brongezondheid.
- Bewaar previews buiten de repo. Werk de afgesproken binaries alleen bij wanneer
  bouwen/installeren daarvan onderdeel van de implementatieopdracht is.

## Bronnen

- [DeepWiki — Bubble Tea](https://deepwiki.com/charmbracelet/bubbletea): model/view,
  inputrouting, asynchroon werk en terminalevents; via MCP gelezen.
- [DeepWiki — Bubbles](https://deepwiki.com/charmbracelet/bubbles): help,
  list/table, textinput en v1→v2-migratie; via MCP gelezen.
- [DeepWiki — Lip Gloss](https://deepwiki.com/charmbracelet/lipgloss): ANSI-/Unicode-
  afmetingen, compositie en terminalkleurprofielen; via MCP gelezen.
- [Officiële Bubble Tea v2-upgradegids](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md),
  geraadpleegd voor declaratieve views en afzonderlijke paste-/muismessages.
- Lokale autoriteit: [go.mod](../go.mod), [app-shell](../internal/app/app.go),
  [commandobar](../internal/ui/components/commandbar.go),
  [Query](../internal/ui/screens/query.go),
  [Profiel](../internal/ui/screens/analysis.go),
  [schermcontract](../internal/ui/screens/contract_test.go) en de genoemde captures.

DeepWiki documenteert hier upstream v2, niet de huidige lokale v1-build.
De aanbevelingen zijn een combinatie van die bouwstenen en de gemeten review;
geen claim dat DeepWiki deze applicatie zelf heeft beoordeeld.
