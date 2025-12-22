package main

import (
	"log"

	"github.com/rivo/tview"

	"nvimwiz/internal/ui"
)

func main() {
	app := tview.NewApplication()
	w, err := ui.New(app)
	if err != nil {
		log.Fatal(err)
	}
	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
