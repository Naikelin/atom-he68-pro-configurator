package hid

import (
	"encoding/binary"
	"fmt"
)

// VIA matrix dimensions used by the VIA frontend definition.
const (
	VIARows = 5
	VIACols = 15
)

// VIAMatrix maps logical VIA [row][col] positions to Fantech physical addresses.
// A value of -1 means the cell is unused.
var VIAMatrix = [VIARows][VIACols]int{
	{0, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 92, 103},
	{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 60, 106},
	{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 76, 105, -1},
	{64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 90, 108, -1},
	{80, 81, 82, 83, 84, 85, 87, 88, 89, 91, -1, -1, -1, -1, -1},
}

// viaAddrToMatrix is the reverse of VIAMatrix.
var viaAddrToMatrix map[int][2]int

func init() {
	viaAddrToMatrix = make(map[int][2]int, VIARows*VIACols)
	for r := 0; r < VIARows; r++ {
		for c := 0; c < VIACols; c++ {
			addr := VIAMatrix[r][c]
			if addr >= 0 {
				viaAddrToMatrix[addr] = [2]int{r, c}
			}
		}
	}
}

// AddrFromVIAMatrix returns the Fantech physical address for a VIA matrix cell.
func AddrFromVIAMatrix(row, col int) (int, bool) {
	if row < 0 || row >= VIARows || col < 0 || col >= VIACols {
		return 0, false
	}
	addr := VIAMatrix[row][col]
	return addr, addr >= 0
}

// VIAMatrixFromAddr returns the VIA matrix position for a Fantech physical address.
func VIAMatrixFromAddr(addr int) (row, col int, ok bool) {
	pos, ok := viaAddrToMatrix[addr]
	if ok {
		return pos[0], pos[1], true
	}
	return 0, 0, false
}

// UsageToVIAKeycode converts a Fantech HID usage to a VIA/QMK keycode.
// Basic keycodes in VIA v13 are identical to HID usage IDs for the 0x00-0xFF range.
func UsageToVIAKeycode(usage byte) uint16 {
	return uint16(usage)
}

// VIAKeycodeToUsage converts a VIA/QMK keycode to a Fantech HID usage.
// Only basic keycodes (0x0000-0x00FF) are supported; advanced QMK codes are rejected.
func VIAKeycodeToUsage(keycode uint16) (byte, bool) {
	if keycode > 0x00FF {
		return 0, false
	}
	return byte(keycode), true
}

// VIAProtocolVersion is the protocol version reported to VIA.
const VIAProtocolVersion = 13

// VIALayerCount is the number of editable layers (base + Fn).
const VIALayerCount = 2

