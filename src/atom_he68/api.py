"""FastAPI backend for the ATOM HE68 PRO configurator."""

from __future__ import annotations

import logging
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Any

from fastapi import FastAPI, HTTPException
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel

from atom_he68.device import AtomHE68Pro, DeviceNotFoundError, HidProtocolError
from atom_he68.keymap import apply_fn_fkeys, set_entry
from atom_he68.protocol import (
    HID_USAGES,
    HID_USAGE_TO_NAME,
    KEY_LAYOUT,
    table_to_mappings,
)

LOG = logging.getLogger(__name__)

# Keep one device instance for the process lifetime.
_device = AtomHE68Pro()

WEB_DIR = Path(__file__).resolve().parent.parent.parent / "web"


class KeymapUpdate(BaseModel):
    bindings: dict[int, str]


class Status(BaseModel):
    connected: bool
    info: dict[str, Any]


@asynccontextmanager
async def lifespan(app: FastAPI):
    logging.basicConfig(level=logging.INFO)
    yield
    _device.disconnect()


app = FastAPI(title="ATOM HE68 PRO Configurator", lifespan=lifespan)


def _ensure_connected() -> None:
    if not _device.connected:
        try:
            _device.connect()
        except DeviceNotFoundError as exc:
            raise HTTPException(status_code=503, detail=str(exc)) from exc
        except RuntimeError as exc:
            raise HTTPException(status_code=500, detail=str(exc)) from exc


@app.get("/api/status", response_model=Status)
async def status() -> dict[str, Any]:
    return {"connected": _device.connected, "info": {}}


@app.post("/api/connect")
async def connect() -> dict[str, bool]:
    try:
        _device.connect()
    except DeviceNotFoundError as exc:
        raise HTTPException(status_code=503, detail=str(exc)) from exc
    except RuntimeError as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc
    return {"connected": True}


@app.post("/api/disconnect")
async def disconnect() -> dict[str, bool]:
    _device.disconnect()
    return {"connected": False}


@app.get("/api/layout")
async def layout() -> list[dict[str, Any]]:
    return [
        {
            "label": key.label,
            "addr": key.addr,
            "x": key.x,
            "y": key.y,
            "w": key.w,
            "h": key.h,
        }
        for key in KEY_LAYOUT
    ]


@app.get("/api/usages")
async def usages() -> dict[str, int]:
    return HID_USAGES


@app.get("/api/keymap/base")
async def get_base_keymap() -> list[dict[str, Any]]:
    _ensure_connected()
    try:
        table = _device.get_base_keymap()
    except HidProtocolError as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    return [_mapping_to_dict(m) for m in table_to_mappings(table)]


@app.get("/api/keymap/fn")
async def get_fn_keymap() -> list[dict[str, Any]]:
    _ensure_connected()
    try:
        table = _device.get_fn_keymap()
    except HidProtocolError as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    return [_mapping_to_dict(m) for m in table_to_mappings(table)]


@app.post("/api/keymap/base")
async def set_base_keymap(update: KeymapUpdate) -> dict[str, str]:
    _ensure_connected()
    try:
        table = bytearray(_device.get_base_keymap())
        _apply_update(table, update.bindings)
        _device.set_base_keymap(bytes(table))
    except (HidProtocolError, KeyError) as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    return {"status": "ok"}


@app.post("/api/keymap/fn")
async def set_fn_keymap(update: KeymapUpdate) -> dict[str, str]:
    _ensure_connected()
    try:
        table = bytearray(_device.get_fn_keymap())
        _apply_update(table, update.bindings)
        _device.set_fn_keymap(bytes(table))
    except (HidProtocolError, KeyError) as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    return {"status": "ok"}


@app.post("/api/keymap/fn/fkeys")
async def apply_fn_fkeys_preset() -> dict[str, str]:
    """Apply Fn+1..= → F1..F12, preserving other Fn-layer bindings."""
    _ensure_connected()
    try:
        table = bytearray(_device.get_fn_keymap())
        apply_fn_fkeys(table)
        _device.set_fn_keymap(bytes(table))
    except (HidProtocolError, KeyError) as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    return {"status": "ok"}


def _apply_update(table: bytearray, bindings: dict[int, str]) -> None:
    for addr, hid_name in bindings.items():
        usage = HID_USAGES.get(hid_name)
        if usage is None:
            raise KeyError(f"unknown HID usage: {hid_name}")
        set_entry(table, addr, usage)


def _mapping_to_dict(mapping: Any) -> dict[str, Any]:
    return {
        "addr": mapping.addr,
        "label": mapping.name,
        "hid_name": mapping.hid_name,
        "hid_usage": mapping.hid_usage,
    }


# Serve the web UI static files.
if WEB_DIR.is_dir():
    app.mount("/", StaticFiles(directory=WEB_DIR, html=True), name="static")
