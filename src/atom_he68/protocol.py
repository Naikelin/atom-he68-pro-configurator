"""Protocol constants and packet builders for the Fantech ATOM HE68 PRO.

Based on reverse-engineering work by Sinshro and contributors
(https://github.com/Sinshro/fantech-atom-he68-pro), used under the MIT license.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import NamedTuple

REPORT_LENGTH = 64
REPORT_ID = 0
VENDOR_ID = 0x0C45
WIRED_PRODUCT_IDS = (0x80CB, 0x80CC)
WIRELESS_PRODUCT_IDS = (0xFEDE,)
USAGE_PAGE = 0xFF68
WIRELESS_USAGE_PAGE = 0xFF60

# Keymap commands
GET_KEY = 0x12
GET_FN_KEY = 0x16
SET_KEY = 0x22
SET_FN_KEY = 0x26

HOST_PREFIX = 0xAA
RESPONSE_PREFIX = 0x55

HEADER_LEN = 8
PAYLOAD_LEN = REPORT_LENGTH - HEADER_LEN  # 56 bytes
KEYMAP_SIZE = 512  # 128 entries * 4 bytes
ENTRY_LEN = 4

# Page/type codes inside a keymap entry
PAGE_KEYBOARD = 0x02

# Standard USB HID keyboard usages (subset)
HID_USAGES: dict[str, int] = {
    "NONE": 0x00,
    "A": 0x04,
    "B": 0x05,
    "C": 0x06,
    "D": 0x07,
    "E": 0x08,
    "F": 0x09,
    "G": 0x0A,
    "H": 0x0B,
    "I": 0x0C,
    "J": 0x0D,
    "K": 0x0E,
    "L": 0x0F,
    "M": 0x10,
    "N": 0x11,
    "O": 0x12,
    "P": 0x13,
    "Q": 0x14,
    "R": 0x15,
    "S": 0x16,
    "T": 0x17,
    "U": 0x18,
    "V": 0x19,
    "W": 0x1A,
    "X": 0x1B,
    "Y": 0x1C,
    "Z": 0x1D,
    "Digit1": 0x1E,
    "Digit2": 0x1F,
    "Digit3": 0x20,
    "Digit4": 0x21,
    "Digit5": 0x22,
    "Digit6": 0x23,
    "Digit7": 0x24,
    "Digit8": 0x25,
    "Digit9": 0x26,
    "Digit0": 0x27,
    "Enter": 0x28,
    "Escape": 0x29,
    "Backspace": 0x2A,
    "Tab": 0x2B,
    "Space": 0x2C,
    "Minus": 0x2D,
    "Equal": 0x2E,
    "BracketLeft": 0x2F,
    "BracketRight": 0x30,
    "Backslash": 0x31,
    "Semicolon": 0x33,
    "Quote": 0x34,
    "Backquote": 0x35,
    "Comma": 0x36,
    "Period": 0x37,
    "Slash": 0x38,
    "CapsLock": 0x39,
    "F1": 0x3A,
    "F2": 0x3B,
    "F3": 0x3C,
    "F4": 0x3D,
    "F5": 0x3E,
    "F6": 0x3F,
    "F7": 0x40,
    "F8": 0x41,
    "F9": 0x42,
    "F10": 0x43,
    "F11": 0x44,
    "F12": 0x45,
    "PrintScreen": 0x46,
    "ScrollLock": 0x47,
    "Pause": 0x48,
    "Insert": 0x49,
    "Home": 0x4A,
    "PageUp": 0x4B,
    "Delete": 0x4C,
    "End": 0x4D,
    "PageDown": 0x4E,
    "ArrowRight": 0x4F,
    "ArrowLeft": 0x50,
    "ArrowDown": 0x51,
    "ArrowUp": 0x52,
    "LeftCtrl": 0xE0,
    "LeftShift": 0xE1,
    "LeftAlt": 0xE2,
    "LeftMeta": 0xE3,
    "RightCtrl": 0xE4,
    "RightShift": 0xE5,
    "RightAlt": 0xE6,
    "RightMeta": 0xE7,
    "Fn": 0xAF,  # non-standard, internal mapping
}

HID_USAGE_TO_NAME: dict[int, str] = {v: k for k, v in HID_USAGES.items()}


class KeyMapEntry(NamedTuple):
    page: int
    param1: int
    usage: int
    param3: int

    @classmethod
    def from_bytes(cls, data: bytes) -> "KeyMapEntry":
        return cls(data[0], data[1], data[2], data[3])

    def to_bytes(self) -> bytes:
        return bytes((self.page, self.param1, self.usage, self.param3))


class KeyMapping(NamedTuple):
    """A key's binding in one layer."""

    addr: int
    name: str
    hid_name: str
    hid_usage: int


