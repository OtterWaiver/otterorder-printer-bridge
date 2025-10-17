package main

import (
	"context"
	"fmt"
	"otter-order-printer-bridge/internal"
)

type App struct {
	ctx           context.Context
	printerServer *internal.PrinterServer
	prefs         *internal.PrefsStore
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.prefs, _ = internal.NewPrefsStore("otter-order-printer-bridge")
	a.printerServer = internal.NewPrinterServer(a.prefs)
	fmt.Println(a.prefs.Path)

	fmt.Println("Printer server started on port 3838")
}

func (a *App) GetServerStatus() string {
	if a.printerServer != nil {
		return "Printer server is running on port 3838"
	}
	return "Printer server is not running"
}

func (a *App) GetPreferences() (internal.Preferences, error) {
	return a.prefs.GetPreferences()
}

func (a *App) SavePreferences(p internal.Preferences) error {
	a.prefs.UpdatePreferences(p)
	return a.printerServer.PrintTestPage()
}
