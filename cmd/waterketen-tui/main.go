// waterketen-tui is the interactive atlas. Run with `waterketen-tui` — it
// reads the same managed store as the waterketen CLI.
package main

import (
	"fmt"
	"os"

	"waterketen/internal/app"
)

func main() {
	if err := app.Run(""); err != nil {
		fmt.Fprintln(os.Stderr, "waterketen-tui:", err)
		os.Exit(1)
	}
}
