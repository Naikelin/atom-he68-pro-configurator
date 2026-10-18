package main

import (
	"context"
	"fmt"

	"github.com/Naikelin/atom-he68-pro-configurator/desktop/internal/hid"
)

// App is the Wails application struct.
type App struct {
	ctx    context.Context
	device hid.Device
}

// NewApp creates a new App.
func NewApp() *App {
	return &App{}
}

// startup is called at application startup.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	StartTray(a)
}

// Connect opens the HID interface.
func (a *App) Connect() error {
	return a.device.Connect()
}

// Disconnect closes the HID interface.
func (a *App) Disconnect() error {
	return a.device.Disconnect()
}

// Connected returns whether the keyboard is connected.
func (a *App) Connected() bool {
	return a.device.Connected()
}

// GetLayout returns the 68-key physical layout.
func (a *App) GetLayout() []hid.PhysicalKey {
	return hid.KeyLayout
}

// GetUsages returns the supported HID usage names and IDs.
func (a *App) GetUsages() map[string]byte {
	return hid.HIDUsages
}

// GetBaseKeymap returns the current base layer mappings.
func (a *App) GetBaseKeymap() ([]hid.Mapping, error) {
	table, err := a.device.GetBaseKeymap()
	if err != nil {
		return nil, err
	}
	return hid.TableToMappings(table), nil
}

// GetFnKeymap returns the current Fn layer mappings.
func (a *App) GetFnKeymap() ([]hid.Mapping, error) {
	table, err := a.device.GetFnKeymap()
	if err != nil {
		return nil, err
	}
	return hid.TableToMappings(table), nil
}

// SetBaseKeymap applies the given address->usage bindings to the base layer.
func (a *App) SetBaseKeymap(bindings map[int]string) error {
	table, err := a.device.GetBaseKeymap()
	if err != nil {
		return err
	}
	if err := applyBindings(table, bindings); err != nil {
		return err
	}
	return a.device.SetBaseKeymap(table)
}

// SetFnKeymap applies the given address->usage bindings to the Fn layer.
func (a *App) SetFnKeymap(bindings map[int]string) error {
	table, err := a.device.GetFnKeymap()
	if err != nil {
		return err
	}
	if err := applyBindings(table, bindings); err != nil {
		return err
	}
	return a.device.SetFnKeymap(table)
}

// ApplyFnFKeysPreset maps Fn + 1..= to F1..F12.
func (a *App) ApplyFnFKeysPreset() error {
	table, err := a.device.GetFnKeymap()
	if err != nil {
		return err
	}
	hid.ApplyFnFKeys(table)
	return a.device.SetFnKeymap(table)
}

// Battery returns the keyboard battery percentage (-1 if unavailable).
func (a *App) Battery() (int, error) {
	return a.device.Battery()
}

// HIDCommand emulates the VIA protocol for the VIA frontend.
// It accepts a VIA command ID and payload bytes, and returns the response bytes.
func (a *App) HIDCommand(command int, payload []int) ([]int, error) {
	if !a.device.Connected() {
		if err := a.device.Connect(); err != nil {
			return nil, err
		}
	}
	payloadBytes := make([]byte, len(payload))
	for i, v := range payload {
		payloadBytes[i] = byte(v)
	}
	resp, err := a.device.HIDCommand(byte(command), payloadBytes)
	if err != nil {
		return nil, err
	}
	respInts := make([]int, len(resp))
	for i, v := range resp {
		respInts[i] = int(v)
	}
	return respInts, nil
}

func applyBindings(table []byte, bindings map[int]string) error {
	for addr, name := range bindings {
		usage, ok := hid.HIDUsages[name]
		if !ok {
			return fmt.Errorf("unknown usage: %s", name)
		}
		hid.SetTableEntry(table, addr, usage)
	}
	return nil
}
