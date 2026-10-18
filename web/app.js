const API = "";

let layout = [];
let usages = {};
let usageNames = [];
let baseMap = {};
let fnMap = {};
let pending = { base: {}, fn: {} };
let selectedAddr = null;
let connected = false;

const $ = (id) => document.getElementById(id);

async function api(method, path, body) {
  const opts = { method, headers: {} };
  if (body) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(`${API}${path}`, opts);
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`${res.status}: ${text}`);
  }
  return res.json();
}

async function init() {
  [layout, usages] = await Promise.all([
    api("GET", "/api/layout"),
    api("GET", "/api/usages"),
  ]);

  usageNames = Object.entries(usages)
    .map(([name, code]) => ({ name, code }))
    .sort((a, b) => a.name.localeCompare(b.name));

  const list = $("usages-list");
  usageNames.forEach(({ name, code }) => {
    const opt = document.createElement("option");
    opt.value = name;
    opt.textContent = `${name} (0x${code.toString(16).toUpperCase().padStart(2, "0")})`;
    list.appendChild(opt);
  });

  renderKeyboard();
  await refreshStatus();
  await refreshKeymaps();

  $("btn-connect").addEventListener("click", connect);
  $("btn-disconnect").addEventListener("click", disconnect);
  $("btn-refresh").addEventListener("click", refreshKeymaps);
  $("btn-assign").addEventListener("click", assignBinding);
  $("btn-clear").addEventListener("click", clearBinding);
  $("btn-save").addEventListener("click", save);
  $("btn-fkeys").addEventListener("click", applyFkeysPreset);
  $("layer-select").addEventListener("change", () => {
    updateKeyLabels();
    updateSidebar();
  });
  $("usage-search").addEventListener("keydown", (e) => {
    if (e.key === "Enter") assignBinding();
  });
}

function renderKeyboard() {
  const grid = $("key-grid");
  grid.innerHTML = "";
  layout.forEach((key) => {
    const el = document.createElement("div");
    el.className = "key";
    el.dataset.addr = key.addr;
    el.style.left = `${key.x}rem`;
    el.style.top = `${key.y}rem`;
    el.style.width = `${key.w}rem`;
    el.style.height = `${key.h}rem`;
    el.innerHTML = `<span class="label">${escapeHtml(key.label)}</span><span class="binding"></span>`;
    el.addEventListener("click", () => selectKey(key.addr));
    grid.appendChild(el);
  });
}

function updateKeyLabels() {
  const layer = $("layer-select").value;
  document.querySelectorAll(".key").forEach((el) => {
    const addr = parseInt(el.dataset.addr, 10);
    const binding = getBinding(layer, addr);
    const other = getBinding(layer === "base" ? "fn" : "base", addr);
    let text = binding || (other ? `Fn:${other}` : "");
    el.querySelector(".binding").textContent = text;
  });
}

function getBinding(layer, addr) {
  const p = pending[layer][addr];
  if (p === "NONE") return "";
  if (p) return p;
  const map = layer === "base" ? baseMap : fnMap;
  return map[addr] || "";
}

function selectKey(addr) {
  selectedAddr = addr;
  document.querySelectorAll(".key").forEach((el) => el.classList.remove("selected"));
  const el = document.querySelector(`.key[data-addr="${addr}"]`);
  if (el) el.classList.add("selected");
  updateSidebar();
}

function updateSidebar() {
  const info = $("selected-info");
  const assign = $("btn-assign");
  const clear = $("btn-clear");
  if (selectedAddr == null) {
    info.textContent = "Select a key to edit.";
    assign.disabled = true;
    clear.disabled = true;
    return;
  }
  const key = layout.find((k) => k.addr === selectedAddr);
  const layer = $("layer-select").value;
  const binding = getBinding(layer, selectedAddr);
  info.innerHTML = `<strong>${escapeHtml(key.label)}</strong> (${layer})<br>Current: <code>${binding || "NONE"}</code>`;
  $("usage-search").value = binding || "";
  assign.disabled = false;
  clear.disabled = false;
  renderPending();
}

function assignBinding() {
  if (selectedAddr == null) return;
  const layer = $("layer-select").value;
  const name = $("usage-search").value.trim();
  if (!name) return;
  if (!usages[name]) {
    alert(`Unknown usage: ${name}`);
    return;
  }
  pending[layer][selectedAddr] = name;
  updateKeyLabels();
  renderPending();
  $("btn-save").disabled = Object.keys(pending.base).length + Object.keys(pending.fn).length === 0;
}

function clearBinding() {
  if (selectedAddr == null) return;
  const layer = $("layer-select").value;
  pending[layer][selectedAddr] = "NONE";
  updateKeyLabels();
  renderPending();
  $("btn-save").disabled = false;
}

function renderPending() {
  const list = $("pending-list");
  list.innerHTML = "";
  for (const layer of ["base", "fn"]) {
    for (const [addr, name] of Object.entries(pending[layer])) {
      const key = layout.find((k) => k.addr === parseInt(addr, 10));
      const li = document.createElement("li");
      li.textContent = `${layer} / ${key ? key.label : addr} → ${name}`;
      list.appendChild(li);
    }
  }
}

async function save() {
  try {
    if (Object.keys(pending.base).length) {
      await api("POST", "/api/keymap/base", { bindings: pending.base });
    }
    if (Object.keys(pending.fn).length) {
      await api("POST", "/api/keymap/fn", { bindings: pending.fn });
    }
    pending = { base: {}, fn: {} };
    $("btn-save").disabled = true;
    await refreshKeymaps();
    alert("Saved to keyboard.");
  } catch (err) {
    alert(`Save failed: ${err.message}`);
  }
}

async function applyFkeysPreset() {
  if (!confirm("This will set Fn+1..= to F1..F12 while preserving other Fn bindings. Continue?")) {
    return;
  }
  try {
    await api("POST", "/api/keymap/fn/fkeys");
    await refreshKeymaps();
    alert("Fn+F-keys preset applied.");
  } catch (err) {
    alert(`Preset failed: ${err.message}`);
  }
}

async function refreshStatus() {
  try {
    const s = await api("GET", "/api/status");
    connected = s.connected;
  } catch {
    connected = false;
  }
  updateStatusUI();
}

function updateStatusUI() {
  const status = $("status");
  status.textContent = connected ? "Connected" : "Disconnected";
  status.classList.toggle("connected", connected);
  $("btn-connect").disabled = connected;
  $("btn-disconnect").disabled = !connected;
}

async function connect() {
  try {
    await api("POST", "/api/connect");
    connected = true;
    updateStatusUI();
    await refreshKeymaps();
  } catch (err) {
    alert(`Connect failed: ${err.message}`);
  }
}

async function disconnect() {
  try {
    await api("POST", "/api/disconnect");
    connected = false;
    updateStatusUI();
  } catch (err) {
    alert(`Disconnect failed: ${err.message}`);
  }
}

async function refreshKeymaps() {
  if (!connected) return;
  try {
    const [base, fn] = await Promise.all([
      api("GET", "/api/keymap/base"),
      api("GET", "/api/keymap/fn"),
    ]);
    baseMap = Object.fromEntries(base.map((m) => [m.addr, m.hid_name]));
    fnMap = Object.fromEntries(fn.map((m) => [m.addr, m.hid_name]));
    updateKeyLabels();
    updateSidebar();
  } catch (err) {
    console.error("refresh keymaps failed", err);
  }
}

function escapeHtml(text) {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

init().catch((err) => {
  console.error("init failed", err);
  alert(`Initialization failed: ${err.message}`);
});
