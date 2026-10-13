package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
	"waterketen/internal/ui/screens"
)

type paletteScope uint8

const (
	paletteGlobal paletteScope = iota
	palettePanel
)

type paletteAction uint8

const (
	paletteNavigate paletteAction = iota
	paletteProbe
	paletteRedraw
	paletteQuit
	paletteDispatchPanel
)

type paletteCommand struct {
	group    string
	label    string
	shortcut string
	action   paletteAction
	target   screens.ScreenID
	targeted bool
	key      tea.KeyPressMsg
}

type commandPalette struct {
	scope     paletteScope
	scopeName string
	query     string
	input     textinput.Model
	cursor    int
	commands  []paletteCommand
}

func newCommandPalette(m Model, scope paletteScope) *commandPalette {
	p := &commandPalette{scope: scope, scopeName: "alle panelen"}
	p.input=components.SearchInput("Zoek · ")
	p.input.SetWidth(max(1,m.width-8))
	if scope == palettePanel {
		p.scopeName = screenTitle(m.screens[m.active])
		p.commands = panelPaletteCommands(m.screens[m.active], false)
		return p
	}

	for _, info := range screens.Screens {
		p.commands = append(p.commands, paletteCommand{
			group:    "Navigatie",
			label:    "Ga naar " + info.Title,
			shortcut: info.Key,
			action:   paletteNavigate,
			target:   info.ID,
			targeted: true,
		})
	}
	if !m.probeRunning() { p.commands=append(p.commands,paletteCommand{group:"Algemeen",label:"Ververs bronnen",shortcut:"v",action:paletteProbe}) }
	if k:=m.kaart();k!=nil&&k.Ready() { p.commands=append(p.commands,paletteCommand{group:"Algemeen",label:"Her-teken kaart",shortcut:"r",action:paletteRedraw,target:screens.ScreenKaart,targeted:true}) }
	p.commands=append(p.commands,paletteCommand{group:"Algemeen",label:"Afsluiten",shortcut:"q",action:paletteQuit})
	for _, screen := range m.screens {
		p.commands = append(p.commands, panelPaletteCommands(screen, true)...)
	}
	return p
}

// panelPaletteCommands exposes actions a screen currently advertises. The
// global palette also lists these, but keeps app-wide keys in its own section.
func panelPaletteCommands(screen screens.Screen, omitAppKeys bool) []paletteCommand {
	var commands []paletteCommand
	bindings:=screen.KeyMap()
	if omitAppKeys { if available,ok:=screen.(interface{ PaletteKeyMap() []key.Binding });ok { bindings=available.PaletteKeyMap() } }
	for _, binding := range bindings {
		if !binding.Enabled() || isBinding(binding, ui.Keys.NextPanel, ui.Keys.PrevPanel,
			ui.Keys.Quit, ui.Keys.Palette, ui.Keys.PanelPalette) {
			continue
		}
		if omitAppKeys && isBinding(binding, ui.Keys.Probe, ui.Keys.Redraw) {
			continue
		}
		if omitAppKeys && isBinding(binding,ui.Keys.Up,ui.Keys.Down,ui.Keys.PanLeft,ui.Keys.PanRight,ui.Keys.PageUp,ui.Keys.PageDown,ui.Keys.Home,ui.Keys.End,ui.Keys.Back,ui.Keys.NextLayer,ui.Keys.PrevLayer,ui.Keys.NextTopic,ui.Keys.PrevTopic) { continue }
		keys := binding.Keys()
		if len(keys) == 0 {
			continue
		}
		msg, ok := keyMessage(keys[0])
		if !ok {
			continue
		}
		help := binding.Help()
		label:=help.Desc
		if omitAppKeys { label=taskLabel(screen.ID(),binding) }
		action := paletteDispatchPanel
		if isBinding(binding, ui.Keys.Probe) {
			action = paletteProbe
		} else if isBinding(binding, ui.Keys.Redraw) {
			action = paletteRedraw
		}
		commands = append(commands, paletteCommand{
			group:    screenTitle(screen),
			label:    label,
			shortcut: help.Key,
			action:   action,
			target:   screen.ID(),
			targeted: true,
			key:      msg,
		})
	}
	return commands
}

