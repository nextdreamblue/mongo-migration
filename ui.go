package main

import (
	"fmt"
	"github.com/gizak/termui"
	"time"
)

func keyboardShortcuts() *termui.List {
	strs := []string{
		"[q] [quit](fg-red)",
		"[d] [debug](switch to log in debug mode)",
		"[s] [start](fg-white,bg-green)"}

	ls := termui.NewList()
	ls.Items = strs
	ls.ItemFgColor = termui.ColorYellow
	ls.BorderLabel = "Press key to action"
	ls.Height = 7
	ls.Width = 25
	ls.Y = 0

	return ls
}

func setupKeyboardHandle(handleMigration HandleMigration, started bool, from InstanceInfo, to InstanceInfo) {
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		handleMigration.Stop = true
		for !handleMigration.Stopped {
			fmt.Print(".")
			time.Sleep(1000 * time.Millisecond)
		}
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/s", func(termui.Event) {
		if !started {
			ImportCollection(&from, &to, &handleMigration)
			started = true
		}
	})
	termui.Handle("/sys/kbd/d", func(termui.Event) {
		handleMigration.LogMode = !handleMigration.LogMode
	})

}
