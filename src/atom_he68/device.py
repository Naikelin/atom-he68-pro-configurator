"""HID transport for the Fantech ATOM HE68 PRO.

Heavily inspired by the MIT-licensed work at
https://github.com/Sinshro/fantech-atom-he68-pro.
"""

from __future__ import annotations

import logging
import os
from typing import Any

from .protocol import (
    ENTRY_LEN,
    KEYMAP_SIZE,
    PAYLOAD_LEN,
    REPORT_ID,
    REPORT_LENGTH,
    USAGE_PAGE,
    VENDOR_ID,
    WIRED_PRODUCT_IDS,
    GET_FN_KEY,
    GET_KEY,
    SET_FN_KEY,
    SET_KEY,
    build_read_request,
    build_write_packets,
    parse_response,
)

LOG = logging.getLogger(__name__)


class DeviceNotFoundError(RuntimeError):
    """No HE68 PRO vendor-defined HID interface is available."""


class HidProtocolError(RuntimeError):
    """Malformed write, timeout, or unexpected response."""


class AtomHE68Pro:
    """Connection to the HE68 PRO vendor-defined HID interface."""

    def __init__(self, ack_timeout_ms: int = 1_000) -> None:
        if ack_timeout_ms < 0:
            raise ValueError("ack_timeout_ms must be non-negative")
        self._hid: Any | None = None
        self._device: Any | None = None
        self._ack_timeout_ms = ack_timeout_ms

    @property
    def connected(self) -> bool:
        return self._device is not None

    def connect(self, path: bytes | str | None = None) -> None:
        """Find and open the vendor-defined HID interface."""
        if self._device is not None:
            return
        try:
            import hid  # type: ignore[import-not-found]
        except ImportError as exc:
            raise RuntimeError("hidapi is required; install it with: pip install hidapi") from exc
        self._hid = hid

        candidates = hid.enumerate(VENDOR_ID)
        matches = [
            c
            for c in candidates
            if c.get("vendor_id") == VENDOR_ID
            and c.get("product_id") in WIRED_PRODUCT_IDS
            and c.get("usage_page") == USAGE_PAGE
        ]

        if path is not None:
            selected_path = os.fsencode(path) if isinstance(path, str) else path
        elif not matches:
            raise DeviceNotFoundError(
                "No ATOM HE68 PRO HID interface found. Check USB connection and permissions."
            )
        elif len(matches) > 1:
            raise DeviceNotFoundError("More than one matching interface found; refusing to guess.")
        else:
            selected_path = matches[0]["path"]

        device = hid.device()
        try:
            device.open_path(selected_path)
        except Exception:
            device.close()
            raise
        self._device = device
        LOG.info("Connected to ATOM HE68 PRO on %r", selected_path)

    def disconnect(self) -> None:
        if self._device is not None:
            self._device.close()
            self._device = None
            LOG.info("Disconnected from ATOM HE68 PRO")

    def __enter__(self) -> "AtomHE68Pro":
        self.connect()
        return self

    def __exit__(self, *exc: object) -> None:
        self.disconnect()

    def _write(self, packet: bytes) -> None:
        if len(packet) != REPORT_LENGTH:
            raise HidProtocolError(f"packet must be {REPORT_LENGTH} bytes")
        device = self._require_device()
        transport = bytes((REPORT_ID,)) + packet
        written = device.write(transport)
        if written != len(transport):
            raise HidProtocolError(f"incomplete HID write: {written}/{len(transport)} bytes")
        LOG.debug("TX: %s", packet.hex())

    def _read(self) -> bytes:
        device = self._require_device()
        raw = bytes(device.read(REPORT_LENGTH + 1, self._ack_timeout_ms))
        if not raw:
            raise HidProtocolError(f"timed out waiting {self._ack_timeout_ms} ms for response")
        packet = raw[1:] if len(raw) == REPORT_LENGTH + 1 and raw[0] == REPORT_ID else raw
        LOG.debug("RX: %s", packet.hex())
        return packet

    def _require_device(self) -> Any:
        if self._device is None:
            raise HidProtocolError("Not connected")
        return self._device

    def _exchange(self, request: bytes) -> bytes:
        self._write(request)
        return self._read()

    def _read_table(self, cmd: int) -> bytes:
        """Read a full 512-byte keymap table from the device."""
        table = bytearray(KEYMAP_SIZE)
        for offset in range(0, KEYMAP_SIZE, PAYLOAD_LEN):
            length = min(PAYLOAD_LEN, KEYMAP_SIZE - offset)
            request = build_read_request(cmd, offset, length)
            response = self._exchange(request)
            resp_offset, payload = parse_response(response)
            if resp_offset != offset:
                raise HidProtocolError(f"response offset mismatch: {resp_offset} != {offset}")
            if len(payload) != length:
                raise HidProtocolError(f"response length mismatch: {len(payload)} != {length}")
            table[offset : offset + length] = payload
        return bytes(table)

    def _write_table(self, cmd: int, table: bytes) -> None:
        """Write a full 512-byte keymap table to the device."""
        packets = build_write_packets(cmd, table)
        for packet in packets:
            response = self._exchange(packet)
            if response[0] != 0x55:
                raise HidProtocolError(f"unexpected response prefix: {response.hex()}")
            LOG.debug("write chunk acknowledged")

    def get_base_keymap(self) -> bytes:
        return self._read_table(GET_KEY)

    def get_fn_keymap(self) -> bytes:
        return self._read_table(GET_FN_KEY)

    def set_base_keymap(self, table: bytes) -> None:
        self._write_table(SET_KEY, table)

    def set_fn_keymap(self, table: bytes) -> None:
        self._write_table(SET_FN_KEY, table)
