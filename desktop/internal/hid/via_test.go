package hid

import (
	"reflect"
	"testing"
)

func TestHIDCommandKeycodesVersion(t *testing.T) {
	var d Device
	resp, err := d.HIDCommand(0x02, []byte{0x06})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0x02, 0x06, 0, 0, 0, 0x09}
	if !reflect.DeepEqual(resp, want) {
		t.Fatalf("got %v, want %v", resp, want)
	}
}

func TestHIDCommandLayoutOptions(t *testing.T) {
	var d Device
	resp, err := d.HIDCommand(0x02, []byte{0x02})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0x02, 0x02, 0, 0, 0, 0}
	if !reflect.DeepEqual(resp, want) {
		t.Fatalf("got %v, want %v", resp, want)
	}
}
