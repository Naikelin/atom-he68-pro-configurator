package main

import (
	_ "embed"
	"fmt"
	"time"

	"fyne.io/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

var (
	trayApp       *App
	batteryItem   *systray.MenuItem
	connectedItem *systray.MenuItem
)

// StartTray runs the system tray icon in a goroutine.
func StartTray(app *App) {
	trayApp = app
	go systray.Run(onTrayReady, onTrayExit)
}

func onTrayReady() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("HE68 PRO")
	systray.SetTooltip("ATOM HE68 PRO Configurator")

	mShow := systray.AddMenuItem("Show", "Open configurator")
	mRefresh := systray.AddMenuItem("Refresh battery", "Query battery now")
	systray.AddSeparator()
	connectedItem = systray.AddMenuItem("Keyboard: unknown", "Connection status")
	connectedItem.Disable()
	batteryItem = systray.AddMenuItem("Battery: unknown", "Battery level")
	batteryItem.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit application")

	updateTrayStatus()

	// Ticker for periodic battery updates.
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			updateTrayStatus()
		}
	}()

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				if trayApp.ctx != nil {
					runtime.WindowShow(trayApp.ctx)
					runtime.WindowUnminimise(trayApp.ctx)
				}
			case <-mRefresh.ClickedCh:
				updateTrayStatus()
			case <-mQuit.ClickedCh:
				if trayApp.ctx != nil {
					runtime.Quit(trayApp.ctx)
				}
				return
			}
		}
	}()
}

func onTrayExit() {}

func updateTrayStatus() {
	if trayApp == nil {
		return
	}
	connected := trayApp.Connected()
	if connected {
		connectedItem.SetTitle("Keyboard: connected")
		level, err := trayApp.Battery()
		if err != nil || level < 0 {
			batteryItem.SetTitle("Battery: N/A")
		} else {
			batteryItem.SetTitle(fmt.Sprintf("Battery: %d%%", level))
		}
	} else {
		connectedItem.SetTitle("Keyboard: disconnected")
		batteryItem.SetTitle("Battery: N/A")
	}
}
