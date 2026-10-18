import './style.css';

import {
    ApplyFnFKeysPreset,
    Battery,
    Connect,
    Connected,
    Disconnect,
    GetBaseKeymap,
    GetFnKeymap,
    GetLayout,
    GetUsages,
    SetBaseKeymap,
    SetFnKeymap,
} from '../wailsjs/go/main/App';

type LayoutKey = {
    label: string;
    addr: number;
    x: number;
    y: number;
    w: number;
    h: number;
};

type Mapping = {
    addr: number;
    label: string;
    hidName: string;
    hidUsage: number;
};

const $ = (id: string) => document.getElementById(id)!;

let layout: LayoutKey[] = [];
let usages: Record<string, number> = {};
let baseMap: Record<number, string> = {};
let fnMap: Record<number, string> = {};
let pending: { base: Record<number, string>; fn: Record<number, string> } = { base: {}, fn: {} };
let selectedAddr: number | null = null;

async function init() {
    [layout, usages] = await Promise.all([GetLayout(), GetUsages()]);

    const list = $('usages-list') as HTMLDataListElement;
    Object.entries(usages)
        .sort(([a], [b]) => a.localeCompare(b))
        .forEach(([name, code]) => {
            const opt = document.createElement('option');
            opt.value = name;
            opt.textContent = `${name} (0x${code.toString(16).toUpperCase().padStart(2, '0')})`;
            list.appendChild(opt);
        });

    renderKeyboard();
    await updateConnectionStatus();
    await refreshBattery();
    await refreshKeymaps();

    $('btn-connect').addEventListener('click', onConnect);
    $('btn-disconnect').addEventListener('click', onDisconnect);
    $('btn-refresh').addEventListener('click', onRefresh);
    $('btn-assign').addEventListener('click', onAssign);
    $('btn-clear').addEventListener('click', onClear);
    $('btn-save').addEventListener('click', onSave);
    $('btn-fkeys').addEventListener('click', onFkeysPreset);
    $('layer-select').addEventListener('change', () => {
        updateKeyLabels();
        updateSidebar();
    });
    $('usage-search').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') onAssign();
    });

    setInterval(refreshBattery, 60000);
}

function renderKeyboard() {
    const grid = $('key-grid');
    grid.innerHTML = '';
    layout.forEach((key) => {
        const el = document.createElement('div');
        el.className = 'key';
        el.dataset.addr = String(key.addr);
        el.style.left = `${key.x}rem`;
        el.style.top = `${key.y}rem`;
        el.style.width = `${key.w}rem`;
        el.style.height = `${key.h}rem`;
        el.innerHTML = `<span class="label">${escapeHtml(key.label)}</span><span class="binding"></span>`;
        el.addEventListener('click', () => selectKey(key.addr));
        grid.appendChild(el);
    });
}

function updateKeyLabels() {
    const layer = ( $('layer-select') as HTMLSelectElement ).value as 'base' | 'fn';
    document.querySelectorAll('.key').forEach((el) => {
        const addr = parseInt((el as HTMLElement).dataset.addr!, 10);
        const binding = getBinding(layer, addr);
        const other = getBinding(layer === 'base' ? 'fn' : 'base', addr);
        const text = binding || (other ? `Fn:${other}` : '');
        el.querySelector('.binding')!.textContent = text;
    });
}

function getBinding(layer: 'base' | 'fn', addr: number): string {
    const p = pending[layer][addr];
    if (p === 'NONE') return '';
    if (p) return p;
    return layer === 'base' ? baseMap[addr] || '' : fnMap[addr] || '';
}

function selectKey(addr: number) {
    selectedAddr = addr;
    document.querySelectorAll('.key').forEach((el) => el.classList.remove('selected'));
    const el = document.querySelector(`.key[data-addr="${addr}"]`);
    if (el) el.classList.add('selected');
    updateSidebar();
}

function updateSidebar() {
    const info = $('selected-info');
    const assign = $('btn-assign') as HTMLButtonElement;
    const clear = $('btn-clear') as HTMLButtonElement;
    if (selectedAddr == null) {
        info.textContent = 'Select a key to edit.';
        assign.disabled = true;
        clear.disabled = true;
        return;
    }
    const key = layout.find((k) => k.addr === selectedAddr);
    const layer = ( $('layer-select') as HTMLSelectElement ).value;
    const binding = getBinding(layer as 'base' | 'fn', selectedAddr);
    info.innerHTML = `<strong>${escapeHtml(key!.label)}</strong> (${layer})<br>Current: <code>${binding || 'NONE'}</code>`;
    ( $('usage-search') as HTMLInputElement ).value = binding || '';
    assign.disabled = false;
    clear.disabled = false;
    renderPending();
}

