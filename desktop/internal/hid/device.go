package hid

import (
	"context"
	"errors"
	"fmt"
	"time"

	go_hid "github.com/telesma-app/hid"
)

// Device wraps a HID connection to the ATOM HE68 PRO.
type Device struct {
	dev *go_hid.Device
}

var (
	ErrNotFound = errors.New("ATOM HE68 PRO not found")
	ErrProtocol = errors.New("HID protocol error")
	AckTimeout  = 2 * time.Second
)

// Connect finds and opens the wired vendor HID interface.
func (d *Device) Connect() error {
	var selected *go_hid.DeviceInfo
	for info, err := range go_hid.Enumerate(
		go_hid.WithVendorID(VendorID),
		go_hid.WithUsagePage(UsagePage),
	) {
		if err != nil {
			return err
		}
		if !containsWiredProductID(info.ProductID) {
			continue
		}
		if selected != nil {
			return fmt.Errorf("%w: more than one matching interface", ErrNotFound)
		}
		selected = info
	}
	if selected == nil {
		return ErrNotFound
	}

	dev, err := go_hid.OpenPath(selected.Path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	d.dev = dev
	return nil
}

// Disconnect closes the HID device handle.
func (d *Device) Disconnect() error {
	if d.dev == nil {
		return nil
	}
	err := d.dev.Close()
	d.dev = nil
	return err
}

// Connected returns true if the device handle is open.
func (d *Device) Connected() bool {
	return d.dev != nil
}

// GetBaseKeymap reads the base layer keymap.
func (d *Device) GetBaseKeymap() ([]byte, error) {
	return d.readTable(CmdGetKey)
}

// GetFnKeymap reads the Fn layer keymap.
func (d *Device) GetFnKeymap() ([]byte, error) {
	return d.readTable(CmdGetFnKey)
}

// SetBaseKeymap writes the base layer keymap.
func (d *Device) SetBaseKeymap(table []byte) error {
	return d.writeTable(CmdSetKey, table)
}

// SetFnKeymap writes the Fn layer keymap.
func (d *Device) SetFnKeymap(table []byte) error {
	return d.writeTable(CmdSetFnKey, table)
}

// Battery queries the keyboard battery percentage. Returns -1 if unavailable.
func (d *Device) Battery() (int, error) {
	req := make([]byte, ReportLength)
	req[0] = HostPrefix
	req[1] = CmdBattery
	req[2] = 0x18
	if err := d.write(req); err != nil {
		return -1, err
	}
	resp, err := d.read(ReportLength)
	if err != nil {
		return -1, err
	}
	if len(resp) < 12 {
		return -1, ErrProtocol
	}
	if resp[0] != 0x55 || resp[1] != 0x10 || resp[2] != 0x18 {
		return -1, ErrProtocol
	}
	encoded := resp[11]
	level := int((encoded>>4)*10 + (encoded & 0x0F))
	if level < 0 || level > 100 {
		return -1, ErrProtocol
	}
	return level, nil
}

func (d *Device) readTable(cmd byte) ([]byte, error) {
	table := make([]byte, KeymapSize)
	for offset := 0; offset < KeymapSize; offset += PayloadLen {
		length := PayloadLen
		if offset+length > KeymapSize {
			length = KeymapSize - offset
		}
		req := BuildReadRequest(cmd, offset, length)
		if err := d.write(req); err != nil {
			return nil, err
		}
		resp, err := d.read(ReportLength)
		if err != nil {
			return nil, err
		}
		respOffset, payload, err := ParseResponse(resp)
		if err != nil {
			return nil, err
		}
		if respOffset != offset {
			return nil, fmt.Errorf("%w: offset mismatch", ErrProtocol)
		}
		if len(payload) != length {
			return nil, fmt.Errorf("%w: length mismatch", ErrProtocol)
		}
		copy(table[offset:], payload)
	}
	return table, nil
}

func (d *Device) writeTable(cmd byte, table []byte) error {
	for _, pkt := range BuildWritePackets(cmd, table) {
		if err := d.write(pkt); err != nil {
			return err
		}
		// Some writes may not return an ACK; drain any pending response with a short timeout.
		ctx, cancel := context.WithTimeout(context.Background(), AckTimeout)
		_, _ = d.dev.Read(ctx, make([]byte, ReportLength+1))
		cancel()
	}
	return nil
}

func (d *Device) write(packet []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), AckTimeout)
	defer cancel()
	// telesma-app/hid expects the report ID as the first byte.
	data := append([]byte{ReportID}, packet...)
	written, err := d.dev.Write(ctx, data)
	if err != nil {
		return err
	}
	if written != len(data) {
		return fmt.Errorf("%w: incomplete write", ErrProtocol)
	}
	return nil
}

func (d *Device) read(size int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), AckTimeout)
	defer cancel()
	buf := make([]byte, size+1) // +1 for optional report-id byte
	n, err := d.dev.Read(ctx, buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("%w: empty response", ErrProtocol)
	}
	// Strip leading report ID if present.
	if n == size+1 && buf[0] == ReportID {
		return buf[1:n], nil
	}
	return buf[:n], nil
}

func containsWiredProductID(pid uint16) bool {
	for _, id := range WiredProductIDs {
		if id == pid {
			return true
		}
	}
	return false
}
