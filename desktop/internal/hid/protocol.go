// Package hid implements the Fantech ATOM HE68 PRO HID protocol.
// Based on reverse-engineering by Sinshro and contributors
// (https://github.com/Sinshro/fantech-atom-he68-pro), MIT licensed.
package hid

import (
	"errors"
	"fmt"
)

const (
	VendorID          = 0x0C45
	WiredProductIDCB  = 0x80CB
	WiredProductIDCC  = 0x80CC
	WirelessProductID = 0xFEDE

	UsagePage          = 0xFF68
	WirelessUsagePage  = 0xFF60
	ReportLength       = 64
	ReportID           = 0
	HeaderLen          = 8
	PayloadLen         = ReportLength - HeaderLen
	KeymapSize         = 512
	EntryLen           = 4
	KeymapEntries      = KeymapSize / EntryLen
	PageKeyboard       = 0x02
	HostPrefix         = 0xAA
	ResponsePrefix     = 0x55

	CmdGetKey    = 0x12
	CmdGetFnKey  = 0x16
	CmdSetKey    = 0x22
	CmdSetFnKey  = 0x26
	CmdBattery   = 0x10
)

var WiredProductIDs = []uint16{WiredProductIDCB, WiredProductIDCC}

// HIDUsages maps friendly names to USB HID keyboard usage IDs.
var HIDUsages = map[string]byte{
	"NONE": 0x00,
	"A": 0x04, "B": 0x05, "C": 0x06, "D": 0x07, "E": 0x08, "F": 0x09,
	"G": 0x0A, "H": 0x0B, "I": 0x0C, "J": 0x0D, "K": 0x0E, "L": 0x0F,
	"M": 0x10, "N": 0x11, "O": 0x12, "P": 0x13, "Q": 0x14, "R": 0x15,
	"S": 0x16, "T": 0x17, "U": 0x18, "V": 0x19, "W": 0x1A, "X": 0x1B,
	"Y": 0x1C, "Z": 0x1D,
	"Digit1": 0x1E, "Digit2": 0x1F, "Digit3": 0x20, "Digit4": 0x21,
	"Digit5": 0x22, "Digit6": 0x23, "Digit7": 0x24, "Digit8": 0x25,
	"Digit9": 0x26, "Digit0": 0x27,
	"Enter": 0x28, "Escape": 0x29, "Backspace": 0x2A, "Tab": 0x2B,
	"Space": 0x2C, "Minus": 0x2D, "Equal": 0x2E,
	"BracketLeft": 0x2F, "BracketRight": 0x30, "Backslash": 0x31,
	"Semicolon": 0x33, "Quote": 0x34, "Backquote": 0x35,
	"Comma": 0x36, "Period": 0x37, "Slash": 0x38, "CapsLock": 0x39,
	"F1": 0x3A, "F2": 0x3B, "F3": 0x3C, "F4": 0x3D, "F5": 0x3E,
	"F6": 0x3F, "F7": 0x40, "F8": 0x41, "F9": 0x42, "F10": 0x43,
	"F11": 0x44, "F12": 0x45,
	"PrintScreen": 0x46, "ScrollLock": 0x47, "Pause": 0x48,
	"Insert": 0x49, "Home": 0x4A, "PageUp": 0x4B, "Delete": 0x4C,
	"End": 0x4D, "PageDown": 0x4E, "ArrowRight": 0x4F, "ArrowLeft": 0x50,
	"ArrowDown": 0x51, "ArrowUp": 0x52,
	"LeftCtrl": 0xE0, "LeftShift": 0xE1, "LeftAlt": 0xE2, "LeftMeta": 0xE3,
	"RightCtrl": 0xE4, "RightShift": 0xE5, "RightAlt": 0xE6, "RightMeta": 0xE7,
	"Fn": 0xAF,
}

// HIDUsageToName is the reverse lookup of HIDUsages.
var HIDUsageToName map[byte]string

func init() {
	HIDUsageToName = make(map[byte]string, len(HIDUsages))
	for name, usage := range HIDUsages {
		HIDUsageToName[usage] = name
	}
}