// HIDCommand emulates a VIA protocol command and returns the response payload.
// The input is the raw VIA command bytes (command ID + payload) and the output
// is the response bytes that VIA expects back.
func (d *Device) HIDCommand(command byte, payload []byte) ([]byte, error) {
	switch command {
	case 0x01: // GET_PROTOCOL_VERSION
		return []byte{command, 0, VIAProtocolVersion}, nil

	case 0x11: // DYNAMIC_KEYMAP_GET_LAYER_COUNT
		return []byte{command, VIALayerCount}, nil

	// The HE68 PRO has no VIA-accessible macro storage; report an empty
	// macro area so VIA initialises cleanly instead of erroring out.
	case 0x0c: // DYNAMIC_KEYMAP_MACRO_GET_COUNT
		return []byte{command, 0}, nil

	case 0x0d: // DYNAMIC_KEYMAP_MACRO_GET_BUFFER_SIZE
		return []byte{command, 0, 0}, nil

	case 0x0e: // DYNAMIC_KEYMAP_MACRO_GET_BUFFER
		size := 0
		if len(payload) >= 3 {
			size = int(payload[2])
		}
		res := make([]byte, 4+size)
		res[0] = command
		if len(payload) >= 3 {
			res[1], res[2], res[3] = payload[0], payload[1], payload[2]
		}
		return res, nil

	case 0x0f: // DYNAMIC_KEYMAP_MACRO_SET_BUFFER
		return append([]byte{command}, payload...), nil

	case 0x10: // DYNAMIC_KEYMAP_MACRO_RESET
		return []byte{command}, nil

	case 0x04: // DYNAMIC_KEYMAP_GET_KEYCODE
		if len(payload) < 3 {
			return nil, fmt.Errorf("%w: get keycode payload too short", ErrProtocol)
		}
		layer, row, col := int(payload[0]), int(payload[1]), int(payload[2])
		keycode, err := d.GetVIAKeycode(layer, row, col)
		if err != nil {
			return nil, err
		}
		hi, lo := byte(keycode>>8), byte(keycode&0xFF)
		return []byte{command, payload[0], payload[1], payload[2], hi, lo}, nil

	case 0x05: // DYNAMIC_KEYMAP_SET_KEYCODE
		if len(payload) < 5 {
			return nil, fmt.Errorf("%w: set keycode payload too short", ErrProtocol)
		}
		layer, row, col := int(payload[0]), int(payload[1]), int(payload[2])
		keycode := binary.BigEndian.Uint16(payload[3:5])
		if err := d.SetVIAKeycode(layer, row, col, keycode); err != nil {
			return nil, err
		}
		return append([]byte{command}, payload...), nil

	case 0x12: // DYNAMIC_KEYMAP_GET_BUFFER
		if len(payload) < 3 {
			return nil, fmt.Errorf("%w: get buffer payload too short", ErrProtocol)
		}
		offset := int(binary.BigEndian.Uint16(payload[0:2]))
		size := int(payload[2])
		data, err := d.GetVIAKeymapBuffer(offset, size)
		if err != nil {
			return nil, err
		}
		// VIA expects 4 header bytes + data
		return append([]byte{command, payload[0], payload[1], payload[2]}, data...), nil

	case 0x13: // DYNAMIC_KEYMAP_SET_BUFFER
		if len(payload) < 3 {
			return nil, fmt.Errorf("%w: set buffer payload too short", ErrProtocol)
		}
		offset := int(binary.BigEndian.Uint16(payload[0:2]))
		size := int(payload[2])
		if len(payload) < 3+size {
			return nil, fmt.Errorf("%w: set buffer data too short", ErrProtocol)
		}
		if err := d.SetVIAKeymapBuffer(offset, payload[3:3+size]); err != nil {
			return nil, err
		}
		return append([]byte{command, payload[0], payload[1], payload[2]}, payload[3:3+size]...), nil

	case 0x02: // GET_KEYBOARD_VALUE
		// Return a 33-byte HID report (matching VIA's packet size).
		// VIA reads result bytes at indices [1+bytes.length .. 1+bytes.length+resultLen).
		valueID := byte(0)
		if len(payload) > 0 {
			valueID = payload[0]
		}
		res := make([]byte, 33) // HID report size
		res[0] = command
		if len(payload) > 0 {
			res[1] = payload[0]
		}
		// KEYCODES_VERSION must be non-zero BCD in the 0x08-0x09 range.
		if valueID == 0x06 {
			res[5] = 0x09 // indices 2-5 = 0x00000009
		}
		return res, nil

	case 0x03: // SET_KEYBOARD_VALUE
		return append([]byte{command}, payload...), nil

	case 0x0a: // EEPROM_RESET
		return []byte{command}, nil

	case 0x0b: // BOOTLOADER_JUMP
		return nil, fmt.Errorf("%w: bootloader jump not supported", ErrProtocol)

	default:
		return nil, fmt.Errorf("%w: unsupported VIA command 0x%02X", ErrProtocol, command)
	}
}

// GetVIAKeycode reads a single keycode from the device for a VIA matrix cell.
// defaultUsage maps a Fantech physical address to its factory-default HID usage.
// It is used to present a sensible base layer in VIA even when the device stores
// 0x00 for uncustomized keys.
var defaultUsage = map[int]byte{
	0: 0x29, 17: 0x1E, 18: 0x1F, 19: 0x20, 20: 0x21, 21: 0x22, 22: 0x23, 23: 0x24,
	24: 0x25, 25: 0x26, 26: 0x27, 27: 0x2D, 28: 0x2E, 92: 0x2A, 103: 0x49,
	32: 0x2B, 33: 0x14, 34: 0x1A, 35: 0x08, 36: 0x15, 37: 0x17, 38: 0x1C, 39: 0x18,
	40: 0x0C, 41: 0x12, 42: 0x13, 43: 0x2F, 44: 0x30, 60: 0x31, 106: 0x4C,
	48: 0x39, 49: 0x04, 50: 0x16, 51: 0x07, 52: 0x09, 53: 0x0A, 54: 0x0B, 55: 0x0D,
	56: 0x0E, 57: 0x0F, 58: 0x33, 59: 0x34, 76: 0x28, 105: 0x4B,
	64: 0xE1, 65: 0x1D, 66: 0x1B, 67: 0x06, 68: 0x19, 69: 0x05, 70: 0x11, 71: 0x10,
	72: 0x36, 73: 0x37, 74: 0x38, 75: 0xE5, 90: 0x52, 108: 0x4E,
	80: 0xE0, 81: 0xE3, 82: 0xE2, 83: 0x2C, 84: 0xE6, 85: 0xAF, 87: 0xE4,
	88: 0x50, 89: 0x51, 91: 0x4F,
}