@dataclass(frozen=True)
class PhysicalKey:
    label: str
    addr: int
    x: float
    y: float
    w: float = 1.0
    h: float = 1.0


# 68-key physical layout derived from the SignalRGB LED map.
# Addresses match the firmware RGB/key table.
KEY_LAYOUT: list[PhysicalKey] = [
    PhysicalKey("Esc", 0, 0, 0),
    PhysicalKey("1", 17, 1, 0),
    PhysicalKey("2", 18, 2, 0),
    PhysicalKey("3", 19, 3, 0),
    PhysicalKey("4", 20, 4, 0),
    PhysicalKey("5", 21, 5, 0),
    PhysicalKey("6", 22, 6, 0),
    PhysicalKey("7", 23, 7, 0),
    PhysicalKey("8", 24, 8, 0),
    PhysicalKey("9", 25, 9, 0),
    PhysicalKey("0", 26, 10, 0),
    PhysicalKey("-", 27, 11, 0),
    PhysicalKey("=", 28, 12, 0),
    PhysicalKey("Backspace", 92, 13, 0, w=2.0),
    PhysicalKey("Insert", 103, 15, 0),
    PhysicalKey("Tab", 32, 0, 1, w=1.5),
    PhysicalKey("Q", 33, 1.5, 1),
    PhysicalKey("W", 34, 2.5, 1),
    PhysicalKey("E", 35, 3.5, 1),
    PhysicalKey("R", 36, 4.5, 1),
    PhysicalKey("T", 37, 5.5, 1),
    PhysicalKey("Y", 38, 6.5, 1),
    PhysicalKey("U", 39, 7.5, 1),
    PhysicalKey("I", 40, 8.5, 1),
    PhysicalKey("O", 41, 9.5, 1),
    PhysicalKey("P", 42, 10.5, 1),
    PhysicalKey("[", 43, 11.5, 1),
    PhysicalKey("]", 44, 12.5, 1),
    PhysicalKey("\\", 60, 13.5, 1, w=1.5),
    PhysicalKey("Del", 106, 15, 1),
    PhysicalKey("Caps", 48, 0, 2, w=1.75),
    PhysicalKey("A", 49, 1.75, 2),
    PhysicalKey("S", 50, 2.75, 2),
    PhysicalKey("D", 51, 3.75, 2),
    PhysicalKey("F", 52, 4.75, 2),
    PhysicalKey("G", 53, 5.75, 2),
    PhysicalKey("H", 54, 6.75, 2),
    PhysicalKey("J", 55, 7.75, 2),
    PhysicalKey("K", 56, 8.75, 2),
    PhysicalKey("L", 57, 9.75, 2),
    PhysicalKey(";", 58, 10.75, 2),
    PhysicalKey("'", 59, 11.75, 2),
    PhysicalKey("Enter", 76, 12.75, 2, w=2.25),
    PhysicalKey("PgUp", 105, 15, 2),
    PhysicalKey("Shift", 64, 0, 3, w=2.25),
    PhysicalKey("Z", 65, 2.25, 3),
    PhysicalKey("X", 66, 3.25, 3),
    PhysicalKey("C", 67, 4.25, 3),
    PhysicalKey("V", 68, 5.25, 3),
    PhysicalKey("B", 69, 6.25, 3),
    PhysicalKey("N", 70, 7.25, 3),
    PhysicalKey("M", 71, 8.25, 3),
    PhysicalKey(",", 72, 9.25, 3),
    PhysicalKey(".", 73, 10.25, 3),
    PhysicalKey("/", 74, 11.25, 3),
    PhysicalKey("Shift", 75, 12.25, 3, w=1.75),
    PhysicalKey("↑", 90, 14, 3),
    PhysicalKey("PgDn", 108, 15, 3),
    PhysicalKey("Ctrl", 80, 0, 4, w=1.25),
    PhysicalKey("Win", 81, 1.25, 4, w=1.25),
    PhysicalKey("Alt", 82, 2.5, 4, w=1.25),
    PhysicalKey("Space", 83, 3.75, 4, w=6.25),
    PhysicalKey("Alt", 84, 10, 4, w=1.25),
    PhysicalKey("Fn", 85, 11.25, 4, w=1.25),
    PhysicalKey("Ctrl", 87, 12.5, 4, w=1.25),
    PhysicalKey("←", 88, 13.75, 4),
    PhysicalKey("↓", 89, 14.75, 4),
    PhysicalKey("→", 91, 15.75, 4),
]

