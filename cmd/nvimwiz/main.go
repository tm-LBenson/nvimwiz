package main

import (
	"log"

	"github.com/rivo/tview"

	"nvimwiz/internal/env"
	"nvimwiz/internal/ui"
)

func main() {
	// Make user-local installs visible inside the running wizard
	// (so you don't need to restart after installing ...). This only affects
	// the nvimwiz process and any children it spawns.
	_, _, _ = env.EnsureLocalBinInPath()

	app := tview.NewApplication()
	w := ui.NewWizard(app)
	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