func taskLabel(id screens.ScreenID,binding key.Binding)string {
	k:=binding.Help().Key
	switch id {
	case screens.ScreenKaart:
		if label:=map[string]string{"/":"Zoek straat","g":"Kies kaartgebied","t":"Kies kaartlagen","L":"Lees kaartlegenda","+":"Zoom in","-":"Zoom uit","f":"Maak kaart passend","o":"Wissel samen/solo","l":"Toon/verberg straatnamen"}[k];label!="" { return label }
	case screens.ScreenQuery:
		if label:=map[string]string{"g":"Kies querygebied","l":"Kies querylagen","ctrl+r":"Herhaal gebiedstelling"}[k];label!="" { return label }
		if k=="enter" { if binding.Help().Desc=="objecten" { return "Bekijk queryobjecten" };return "Voer gebiedstelling uit" }
	case screens.ScreenProfiel:
		if k=="s" { return "Kies BOB-straat" }
	case screens.ScreenLagen:
		if k=="l" { return "Toon lokale/alle lagen" };if k=="/" { return "Filter lagen" };if k=="enter" { return "Lees laagdetails" }
	case screens.ScreenGebieden:
		if k=="/" { return "Filter gebieden" };if k=="enter" { return "Lees gebiedsdetails" }
	case screens.ScreenBronnen:
		if k=="enter" { return "Lees brondetails en licentie" }
	case screens.ScreenOntdek:
		if k=="enter" { return "Lees geselecteerde uitleg" };if k=="t" { return "Toon/verberg technische uitleg" }
	}
	return binding.Help().Desc
}

