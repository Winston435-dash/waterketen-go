package components

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"

	"waterketen/internal/ui"
)

// NewSpinner builds the app's single busy spinner from the theme tokens. The
// app owns ONE spinner Model for the whole program and pushes its frame down
// to screens; screens never start their own tick chain.
func NewSpinner() spinner.Model {
	frames := make([]string, 0, len(ui.SpinnerFrames))
	for _, r := range ui.SpinnerFrames {
		frames = append(frames, string(r))
	}
	return spinner.New(
		spinner.WithSpinner(spinner.Spinner{
			Frames: frames,
			FPS:    time.Second / ui.SpinnerFPS,
		}),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(ui.ColorPrimary)),
	)
}

// NewProgress builds the app's progress bar: Bubbles' own progress component,
// so the fill glides to its new value on a spring (harmonica) instead of
// snapping, and the fill is a gradient rather than a flat colour. width is the
// number of cells the bar occupies; the caller clamps it to the panel.
func NewProgress(width int) progress.Model {
	return progress.New(
		progress.WithColors(lipgloss.Color(ui.GradientBarA), lipgloss.Color(ui.GradientBarB)),
		progress.WithWidth(width),
		progress.WithoutPercentage(), // the panel shows its own "n/totaal" count
		progress.WithFillCharacters('█', '░'),
	)
}

// Busy is the app's single "work in progress" line: a live spinner frame and a
// muted Dutch label. The frame is expected pre-styled (the app passes
// spinner.Model.View()); Busy only adds the gap and the label.
func Busy(frame, label string) string {
	if frame == "" {
		return Muted(label)
	}
	return frame + " " + Muted(label)
}

// ProgressCount is the honest counter that sits beside the bar: real units
// finished of the real total. It never invents a percentage.
func ProgressCount(done, total int) string {
	if total <= 0 {
		total = 1
	}
	done = min(max(done, 0), total)
	return fmt.Sprintf("%d/%d", done, total)
}
