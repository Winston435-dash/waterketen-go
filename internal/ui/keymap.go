package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap is the single source of truth for every binding in the app. Screens
// declare which of these they use; the CommandBar renders them. A key that is
// not in this table does not exist.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	NextPanel key.Binding
	PrevPanel key.Binding
	NextLayer key.Binding
	PrevLayer key.Binding

	Probe        key.Binding
	Redraw       key.Binding
	Back         key.Binding
	Quit         key.Binding
	Enter        key.Binding
	Palette      key.Binding
	PanelPalette key.Binding

	// Explanatory catalog only.
	NextTopic key.Binding
	PrevTopic key.Binding
	Technical key.Binding

	// Data query only.
	QueryLayers key.Binding
	RunQuery    key.Binding
	Inspect     key.Binding

	// Atlas (kaart) only
	PanLeft  key.Binding
	PanRight key.Binding
	ZoomIn   key.Binding
	ZoomOut  key.Binding
	Fit      key.Binding
	Overlay  key.Binding
	Labels   key.Binding
	Search   key.Binding
	Area     key.Binding
	Layers   key.Binding
}

// Keys is the app-wide binding table. Descriptions are Dutch, matching the UI.
var Keys = KeyMap{
	Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "omhoog")),
	Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "omlaag")),
	PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "pagina ↑")),
	PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "pagina ↓")),
	Home:     key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "begin")),
	End:      key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "eind")),

	NextPanel: key.NewBinding(key.WithKeys("tab", "j"), key.WithHelp("j", "volgend paneel")),
	PrevPanel: key.NewBinding(key.WithKeys("shift+tab", "k"), key.WithHelp("k", "vorig paneel")),
	NextLayer: key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "laag →")),
	PrevLayer: key.NewBinding(key.WithKeys("["), key.WithHelp("[", "← laag")),

	Probe:  key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "ververs")),
	Redraw: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "herteken")),
	Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "terug")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "afsluiten")),
	// Enter confirms a picker/overlay. It is NOT a global key: outside an
	// overlay it does nothing, and a zero-value binding here would make every
	// enter in the picker silently do nothing.
	Enter:        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "kies")),
	Palette:      key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "opdrachten")),
	PanelPalette: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "paneelopdrachten")),

	NextTopic: key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "categorie →")),
	PrevTopic: key.NewBinding(key.WithKeys("["), key.WithHelp("[", "← categorie")),
	Technical: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "techniek")),

	QueryLayers: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "lagen")),
	RunQuery:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "uitvoeren")),
	Inspect:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "objecten")),

	PanLeft:  key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "west")),
	PanRight: key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "oost")),
	ZoomIn:   key.NewBinding(key.WithKeys("+", "="), key.WithHelp("+", "zoom in")),
	ZoomOut:  key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "zoom uit")),
	Fit:      key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "passend")),
	Overlay:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "lagen samen")),
	Labels:   key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "straatnamen")),
	Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "straat zoeken")),
	Area:     key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "gebied")),
	Layers:   key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "lagen")),
}