func isBinding(binding key.Binding, candidates ...key.Binding) bool {
	keys := binding.Keys()
	for _, candidate := range candidates {
		other := candidate.Keys()
		if len(keys) != len(other) {
			continue
		}
		match := true
		for i := range keys {
			if keys[i] != other[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// keyNames maps the key names a binding's Keys() reports back to their key
// types, so the palette can replay a command's shortcut as a real key message.
var keyNames = map[string]rune{
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"enter":     tea.KeyEnter,
	"esc":       tea.KeyEsc,
	"backspace": tea.KeyBackspace,
	"space":     tea.KeySpace,
	" ":         tea.KeySpace,
	"tab":       tea.KeyTab,
}

func keyMessage(name string) (tea.KeyPressMsg, bool) {
	var mod tea.KeyMod
	for _, prefix := range []struct {
		name string
		mod  tea.KeyMod
	}{{"ctrl+", tea.ModCtrl}, {"shift+", tea.ModShift}, {"alt+", tea.ModAlt}} {
		if strings.HasPrefix(name, prefix.name) {
			mod |= prefix.mod
			name = strings.TrimPrefix(name, prefix.name)
		}
	}
	if code, ok := keyNames[name]; ok {
		return tea.KeyPressMsg{Code: code, Mod: mod}, true
	}
	if runes := []rune(name); len(runes) == 1 {
		text := name
		if mod != 0 {
			text = ""
		}
		return tea.KeyPressMsg{Code: runes[0], Text: text, Mod: mod}, true
	}
	return tea.KeyPressMsg{}, false
}

func (p *commandPalette) filteredCommands() []paletteCommand {
	query := strings.ToLower(strings.TrimSpace(p.query))
	if query == "" {
		return p.commands
	}
	filtered := make([]paletteCommand, 0, len(p.commands))
	for _, command := range p.commands {
		searchable := strings.ToLower(command.group + " " + command.label + " " + command.shortcut)
		if strings.Contains(searchable, query) {
			filtered = append(filtered, command)
		}
	}
	return filtered
}

func (p *commandPalette) selectedCommand() (paletteCommand, bool) {
	commands := p.filteredCommands()
	if len(commands) == 0 || p.cursor < 0 || p.cursor >= len(commands) {
		return paletteCommand{}, false
	}
	return commands[p.cursor], true
}

func (p *commandPalette) move(delta int) {
	commands := p.filteredCommands()
	if len(commands) == 0 {
		p.cursor = 0
		return
	}
	p.cursor = (p.cursor + delta + len(commands)) % len(commands)
}

func (m Model) updatePalette(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	p := m.palette
	switch msg.String() {
	case "esc":
		m.palette = nil
		return m, nil
	case "ctrl+p":
		if p.scope == palettePanel {
			m.palette = newCommandPalette(m, paletteGlobal)
		}
		return m, nil
	case "up":
		p.move(-1)
	case "down":
		p.move(1)
	case "enter":
		command, ok := p.selectedCommand()
		if !ok {
			return m, nil
		}
		m.palette = nil
		return m.executePaletteCommand(command)
	default:
		// screens.TypedText, not a KeyRunes test: a SPACE arrives as KeySpace
		// and would otherwise be silently dropped while typing a command.
		old:=p.query
		p.input,_=p.input.Update(msg);p.query=p.input.Value()
		if old!=p.query { p.cursor=0 }
	}
	return m, nil
}

func (m Model) executePaletteCommand(command paletteCommand) (tea.Model, tea.Cmd) {
	index := -1
	if command.targeted {
		index = m.screenIndex(command.target)
		if index < 0 {
			return m, nil
		}
	}
	switch command.action {
	case paletteNavigate:
		m.active = index
		return m, m.afterScreenUpdate()
	case paletteProbe:
		return m, m.beginProbe()
	case paletteRedraw:
		m.active = index
		return m, m.redrawMap()
	case paletteQuit:
		return m, tea.Quit
	case paletteDispatchPanel:
		m.active = index
		next, cmd := m.screens[index].Update(command.key)
		m.screens[index] = next
		return m, tea.Batch(cmd, m.afterScreenUpdate())
	default:
		return m, nil
	}
}

func (m Model) screenIndex(id screens.ScreenID) int {
	for i, screen := range m.screens {
		if screen.ID() == id {
			return i
		}
	}
	return -1
}

func (p *commandPalette) View(width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}
	p.input.SetWidth(max(1,width-8))
	prompt := p.input.View()
	if height == 1 {
		return components.SingleLine(prompt, width)
	}
	if height == 2 {
		return components.SingleLine(prompt, width) + "\n" + components.SingleLine(
			components.Muted("enter uitvoeren · esc terug"), width)
	}

	commands := p.filteredCommands()
	listRows := height - 2
	start := 0
	if p.cursor >= listRows {
		start = p.cursor - listRows + 1
	}
	lines := make([]string, 0, height)
	lines = append(lines, components.SingleLine(prompt, width))
	if len(commands) == 0 {
		empty := "Geen opdrachten gevonden"
		if len(p.commands) == 0 && p.scope == palettePanel {
			empty = "Geen paneelopdrachten"
		}
		lines = append(lines, components.SingleLine(components.Muted(empty), width))
	} else {
		for i := start; i < len(commands) && len(lines) < height-1; i++ {
			command := commands[i]
			label := command.group + " · " + command.label
			if command.shortcut != "" {
				label += "  · " + command.shortcut
			}
			if i == p.cursor {
				lines = append(lines, components.SingleLine(components.Selected("› "+label, width), width))
			} else {
				lines = append(lines, components.SingleLine("  "+label, width))
			}
		}
	}
	lines = append(lines, components.SingleLine(components.Muted("↑↓ kiezen · enter uitvoeren · esc terug"), width))
	return strings.Join(lines, "\n")
}