func (d *Device) GetVIAKeycode(layer, row, col int) (uint16, error) {
	addr, ok := AddrFromVIAMatrix(row, col)
	if !ok {
		return 0, nil // unused cell
	}
	table, err := d.getLayerTable(layer)
	if err != nil {
		return 0, err
	}
	entry := TableEntry(table, addr)
	usage := entry.Usage
	if usage == 0 && layer == 0 {
		usage = defaultUsage[addr]
	}
	return UsageToVIAKeycode(usage), nil
}

// SetVIAKeycode writes a single keycode to the device for a VIA matrix cell.
func (d *Device) SetVIAKeycode(layer, row, col int, keycode uint16) error {
	addr, ok := AddrFromVIAMatrix(row, col)
	if !ok {
		return nil // unused cell, ignore
	}
	usage, ok := VIAKeycodeToUsage(keycode)
	if !ok {
		return fmt.Errorf("%w: unsupported VIA keycode 0x%04X", ErrProtocol, keycode)
	}
	table, err := d.getLayerTable(layer)
	if err != nil {
		return err
	}
	SetTableEntry(table, addr, usage)
	return d.setLayerTable(layer, table)
}

// viaLayerBytes is the size in bytes of a single layer's flat keymap.
const viaLayerBytes = VIARows * VIACols * 2

// viaFullBytes is the size in bytes of the whole flat keymap (all layers).
const viaFullBytes = VIALayerCount * viaLayerBytes

// absToCell maps an absolute byte offset in the flat keymap buffer to the
// layer, row and column it belongs to. Each keycode is two bytes, so the
// layer boundary (150) is even and a pair never straddles two layers.
func absToCell(abs int) (layer, row, col int) {
	layer = abs / viaLayerBytes
	local := abs % viaLayerBytes
	cellIdx := local / 2
	return layer, cellIdx / VIACols, cellIdx % VIACols
}

// GetVIAKeymapBuffer reads a chunk of the flat VIA keymap buffer.
// VIA addresses the buffer as a single flat array where each layer occupies
// VIARows*VIACols*2 bytes, so a chunk may span a layer boundary.
func (d *Device) GetVIAKeymapBuffer(offset, size int) ([]byte, error) {
	if offset < 0 || size < 0 || offset+size > viaFullBytes {
		return nil, fmt.Errorf("%w: buffer out of bounds", ErrProtocol)
	}
	buf := make([]byte, size)
	cache := make(map[int][]byte, VIALayerCount)
	for i := 0; i+2 <= size; i += 2 {
		layer, r, c := absToCell(offset + i)
		addr, ok := AddrFromVIAMatrix(r, c)
		if !ok {
			continue // unused cell, leave as 0
		}
		table, ok := cache[layer]
		if !ok {
			var err error
			table, err = d.getLayerTable(layer)
			if err != nil {
				return nil, err
			}
			cache[layer] = table
		}
		entry := TableEntry(table, addr)
		usage := entry.Usage
		if usage == 0 && layer == 0 {
			usage = defaultUsage[addr]
		}
		keycode := UsageToVIAKeycode(usage)
		buf[i] = byte(keycode >> 8)
		buf[i+1] = byte(keycode & 0xFF)
	}
	return buf, nil
}

// SetVIAKeymapBuffer writes a chunk of the flat VIA keymap buffer.
// Chunks may span a layer boundary; every touched layer is flushed once.
func (d *Device) SetVIAKeymapBuffer(offset int, data []byte) error {
	if offset < 0 || offset+len(data) > viaFullBytes {
		return fmt.Errorf("%w: buffer out of bounds", ErrProtocol)
	}
	tables := make(map[int][]byte, VIALayerCount)
	var touched []int
	for i := 0; i+2 <= len(data); i += 2 {
		layer, r, c := absToCell(offset + i)
		addr, ok := AddrFromVIAMatrix(r, c)
		if !ok {
			continue
		}
		keycode := uint16(data[i])<<8 | uint16(data[i+1])
		usage, ok := VIAKeycodeToUsage(keycode)
		if !ok {
			continue // skip unsupported keycodes
		}
		table, ok := tables[layer]
		if !ok {
			var err error
			table, err = d.getLayerTable(layer)
			if err != nil {
				return err
			}
			tables[layer] = table
			touched = append(touched, layer)
		}
		SetTableEntry(table, addr, usage)
	}
	for _, layer := range touched {
		if err := d.setLayerTable(layer, tables[layer]); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) getLayerTable(layer int) ([]byte, error) {
	switch layer {
	case 0:
		return d.GetBaseKeymap()
	case 1:
		return d.GetFnKeymap()
	default:
		return nil, fmt.Errorf("%w: invalid layer %d", ErrProtocol, layer)
	}
}

func (d *Device) setLayerTable(layer int, table []byte) error {
	switch layer {
	case 0:
		return d.SetBaseKeymap(table)
	case 1:
		return d.SetFnKeymap(table)
	default:
		return fmt.Errorf("%w: invalid layer %d", ErrProtocol, layer)
	}
}
