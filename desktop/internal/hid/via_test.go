package hid

import (
	"testing"
)

func TestHIDCommandKeycodesVersion(t *testing.T) {
	var d Device
	resp, err := d.HIDCommand(0x02, []byte{0x06})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 33 {
		t.Fatalf("expected 33 bytes, got %d", len(resp))
	}
	if resp[0] != 0x02 || resp[1] != 0x06 {
		t.Fatalf("bad header: %v", resp[:2])
	}
	if resp[2] != 0 || resp[3] != 0 || resp[4] != 0 || resp[5] != 0x09 {
		t.Fatalf("bad version bytes: %v", resp[2:6])
	}
}

func TestHIDCommandLayoutOptions(t *testing.T) {
	var d Device
	resp, err := d.HIDCommand(0x02, []byte{0x02})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 33 {
		t.Fatalf("expected 33 bytes, got %d", len(resp))
	}
	if resp[0] != 0x02 || resp[1] != 0x02 {
		t.Fatalf("bad header: %v", resp[:2])
	}
}

func TestHIDCommandLayerCount(t *testing.T) {
	var d Device
	resp, err := d.HIDCommand(0x11, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 2 || resp[0] != 0x11 || resp[1] != VIALayerCount {
		t.Fatalf("bad layer count response: %v", resp)
	}
}

func TestHIDCommandMacroCommands(t *testing.T) {
	var d Device

	// Macro count: [0x0c, count].
	resp, err := d.HIDCommand(0x0c, nil)
	if err != nil {
		t.Fatalf("macro count: unexpected error: %v", err)
	}
	if len(resp) != 2 || resp[0] != 0x0c || resp[1] != 0 {
		t.Fatalf("bad macro count response: %v", resp)
	}

	// Macro buffer size: [0x0d, hi, lo] = 0.
	resp, err = d.HIDCommand(0x0d, nil)
	if err != nil {
		t.Fatalf("macro buffer size: unexpected error: %v", err)
	}
	if len(resp) != 3 || resp[0] != 0x0d || resp[1] != 0 || resp[2] != 0 {
		t.Fatalf("bad macro buffer size response: %v", resp)
	}

	// Macro buffer read echoes the 4-byte header plus `size` data bytes.
	resp, err = d.HIDCommand(0x0e, []byte{0x00, 0x00, 0x04})
	if err != nil {
		t.Fatalf("macro buffer: unexpected error: %v", err)
	}
	if len(resp) != 4+4 || resp[0] != 0x0e || resp[3] != 0x04 {
		t.Fatalf("bad macro buffer response: %v", resp)
	}

	// Macro buffer write echoes the command.
	resp, err = d.HIDCommand(0x0f, []byte{0x00, 0x01, 0x02, 0xAA, 0xBB})
	if err != nil {
		t.Fatalf("macro set buffer: unexpected error: %v", err)
	}
	if len(resp) != 6 || resp[0] != 0x0f || resp[4] != 0xAA || resp[5] != 0xBB {
		t.Fatalf("bad macro set buffer response: %v", resp)
	}

	// Macro reset.
	resp, err = d.HIDCommand(0x10, nil)
	if err != nil {
		t.Fatalf("macro reset: unexpected error: %v", err)
	}
	if len(resp) != 1 || resp[0] != 0x10 {
		t.Fatalf("bad macro reset response: %v", resp)
	}
}

// TestAbsToCellLayerBoundary verifies the flat-buffer addressing used by VIA's
// keymap reads/writes, especially around the layer boundary at byte 150.
func TestAbsToCellLayerBoundary(t *testing.T) {
	cases := []struct {
		abs             int
		layer, row, col int
	}{
		{0, 0, 0, 0},
		{2, 0, 0, 1},
		{148, 0, 4, 14}, // last keycode of layer 0
		{150, 1, 0, 0},  // first keycode of layer 1 (same chunk as 148)
		{298, 1, 4, 14}, // last keycode of layer 1
	}
	for _, c := range cases {
		layer, row, col := absToCell(c.abs)
		if layer != c.layer || row != c.row || col != c.col {
			t.Errorf("absToCell(%d) = (%d,%d,%d), want (%d,%d,%d)",
				c.abs, layer, row, col, c.layer, c.row, c.col)
		}
	}
}

func TestVIAKeymapBufferSizes(t *testing.T) {
	if viaLayerBytes != VIARows*VIACols*2 {
		t.Fatalf("viaLayerBytes = %d, want %d", viaLayerBytes, VIARows*VIACols*2)
	}
	if viaFullBytes != VIALayerCount*viaLayerBytes {
		t.Fatalf("viaFullBytes = %d, want %d", viaFullBytes, VIALayerCount*viaLayerBytes)
	}
}
