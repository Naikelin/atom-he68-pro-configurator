"""High-level keymap operations for the ATOM HE68 PRO."""

from __future__ import annotations

from dataclasses import dataclass

from .protocol import (
    ENTRY_LEN,
    HID_USAGES,
    KEYMAP_SIZE,
    NUMBER_ROW_ADDRESSES,
    PAGE_KEYBOARD,
    KeyMapEntry,
    build_keymap_entry,
)


@dataclass
class LayerBinding:
    addr: int
    hid_name: str


def make_empty_table() -> bytes:
    """Return a 512-byte table with all entries set to NONE (0x00)."""
    return bytes(KEYMAP_SIZE)


def set_entry(table: bytearray, addr: int, usage: int, page: int = PAGE_KEYBOARD) -> None:
    """Write a keymap entry at the given physical address."""
    base = addr * ENTRY_LEN
    table[base : base + ENTRY_LEN] = build_keymap_entry(usage, page)


def get_entry(table: bytes, addr: int) -> KeyMapEntry:
    base = addr * ENTRY_LEN
    return KeyMapEntry.from_bytes(table[base : base + ENTRY_LEN])


def apply_bindings(table: bytearray, bindings: list[LayerBinding]) -> None:
    for binding in bindings:
        usage = HID_USAGES.get(binding.hid_name)
        if usage is None:
            raise KeyError(f"unknown HID usage name: {binding.hid_name}")
        set_entry(table, binding.addr, usage)


def apply_fn_fkeys(table: bytearray) -> None:
    """Map the number row (1..=) to F1..F12 in the Fn layer."""
    for index, addr in enumerate(NUMBER_ROW_ADDRESSES):
        f_usage = HID_USAGES[f"F{index + 1}"]
        set_entry(table, addr, f_usage)


def table_from_bindings(bindings: dict[int, str]) -> bytes:
    """Create a full table from {address: hid_name}. Missing addresses are NONE."""
    table = bytearray(make_empty_table())
    for addr, hid_name in bindings.items():
        usage = HID_USAGES[hid_name]
        set_entry(table, addr, usage)
    return bytes(table)


def table_to_bindings(table: bytes) -> dict[int, int]:
    """Return {address: hid_usage} for all non-NONE entries."""
    bindings: dict[int, int] = {}
    for addr in range(KEYMAP_SIZE // ENTRY_LEN):
        entry = get_entry(table, addr)
        if entry.usage != 0:
            bindings[addr] = entry.usage
    return bindings


def make_default_fn_fkeys_table() -> bytes:
    """Convenience table that only adds Fn+F1..F12 to an otherwise empty Fn layer."""
    table = bytearray(make_empty_table())
    apply_fn_fkeys(table)
    return bytes(table)
