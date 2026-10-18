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