# Addresses that belong to the number row (1..=), used for the Fn+F-keys preset.
NUMBER_ROW_ADDRESSES: list[int] = [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28]


def build_keymap_entry(usage: int, page: int = PAGE_KEYBOARD) -> bytes:
    """Encode one 4-byte keymap entry for a keyboard usage."""
    return KeyMapEntry(page, 0x00, usage, 0x00).to_bytes()


def build_read_request(cmd: int, offset: int, length: int) -> bytes:
    """Build a single read-request packet."""
    if length > PAYLOAD_LEN:
        raise ValueError(f"length cannot exceed {PAYLOAD_LEN}")
    pkt = bytearray(REPORT_LENGTH)
    pkt[0] = HOST_PREFIX
    pkt[1] = cmd
    pkt[2] = length
    pkt[3] = offset & 0xFF
    pkt[4] = (offset >> 8) & 0xFF
    pkt[5] = 0
    pkt[6] = 0
    pkt[7] = 0
    return bytes(pkt)


def build_write_packets(cmd: int, table: bytes) -> list[bytes]:
    """Split a 512-byte keymap table into 64-byte HID reports."""
    if len(table) != KEYMAP_SIZE:
        raise ValueError(f"table must be {KEYMAP_SIZE} bytes, got {len(table)}")
    packets: list[bytes] = []
    for offset in range(0, KEYMAP_SIZE, PAYLOAD_LEN):
        chunk = table[offset : offset + PAYLOAD_LEN]
        is_last = offset + PAYLOAD_LEN >= KEYMAP_SIZE
        pkt = bytearray(REPORT_LENGTH)
        pkt[0] = HOST_PREFIX
        pkt[1] = cmd
        pkt[2] = len(chunk)
        pkt[3] = offset & 0xFF
        pkt[4] = (offset >> 8) & 0xFF
        pkt[5] = 0
        pkt[6] = 1 if is_last else 0
        pkt[7] = 0
        pkt[HEADER_LEN : HEADER_LEN + len(chunk)] = chunk
        packets.append(bytes(pkt))
    return packets


def parse_response(packet: bytes) -> tuple[int, bytes]:
    """Parse a response packet: returns (offset, payload)."""
    if len(packet) < HEADER_LEN:
        raise ValueError("response packet too short")
    if packet[0] != RESPONSE_PREFIX:
        raise ValueError(f"response must start with 0x{RESPONSE_PREFIX:02X}, got {packet[0]:02X}")
    offset = packet[3] | (packet[4] << 8)
    payload_len = packet[2]
    payload = packet[HEADER_LEN : HEADER_LEN + payload_len]
    return offset, payload


def table_to_mappings(table: bytes) -> list[KeyMapping]:
    """Convert a 512-byte keymap table into human-readable mappings."""
    if len(table) != KEYMAP_SIZE:
        raise ValueError(f"table must be {KEYMAP_SIZE} bytes")
    mappings: list[KeyMapping] = []
    for key in KEY_LAYOUT:
        addr = key.addr
        entry = KeyMapEntry.from_bytes(table[addr * ENTRY_LEN : (addr + 1) * ENTRY_LEN])
        name = HID_USAGE_TO_NAME.get(entry.usage, f"0x{entry.usage:02X}")
        mappings.append(KeyMapping(addr, key.label, name, entry.usage))
    return mappings
