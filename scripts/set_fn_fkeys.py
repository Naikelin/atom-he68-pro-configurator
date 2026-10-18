"""CLI script to set Fn + number row to F1..F12 on the ATOM HE68 PRO."""

from __future__ import annotations

import argparse
import logging
import sys

from atom_he68.device import AtomHE68Pro, DeviceNotFoundError, HidProtocolError
from atom_he68.keymap import apply_fn_fkeys
from atom_he68.protocol import NUMBER_ROW_ADDRESSES, table_to_mappings

logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")
LOG = logging.getLogger(__name__)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Set Fn+1..= to emit F1..F12 on the ATOM HE68 PRO."
    )
    parser.add_argument("--dry-run", action="store_true", help="do not write to the keyboard")
    parser.add_argument("--yes", action="store_true", help="skip confirmation prompt")
    args = parser.parse_args()

    try:
        with AtomHE68Pro() as kbd:
            LOG.info("Reading current Fn layer...")
            table = bytearray(kbd.get_fn_keymap())
            current = [m for m in table_to_mappings(table) if m.addr in NUMBER_ROW_ADDRESSES]
            LOG.info("Current Fn number row:")
            for m in current:
                LOG.info("  %-10s (addr %d) -> %s", m.name, m.addr, m.hid_name)

            apply_fn_fkeys(table)
            new = [m for m in table_to_mappings(bytes(table)) if m.addr in NUMBER_ROW_ADDRESSES]
            LOG.info("New Fn number row:")
            for m in new:
                LOG.info("  %-10s (addr %d) -> %s", m.name, m.addr, m.hid_name)

            if args.dry_run:
                LOG.info("Dry-run: no data written.")
                return 0

            if not args.yes:
                answer = input("Write new Fn layer to the keyboard? [y/N] ")
                if answer.strip().lower() not in ("y", "yes"):
                    LOG.info("Cancelled.")
                    return 0

            LOG.info("Writing Fn layer...")
            kbd.set_fn_keymap(bytes(table))
            LOG.info("Done. Try Fn + 1..= now.")
    except DeviceNotFoundError as exc:
        LOG.error("Device not found: %s", exc)
        return 1
    except HidProtocolError as exc:
        LOG.error("Protocol error: %s", exc)
        return 1
    except KeyboardInterrupt:
        LOG.info("Interrupted.")
        return 130
    return 0


if __name__ == "__main__":
    sys.exit(main())
