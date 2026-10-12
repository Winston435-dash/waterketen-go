# Waterketen Go rewrite — migratieplan

**Beslissing:** volledige Go rewrite van `waterketen-cli` + `waterketen-tui`
(optie C). Stack: Bubble Tea (architectuur) · Bubbles (componenten) · Lip Gloss
(design tokens) · `ui/` kit (project-design-taal).

## Kernprincipe (uit de instructies)

> Never create borders/colors directly. Use components from `ui/components`.
> Use theme tokens only. Every screen has Header → Content → CommandBar.

Dat betekent: de agent (Codex, of wij) mag nooit `lipgloss.NewStyle()` met een
hex-kleur in een screen schrijven — alleen `ui.Panel(...)`, `ui.StatusBadge(...)`,
tokens als `theme.ColorPrimary` / `theme.SpacingSM`.

## Voorgestelde structuur

```
~/projects/waterketen-go/
├── go.mod
├── cmd/
│   ├── waterketen/main.go          # CLI (cobra)
│   └── waterketen-tui/main.go      # TUI entry
├── internal/
│   ├── core/                       # pure GIS-logica, geen UI
│   │   ├── sources.go, pipeline.go, analysis.go, health.go,
│   │   ├── report.go, braille.go, kaart.go, library.go …
│   └── ui/
│       ├── theme.go                # design tokens (kleuren, spacing, borders)
│       ├── keymap.go               # centrale keymap
│       ├── components/
│       │   ├── panel.go            # Panel: titel + border + padding
│       │   ├── commandbar.go       # CommandBar: footer met acties
│       │   ├── status.go           # StatusBadge: live/dead/warn
│       │   ├── table.go            # DataTable bovenop bubbles/table
│       │   ├── empty.go            # EmptyState (FETCH_HINT-equivalent)
│       │   └── canvas.go           # BrailleCanvas (atlas-rendering)
│       └── screens/
│           ├── overview.go         # Bronnen
│           ├── sources.go          # Lagen + Gebieden
│           ├── query.go            # Analyse
│           ├── atlas.go            # Kaart (5)
│           └── profile.go          # Profiel (6)
└── README.md
```

## GIS in Go — eerlijk over de grens

| Functie | Go-strategie |
|---|---|
| Braille-atlas, overzichtskaart, profiel | **native port** — alle rendercode is index-matig → strings |
| WFS probing (health) | native HTTP + XML |
| Lagen/gebieden/bronnen panelen | native (metadata-JSON lezen) |
| Analyse (diepte, verval, lengte) | native op metadata/attributen |
| `fetch` (GWSW WFS-paging, clip op CBS-grens) | **port** — WFS-paging is HTTP+XML, clip = bbox (geen echte overlay nodig) |
| `export` gpkg/geojson | port: geojson native; gpkg via GDAL `ogr2ogr` shell-uit of weglaten in v1 |
| `kaart` SVG/PNG-render, `bibliotheek bouw` (osmium+ogr2ogr) | **buiten scope v1** — blijft woonplaats van de Python-CLI, die bestaat en werkt |

V1-doel: een TUI die er goed uitziet en leest uit `~/.local/share/waterketen/`
(dezelfde managed data), plus een Go `waterketen` CLI met `sources`, `status`,
`analyse`, `braille`. De Python `fetch`/`bibliotheek` blijven de producer van de
data tot de Go-fetch er is — beide tools praten hetzelfde opslagformaat.

## Werkwijze

- Bestaand Python-project **niet aanraken** (git werkboom is vies, maar dat is
  bewust ongebcommit werk — commit daar hoort bij dat proces, niet bij dit).
- Nieuwe repo `waterketen-go`; geen productie-data migratie nodig want gedeelde
  directory.
- TDD waar het kan: core-pakketten krijgen Go tests; screens krijgen
  golden-file tests op de gerenderde output (teatest).

## Open vragen voor later

- `library/`-atlas (osmium-import) naar Go? **Lezen is gedaan**, bouwen niet:
  `core/library_atlas.go` leest viewports uit de snapshot (R-tree, geen cgo) met
  het gebiedsregister uit `areas.gpkg`. `bibliotheek bouw` blijft Python — dat is
  een eenmalige osmium+ogr2ogr-import. Twee bewuste afwijkingen van
  `library_atlas.py`: geen polygoon-clip per viewport (de snapshot is al op
  Rotterdam geclipt) en geen `simplify` (de renderer laat dots per cel vallen).
  Gevolg: een viewport kan méér objecten tonen dan Python voor hetzelfde venster,
  en dat staat als waarschuwing in de kaart.

## Bronnenregister: twee tools, één waarheid per bron

De bronlijsten lopen bewust **niet** meer één-op-één: de Python-CLI heeft 3
bedrade bronnen (`gws_netwerk`, `damo_leiding`, `vewin`), de Go-CLI heeft er 17
(inclusief afzonderlijke DAMO-/RIONED-objecttypen en vier IMWA-bronlagen). De
licentie-notie is in **beide** tools dezelfde eis, met een test:
`internal/core/sources_test.go` en `tests/test_health.py`. Trek een nieuwe bron
in Python bij in `tui/app.py` (`WIRED_SOURCES`) en in Go in
`internal/core/sources.go` — inclusief `license` en, voor WFS, `type_name`.
