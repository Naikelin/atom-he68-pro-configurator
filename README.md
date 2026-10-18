# ATOM HE68 PRO Configurator

Unofficial configurator for the Fantech ATOM HE68 PRO (MK922). Includes a Python prototype and a native Windows desktop app built with Wails.

Based on reverse-engineering work by [Sinshro/fantech-atom-he68-pro](https://github.com/Sinshro/fantech-atom-he68-pro) (MIT license).

> **License:** This project is licensed under the GNU General Public License v3.0 or later (GPL-3.0+). The desktop app bundles the [VIA](https://github.com/the-via/app) frontend, which is also GPL-3.0.

## Features

- Web UI editor for base and Fn layers (VIA-like).
- Read and write keymaps through the HID protocol.
- One-click preset to set **Fn + 1..=** → **F1..F12**.
- Python library and CLI scripts.

## Install

```powershell
python -m venv .venv
.venv\Scripts\Activate.ps1
pip install -e .
```

## Quick fixes

### Set Fn + number row to F1..F12

```powershell
.venv\Scripts\Activate.ps1
python scripts/set_fn_fkeys.py
```

Add `--dry-run` to preview changes without writing, or `--yes` to skip confirmation.

## Run the web configurator

```powershell
.venv\Scripts\Activate.ps1
uvicorn atom_he68.api:app --reload
```

Then open http://localhost:8000 in a Chromium-based browser.

## Safety

- This tool writes to your keyboard's runtime configuration, not its firmware. A factory reset from Fantech's official software should restore defaults.
- Do not run this while the official Fantech configurator is open; only one app can own the vendor HID endpoint.
- Test changes on hardware you control and keep backups of working configurations.

## Desktop app (Wails + Go + TypeScript)

The native desktop app is in `desktop/`.

```powershell
cd desktop
wails build
```

The executable will be at `desktop/build/bin/atom-he68-configurator.exe`.

Features:
- Native window with VIA-like keymap editor.
- System tray icon showing battery level.
- One-click Fn+F-keys preset.

## Project structure

- `desktop/` — Wails desktop app (Go backend + TypeScript frontend).
- `src/atom_he68/` — Python prototype library.
- `web/` — Static frontend for the Python prototype.
- `scripts/` — CLI utilities.
- `docs/` — Protocol notes.