// PhysicalKey describes one key on the 68-key layout.
type PhysicalKey struct {
	Label string  `json:"label"`
	Addr  int     `json:"addr"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
}

// KeyLayout is the 68-key physical layout derived from the SignalRGB LED map.
var KeyLayout = []PhysicalKey{
	{Label: "Esc", Addr: 0, X: 0, Y: 0},
	{Label: "1", Addr: 17, X: 1, Y: 0}, {Label: "2", Addr: 18, X: 2, Y: 0},
	{Label: "3", Addr: 19, X: 3, Y: 0}, {Label: "4", Addr: 20, X: 4, Y: 0},
	{Label: "5", Addr: 21, X: 5, Y: 0}, {Label: "6", Addr: 22, X: 6, Y: 0},
	{Label: "7", Addr: 23, X: 7, Y: 0}, {Label: "8", Addr: 24, X: 8, Y: 0},
	{Label: "9", Addr: 25, X: 9, Y: 0}, {Label: "0", Addr: 26, X: 10, Y: 0},
	{Label: "-", Addr: 27, X: 11, Y: 0}, {Label: "=", Addr: 28, X: 12, Y: 0},
	{Label: "Backspace", Addr: 92, X: 13, Y: 0, W: 2},
	{Label: "Insert", Addr: 103, X: 15, Y: 0},

	{Label: "Tab", Addr: 32, X: 0, Y: 1, W: 1.5},
	{Label: "Q", Addr: 33, X: 1.5, Y: 1}, {Label: "W", Addr: 34, X: 2.5, Y: 1},
	{Label: "E", Addr: 35, X: 3.5, Y: 1}, {Label: "R", Addr: 36, X: 4.5, Y: 1},
	{Label: "T", Addr: 37, X: 5.5, Y: 1}, {Label: "Y", Addr: 38, X: 6.5, Y: 1},
	{Label: "U", Addr: 39, X: 7.5, Y: 1}, {Label: "I", Addr: 40, X: 8.5, Y: 1},
	{Label: "O", Addr: 41, X: 9.5, Y: 1}, {Label: "P", Addr: 42, X: 10.5, Y: 1},
	{Label: "[", Addr: 43, X: 11.5, Y: 1}, {Label: "]", Addr: 44, X: 12.5, Y: 1},
	{Label: "\\", Addr: 60, X: 13.5, Y: 1, W: 1.5},
	{Label: "Del", Addr: 106, X: 15, Y: 1},

	{Label: "Caps", Addr: 48, X: 0, Y: 2, W: 1.75},
	{Label: "A", Addr: 49, X: 1.75, Y: 2}, {Label: "S", Addr: 50, X: 2.75, Y: 2},
	{Label: "D", Addr: 51, X: 3.75, Y: 2}, {Label: "F", Addr: 52, X: 4.75, Y: 2},
	{Label: "G", Addr: 53, X: 5.75, Y: 2}, {Label: "H", Addr: 54, X: 6.75, Y: 2},
	{Label: "J", Addr: 55, X: 7.75, Y: 2}, {Label: "K", Addr: 56, X: 8.75, Y: 2},
	{Label: "L", Addr: 57, X: 9.75, Y: 2}, {Label: ";", Addr: 58, X: 10.75, Y: 2},
	{Label: "'", Addr: 59, X: 11.75, Y: 2},
	{Label: "Enter", Addr: 76, X: 12.75, Y: 2, W: 2.25},
	{Label: "PgUp", Addr: 105, X: 15, Y: 2},

	{Label: "Shift", Addr: 64, X: 0, Y: 3, W: 2.25},
	{Label: "Z", Addr: 65, X: 2.25, Y: 3}, {Label: "X", Addr: 66, X: 3.25, Y: 3},
	{Label: "C", Addr: 67, X: 4.25, Y: 3}, {Label: "V", Addr: 68, X: 5.25, Y: 3},
	{Label: "B", Addr: 69, X: 6.25, Y: 3}, {Label: "N", Addr: 70, X: 7.25, Y: 3},
	{Label: "M", Addr: 71, X: 8.25, Y: 3}, {Label: ",", Addr: 72, X: 9.25, Y: 3},
	{Label: ".", Addr: 73, X: 10.25, Y: 3}, {Label: "/", Addr: 74, X: 11.25, Y: 3},
	{Label: "Shift", Addr: 75, X: 12.25, Y: 3, W: 1.75},
	{Label: "↑", Addr: 90, X: 14, Y: 3},
	{Label: "PgDn", Addr: 108, X: 15, Y: 3},

	{Label: "Ctrl", Addr: 80, X: 0, Y: 4, W: 1.25},
	{Label: "Win", Addr: 81, X: 1.25, Y: 4, W: 1.25},
	{Label: "Alt", Addr: 82, X: 2.5, Y: 4, W: 1.25},
	{Label: "Space", Addr: 83, X: 3.75, Y: 4, W: 6.25},
	{Label: "Alt", Addr: 84, X: 10, Y: 4, W: 1.25},
	{Label: "Fn", Addr: 85, X: 11.25, Y: 4, W: 1.25},
	{Label: "Ctrl", Addr: 87, X: 12.5, Y: 4, W: 1.25},
	{Label: "←", Addr: 88, X: 13.75, Y: 4},
	{Label: "↓", Addr: 89, X: 14.75, Y: 4},
	{Label: "→", Addr: 91, X: 15.75, Y: 4},
}

// NumberRowAddresses are the physical addresses of 1..= on the top row.
var NumberRowAddresses = []int{17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28}

// KeyMapEntry is one 4-byte entry in the keymap table.
type KeyMapEntry struct {
	Page   byte
	Param1 byte
	Usage  byte
	Param3 byte
}

// FromBytes decodes a keymap entry.
func (e *KeyMapEntry) FromBytes(b []byte) {
	e.Page = b[0]
	e.Param1 = b[1]
	e.Usage = b[2]
	e.Param3 = b[3]
}

// ToBytes encodes a keymap entry.
func (e KeyMapEntry) ToBytes() []byte {
	return []byte{e.Page, e.Param1, e.Usage, e.Param3}
}

// BuildKeymapEntry creates a keyboard keymap entry for a usage ID.
func BuildKeymapEntry(usage byte) []byte {
	return KeyMapEntry{Page: PageKeyboard, Param1: 0, Usage: usage, Param3: 0}.ToBytes()
}

// BuildReadRequest builds a single read-request packet.
func BuildReadRequest(cmd byte, offset, length int) []byte {
	if length > PayloadLen {
		panic("read length exceeds payload size")
	}
	pkt := make([]byte, ReportLength)
	pkt[0] = HostPrefix
	pkt[1] = cmd
	pkt[2] = byte(length)
	pkt[3] = byte(offset & 0xFF)
	pkt[4] = byte((offset >> 8) & 0xFF)
	pkt[5] = 0
	pkt[6] = 0
	pkt[7] = 0
	return pkt
}

// BuildWritePackets splits a 512-byte keymap table into 64-byte HID reports.
func BuildWritePackets(cmd byte, table []byte) [][]byte {
	if len(table) != KeymapSize {
		panic(fmt.Sprintf("table must be %d bytes", KeymapSize))
	}
	var packets [][]byte
	for offset := 0; offset < KeymapSize; offset += PayloadLen {
		chunkEnd := offset + PayloadLen
		if chunkEnd > KeymapSize {
			chunkEnd = KeymapSize
		}
		chunk := table[offset:chunkEnd]
		isLast := chunkEnd >= KeymapSize
		pkt := make([]byte, ReportLength)
		pkt[0] = HostPrefix
		pkt[1] = cmd
		pkt[2] = byte(len(chunk))
		pkt[3] = byte(offset & 0xFF)
		pkt[4] = byte((offset >> 8) & 0xFF)
		pkt[5] = 0
		if isLast {
			pkt[6] = 1
		}
		pkt[7] = 0
		copy(pkt[HeaderLen:], chunk)
		packets = append(packets, pkt)
	}
	return packets
}

// ParseResponse parses a response packet and returns (offset, payload, error).
func ParseResponse(packet []byte) (int, []byte, error) {
	if len(packet) < HeaderLen {
		return 0, nil, errors.New("response too short")
	}
	if packet[0] != ResponsePrefix {
		return 0, nil, fmt.Errorf("unexpected response prefix: 0x%02X", packet[0])
	}
	offset := int(packet[3]) | (int(packet[4]) << 8)
	length := int(packet[2])
	if HeaderLen+length > len(packet) {
		return 0, nil, errors.New("response length exceeds packet size")
	}
	return offset, packet[HeaderLen : HeaderLen+length], nil
}

// TableEntry returns the keymap entry at the given physical address.
func TableEntry(table []byte, addr int) KeyMapEntry {
	base := addr * EntryLen
	var e KeyMapEntry
	e.FromBytes(table[base : base+EntryLen])
	return e
}

// SetTableEntry writes a keyboard usage at the given physical address.
func SetTableEntry(table []byte, addr int, usage byte) {
	base := addr * EntryLen
	copy(table[base:], BuildKeymapEntry(usage))
}

// Mapping is a human-readable key binding in one layer.
type Mapping struct {
	Addr     int    `json:"addr"`
	Label    string `json:"label"`
	HIDName  string `json:"hidName"`
	HIDUsage byte   `json:"hidUsage"`
}

// TableToMappings converts a keymap table to mappings for keys in KeyLayout.
func TableToMappings(table []byte) []Mapping {
	var out []Mapping
	for _, key := range KeyLayout {
		e := TableEntry(table, key.Addr)
		name := HIDUsageToName[e.Usage]
		if name == "" {
			name = fmt.Sprintf("0x%02X", e.Usage)
		}
		out = append(out, Mapping{
			Addr:     key.Addr,
			Label:    key.Label,
			HIDName:  name,
			HIDUsage: e.Usage,
		})
	}
	return out
}

// ApplyFnFKeys maps the number row (1..=) to F1..F12 in place on the Fn table.
func ApplyFnFKeys(table []byte) {
	for i, addr := range NumberRowAddresses {
		usage := byte(0x3A + i) // F1..F12
		SetTableEntry(table, addr, usage)
	}
}

// EmptyKeymap returns a 512-byte table filled with NONE entries.
func EmptyKeymap() []byte {
	return make([]byte, KeymapSize)
}