function onAssign() {
    if (selectedAddr == null) return;
    const layer = ( $('layer-select') as HTMLSelectElement ).value as 'base' | 'fn';
    const name = ( $('usage-search') as HTMLInputElement ).value.trim();
    if (!name) return;
    if (!usages[name]) {
        alert(`Unknown usage: ${name}`);
        return;
    }
    pending[layer][selectedAddr] = name;
    updateKeyLabels();
    renderPending();
    updateSaveButton();
}

function onClear() {
    if (selectedAddr == null) return;
    const layer = ( $('layer-select') as HTMLSelectElement ).value as 'base' | 'fn';
    pending[layer][selectedAddr] = 'NONE';
    updateKeyLabels();
    renderPending();
    updateSaveButton();
}

function renderPending() {
    const list = $('pending-list');
    list.innerHTML = '';
    for (const layer of ['base', 'fn'] as const) {
        for (const [addr, name] of Object.entries(pending[layer])) {
            const key = layout.find((k) => k.addr === parseInt(addr, 10));
            const li = document.createElement('li');
            li.textContent = `${layer} / ${key ? key.label : addr} → ${name}`;
            list.appendChild(li);
        }
    }
}

function updateSaveButton() {
    const hasPending = Object.keys(pending.base).length + Object.keys(pending.fn).length > 0;
    ( $('btn-save') as HTMLButtonElement ).disabled = !hasPending;
}

async function onSave() {
    try {
        if (Object.keys(pending.base).length) {
            await SetBaseKeymap(pending.base);
        }
        if (Object.keys(pending.fn).length) {
            await SetFnKeymap(pending.fn);
        }
        pending = { base: {}, fn: {} };
        updateSaveButton();
        await refreshKeymaps();
        alert('Saved to keyboard.');
    } catch (err) {
        alert(`Save failed: ${err}`);
    }
}

async function onFkeysPreset() {
    if (!confirm('This will set Fn+1..= to F1..F12 while preserving other Fn bindings. Continue?')) return;
    try {
        await ApplyFnFKeysPreset();
        await refreshKeymaps();
        alert('Fn+F-keys preset applied.');
    } catch (err) {
        alert(`Preset failed: ${err}`);
    }
}

async function onConnect() {
    try {
        await Connect();
        await updateConnectionStatus();
        await refreshKeymaps();
        await refreshBattery();
    } catch (err) {
        alert(`Connect failed: ${err}`);
    }
}

async function onDisconnect() {
    try {
        await Disconnect();
        await updateConnectionStatus();
    } catch (err) {
        alert(`Disconnect failed: ${err}`);
    }
}

async function onRefresh() {
    await refreshKeymaps();
    await refreshBattery();
}

async function updateConnectionStatus() {
    const connected = await Connected();
    const status = $('status');
    status.textContent = connected ? 'Connected' : 'Disconnected';
    status.classList.toggle('connected', connected);
    ( $('btn-connect') as HTMLButtonElement ).disabled = connected;
    ( $('btn-disconnect') as HTMLButtonElement ).disabled = !connected;
}

async function refreshKeymaps() {
    const connected = await Connected();
    if (!connected) return;
    try {
        const [base, fn] = await Promise.all([GetBaseKeymap(), GetFnKeymap()]);
        baseMap = Object.fromEntries(base.map((m: Mapping) => [m.addr, m.hidName]));
        fnMap = Object.fromEntries(fn.map((m: Mapping) => [m.addr, m.hidName]));
        updateKeyLabels();
        updateSidebar();
    } catch (err) {
        console.error('refresh keymaps failed', err);
    }
}

async function refreshBattery() {
    const connected = await Connected();
    if (!connected) {
        $('battery').textContent = '';
        return;
    }
    try {
        const level = await Battery();
        $('battery').textContent = level >= 0 ? `Battery: ${level}%` : '';
    } catch {
        $('battery').textContent = '';
    }
}

function escapeHtml(text: string) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

init().catch((err) => {
    console.error('init failed', err);
    alert(`Initialization failed: ${err}`);
});
