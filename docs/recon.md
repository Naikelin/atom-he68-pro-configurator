# Reconocimiento — Fantech ATOM HE68 PRO (MK922)

Resumen de la investigación del repo de referencia `Sinshro/fantech-atom-he68-pro` y la web oficial.

## 1. Identificación HID

| Campo | Valor cableado | Notas |
|---|---|---|
| VID | `0x0C45` | Fantech |
| PID | `0x80CB` | HE68 PRO wired |
| Usage page | `0xFF68` | Vendor-defined |
| Report ID | `0` | |
| Longitud reporte | 64 bytes | 32 bytes en 2.4 GHz |

## 2. Comandos del protocolo

| Comando | Valor | Función |
|---|---|---|
| `GET_KEY` | `0x12` | Leer capa base |
| `GET_FN_KEY` | `0x16` | Leer capa Fn |
| `SET_KEY` | `0x22` | Escribir capa base |
| `SET_FN_KEY` | `0x26` | Escribir capa Fn |

## 3. Formato de paquete

Reporte de 64 bytes:

```
[0] = 0xAA (host) / 0x55 (respuesta)
[1] = comando
[2] = bytes de payload en este paquete
[3-4] = offset little-endian
[5] = 0
[6] = 1 si es el último paquete, 0 si no
[7] = 0
[8..63] = payload (hasta 56 bytes)
```

La tabla de keymap es **512 bytes** = 128 entradas × 4 bytes.

Cada entrada:

```
[0] = page/type (0x02 = KEYBOARD)
[1] = parámetro 1
[2] = HID usage ID
[3] = parámetro 3
```

## 4. Fila numérica y F-keys

| Tecla | Dirección física | HID por defecto | HID Fn objetivo |
|---|---|---|---|
| `1` | 17 | 30 (Digit1) | 58 (F1) |
| `2` | 18 | 31 (Digit2) | 59 (F2) |
| `3` | 19 | 32 (Digit3) | 60 (F3) |
| `4` | 20 | 33 (Digit4) | 61 (F4) |
| `5` | 21 | 34 (Digit5) | 62 (F5) |
| `6` | 22 | 35 (Digit6) | 63 (F6) |
| `7` | 23 | 36 (Digit7) | 64 (F7) |
| `8` | 24 | 37 (Digit8) | 65 (F8) |
| `9` | 25 | 38 (Digit9) | 66 (F9) |
| `0` | 26 | 39 (Digit0) | 67 (F10) |
| `-` | 27 | 45 (Minus) | 68 (F11) |
| `=` | 28 | 46 (Equal) | 69 (F12) |

La tecla **Fn** es la dirección física `85`.

## 5. Solución Fn + F1..F12

Escribir la **capa Fn** (`SET_FN_KEY`, `0x26`) con las entradas `17..28` apuntando a los HID usages `58..69`. Para no perder otros atajos Fn, leer primero la tabla actual con `GET_FN_KEY` (`0x16`) y modificar solo las entradas necesarias.

## 6. Incertidumbres

- No hay captura física de `SET_FN_KEY`; se infiere de `SET_KEY` y de la web oficial.
- Falta confirmar ACK exacto de las escrituras de keymap.
- Perfiles: aún no se conoce el comando de cambio de perfil.
- VIA/QMK: no hay evidencia de compatibilidad; la vía práctica es el protocolo propietario.

## 7. Créditos

Basado en el trabajo de `Sinshro/fantech-atom-he68-pro` (licencia MIT).
