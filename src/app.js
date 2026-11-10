const $ = (selector) => document.querySelector(selector);
const fileInput = $('#file-input');
const editor = $('#html-input');
const dropzone = $('#dropzone');
const status = $('#status');
let currentName = 'Pasted HTML';
let lastReport = null;
let wasmReady = false;

function escapeHTML(value) {
  return String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
}

async function startEngine() {
  try {
    const go = new Go();
    const response = await fetch('./audit.wasm');
    if (!response.ok) throw new Error(`WASM request failed (${response.status})`);
    const bytes = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, go.importObject);
    go.run(result.instance);
    const startedAt = Date.now();
    while (typeof window.markupMedicAudit !== 'function' && Date.now() - startedAt < 3000) {
      await new Promise((resolve) => setTimeout(resolve, 20));
    }
    if (typeof window.markupMedicAudit !== 'function') throw new Error('WASM audit engine did not initialize');
    wasmReady = true;
    status.textContent = 'Go WebAssembly engine ready · source stays on this device.';
  } catch (error) {
    status.textContent = `Could not start the local audit engine: ${error.message}. Serve the built site over HTTP rather than opening the file directly.`;
  }
}

function collectDocument(document) {
  const elements = [...document.querySelectorAll('*')].map((node) => ({
    tag: node.tagName.toLowerCase(),
    attrs: Object.fromEntries([...node.attributes].map((item) => [item.name, item.value])),
    text: (node.innerText || node.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 160)
  }));
  const labelledIds = [...document.querySelectorAll('label[for]')].map((label) => label.htmlFor);
  for (const label of document.querySelectorAll('label')) {
    const control = label.querySelector('input,select,textarea');
    if (control?.id) labelledIds.push(control.id);
  }
  return {
    elements,
    metadata: {
      title: Boolean(document.title.trim()),
      viewport: Boolean(document.querySelector('meta[name="viewport"]')),
      labelledIds
    }
  };
}

function scan(source, name) {
  if (!source.trim()) { status.textContent = 'Add HTML or choose a file before running the audit.'; editor.focus(); return; }
  if (source.length > 2_000_000) { status.textContent = 'This file is larger than 2 MB. Choose a smaller HTML document.'; return; }
  if (!wasmReady) { status.textContent = 'The Go WebAssembly engine is still loading. Try again in a moment.'; return; }
  const parsed = new DOMParser().parseFromString(source, 'text/html');
  const collected = collectDocument(parsed);
  try {
    lastReport = JSON.parse(window.markupMedicAudit(JSON.stringify(collected.elements), JSON.stringify(collected.metadata)));
  } catch {
    status.textContent = 'The audit engine returned an invalid report.';
    return;
  }
  if (lastReport.error) { status.textContent = lastReport.error; return; }
  currentName = name;
  render(lastReport);
  status.textContent = `Audit complete · ${name} · processed locally in your browser with Go WebAssembly`;
}

function render(report) {
  const { errors, warnings, elements } = report.summary;
  $('#error-count').textContent = errors;
  $('#warning-count').textContent = warnings;
  $('#element-count').textContent = elements;
  $('#empty-state').hidden = true;
  $('#report').hidden = false;
  const list = $('#finding-list');
  if (!report.findings.length) {
    list.innerHTML = '<div class="clean-state"><span class="clean-icon">✓</span><h3>No issues detected by these checks</h3><p>This is a useful baseline, not a complete WCAG conformance assessment.</p></div>';
    return;
  }
  list.innerHTML = report.findings.map((finding) => `
    <article class="finding ${finding.severity}">
      <div class="finding-top"><span class="severity">${finding.severity}</span><code>${escapeHTML(finding.rule)}</code></div>
      <h3>${escapeHTML(finding.title)}</h3><p>${escapeHTML(finding.detail)}</p>
      ${finding.index === null ? '' : `<span class="location">Element ${finding.index + 1}</span>`}
    </article>`).join('');
}

$('#run-audit').addEventListener('click', () => scan(editor.value, currentName));
$('#choose-file').addEventListener('click', () => fileInput.click());
fileInput.addEventListener('change', async () => {
  const file = fileInput.files?.[0];
  if (!file) return;
  if (!/\.html?$/i.test(file.name) && file.type !== 'text/html') { status.textContent = 'Please choose an .html or .htm file.'; fileInput.value = ''; return; }
  if (file.size > 2_000_000) { status.textContent = 'This file is larger than 2 MB. Choose a smaller HTML document.'; fileInput.value = ''; return; }
  editor.value = await file.text();
  currentName = file.name;
  status.textContent = `${file.name} loaded · not uploaded`;
  scan(editor.value, currentName);
});
for (const eventName of ['dragenter', 'dragover']) {
  dropzone.addEventListener(eventName, (event) => { event.preventDefault(); dropzone.classList.add('dragging'); });
}
for (const eventName of ['dragleave', 'drop']) {
  dropzone.addEventListener(eventName, (event) => { event.preventDefault(); dropzone.classList.remove('dragging'); });
}
dropzone.addEventListener('drop', async (event) => {
  const file = event.dataTransfer.files?.[0];
  if (!file) return;
  if (!/\.html?$/i.test(file.name) && file.type !== 'text/html') { status.textContent = 'Please drop an .html or .htm file.'; return; }
  if (file.size > 2_000_000) { status.textContent = 'This file is larger than 2 MB.'; return; }
  editor.value = await file.text();
  currentName = file.name;
  scan(editor.value, currentName);
});
$('#load-sample').addEventListener('click', () => {
  currentName = 'sample-page.html';
  editor.value = `<!doctype html><html lang="en"><head><title>Sample page</title></head><body><div class="hero"><img src="hero.jpg"><h1>Welcome</h1><h3>Our services</h3><a href="/start"></a><button></button><input type="email" placeholder="Email"><a href="/docs" target="_blank">Read docs</a></div></body></html>`;
  scan(editor.value, currentName);
});
$('#clear').addEventListener('click', () => {
  editor.value = '';
  fileInput.value = '';
  currentName = 'Pasted HTML';
  lastReport = null;
  $('#report').hidden = true;
  $('#empty-state').hidden = false;
  status.textContent = 'Workspace cleared.';
  editor.focus();
});
$('#export-report').addEventListener('click', () => {
  if (!lastReport) return;
  const payload = { tool: 'MarkupMedic', version: '1.0.0', file: currentName, generatedAt: new Date().toISOString(), ...lastReport };
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = 'markup-medic-report.json';
  link.click();
  URL.revokeObjectURL(url);
});
startEngine();